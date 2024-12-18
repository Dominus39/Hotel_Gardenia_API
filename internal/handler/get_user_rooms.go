package handler

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// GetUserRooms godoc
// @Summary Get booked rooms for the authenticated user
// @Description Fetch all rooms currently booked by the authenticated user, including payment status.
// @Tags Rooms
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{} "List of booked rooms with payment status"
// @Failure 401 {object} map[string]string "Unauthorized access"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /rooms/booked [get]
func GetUserRooms(c echo.Context) error {
	// Extract user claims from JWT
	user := c.Get("user")
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Unauthorized access"})
	}

	// Parse user claims as jwt.MapClaims
	claims, ok := user.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to parse user claims"})
	}

	// Extract user ID from claims
	userIDFloat, ok := claims["id"].(float64)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "User ID not found in claims"})
	}
	userID := int(userIDFloat)

	// Query the bookings for the authenticated user and preload room, category, and payment data
	var bookings []entity.Booking
	if err := config.DB.Preload("Room.Category").
		Preload("Payment"). // Load associated payment data
		Where("user_id = ?", userID).
		Find(&bookings).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch booked rooms"})
	}

	// Build the response
	var response []map[string]interface{}
	for _, booking := range bookings {
		// Check payment status for the current booking
		var payment entity.Payment
		isPaid := false
		if err := config.DB.Where("booking_id = ?", booking.ID).First(&payment).Error; err == nil {
			isPaid = payment.IsPaid
		}

		// Append booking details to the response
		response = append(response, map[string]interface{}{
			"booking_id":  booking.ID,
			"room_id":     booking.Room.ID,
			"room_name":   booking.Room.Name,
			"category":    booking.Room.Category.Name,
			"price":       booking.Room.Category.Price,
			"start_date":  booking.StartDate,
			"end_date":    booking.EndDate,
			"total_price": booking.TotalPrice,
			"is_paid":     isPaid,
		})
	}

	// Return the list of booked rooms
	return c.JSON(http.StatusOK, response)
}
