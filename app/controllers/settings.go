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

func (controller *Controller) GetTemplates(c echo.Context) error {

	ctx, span := controller.Tracer.Start(c.Request().Context(), "GetTemplates")
	defer span.End()

	dbUtils := goutils.Db{DBSlave: controller.DBSlave, Context: ctx}

	sqlQuery := "SELECT name, message FROM sms_template ORDER BY name"

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

	var res []models.Template

	for rows.Next() {

		var name, message sql.NullString
		err = rows.Scan(&name, &message)
		if err != nil {

			logrus.WithContext(ctx).
				WithFields(logrus.Fields{
					constants.DESCRIPTION: "error scanning sms template",
				}).
				Warn(err.Error())

			continue
		}

		res = append(res, models.Template{
			Name:    name.String,
			Content: message.String,
		})

	}

	return RespondRaw(c, http.StatusOK, res)

}

func (controller *Controller) UpdateTemplates(c echo.Context) error {

	ctx, span := controller.Tracer.Start(c.Request().Context(), "UpdateSettings")
	defer span.End()

	u := new(models.Template)
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
		"message": u.Content,
	}

	condition := map[string]interface{}{
		"name": u.Name,
	}

	_, err := dbUtils.UpdateWithContext("sms_template", condition, inserts)
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error updating setting ",
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
		Message: "template updated successfully",
	})

}

// Keywords CRUD

func (controller *Controller) GetKeywords(c echo.Context) error {
	ctx, span := controller.Tracer.Start(c.Request().Context(), "GetKeywords")
	defer span.End()

	dbUtils := goutils.Db{DBSlave: controller.DBSlave, Context: ctx}

	sqlQuery := "SELECT id, category_id, keyword, is_active FROM sms_keywords ORDER BY category_id, keyword"

	dbUtils.SetQuery(sqlQuery)

	rows, err := dbUtils.FetchSlaveWithContext()

	if err == sql.ErrNoRows {
		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "No keywords found",
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
				constants.DESCRIPTION: "error retrieving keywords",
				constants.DATA:        sqlQuery,
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "internal server error",
		})
	}

	defer rows.Close()

	var res []models.Keyword

	for rows.Next() {
		var keyword models.Keyword
		err = rows.Scan(&keyword.ID, &keyword.CategoryID, &keyword.Keyword, &keyword.IsActive)
		if err != nil {
			logrus.WithContext(ctx).
				WithFields(logrus.Fields{
					constants.DESCRIPTION: "error scanning keyword",
				}).
				Warn(err.Error())
			continue
		}
		res = append(res, keyword)
	}

	return RespondRaw(c, http.StatusOK, res)
}

func (controller *Controller) CreateKeyword(c echo.Context) error {
	ctx, span := controller.Tracer.Start(c.Request().Context(), "CreateKeyword")
	defer span.End()

	u := new(models.Keyword)
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
		"category_id": u.CategoryID,
		"keyword":     u.Keyword,
		"is_active":   u.IsActive,
	}

	_, err := dbUtils.UpsertWithContext("sms_keywords", inserts, nil)
	if err != nil {
		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error creating keyword",
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
		Message: "keyword created successfully",
	})
}

func (controller *Controller) UpdateKeyword(c echo.Context) error {
	ctx, span := controller.Tracer.Start(c.Request().Context(), "UpdateKeyword")
	defer span.End()

	id, _ := strconv.Atoi(c.Param("id"))

	if id < 1 {
		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Invalid id",
		})
	}

	u := new(models.Keyword)
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

	updates := map[string]interface{}{
		"category_id": u.CategoryID,
		"keyword":     u.Keyword,
		"is_active":   u.IsActive,
	}

	condition := map[string]interface{}{
		"id": id,
	}

	_, err := dbUtils.UpdateWithContext("sms_keywords", condition, updates)
	if err != nil {
		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error updating keyword",
				constants.DATA:        updates,
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "Internal server error",
		})
	}

	return RespondRaw(c, http.StatusOK, models.SuccessResponse{
		Status:  http.StatusOK,
		Message: "keyword updated successfully",
	})
}

func (controller *Controller) DeleteKeyword(c echo.Context) error {
	ctx, span := controller.Tracer.Start(c.Request().Context(), "DeleteKeyword")
	defer span.End()

	id, _ := strconv.Atoi(c.Param("id"))

	if id < 1 {
		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Invalid id",
		})
	}

	dbUtils := goutils.Db{DB: controller.DB, Context: ctx}

	condition := map[string]interface{}{
		"id": id,
	}

	_, err := dbUtils.DeleteWithContext("sms_keywords", condition)
	if err != nil {
		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error deleting keyword",
				constants.DATA:        id,
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "Internal server error",
		})
	}

	return RespondRaw(c, http.StatusOK, models.SuccessResponse{
		Status:  http.StatusOK,
		Message: "keyword deleted successfully",
	})
}

// Categories CRUD

func (controller *Controller) GetCategories(c echo.Context) error {
	ctx, span := controller.Tracer.Start(c.Request().Context(), "GetCategories")
	defer span.End()

	dbUtilslave := goutils.Db{DBSlave: controller.DBSlave, Context: ctx}

	sqlQuery := "SELECT id, category_name, description, is_active FROM sms_categories ORDER BY category_name ASC"

	dbUtilslave.SetQuery(sqlQuery)

	rows, err := dbUtilslave.FetchSlaveWithContext()

	if err == sql.ErrNoRows {
		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "No categories found",
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
				constants.DESCRIPTION: "error retrieving categories",
				constants.DATA:        sqlQuery,
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "internal server error",
		})
	}

	defer rows.Close()

	var res []models.Category

	for rows.Next() {
		var category models.Category
		err = rows.Scan(&category.ID, &category.Name, &category.Description, &category.IsActive)
		if err != nil {
			logrus.WithContext(ctx).
				WithFields(logrus.Fields{
					constants.DESCRIPTION: "error scanning category",
				}).
				Warn(err.Error())
			continue
		}
		res = append(res, category)
	}

	return RespondRaw(c, http.StatusOK, res)
}

func (controller *Controller) CreateCategory(c echo.Context) error {
	ctx, span := controller.Tracer.Start(c.Request().Context(), "CreateCategory")
	defer span.End()

	u := new(models.Category)
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
		"category_name": u.Name,
		"description":   u.Description,
		"is_active":     u.IsActive,
	}

	_, err := dbUtils.UpsertWithContext("sms_categories", inserts, nil)
	if err != nil {
		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error creating category",
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
		Message: "category created successfully",
	})
}

func (controller *Controller) UpdateCategory(c echo.Context) error {
	ctx, span := controller.Tracer.Start(c.Request().Context(), "UpdateCategory")
	defer span.End()

	id, _ := strconv.Atoi(c.Param("id"))

	if id < 1 {
		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Invalid id",
		})
	}

	u := new(models.Category)
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

	updates := map[string]interface{}{
		"category_name": u.Name,
		"description":   u.Description,
		"is_active":     u.IsActive,
	}

	condition := map[string]interface{}{
		"id": id,
	}

	_, err := dbUtils.UpdateWithContext("sms_categories", condition, updates)
	if err != nil {
		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error updating category",
				constants.DATA:        updates,
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "Internal server error",
		})
	}

	return RespondRaw(c, http.StatusOK, models.SuccessResponse{
		Status:  http.StatusOK,
		Message: "category updated successfully",
	})
}

func (controller *Controller) DeleteCategory(c echo.Context) error {
	ctx, span := controller.Tracer.Start(c.Request().Context(), "DeleteCategory")
	defer span.End()

	id, _ := strconv.Atoi(c.Param("id"))

	if id < 1 {
		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Invalid id",
		})
	}

	dbUtils := goutils.Db{DB: controller.DB, Context: ctx}

	condition := map[string]interface{}{
		"id": id,
	}

	_, err := dbUtils.DeleteWithContext("sms_categories", condition)
	if err != nil {
		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error deleting category",
				constants.DATA:        id,
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "Internal server error",
		})
	}

	return RespondRaw(c, http.StatusOK, models.SuccessResponse{
		Status:  http.StatusOK,
		Message: "category deleted successfully",
	})
}

// UpdateKeywordStatus updates the active status of a keyword
func (controller *Controller) UpdateKeywordStatus(c echo.Context) error {
	ctx, span := controller.Tracer.Start(c.Request().Context(), "UpdateKeywordStatus")
	defer span.End()

	id, _ := strconv.Atoi(c.Param("id"))

	if id < 1 {
		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Invalid id",
		})
	}

	u := new(models.UpdateStatus)
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

	updates := map[string]interface{}{
		"is_active": u.IsActive,
	}

	condition := map[string]interface{}{
		"id": id,
	}

	_, err := dbUtils.UpdateWithContext("sms_keywords", condition, updates)
	if err != nil {
		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error updating keyword status",
				constants.DATA:        updates,
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "Internal server error",
		})
	}

	return RespondRaw(c, http.StatusOK, models.SuccessResponse{
		Status:  http.StatusOK,
		Message: "keyword status updated successfully",
	})
}

// UpdateCategoryStatus updates the active status of a category
func (controller *Controller) UpdateCategoryStatus(c echo.Context) error {
	ctx, span := controller.Tracer.Start(c.Request().Context(), "UpdateCategoryStatus")
	defer span.End()

	id, _ := strconv.Atoi(c.Param("id"))

	if id < 1 {
		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Invalid id",
		})
	}

	u := new(models.UpdateStatus)
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

	updates := map[string]interface{}{
		"is_active": u.IsActive,
	}

	condition := map[string]interface{}{
		"id": id,
	}

	_, err := dbUtils.UpdateWithContext("sms_categories", condition, updates)
	if err != nil {
		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error updating category status",
				constants.DATA:        updates,
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "Internal server error",
		})
	}

	return RespondRaw(c, http.StatusOK, models.SuccessResponse{
		Status:  http.StatusOK,
		Message: "category status updated successfully",
	})
}
