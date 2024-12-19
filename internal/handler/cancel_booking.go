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
// @Param id path int true "Booking ID"
// @Success 200 {object} map[string]string "Cancellation Successful"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 403 {object} map[string]string "Not authorized to cancel this booking"
// @Failure 404 {object} map[string]string "Booking not found"
// @Failure 500 {object} map[string]string "Cancellation failed"
// @Router /rooms/cancel/{id} [delete]
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
		ActivityType: "Cancel Booking",
		ReferenceID:  booking.ID,
	}

	if err := config.DB.Create(&log).Error; err != nil {
		// Log the error internally without breaking the response
		c.Logger().Error("Failed to log user history: ", err)
	}
	response := map[string]interface{}{
		"message": "Booking successfully cancelled",
	}

	if refundAmount > 0 {
		response["refund_amount"] = refundAmount
	}

	// Return success response
	return c.JSON(http.StatusOK, response)
}
