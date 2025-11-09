package router

import (
	"github.com/labstack/echo/v4"
)


//	UpdateOutcomeAlias create sms market alias
//
// @Summary create sms market alias
// @Description This API creates  sms market alias
// @Tags admin - markets
// @Security ApiKeyAuth
// @Param request body models.CreateMarketAlias true "MarketAlias Details"
// @Accept json
// @Produce json
// @Success      201  {object}  models.SuccessResponse " created successfully"
// @Failure      400  {object}  models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure      401  {object}  models.ErrorResponse "Authorization error"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router /outcome/alias [post]
func (a *App) CreateOutcomeAlias(c echo.Context) error {

	return a.Controller.CreateOutcomeAlias(c)
}

//	UpdateOutcomeAlias update sms market alias
//
// @Summary Update sms market alias
// @Description This API updates  sms market alias
// @Tags admin - markets
// @Security ApiKeyAuth
// @Param request body models.CreateMarketAlias true "MarketAlias Details"
// @Param id   path int true "SMS Alias ID"
// @Accept json
// @Produce json
// @Success      201  {object}  models.SuccessResponse " created successfully"
// @Failure      400  {object}  models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure      401  {object}  models.ErrorResponse "Authorization error"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router /outcome/alias/{id} [put]
func (a *App) UpdateOutcomeAlias(c echo.Context) error {

	return a.Controller.UpdateOutcomeAlias(c)
}


//	GetOutcomeAlias get sms markets alias
//
// @Summary Get sms markets alias
// @Description This API gets all sms markets alias
// @Tags admin - markets
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success      200  {array}  models.MarketAlias "MarketAlias array "
// @Failure      400  {object}  models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure      401  {object}  models.ErrorResponse "Authorization error"
// @Failure      404  {object}  models.ErrorResponse "Template not found"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router /outcome/alias [get]
func (a *App) GetOutcomeAlias(c echo.Context) error {

	return a.Controller.GetOutcomeAlias(c)
}

//	DeleteOutcomeAlias delete sms market alias
//
// @Summary delete sms market alias
// @Description This API deletes  sms market alias
// @Tags admin - markets
// @Security ApiKeyAuth
// @Param id   path int true "SMS Alias ID"
// @Accept json
// @Produce json
// @Success      201  {object}  models.SuccessResponse " deleted successfully"
// @Failure      400  {object}  models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure      401  {object}  models.ErrorResponse "Authorization error"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router /outcome/alias/{id} [delete]
func (a *App) DeleteOutcomeAlias(c echo.Context) error {

	return a.Controller.DeleteOutcomeAlias(c)
}