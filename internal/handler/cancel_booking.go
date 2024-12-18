package handler

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// CancelBooking godoc
// @Summary Cancel a booking
// @Description Cancel a user's booking by booking ID. Only the owner of the booking can cancel it.
// @Tags Rooms
// @Accept json
// @Produce json
// @Param cancel body CancelBookingRequest true "Cancel Booking Request"
// @Success 200 {object} map[string]string "Cancellation Successful"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 403 {object} map[string]string "Not authorized to cancel this booking"
// @Failure 404 {object} map[string]string "Booking not found"
// @Failure 500 {object} map[string]string "Cancellation failed"
// @Router /rooms/cancel [post]
func CancelBooking(c echo.Context) error {
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

	// Parse booking ID from path parameters
	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid booking ID"})
	}

	// Find the booking by ID
	var booking entity.Booking
	if err := config.DB.Preload("Room").First(&booking, bookingID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Booking not found"})
	}

	// Check if the booking belongs to the logged-in user
	if booking.UserID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"message": "You are not authorized to cancel this booking"})
	}

	// Cancel the booking: delete the record
	if err := config.DB.Delete(&booking).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Cancellation failed"})
	}

	// Increment the stock of the room
	booking.Room.Stock += 1
	if err := config.DB.Save(&booking.Room).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update room stock"})
	}

	// Log the cancellation in user history
	log := entity.UserHistory{
		UserID:       userID,
		Description:  "Canceled booking for room " + booking.Room.Name,
		ActivityType: "CANCELLATION",
		ReferenceID:  booking.ID,
	}

	if err := config.DB.Create(&log).Error; err != nil {
		// Log the error internally without breaking the response
		c.Logger().Error("Failed to log user history: ", err)
	}

	// Return success response
	return c.JSON(http.StatusOK, map[string]string{"message": "Booking canceled successfully"})
}
