package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

// Create creates a new verbal question with the data provided in the request payload and saves it in the database.
//
// Example Request:
// POST /verbal-questions
// Content-Type: application/json//
//
//	{
//	    "text": "What is GPT-3?",
//	    "answer": "GPT-3 is a state-of-the-art language model developed by OpenAI.",
//	    "difficulty": 3,
//	    "competence": "General Knowledge",
//	    "framed_as": "Question",
//	    "type": "Verbal"
//	}
//
// Example Response:
// HTTP/1.1 201 Created
// Content-Type: application/json//
//
//	{
//	    "id": 1,
//	    "text": "What is GPT-3?",
//	    "answer": "GPT-3 is a state-of-the-art language model developed by OpenAI.",
//	    "difficulty": 3,
//	    "competence": "General Knowledge",
//	    "framed_as": "Question",
//	    "type": "Verbal"
//	}
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the created question data.
func (h *VerbalQuestionHandler) Create(c echo.Context) error {
	var q models.VerbalQuestionRequest
	if err := c.Bind(&q); err != nil {
		println(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	ctx := c.Request().Context()
	err := h.Service.Create(ctx, &q)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create question")
	}
	return c.JSON(http.StatusCreated, q)
}

// Get retrieves a verbal question from the database by its ID and returns its data.
//
// Example Request:
// GET /verbal-questions/1
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
//	{
//	    "id": 1,
//	    "text": "What is GPT-3?",
//	    "answer": "GPT-3 is a state-of-the-art language model developed by OpenAI.",
//	    "difficulty": 3,
//	    "competence": "General Knowledge",
//	    "framed_as": "Question",
//	    "type": "Verbal"
//	}
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the question data.
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
		println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get question")
	}
	return c.JSON(http.StatusOK, q)
}

func (h *VerbalQuestionHandler) GetAll(c echo.Context) error {
	idsParam := c.QueryParam("ids")
	// idsParam is a string like "[31,63]" so we need to convert it into an array of ints
	idsParam = strings.Trim(idsParam, "[]")
	idStrings := strings.Split(idsParam, ",")
	ids := make([]int, 0, len(idStrings))
	for _, idString := range idStrings {
		if len(idString) > 0 {
			id, err := strconv.Atoi(strings.TrimSpace(idString))
			if err != nil {
				return err
			}
			ids = append(ids, id)
		}
	}
	ctx := c.Request().Context()
	q, err := h.Service.GetByIDs(ctx, ids)
	if err != nil {
		if err == echo.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Question not found with id "+c.Param("id"))
		}
		println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get question")
	}
	return c.JSON(http.StatusOK, q)
}

func (h *VerbalQuestionHandler) GetRandomQuestions(c echo.Context) error {
	// Bind the request payload to the req struct
	req := models.RandomQuestionsRequest{}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	// Call the service function
	questions, err := h.Service.Random(c.Request().Context(), req.Limit, req.QuestionType, req.Competence, req.FramedAs, req.Difficulty, req.ExcludeIDs)
	if err != nil {
		fmt.Println(err.Error())
		if err == echo.ErrNotFound {
			// If a 404 error was returned, return an empty array
			return c.JSON(404, "No more questions for given criteria")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve random questions")
	}
	return c.JSON(http.StatusOK, questions)
}

func (h *VerbalQuestionHandler) GetAdaptiveQuestions(c echo.Context) error {
	// Bind the request payload to the req struct
	// Call the service function
	ctx := c.Request().Context()
	u, err := getUserClaims(c)
	if err != nil {
		return err
	}
	qidsToAvoidParam := c.QueryParam("questions")
	qidStrArr := strings.Split(strings.Trim(qidsToAvoidParam, "[]"), ",")
	qIds := make([]int, 0, len(qidStrArr))
	for _, idString := range qidStrArr {
		if len(idString) > 0 {
			id, err := strconv.Atoi(strings.TrimSpace(idString))
			if err != nil {
				return err
			}
			qIds = append(qIds, id)
		}
	}
	questions, err := h.Service.GetAdaptiveQuestions(ctx, u.Token, 5, qIds)
	if err != nil {
		fmt.Println(err.Error())
		if err == echo.ErrNotFound {
			// If a 404 error was returned, return an empty array
			return c.JSON(404, "No more questions for given criteria")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve random questions")
	}
	return c.JSON(http.StatusOK, questions)
}

func (h *VerbalQuestionHandler) GetQuestionsOnVocab(c echo.Context) error {
	ctx := c.Request().Context()
	u, err := getUserClaims(c)
	if err != nil {
		return err
	}
	wordsIdsParam := c.QueryParam("words")
	qidsToAvoidParam := c.QueryParam("questions")
	wordStrArr := strings.Split(strings.Trim(wordsIdsParam, "[]"), ",")
	qidStrArr := strings.Split(strings.Trim(qidsToAvoidParam, "[]"), ",")
	wordIds := make([]int, 0, len(wordStrArr))
	qIds := make([]int, 0, len(qidStrArr))
	for _, idString := range wordStrArr {
		if len(idString) > 0 {
			id, err := strconv.Atoi(strings.TrimSpace(idString))
			if err != nil {
				return err
			}
			wordIds = append(wordIds, id)
		}
	}
	for _, idString := range qidStrArr {
		if len(idString) > 0 {
			id, err := strconv.Atoi(strings.TrimSpace(idString))
			if err != nil {
				return err
			}
			qIds = append(qIds, id)
		}
	}
	questions, err := h.Service.GetQuestionsOnVocab(ctx, u.Token, qIds, wordIds)
	if err != nil {
		fmt.Println(err.Error())
		if err == echo.ErrNotFound {
			// If a 404 error was returned, return an empty array
			return c.JSON(404, "No more questions for given criteria")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve random questions")
	}
	return c.JSON(http.StatusOK, questions)
}
