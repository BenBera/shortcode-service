package controllers

import (
	"database/sql"
	"github.com/BenBera/shortcode-service/app/constants"
	"github.com/BenBera/shortcode-service/app/models"
	"github.com/labstack/echo/v4"
	goutils "github.com/mudphilo/go-utils"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func (controller *Controller) GetOutcomeAlias(c echo.Context) error {

	ctx, span := controller.Tracer.Start(c.Request().Context(), "GetOutcomeAlias")
	defer span.End()

	dbUtils := goutils.Db{DBSlave: controller.DBSlave, Context: ctx}

	sqlQuery := "SELECT s.id,s.market_id,s.outcome_id,s.specifier,s.alias,m.market_name FROM sms_market s LEFT JOIN markets m ON s.market_id = m.market_id ORDER BY s.market_id ASC"

	dbUtils.SetQuery(sqlQuery)

	rows, err := dbUtils.FetchSlaveWithContext()

	if err == sql.ErrNoRows {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "No templates sets",
				constants.DATA:        sqlQuery,
			}).
			Warn(err.Error())

		return RespondRaw(c, http.StatusNotFound, models.ErrorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: "not available",
		})
	}

	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error retrieving bet limits ",
				constants.DATA:        sqlQuery,
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "internal server error",
		})

	}

	defer rows.Close()

	var res []models.MarketAlias

	for rows.Next() {

		var outcome_id, specifier, alias, market_name sql.NullString
		var id, market_id sql.NullInt64

		err = rows.Scan(&id, &market_id, &outcome_id, &specifier, &alias, &market_name)
		if err != nil {

			logrus.WithContext(ctx).
				WithFields(logrus.Fields{
					constants.DESCRIPTION: "error scanning sms template",
				}).
				Warn(err.Error())

			continue
		}

		res = append(res, models.MarketAlias{
			ID:         id.Int64,
			MarketID:   market_id.Int64,
			Specifier:  specifier.String,
			OutcomeID:  outcome_id.String,
			Alias:      alias.String,
			MarketName: market_name.String,
		})

	}

	return RespondRaw(c, http.StatusOK, res)

}

func (controller *Controller) CreateOutcomeAlias(c echo.Context) error {

	ctx, span := controller.Tracer.Start(c.Request().Context(), "CreateOutcomeAlias")
	defer span.End()

	u := new(models.CreateMarketAlias)
	if err := c.Bind(u); err != nil {

		logrus.WithContext(ctx).
			WithError(err).
			WithField("description", "Error binding data").
			Error(err.Error())

		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Invalid values",
		})
	}

	dbUtils := goutils.Db{DB: controller.DB, Context: ctx}

	inserts := map[string]interface{}{
		"market_id":  u.MarketID,
		"specifier":  u.Specifier,
		"outcome_id": u.OutcomeID,
		"alias":      u.Alias,
	}

	_, err := dbUtils.UpsertWithContext("sms_market", inserts, nil)
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error creating market alias ",
				constants.DATA:        inserts,
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "Internal server error",
		})
	}

	return RespondRaw(c, http.StatusCreated, models.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "market alias creating successfully",
	})

}

func (controller *Controller) UpdateOutcomeAlias(c echo.Context) error {

	ctx, span := controller.Tracer.Start(c.Request().Context(), "UpdateOutcomeAlias")
	defer span.End()

	id, _ := strconv.Atoi(c.Param("id"))

	if id < 1 {

		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Invalid values",
		})
	}

	u := new(models.CreateMarketAlias)
	if err := c.Bind(u); err != nil {

		logrus.WithContext(ctx).
			WithError(err).
			WithField("description", "Error binding data").
			Error(err.Error())

		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Invalid values",
		})
	}

	dbUtils := goutils.Db{DB: controller.DB, Context: ctx}

	inserts := map[string]interface{}{
		"market_id":  u.MarketID,
		"specifier":  u.Specifier,
		"outcome_id": u.OutcomeID,
		"alias":      u.Alias,
	}

	condition := map[string]interface{}{
		"id": id,
	}

	_, err := dbUtils.UpdateWithContext("sms_market", condition, inserts)
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error updating market alias ",
				constants.DATA:        inserts,
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "Internal server error",
		})
	}

	return RespondRaw(c, http.StatusCreated, models.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "market alias updated successfully",
	})

}

func (controller *Controller) DeleteOutcomeAlias(c echo.Context) error {

	ctx, span := controller.Tracer.Start(c.Request().Context(), "DeleteOutcomeAlias")
	defer span.End()

	id, _ := strconv.Atoi(c.Param("id"))

	if id < 1 {

		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Invalid values",
		})
	}

	dbUtils := goutils.Db{DB: controller.DB, Context: ctx}

	condition := map[string]interface{}{
		"id": id,
	}

	_, err := dbUtils.DeleteWithContext("sms_market", condition)
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error deleting market alias ",
				constants.DATA:        id,
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "Internal server error",
		})
	}

	return RespondRaw(c, http.StatusCreated, models.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "market alias deleted successfully",
	})

}
