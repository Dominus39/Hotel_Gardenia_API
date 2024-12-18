package handler

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// TopUpBalance godoc
// @Summary Top up user balance
// @Description Authenticated users can top up their balance.
// @Tags Users
// @Accept json
// @Produce json
// @Param topup body TopUpRequest true "Top-Up Request"
// @Success 200 {object} map[string]interface{} "Top-Up Successful"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 500 {object} map[string]string "Top-Up failed"
// @Router /users/topup [post]
func TopUpBalance(c echo.Context) error {
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

	// Bind and validate the request
	var req entity.TopUpRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request parameters"})
	}

	if req.Amount <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Amount must be greater than zero"})
	}

	// Find the user by ID
	var userEntity entity.User
	if err := config.DB.First(&userEntity, userID).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "User not found"})
	}

	// Update user balance
	userEntity.Balance += req.Amount
	if err := config.DB.Save(&userEntity).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update balance"})
	}

	// Create a new payment record for the top-up
	payment := entity.PaymentForTopUp{
		UserID:    userID,
		Amount:    req.Amount,
		CreatedAt: time.Now(),
	}
	if err := config.DB.Create(&payment).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to record top-up payment"})
	}

	// Log the top-up in user history
	log := entity.UserHistory{
		UserID:       userID,
		Description:  "Top-up of " + fmt.Sprintf("%.2f", req.Amount),
		ActivityType: "TOPUP",
		ReferenceID:  payment.ID,
	}

	if err := config.DB.Create(&log).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to record user history"})
	}

	// Return success response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Top-Up successful",
		"balance": userEntity.Balance,
	})
}
