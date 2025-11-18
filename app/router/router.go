package router

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/BenBera/shortcode-service/app/auth"
	"github.com/BenBera/shortcode-service/app/controllers"
	db "github.com/BenBera/shortcode-service/app/database"
	"github.com/BenBera/shortcode-service/app/grpc/shortcode"
	_ "github.com/BenBera/shortcode-service/docs"
	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// router and DB instance
type App struct {
	E               *echo.Echo
	DB              *sql.DB
	DBSlave         *sql.DB
	RedisConn       *redis.Client
	GlobalRedisConn *redis.Client
	Controller      *controllers.Controller
	shortcode.UnimplementedShortcodeServer
}

//var defaultConfigPath = "/gaming/application/config/config.ini"

// Initialize initializes the app with predefined configuration
func (a *App) Initialize(tr trace.Tracer, ctx context.Context, dbInstance *sql.DB, dbInstanceSlave *sql.DB) {

	// The passed ctx carries the parent span (main).
	// That is how OpenTelemetry manages span relations.
	ctx, _ = tr.Start(ctx, "Initialize")

	// get rabbitMQ connetion
	a.DB = dbInstance
	a.RedisConn = db.RedisClient()
	a.DBSlave = dbInstanceSlave
	a.GlobalRedisConn = db.GlobalRedisClient()

	//cronjob := crontask.Crontask{
	//	RabbitMQConn: db.GetRabbitMQConnection(),
	//	DB:           a.DB,
	//	DBSlave:      a.DBSlave,
	//	RedisConn:    a.RedisConn,
	//	Tracer:       tr,
	//}

	//	go cronjob.SetupJobs(ctx)

	const (
		maxRetries = 5
		retryDelay = 5 * time.Second
	)

	var controller controllers.Controller
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		controller = controllers.Controller{
			RabbitMQConn: db.GetRabbitMQConnection(),
			DB:           a.DB,
			DBSlave:      a.DBSlave,
			RedisConn:    a.RedisConn,
			Tracer:       tr,
		}
		if err == nil {
			break
		}

		log.Printf("Attempt %d: Failed to initialize controller: %v", retry+1, err)
		if retry < maxRetries-1 {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		log.Fatalf("Failed to initialize controller after %d attempts. Exiting program.", maxRetries)
	}

	a.Controller = &controller

	go controllers.SDPRefreshAccessToken(a.RedisConn)

	a.setRouters()
}

// setRouters sets the all required router
func (a *App) setRouters() {

	// init webserver
	a.E = echo.New()
	a.E.Static("/doc", "api")

	a.E.Use(middleware.Gzip())
	a.E.IPExtractor = echo.ExtractIPFromXFFHeader()
	// add recovery middleware to make the system null safe
	a.E.Use(middleware.Recover()) // change due to swagger
	a.E.Use(session.Middleware(sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))))

	a.E.Use(middleware.RequestID())

	a.E.Use(otelecho.Middleware("shortcode-service"))

	// setup log format and parameters to log for every request

	// Instrument logrus.
	logrus.AddHook(otellogrus.NewHook(otellogrus.WithLevels(
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	)))

	a.E.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {

			req := c.Request()
			res := c.Response()
			start := values.StartTime
			startMicro := start.UnixMicro()

			stop := time.Now()
			stopMicro := stop.UnixMicro()

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {

				id = res.Header().Get(echo.HeaderXRequestID)
			}

			reqSize := req.Header.Get(echo.HeaderContentLength)
			if reqSize == "" {

				reqSize = "0"
			}

			traceID := req.Header.Get("trace-id")
			if traceID == "" {

				traceID = "0"
			}

			service, _ := os.Hostname()

			logrus.WithContext(c.Request().Context()).WithFields(logrus.Fields{
				"service":  service,
				"id":       id,
				"ip":       c.RealIP(),
				"time":     stop.Format(time.RFC3339),
				"host":     req.Host,
				"method":   req.Method,
				"uri":      req.RequestURI,
				"status":   res.Status,
				"size":     reqSize,
				"referer":  req.Referer(),
				"ua":       req.UserAgent(),
				"ttl":      stopMicro - startMicro,
				"trace-id": traceID,
			}).Info("API Response")

			return nil
		},
	}))

	allowedMethods := []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions}
	AllowOrigins := []string{"*"}

	//setup CORS
	corsConfig := middleware.CORSConfig{
		AllowOrigins: AllowOrigins, // in production limit this to only known hosts
		AllowHeaders: AllowOrigins,
		AllowMethods: allowedMethods,
	}

	a.E.Use(middleware.CORSWithConfig(corsConfig))

	// callback
	a.E.POST("/inbox", a.Inbox)
	a.E.POST("/incoming/ussd", auth.Authenticate(a.IncomingUSSD, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "write"))
	a.E.POST("/sdp", a.SDPIncomingSMS)
	a.E.POST("/dlr", a.SDPShortcodeDLR)
	a.E.PUT("/templates", auth.Authenticate(a.UpdateTemplates, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "write"))
	a.E.GET("/templates", auth.Authenticate(a.GetTemplates, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "read"))
	a.E.GET("/ussd/hops", auth.Authenticate(a.UssdHops, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "read"))

	a.E.POST("/outcome/alias", auth.Authenticate(a.CreateOutcomeAlias, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "write"))
	a.E.PUT("/outcome/alias/:id", auth.Authenticate(a.UpdateOutcomeAlias, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "update"))
	a.E.GET("/outcome/alias", auth.Authenticate(a.GetOutcomeAlias, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "read"))
	a.E.DELETE("/outcome/alias/:id", auth.Authenticate(a.DeleteOutcomeAlias, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "delete"))

	a.E.POST("/keyword", auth.Authenticate(a.CreateKeyword, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "write"))
	a.E.PATCH("/keyword/:id", auth.Authenticate(a.UpdateKeywordStatus, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "update"))
	a.E.PUT("/keyword/:id", auth.Authenticate(a.UpdateKeyword, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "update"))
	a.E.GET("/keyword", auth.Authenticate(a.GetKeywords, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "read"))
	a.E.DELETE("/keyword/:id", auth.Authenticate(a.DeleteKeyword, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "delete"))

	a.E.POST("/category", auth.Authenticate(a.CreateCategory, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "write"))
	a.E.PATCH("/category/:id", auth.Authenticate(a.UpdateCategoryStatus, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "update"))
	a.E.PUT("/category/:id", auth.Authenticate(a.UpdateCategory, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "update"))
	a.E.GET("/category", auth.Authenticate(a.GetCategories, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "read"))
	a.E.DELETE("/category/:id", auth.Authenticate(a.DeleteCategory, a.GlobalRedisConn, a.Controller.Tracer, "shortcode", "delete"))

	a.E.GET("/docs/*", echoSwagger.WrapHandler)

	//status
	a.E.POST("/", a.GetStatus)
	a.E.GET("/", a.GetStatus)
}

// Run the app on it's router
func (a *App) Run() {

	server := fmt.Sprintf("%s:%s", os.Getenv("system_host"), os.Getenv("system_port"))
	log.Printf(" HTTP listening on %s ", server)
	a.E.Logger.Fatal(a.E.Start(server))
}
func (a *App) GRPC() {

	server := fmt.Sprintf("%s:%s", os.Getenv("system_host"), os.Getenv("system_grpc_port"))

	lis, err := net.Listen("tcp", server)
	if err != nil {

		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	shortcode.RegisterShortcodeServer(s, a)
	log.Printf("GRPC server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {

		log.Fatalf("failed to serve: %v", err)
	}
}

func getGrpcConnWithFallback(target string) *grpc.ClientConn {
	// Try insecure connection first
	conn, err := getGrpcConn(target)
	if err == nil {
		return conn
	}

	// If insecure fails, try TLS
	log.Printf("Insecure connection failed, attempting TLS: %v", err)
	return getGrpcConnTls(target)
}

func getGrpcConn(target string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	log.Printf("target: %s", target)

	if err != nil {
		log.Printf("target: %s [%s]", target, err.Error())
		return nil, err
	}

	return conn, nil
}

func getGrpcConnTls(target string) *grpc.ClientConn {
	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		log.Printf("Failed to load system root CA pool: %v", err)
		return nil
	}

	cred := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target,
		grpc.WithTransportCredentials(cred),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)

	if err != nil {
		log.Printf("Failed to connect to %s: %v", target, err)
		return nil
	}

	return conn
}
