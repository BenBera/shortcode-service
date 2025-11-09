package router

import (
	"github.com/labstack/echo/v4"
)

// CreateKeyword creates a new SMS keyword
//
// @Summary Create SMS keyword
// @Description This API creates a new SMS keyword
// @Tags admin - Keywords
// @Security ApiKeyAuth
// @Param request body models.Keyword true "Keyword Details"
// @Accept json
// @Produce json
// @Success 201 {object} models.SuccessResponse "Created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure 401 {object} models.ErrorResponse "Authorization error"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /keyword [post]
func (a *App) CreateKeyword(c echo.Context) error {
	return a.Controller.CreateKeyword(c)
}

// UpdateKeyword updates an existing SMS keyword
//
// @Summary Update SMS keyword
// @Description This API updates an existing SMS keyword
// @Tags admin - Keywords
// @Security ApiKeyAuth
// @Param request body models.Keyword true "Keyword Details"
// @Param id path int true "SMS Keyword ID"
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse "Updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure 401 {object} models.ErrorResponse "Authorization error"
// @Failure 404 {object} models.ErrorResponse "Keyword not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /keyword/{id} [put]
func (a *App) UpdateKeyword(c echo.Context) error {
	return a.Controller.UpdateKeyword(c)
}

// UpdateKeywordStatus updates the status of an SMS keyword
//
// @Summary Update SMS keyword status
// @Description This API updates the status of an SMS keyword
// @Tags admin - Keywords
// @Security ApiKeyAuth
// @Param request body models.UpdateStatus true "Keyword Status"
// @Param id path int true "SMS Keyword ID"
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse "Status updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure 401 {object} models.ErrorResponse "Authorization error"
// @Failure 404 {object} models.ErrorResponse "Keyword not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /keyword/{id} [patch]
func (a *App) UpdateKeywordStatus(c echo.Context) error {
	return a.Controller.UpdateKeywordStatus(c)
}

// GetKeywords retrieves all SMS keywords
//
// @Summary Get all SMS keywords
// @Description This API retrieves all SMS keywords
// @Tags admin - Keywords
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {array} models.Keyword "Array of keywords"
// @Failure 401 {object} models.ErrorResponse "Authorization error"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /keyword [get]
func (a *App) GetKeywords(c echo.Context) error {
	return a.Controller.GetKeywords(c)
}

// DeleteKeyword deletes an SMS keyword
//
// @Summary Delete SMS keyword
// @Description This API deletes an SMS keyword
// @Tags admin - Keywords
// @Security ApiKeyAuth
// @Param id path int true "SMS Keyword ID"
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse "Deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid ID"
// @Failure 401 {object} models.ErrorResponse "Authorization error"
// @Failure 404 {object} models.ErrorResponse "Keyword not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /keyword/{id} [delete]
func (a *App) DeleteKeyword(c echo.Context) error {
	return a.Controller.DeleteKeyword(c)
}

// CreateCategory creates a new SMS keyword category
//
// @Summary Create SMS keyword category
// @Description This API creates a new SMS keyword category
// @Tags admin - Categories
// @Security ApiKeyAuth
// @Param request body models.Category true "Category Details"
// @Accept json
// @Produce json
// @Success 201 {object} models.SuccessResponse "Created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure 401 {object} models.ErrorResponse "Authorization error"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /category [post]
func (a *App) CreateCategory(c echo.Context) error {
	return a.Controller.CreateCategory(c)
}

// UpdateCategory updates an existing SMS keyword category
//
// @Summary Update SMS keyword category
// @Description This API updates an existing SMS keyword category
// @Tags admin - Categories
// @Security ApiKeyAuth
// @Param request body models.Category true "Category Details"
// @Param id path int true "SMS Category ID"
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse "Updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure 401 {object} models.ErrorResponse "Authorization error"
// @Failure 404 {object} models.ErrorResponse "Category not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /category/{id} [put]
func (a *App) UpdateCategory(c echo.Context) error {
	return a.Controller.UpdateCategory(c)
}

// UpdateCategoryStatus updates the status of an SMS keyword category
//
// @Summary Update SMS keyword category status
// @Description This API updates the status of an SMS keyword category
// @Tags admin - Categories
// @Security ApiKeyAuth
// @Param request body models.UpdateStatus true "Category Status"
// @Param id path int true "SMS Category ID"
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse "Status updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure 401 {object} models.ErrorResponse "Authorization error"
// @Failure 404 {object} models.ErrorResponse "Category not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /category/{id} [patch]
func (a *App) UpdateCategoryStatus(c echo.Context) error {
	return a.Controller.UpdateCategoryStatus(c)
}

// GetCategories retrieves all SMS keyword categories
//
// @Summary Get all SMS keyword categories
// @Description This API retrieves all SMS keyword categories
// @Tags admin - Categories
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {array} models.Category "Array of categories"
// @Failure 401 {object} models.ErrorResponse "Authorization error"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /category [get]
func (a *App) GetCategories(c echo.Context) error {
	return a.Controller.GetCategories(c)
}

// DeleteCategory deletes an SMS keyword category
//
// @Summary Delete SMS keyword category
// @Description This API deletes an SMS keyword category
// @Tags admin - Categories
// @Security ApiKeyAuth
// @Param id path int true "SMS Category ID"
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse "Deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid ID"
// @Failure 401 {object} models.ErrorResponse "Authorization error"
// @Failure 404 {object} models.ErrorResponse "Category not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /category/{id} [delete]
func (a *App) DeleteCategory(c echo.Context) error {
	return a.Controller.DeleteCategory(c)
}
