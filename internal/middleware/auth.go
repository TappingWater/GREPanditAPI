package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/jwk"
)

/**
* Middleware that is used to authenticate the JWT token sent
* by the front end with each request. It confirms that there is
* an authorization header and it is issued from the user pool based
* on the available public keyset
**/
func JWTAuthMiddleware(jwkSet jwk.Set) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "No bearer token found within Authorization header")
			}
			// Remove "Bearer " prefix from the token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid access token format")
			}
			// Parse and verify the token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				kid, ok := token.Header["kid"].(string)
				if !ok {
					return nil, fmt.Errorf("invalid token")
				}
				key, ok := jwkSet.LookupKeyID(kid)
				if !ok {
					return nil, fmt.Errorf("unknown key ID")
				}
				var pubkey interface{}
				if err := key.Raw(&pubkey); err != nil {
					return nil, fmt.Errorf("failed to create public key: %s", err)
				}
				return pubkey, nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired access token")
			}
			// Store the token claims in the context
			c.Set("user", token.Claims)
			return next(c)
		}
	}
}
