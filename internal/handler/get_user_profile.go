package handler

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// UserProfile godoc
// @Summary Get user profile
// @Description Retrieve the authenticated user's profile, including their balance.
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "User profile retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized access"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/profile [get]
func UserProfile(c echo.Context) error {
	// Extract user claims from JWT
	user := c.Get("user")
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"message": "Unauthorized access"})
	}

	// Parse user claims as jwt.MapClaims
	claims, ok := user.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "Failed to parse user claims"})
	}

	// Extract user ID from claims
	userIDFloat, ok := claims["id"].(float64)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "User ID not found in claims"})
	}
	userID := int(userIDFloat)

	// Retrieve the user from the database
	var userProfile entity.User
	if err := config.DB.First(&userProfile, userID).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "Failed to retrieve user profile"})
	}

	// Build the response
	response := entity.UserProfileResponse{
		ID:      userProfile.ID,
		Name:    userProfile.Name,
		Email:   userProfile.Email,
		Balance: userProfile.Balance,
	}

	return c.JSON(http.StatusOK, response)
}
