package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"grepandit.com/api/internal/models"
	"grepandit.com/api/internal/services"
)

type VerbalQuestionHandler struct {
	Service *services.VerbalQuestionService
}

func NewVerbalQuestionHandler(s *services.VerbalQuestionService) *VerbalQuestionHandler {
	return &VerbalQuestionHandler{Service: s}
}

func (h *VerbalQuestionHandler) Create(c echo.Context) error {
	var q models.VerbalQuestion
	if err := c.Bind(&q); err != nil {
		return err
	}
	err := h.Service.Create(&q)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create question")
	}
	return c.JSON(http.StatusCreated, q)
}

func (h *VerbalQuestionHandler) Get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}
	q, err := h.Service.Get(id)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get question"+err.Error())
	}
	return c.JSON(http.StatusOK, q)
}

func (h *VerbalQuestionHandler) Update(c echo.Context) error {
	var q models.VerbalQuestion
	if err := c.Bind(&q); err != nil {
		return err
	}
	err := h.Service.Update(&q)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update question")
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *VerbalQuestionHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}
	err = h.Service.Delete(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete question")
	}
	return c.NoContent(http.StatusNoContent)
}
