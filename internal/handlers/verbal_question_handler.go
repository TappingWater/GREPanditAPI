package handlers

import (
	"fmt"
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
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get question")
	}
	return c.JSON(http.StatusOK, q)
}

// Count retrieves the total number of verbal questions in the database.
//
// Example Request:
// GET /verbal-questions/count
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
// 10
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the total count of questions.
func (h *VerbalQuestionHandler) Count(c echo.Context) error {
	count, err := h.Service.Count(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get question count")
	}
	return c.JSON(http.StatusOK, count)
}

// GetRandomQuestions retrieves a specified number of random verbal questions from the database, filtered by various criteria.
//
// Example Request:
// POST /verbal-questions/random
// Content-Type: application/json
//
//	{
//	    "limit": 5,
//	    "question_type": "Verbal",
//	    "competence": "General Knowledge",
//	    "framed_as": "Question",
//	    "difficulty": 3,
//	    "exclude_ids": [1, 2, 3]
//	}
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
// [
//
//	{
//	    "id": 4,
//	    "text": "What is the capital of France?",
//	    "answer": "Paris"
//	    "difficulty": 2,
//	    "competence": "General Knowledge",
//	    "framed_as": "Question",
//	    "type": "Verbal"
//	},
//	{
//	    "id": 5,
//	    "text": "What is the largest country in the world by land area?",
//	    "answer": "Russia",
//	    "difficulty": 2,
//	    "competence": "General Knowledge",
//	    "framed_as": "Question",
//	    "type": "Verbal"
//	},
//	{
//	    "id": 6,
//	    "text": "Who is the current Prime Minister of the United Kingdom?",
//	    "answer": "Boris Johnson",
//	    "difficulty": 2,
//	    "competence": "General Knowledge",
//	    "framed_as": "Question",
//	    "type": "Verbal"
//	},
//	{
//	    "id": 7,
//	    "text": "What is the largest planet in our solar system?",
//	    "answer": "Jupiter",
//	    "difficulty": 2,
//	    "competence": "General Knowledge",
//	    "framed_as": "Question",
//	    "type": "Verbal"
//	},
//	{
//	    "id": 8,
//	    "text": "What is the smallest country in the world by land area?",
//	    "answer": "Vatican City",
//	    "difficulty": 2,
//	    "competence": "General Knowledge",
//	    "framed_as": "Question",
//	    "type": "Verbal"
//	}
//
// ]
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the array of random questions data.
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
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve random questions")
	}

	return c.JSON(http.StatusOK, questions)
}
