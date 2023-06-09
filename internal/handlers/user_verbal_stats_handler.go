package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"grepandit.com/api/internal/models"
	"grepandit.com/api/internal/services"
)

type UserVerbalStatHandler struct {
	Service *services.UserVerbalStatsService
}

func NewUserVerbalStatHandler(s *services.UserVerbalStatsService) *UserVerbalStatHandler {
	return &UserVerbalStatHandler{Service: s}
}

// Create adds a new user verbal stat to the database with the given data.
//
// Example Request:
// POST /user-verbal-stats
// Content-Type: application/json
//
//	{
//	    "userToken": "abc123",
//	    "questionID": 1,
//	    "correct": true,
//	    "answers": ["option1", "option2"],
//	    "date": "2023-06-09"
//	}
//
// Example Response:
// HTTP/1.1 201 Created
// Content-Type: application/json
//
//	{
//	    "id": 1,
//	    "userToken": "abc123",
//	    "questionID": 1,
//	    "correct": true,
//	    "answers": ["option1", "option2"],
//	    "date": "2023-06-09"
//	}
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the created user verbal stat data.
func (h *UserVerbalStatHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var stat models.UserVerbalStat
	if err := c.Bind(&stat); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	if stat.UserToken == "" || stat.QuestionID == 0 || len(stat.Answers) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body. Requires userToken, questionID, correct, answers, and date")
	}
	err := h.Service.Create(ctx, &stat, nil) // Pass nil for wordIDs as it is not used in this handler
	if err != nil {
		fmt.Println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user verbal stat")
	}
	return c.JSON(http.StatusCreated, stat)
}

// GetMarkedWordsByUserToken retrieves all marked words for a user token from the database.
//
// Example Request:
// GET /user-verbal-stats/marked-words?userToken=abc123
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
//	[
//	    {
//	        "id": 1,
//	        "userToken": "abc123",
//	        "wordID": 1
//	    },
//	    {
//	        "id": 2,
//	        "userToken": "abc123",
//	        "wordID": 2
//	    }
//	]
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the marked words data.
func (h *UserVerbalStatHandler) GetMarkedWordsByUserToken(c echo.Context) error {
	ctx := c.Request().Context()
	userToken := c.QueryParam("userToken")
	markedWords, err := h.Service.GetMarkedWordsByUserToken(ctx, userToken)
	if err != nil {
		fmt.Println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get marked words")
	}
	return c.JSON(http.StatusOK, markedWords)
}

// GetMarkedVerbalQuestionsByUserToken retrieves all marked verbal questions for a user token from the database.
//
// Example Request:
// GET /user-verbal-stats/marked-questions?userToken=abc123
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
//	[
//	    {
//	        "id": 1,
//	        "userToken": "abc123",
//	        "verbalQuestionID": 1
//	    },
//	    {
//	        "id": 2,
//	        "userToken": "abc123",
//	        "verbalQuestionID": 2
//	    }
//	]
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the marked verbal questions data.
func (h *UserVerbalStatHandler) GetMarkedVerbalQuestionsByUserToken(c echo.Context) error {
	ctx := c.Request().Context()
	userToken := c.QueryParam("userToken")
	markedQuestions, err := h.Service.GetMarkedVerbalQuestionsByUserToken(ctx, userToken)
	if err != nil {
		fmt.Println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get marked verbal questions")
	}
	return c.JSON(http.StatusOK, markedQuestions)
}

// GetVerbalStatsByUserToken retrieves all verbal stats for a user token from the database.
//
// Example Request:
// GET /user-verbal-stats?userToken=abc123
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
//	[
//	    {
//	        "id": 1,
//	        "userToken": "abc123",
//	        "questionID": 1,
//	        "correct": true,
//	        "answers": ["option1", "option2"],
//	        "date": "2023-06-09"
//	    },
//	    {
//	        "id": 2,
//	        "userToken": "abc123",
//	        "questionID": 2,
//	        "correct": false,
//	        "answers": ["option1"],
//	        "date": "2023-06-10"
//	    }
//	]
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the verbal stats data.
func (h *UserVerbalStatHandler) GetVerbalStatsByUserToken(c echo.Context) error {
	ctx := c.Request().Context()
	userToken := c.QueryParam("userToken")
	verbalStats, err := h.Service.GetVerbalStatsByUserToken(ctx, userToken)
	if err != nil {
		fmt.Println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get verbal stats")
	}
	return c.JSON(http.StatusOK, verbalStats)
}
