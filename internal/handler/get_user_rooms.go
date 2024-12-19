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
		Where("user_id = ?", userID).
		Find(&bookings).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch booked rooms"})
	}

	// Build the response
	var response []entity.GetUserRoomsResponse
	for _, booking := range bookings {
		// Check payment status for the current booking
		var payment entity.Booking
		isPaid := false
		if err := config.DB.Where("id = ?", booking.ID).First(&payment).Error; err == nil {
			isPaid = payment.IsPaid
		}

		// Append booking details to the response
		response = append(response, entity.GetUserRoomsResponse{
			BookingID:  booking.ID,
			RoomID:     booking.Room.ID,
			RoomName:   booking.Room.Name,
			Category:   booking.Room.Category.Name,
			Price:      booking.Room.Category.Price,
			StartDate:  booking.StartDate,
			EndDate:    booking.EndDate,
			TotalPrice: booking.TotalPrice,
			IsPaid:     isPaid,
		})
	}

	// Return the list of booked rooms
	return c.JSON(http.StatusOK, response)
}
