package router

import (
	"bitbucket.org/maybets/shortcode-service/app/database"
	"bitbucket.org/maybets/shortcode-service/app/grpc/fixture"
	"bitbucket.org/maybets/shortcode-service/app/grpc/identity"
	"bitbucket.org/maybets/shortcode-service/app/grpc/jackpot"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

func (a *App) GetStatus(c echo.Context) error {

	ctx := context.TODO()
	defer ctx.Done()

	status := make(map[string]interface{})

	statusCode := http.StatusOK

	res, err := a.Controller.FixtureServiceClient.Ping(ctx,&fixture.PingRequest{})
	if err != nil {

		log.Printf("error checking fixture service health %s ",err.Error())
		status["fixture-service"] = fmt.Sprintf("error checking fixture service health %s ",err.Error())
		statusCode = http.StatusInternalServerError

	} else {

		if res.Status != http.StatusOK {

			log.Printf("error checking FixtureServiceClient health %s ",res.Data)
			status["fixture-service"] = res.Data
			statusCode = int(res.Status)

		}
	}


	res1, err := a.Controller.IdentityServiceClient.Ping(ctx,&identity.PingRequest{})
	if err != nil {

		log.Printf("error checking fixture service health %s ",err.Error())
		status["identity-service"] = fmt.Sprintf("error checking identity service health %s ",err.Error())
		statusCode = http.StatusInternalServerError

	} else {

		if res1.Status != http.StatusOK {

			log.Printf("error checking IdentityServiceClient health %s ",res1.Data)
			status["identity-service"] = res1.Data
			statusCode = int(res.Status)

		}
	}

	res2, err := a.Controller.JackpotServiceClient.Ping(ctx,&jackpot.JackpotPing{})
	if err != nil {

		log.Printf("error checking jackpot service health %s ",err.Error())
		status["odds-service"] = fmt.Sprintf("error checking jackpot service health %s ",err.Error())
		statusCode = http.StatusInternalServerError

	} else {

		if res2.Status != http.StatusOK {

			log.Printf("error checking JackpotServiceClient health %s ",res2.Data)
			status["jackpot-service"] = res2.Data
			statusCode = int(res.Status)

		}
	}

	st, re := database.CheckConnectionStatus(c.Request().Context(), a.DB)

	if st > statusCode {

		statusCode = st
	}

	for k, v := range re {

		status[k] = v
	}

	return c.JSON(statusCode, status)

}
