package handler

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// UpdateBooking godoc
// @Summary Update a booked room or duration
// @Description Allows users to change their booked room or the number of booking days.
// @Tags Bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param update body UpdateBookingRequest true "Update Booking Request"
// @Success 200 {object} map[string]string "Booking successfully updated"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Booking or room not found"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /bookings/{id} [put]
func UpdateBooking(c echo.Context) error {
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

	// Parse booking ID from path
	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid booking ID"})
	}

	// Parse the update request body
	var req entity.UpdateBookingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}

	// Find the existing booking
	var booking entity.Booking
	if err := config.DB.Preload("Room.Category").
		Where("id = ? AND user_id = ?", bookingID, userID).
		First(&booking).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Booking not found"})
	}

	// If the booking has been paid, refund the payment and reset IsPaid
	var refundAmount float64
	if booking.IsPaid {
		// Find the payment record
		var payment entity.PaymentForBooking
		if err := config.DB.Where("booking_id = ?", booking.ID).First(&payment).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Payment record not found"})
		}

		// Refund the payment to the user
		var userAcc entity.User
		if err := config.DB.First(&userAcc, userID).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "User not found"})
		}

		// Refund the total price to the user's balance
		userAcc.Balance += payment.Amount
		refundAmount = payment.Amount
		if err := config.DB.Save(&userAcc).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to refund user balance"})
		}

		// Reset the IsPaid status and delete the payment record
		booking.IsPaid = false
		if err := config.DB.Save(&booking).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update booking status"})
		}
		if err := config.DB.Delete(&payment).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to delete payment record"})
		}
	}

	// Adjust stock if the room is changed
	if req.NewRoomID != 0 && req.NewRoomID != booking.RoomID {
		// Check if the new room exists and has enough stock
		var newRoom entity.Room
		if err := config.DB.Preload("Category").
			Where("id = ?", req.NewRoomID).
			First(&newRoom).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "New room not found"})
		}
		if newRoom.Stock <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "New room is fully booked"})
		}

		// Update stock for the old and new rooms
		oldRoom := booking.Room
		oldRoom.Stock++
		newRoom.Stock--
		if err := config.DB.Save(&oldRoom).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update old room stock"})
		}
		if err := config.DB.Save(&newRoom).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update new room stock"})
		}

		// Update booking with the new room
		booking.RoomID = req.NewRoomID
		booking.Room = newRoom
	}

	// Update duration and start date if requested
	if req.NewDays > 0 {
		booking.EndDate = booking.StartDate.AddDate(0, 0, req.NewDays)
		booking.TotalPrice = float64(req.NewDays) * booking.Room.Category.Price
	}
	if !req.StartDate.IsZero() {
		booking.StartDate = req.StartDate
		booking.EndDate = req.StartDate.AddDate(0, 0, req.NewDays)
	}

	// Save the updated booking
	if err := config.DB.Save(&booking).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update booking"})
	}

	// Log this action in user history
	history := entity.UserHistory{
		UserID:       userID,
		Description:  "Updated booking for room: " + booking.Room.Name,
		ActivityType: "update_booking",
		ReferenceID:  booking.ID,
	}
	if err := config.DB.Create(&history).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to log user history"})
	}

	// Build the response
	response := map[string]interface{}{
		"message":     "Booking successfully updated",
		"room":        booking.Room.Name,
		"total_price": booking.TotalPrice,
	}
	if refundAmount > 0 {
		response["refund_amount"] = refundAmount
	}
	return c.JSON(http.StatusOK, response)
}
