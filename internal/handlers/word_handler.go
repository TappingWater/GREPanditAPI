// handlers/word_handler.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"grepandit.com/api/internal/models"
	"grepandit.com/api/internal/services"
)

type WordHandler struct {
	Service *services.WordService
}

func NewWordHandler(s *services.WordService) *WordHandler {
	return &WordHandler{Service: s}
}

func (h *WordHandler) Create(c echo.Context) error {
	var w models.Word
	if err := c.Bind(&w); err != nil {
		return err
	}
	err := h.Service.Create(&w)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create word")
	}
	return c.JSON(http.StatusCreated, w)
}

func (h *WordHandler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}
	w, err := h.Service.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get word")
	}
	return c.JSON(http.StatusOK, w)
}

func (h *WordHandler) GetByWord(c echo.Context) error {
	word := c.Param("word")
	w, err := h.Service.GetByWord(word)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get word")
	}
	return c.JSON(http.StatusOK, w)
}

func (h *WordHandler) Update(c echo.Context) error {
	var w models.Word
	if err := c.Bind(&w); err != nil {
		return err
	}
	err := h.Service.Update(&w)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update word")
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *WordHandler) Delete(c echo.Context) error {
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
