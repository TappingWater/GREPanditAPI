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

// Create creates a new paragraph with the given text
// Example Request:
// POST /paragraph
//
//	{
//	    "text": "This is a new paragraph"
//	}
//
// Example Response:
// HTTP/1.1 201 Created
//
//	{
//	    "id": 1,
//	    "text": "This is a new paragraph"
//	}
func (h *ParagraphHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var p models.Paragraph
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	if p.Text == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Text field is required")
	}
	err := h.Service.Create(ctx, &p)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create paragraph")
	}
	return c.JSON(http.StatusCreated, p)
}

// Get retrieves a paragraph with the given ID
// Example Request:
// GET /paragraph/1
// Example Response:
// HTTP/1.1 200 OK
//
//	{
//	    "id": 1,
//	    "text": "This is a new paragraph"
//	}
func (h *ParagraphHandler) Get(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}
	p, err := h.Service.GetByID(ctx, id)
	if err != nil {
		if err == echo.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Paragraph not found with id "+c.Param("id"))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get paragraph")
	}
	return c.JSON(http.StatusOK, p)
}
