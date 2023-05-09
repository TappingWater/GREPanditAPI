// handlers/paragraph_handler.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"grepandit.com/api/internal/models"
	"grepandit.com/api/internal/services"
)

type ParagraphHandler struct {
	Service *services.ParagraphService
}

func NewParagraphHandler(s *services.ParagraphService) *ParagraphHandler {
	return &ParagraphHandler{Service: s}
}

func (h *ParagraphHandler) Create(c echo.Context) error {
	var p models.Paragraph
	if err := c.Bind(&p); err != nil {
		return err
	}
	err := h.Service.Create(&p)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create paragraph")
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *ParagraphHandler) Get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}
	w, err := h.Service.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get paragraph")
	}
	return c.JSON(http.StatusOK, w)
}

func (h *ParagraphHandler) Update(c echo.Context) error {
	var p models.Paragraph
	if err := c.Bind(&p); err != nil {
		return err
	}
	err := h.Service.Update(&p)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update word")
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *ParagraphHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}
	err = h.Service.Delete(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete word")
	}
	return c.NoContent(http.StatusNoContent)
}
