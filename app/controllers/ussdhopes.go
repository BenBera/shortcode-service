package controllers

import (
	"encoding/json"

	"github.com/BenBera/shortcode-service/app/library"
	"github.com/labstack/echo/v4"
	goutils "github.com/mudphilo/go-utils"
	goutilsmodels "github.com/mudphilo/go-utils/models"
)

func (controller *Controller) GetUssdHopsAsTable(c echo.Context) error {

	payload := goutils.GetJSONRawBody(c)

	body, _ := json.Marshal(payload)
	search := new(goutilsmodels.VueTable)
	_ = json.Unmarshal(body, search)

	httpStatus, resp, _ := library.GetUssdhopsAsTable(controller.DBSlave, c, *search)
	return RespondRaw(c, httpStatus, resp)

}
