package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"grepandit.com/api/internal/models"
	"grepandit.com/api/internal/services"
)

type UserHandler struct {
	Service *services.UserService
}

func NewUserHandler(s *services.UserService) *UserHandler {
	return &UserHandler{Service: s}
}

// Create adds a new user to the database with the given data.
//
// Example Request:
// POST /users
// Content-Type: application/json
//
//	{
//	    "token": "abc123",
//	    "email": "user@example.com"
//	}
//
// Example Response:
// HTTP/1.1 201 Created
// Content-Type: application/json
//
//	{
//	    "id": 1,
//	    "token": "abc123",
//	    "email": "user@example.com"
//	}
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the created user data.
func (h *UserHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var u models.User
	if err := c.Bind(&u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	if u.Token == "" || u.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body. Requires token and email")
	}
	err := h.Service.Create(ctx, &u)
	if err != nil {
		fmt.Println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}
	return c.JSON(http.StatusCreated, u)
}

// Update updates an existing user in the database with the given data.
//
// Example Request:
// PUT /users/1
// Content-Type: application/json
//
//	{
//	    "token": "abc123",
//	    "email": "newemail@example.com"
//	}
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
//	{
//	    "id": 1,
//	    "token": "abc123",
//	    "email": "newemail@example.com"
//	}
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the updated user data.
func (h *UserHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}
	var u models.User
	if err := c.Bind(&u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	if u.Token == "" || u.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body. Requires token and email")
	}
	err = h.Service.Update(ctx, id, &u)
	if err != nil {
		fmt.Println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user")
	}
	return c.JSON(http.StatusOK, u)
}

// GetByID retrieves a user from the database by its ID and returns its data.
//
// Example Request:
// GET /users/1
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
//	{
//	    "id": 1,
//	    "token": "abc123",
//	    "email": "user@example.com"
//	}
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the user data.
func (h *UserHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}
	u, err := h.Service.GetByID(ctx, id)
	if err != nil {
		fmt.Println(err.Error())
		if err == echo.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User not found with id "+c.Param("id"))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user")
	}
	return c.JSON(http.StatusOK, u)
}

// GetByEmail retrieves a user from the database by its email and returns its data.
//
// Example Request:
// GET /users?email=user@example.com
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
//	{
//	    "id": 1,
//	    "token": "abc123",
//	    "email": "user@example.com"
//	}
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the user data.
func (h *UserHandler) GetByEmail(c echo.Context) error {
	ctx := c.Request().Context()
	email := c.QueryParam("email")
	u, err := h.Service.GetByEmail(ctx, email)
	if err != nil {
		fmt.Println(err.Error())
		if err == echo.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User not found with email "+email)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user")
	}
	return c.JSON(http.StatusOK, u)
}

// GetByUserToken retrieves a user from the database by its user token and returns its data.
//
// Example Request:
// GET /users?token=abc123
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
//	{
//	    "id": 1,
//	    "token": "abc123",
//	    "email": "user@example.com"
//	}
//
// @param c An echo.Context instance.
// @return An error response or a JSON response with the user data.
func (h *UserHandler) GetByUserToken(c echo.Context) error {
	ctx := c.Request().Context()
	token := c.QueryParam("token")
	u, err := h.Service.GetByUserToken(ctx, token)
	if err != nil {
		fmt.Println(err.Error())
		if err == echo.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User not found with token "+token)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user")
	}
	return c.JSON(http.StatusOK, u)
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
func (h *UserHandler) GetMarkedWordsByUserToken(c echo.Context) error {
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
func (h *UserHandler) GetMarkedVerbalQuestionsByUserToken(c echo.Context) error {
	ctx := c.Request().Context()
	userToken := c.QueryParam("userToken")
	markedQuestions, err := h.Service.GetMarkedVerbalQuestionsByUserToken(ctx, userToken)
	if err != nil {
		fmt.Println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get marked verbal questions")
	}
	return c.JSON(http.StatusOK, markedQuestions)
}

// AddMarkedWords adds marked words for a user to the database.
func (h *UserHandler) AddMarkedWords(c echo.Context) error {
	ctx := c.Request().Context()
	// Parse the request body into a marked words model
	var requestBody struct {
		UserToken string `json:"userToken"`
		WordIDs   []int  `json:"wordIDs"`
	}
	if err := c.Bind(&requestBody); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	// Call the service method to add the marked words
	err := h.Service.AddMarkedWords(ctx, requestBody.UserToken, requestBody.WordIDs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add marked words")
	}
	return c.JSON(http.StatusCreated, requestBody)
}

// AddMarkedQuestions adds marked questions for a user to the database.
func (h *UserHandler) AddMarkedQuestions(c echo.Context) error {
	ctx := c.Request().Context()
	// Parse the request body into a marked questions model
	var requestBody struct {
		UserToken   string `json:"userToken"`
		QuestionIDs []int  `json:"questionIDs"`
	}
	if err := c.Bind(&requestBody); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	// Call the service method to add the marked questions
	err := h.Service.AddMarkedQuestions(ctx, requestBody.UserToken, requestBody.QuestionIDs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add marked questions")
	}
	return c.JSON(http.StatusCreated, requestBody)
}