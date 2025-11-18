package router

import (
	"context"

	"github.com/BenBera/shortcode-service/app/database"

	"net/http"

	"github.com/labstack/echo/v4"
)

func (a *App) GetStatus(c echo.Context) error {

	ctx := context.TODO()
	defer ctx.Done()

	status := make(map[string]interface{})

	statusCode := http.StatusOK

	st, re := database.CheckConnectionStatus(c.Request().Context(), a.DB)

	if st > statusCode {

		statusCode = st
	}

	for k, v := range re {

		status[k] = v
	}

	return c.JSON(statusCode, status)

}
