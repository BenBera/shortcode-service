package router

import (
	"github.com/labstack/echo/v4"
)

//	IncomingUSSD receives incoming USSD REQUEST
//
// @Summary  receives incoming ussd from Intouch Vas
// @Description This API  receives ussd from Intouch Vas,
// @Tags callback
// @Param request body models.USSD true "ussd Details"
// @Accept json
// @Produce json
// @Success      200  {object}  models.UssdResponse "USSD processed successfully"
// @Failure      400  {object}  models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Failure      502  {object}  models.ErrorResponse "Unable to communicate with other micro services"
// @Router /incoming/ussd [post]
func (a *App) IncomingUSSD(c echo.Context) error {

	return a.Controller.IncomingUSSD(c)
}

// UssdHopes returns paginated USSD hopes as a table
//
// @Summary      Fetch USSD hopes as a Vue-compatible table
// @Description  Retrieves paginated, searchable USSD hope data with optional CSV download
// @Tags         ussd
// @Accept       json
// @Produce      json
// @Param        request body models.VueTable true "Vue table filter and pagination options"
// @Success      200 {object} interface{} "Paginated USSD hopes or downloadable CSV"
// @Failure      400 {object} models.ErrorResponse "Invalid request body or parameters"
// @Failure      500 {object} models.ErrorResponse "Internal server error"
// @Router       /ussd/hops [post]
func (a *App) UssdHops(c echo.Context) error {
	return a.Controller.GetUssdHopsAsTable(c)
}


