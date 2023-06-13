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

/**
* Used to create a datapoint that represents the performance of a user
* for a particular question at a particular time in a many to many table.
**/
func (h *UserVerbalStatHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	u, err := getUserClaims(c)
	if err != nil {
		return err
	}
	var stat models.UserVerbalStat
	if err := c.Bind(&stat); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	if stat.QuestionID <= 0 || len(stat.Answers) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body. Requires questionID, correct, answers, and date")
	}
	err = h.Service.Create(ctx, &stat, u.Token)
	if err != nil {
		fmt.Println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user verbal stat")
	}
	return c.JSON(http.StatusCreated, stat)
}

/**
* Function that is used to get the verbal stats for a particular user
* token
**/
func (h *UserVerbalStatHandler) GetVerbalStatsByUserToken(c echo.Context) error {
	ctx := c.Request().Context()
	u, err := getUserClaims(c)
	if err != nil {
		return err
	}
	verbalStats, err := h.Service.GetVerbalStatsByUserToken(ctx, u.Token)
	if err != nil {
		fmt.Println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get verbal stats")
	}
	return c.JSON(http.StatusOK, verbalStats)
}
