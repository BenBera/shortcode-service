package main

import (
	"bitbucket.org/maybets/shortcode-service/docs"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"bitbucket.org/maybets/shortcode-service/app/database"
	"bitbucket.org/maybets/shortcode-service/app/router"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
)

// @termsOfService https://maybets.com/terms

// @contact.name API Support
// @contact.url https://maybets.com/contact-us
// @contact.email tech@maybets.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name api-key

func main() {

	docs.SwaggerInfo.Title = "Shortcode Service API"
	docs.SwaggerInfo.Description = "This API documents exposes all the available API endpoints for Shortcode service"
	docs.SwaggerInfo.Version = "3.0"
	docs.SwaggerInfo.Host = os.Getenv("base_url")
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{os.Getenv("scheme")}

	ctx := context.Background()

	// Configure OpenTelemetry with sensible defaults.
	uptrace.ConfigureOpentelemetry(
		// copy your project DSN here or use UPTRACE_DSN env var
		//uptrace.WithDSN(os.Getenv("UPTRACE_DSN")),

		uptrace.WithServiceName("shortcode-service"),
		uptrace.WithServiceVersion("v1.0.0"),
		uptrace.WithDeploymentEnvironment("production"),
	)

	// Send buffered spans and free resources.
	defer uptrace.Shutdown(ctx)

	// Create a tracer. Usually, tracer is a global variable.
	tracer := otel.Tracer("shortcode-service")

	// Create a root span (a trace) to measure some operation.
	ctx, main := tracer.Start(ctx, "shortcode-service")
	// End the span when the operation we are measuring is done.
	defer main.End()

	fmt.Printf("trace: %s\n", uptrace.TraceURL(main))

	//setup database
	dbInstance := database.DbInstance()

	driver, err := mysql.WithInstance(dbInstance, &mysql.Config{})
	if err != nil {

		logrus.Panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file:///%s/migrations", GetRootPath()), "mysql", driver)
	if err != nil {

		logrus.Errorf("migration setup error %s ", err.Error())
	}

	err = m.Up() // or m.Step(2) if you want to explicitly set the number of migrations to run
	if err != nil {

		logrus.Errorf("migration error %s ", err.Error())
	}

	// setup consumers
	var a router.App
	a.Initialize(tracer, ctx, dbInstance, database.DbInstanceSlave())
	go a.GRPC()
	a.Run()

}

func GetRootPath() string {

	_, b, _, _ := runtime.Caller(0)

	// Root folder of this project
	return filepath.Join(filepath.Dir(b), "./")
}
