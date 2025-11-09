package router

import (
	"github.com/labstack/echo/v4"
)

//	UpdateTemplates update sms templates
//
// @Summary Update sms templates
// @Description This API updates  sms templates
// @Tags admin - templates
// @Security ApiKeyAuth
// @Param request body models.Template true "Template Details"
// @Accept json
// @Produce json
// @Success      201  {object}  models.SuccessResponse "Settings created successfully"
// @Failure      400  {object}  models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure      401  {object}  models.ErrorResponse "Authorization error"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router /templates [put]
func (a *App) UpdateTemplates(c echo.Context) error {

	return a.Controller.UpdateTemplates(c)
}


//	GetTemplates get sms templates
//
// @Summary Get sms templates
// @Description This API sms templates used to send back responses to the customer
// @Tags admin - templates
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success      200  {array}  models.Template "Templates array "
// @Failure      400  {object}  models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure      401  {object}  models.ErrorResponse "Authorization error"
// @Failure      404  {object}  models.ErrorResponse "Template not found"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router /templates [get]
func (a *App) GetTemplates(c echo.Context) error {

	return a.Controller.GetTemplates(c)
}
