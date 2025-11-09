package router

import (
	"github.com/labstack/echo/v4"
)

//	Inbox receives incoming SMS
//
// @Summary  receives incoming SMS
// @Description This API  receives incoming SMS,
// @Tags callback
// @Param request body models.Inbox true "Message Details"
// @Accept json
// @Produce json
// @Success      200  {object}  models.BetSuccessResponse "SMS processed successfully"
// @Failure      400  {object}  models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Failure      502  {object}  models.ErrorResponse "Unable to communicate with other micro services"
// @Router /inbox [post]
func (a *App) Inbox(c echo.Context) error {

	return a.Controller.Inbox(c)
}

//	SDPIncomingSMS receives incoming SMS from SDP
//
// @Summary  receives incoming SMS from SDP
// @Description This API  receives incoming SMS from SDP,
// @Tags callback
// @Param request body models.InboxPayloadInteractive true "Message Details"
// @Accept json
// @Produce json
// @Success      200  {object}  models.BetSuccessResponse "SMS processed successfully"
// @Failure      400  {object}  models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Failure      502  {object}  models.ErrorResponse "Unable to communicate with other micro services"
// @Router /sdp [post]

func (a *App) SDPIncomingSMS(c echo.Context) error {

	return a.Controller.SDPIncomingSMS(c)
}

//	SDPShortcodeDLRS receives incoming SMS dlr from INTOUCH VAS
//
// @Summary  receives incoming SMS dlr from INTOUCH VAS
// @Description This API  receives incoming SMS dlr from INTOUCH VAS,
// @Tags callback
// @Param request body models.InboxDLR true "Message DLR Details"
// @Accept json
// @Produce json
// @Success      200  {object}  models.BetSuccessResponse "SMS processed successfully"
// @Failure      400  {object}  models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Failure      502  {object}  models.ErrorResponse "Unable to communicate with other micro services"
// @Router /dlr [post]

func (a *App) SDPShortcodeDLR(c echo.Context) error {

	return a.Controller.Inboxdlr(c, a.DB)
}
