package handlers

import (
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"grepandit.com/api/internal/models"
)

/**
* Extract user info as string values from the passed access
* token
**/
func getUserClaims(c echo.Context) (models.User, error) {
	claims := c.Get("user")
	if claims == nil {
		return models.User{}, echo.NewHTTPError(http.StatusUnauthorized, "No user context available")
	}
	claimsMap, ok := claims.(jwt.MapClaims)
	if !ok {
		return models.User{}, echo.NewHTTPError(http.StatusInternalServerError, "User context is of the wrong type")
	}
	sub, subExists := claimsMap["sub"]
	if !subExists {
		return models.User{}, echo.NewHTTPError(http.StatusInternalServerError, "No user ID available in the user context")
	}
	email, emailExists := claimsMap["email"]
	if !emailExists {
		return models.User{}, echo.NewHTTPError(http.StatusInternalServerError, "No email available in the user context")
	}
	u := models.User{
		Token: sub.(string),
		Email: email.(string),
	}
	return u, nil
}
