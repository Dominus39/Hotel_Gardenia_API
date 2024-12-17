package handler

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	"github.com/labstack/echo/v4"
)

// GetUserActivities godoc
// @Summary Get user activities
// @Description Retrieve all activities of the current user from the database.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success message and list of user activities"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve activities"
// @Router /activities [get]
func GetHistory(c echo.Context) error {
	// Retrieve the current user from the context (set by the JWT middleware)
	userClaims, ok := c.Get("user").(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": "Failed to parse user claims",
		})
	}

	// Extract user ID from claims
	currentUserID, ok := userClaims["id"].(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": "User ID not found in claims",
		})
	}

	// Query user histories from the database
	var history []entity.UserHistory
	if err := config.DB.Where("user_id = ?", int(currentUserID)).Order("created_at desc").Find(&history).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to retrieve history",
		})
	}

	// Prepare the response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "User history retrieved successfully",
		"history": history,
	})
}
