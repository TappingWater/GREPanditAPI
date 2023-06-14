package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"grepandit.com/api/internal/services"
)

type UserHandler struct {
	Service *services.UserService
}

func NewUserHandler(s *services.UserService) *UserHandler {
	return &UserHandler{Service: s}
}

/**
* Create adds a new user to the database with the given data.
* If user exits sends a 409 error
*
* @param c An echo.Context instance.
* @return An error response or a JSON response with the created user data.
**/
func (h *UserHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	u, err := getUserClaims(c)
	if err != nil {
		return err
	}
	err = h.Service.Create(ctx, &u)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return echo.NewHTTPError(409, "User already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}
	return c.JSON(http.StatusCreated, u)
}

/**
* Retrieves a user resource based on the passed access token.
* The claims contains the sub token which is the user ID that is
* used to retrieve user data
**/
func (h *UserHandler) Get(c echo.Context) error {
	ctx := c.Request().Context()
	user, err := getUserClaims(c)
	if err != nil {
		return err
	}
	u, err := h.Service.Get(ctx, user.Token)
	if err != nil {
		fmt.Println(err.Error())
		if err == echo.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User not found with id "+c.Param("id"))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user")
	}
	return c.JSON(http.StatusOK, u)
}

/**
* Retrieves a list of words that have been marked by the user based
* on the sub claims passed from the access token
**/
func (h *UserHandler) GetMarkedWordsByUserToken(c echo.Context) error {
	ctx := c.Request().Context()
	user, err := getUserClaims(c)
	if err != nil {
		return err
	}
	markedWords, err := h.Service.GetMarkedWordsByUserToken(ctx, user.Token)
	if err != nil {
		fmt.Println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get marked words")
	}
	return c.JSON(http.StatusOK, markedWords)
}

/**
* Retrieves a list of verbal questions that have been marked by a user based
* on the sub claims within the access token
**/
func (h *UserHandler) GetMarkedVerbalQuestionsByUserToken(c echo.Context) error {
	ctx := c.Request().Context()
	user, err := getUserClaims(c)
	if err != nil {
		return err
	}
	markedQuestions, err := h.Service.GetMarkedVerbalQuestionsByUserToken(ctx, user.Token)
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
		WordIDs []int `json:"words"`
	}
	if err := c.Bind(&requestBody); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	user, err := getUserClaims(c)
	if err != nil {
		return err
	}
	// Call the service method to add the marked words
	err = h.Service.AddMarkedWords(ctx, user.Token, requestBody.WordIDs)
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
		QuestionIDs []int `json:"questions"`
	}
	if err := c.Bind(&requestBody); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	user, err := getUserClaims(c)
	if err != nil {
		return err
	}
	// Call the service method to add the marked questions
	err = h.Service.AddMarkedQuestions(ctx, user.Token, requestBody.QuestionIDs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add marked questions")
	}
	return c.JSON(http.StatusCreated, requestBody)
}

// Remove marked words for a user from the database
func (h *UserHandler) RemoveMarkedWords(c echo.Context) error {
	ctx := c.Request().Context()
	// Parse the request body into a marked words model
	var requestBody struct {
		WordIDs []int `json:"words"`
	}
	if err := c.Bind(&requestBody); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	user, err := getUserClaims(c)
	if err != nil {
		return err
	}
	// Call the service method to add the marked words
	err = h.Service.RemoveMarkedWords(ctx, user.Token, requestBody.WordIDs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add marked words")
	}
	return c.JSON(http.StatusCreated, requestBody)
}

// Remove marked words for a user from the database
func (h *UserHandler) RemoveMarkedQuestions(c echo.Context) error {
	ctx := c.Request().Context()
	// Parse the request body into a marked words model
	var requestBody struct {
		QuestionIDs []int `json:"questions"`
	}
	if err := c.Bind(&requestBody); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	user, err := getUserClaims(c)
	if err != nil {
		return err
	}
	// Call the service method to add the marked words
	err = h.Service.RemoveMarkedQuestions(ctx, user.Token, requestBody.QuestionIDs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add marked words")
	}
	return c.JSON(http.StatusCreated, requestBody)
}
