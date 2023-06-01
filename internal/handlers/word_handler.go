// handlers/word_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgconn"
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

// Create adds a new word to the database with the given data.
//
// Example Request:
// POST /words
// Content-Type: application/json
//
//	{
//	    "word": "abacus",
//	    "meanings": ["a frame with beads for doing arithmetic"],
//	    "examples": ["He used an abacus to do his calculations."]
//	}
//
// Example Response:
// HTTP/1.1 201 Created
// Content-Type: application/json
//
//	{
//	    "id": 1,
//	    "word": "abacus",
//	    "meanings": ["a frame with beads for doing arithmetic"],
//	    "examples": ["He used an abacus to do his calculations."]
//	}
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the created word data.
func (h *WordHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var w models.Word
	if err := c.Bind(&w); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	if w.Word == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body. Requires meaning, word and examples for each meaning")
	}
	err := h.Service.Create(ctx, &w)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return echo.NewHTTPError(http.StatusConflict, "Word already exists")
			}
		}
		fmt.Println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create word")
	}
	return c.JSON(http.StatusCreated, w)
}

// GetByID retrieves a word from the database by its ID and returns its data.
//
// Example Request:
// GET /words/1
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
//	{
//	    "id": 1,
//	    "word": "abacus",
//	    "meanings": ["a frame with beads for doing arithmetic"],
//	    "examples": ["He used an abacus to do his calculations."]
//	}
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the word data.
func (h *WordHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}
	w, err := h.Service.GetByID(ctx, id)
	if err != nil {
		fmt.Println(err.Error())
		if err == echo.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Word not found with id "+c.Param("id"))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get word")
	}
	return c.JSON(http.StatusOK, w)
}

// GetByWord retrieves a word from the database by its word string and returns its data.
//
// Example Request:
// GET /words/abacus
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
//	{
//	    "id": 1,
//	    "word": "abacus",
//	    "meanings": ["a frame with beads for doing arithmetic"],
//	    "examples": ["He used an abacus to do his calculations."]
//	}
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the word data.
func (h *WordHandler) GetByWord(c echo.Context) error {
	ctx := c.Request().Context()
	word := c.Param("word")
	w, err := h.Service.GetByWord(ctx, word)
	if err != nil {
		fmt.Println(err.Error())
		if err == echo.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "No record found for word "+c.Param("word"))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get word")
	}
	return c.JSON(http.StatusOK, w)
}
