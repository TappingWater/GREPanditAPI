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
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	ctx := c.Request().Context()
	err := h.Service.Create(ctx, &q)
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
	ctx := c.Request().Context()
	q, err := h.Service.GetByID(ctx, id)
	if err != nil {
		if err == echo.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Question not found with id "+c.Param("id"))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get question")
	}
	return c.JSON(http.StatusOK, q)
}

func (h *VerbalQuestionHandler) Count(c echo.Context) error {
	count, err := h.Service.Count(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get question count")
	}
	return c.JSON(http.StatusOK, count)
}

type RandomQuestionsRequest struct {
	Limit        int   `json:"limit"`
	QuestionType int   `json:"question_type,omitempty"`
	Competence   int   `json:"competence,omitempty"`
	FramedAs     int   `json:"framed_as,omitempty"`
	Difficulty   int   `json:"difficulty,omitempty"`
	ExcludeIDs   []int `json:"exclude_ids,omitempty"`
}

func (h *VerbalQuestionHandler) GetRandomQuestions(c echo.Context) error {
	req := RandomQuestionsRequest{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	questions, err := h.Service.Random(c.Request().Context(), req.Limit, req.QuestionType, req.Competence, req.FramedAs, req.Difficulty, req.ExcludeIDs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve random questions")
	}

	return c.JSON(http.StatusOK, questions)
}
