package handler

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// PayBooking godoc
// @Summary Pay for a booking
// @Description Pay the total price of a booking with the user's balance and mark it as paid.
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} map[string]interface{} "Payment successful"
// @Failure 400 {object} map[string]string "Insufficient balance or invalid request"
// @Failure 404 {object} map[string]string "Booking not found"
// @Failure 500 {object} map[string]string "Payment failed"
// @Router /rooms/payment/{id} [post]
func PayBooking(c echo.Context) error {
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

	// Parse the booking ID from the path
	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid booking ID"})
	}

	// Find the booking by ID and ensure it belongs to the authenticated user
	var booking entity.Booking
	if err := config.DB.Where("id = ? AND user_id = ?", bookingID, userID).Preload("Room").First(&booking).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Booking not found"})
	}

	// Check if a payment already exists for this booking
	var payment entity.Payment
	paymentExists := config.DB.Where("booking_id = ?", bookingID).First(&payment).Error == nil

	// If payment exists and is already paid, return an error
	if paymentExists && payment.IsPaid {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Booking is already paid"})
	}

	// Find the user's account
	var userAcc entity.User
	if err := config.DB.First(&userAcc, userID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "User not found"})
	}

	// Check if the user has enough balance
	if userAcc.Balance < booking.TotalPrice {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Insufficient balance"})
	}

	// Deduct the total price from the user's balance
	userAcc.Balance -= booking.TotalPrice
	if err := config.DB.Save(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update user balance"})
	}

	// Create or update the payment record
	if paymentExists {
		payment.IsPaid = true
		payment.PaidAt = timePtr(time.Now())
		if err := config.DB.Save(&payment).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update payment record"})
		}
	} else {
		payment = entity.Payment{
			BookingID: booking.ID,
			Amount:    booking.TotalPrice,
			IsPaid:    true,
			PaidAt:    timePtr(time.Now()),
		}
		if err := config.DB.Create(&payment).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to create payment record"})
		}
	}

	// Add a record to the user's history
	userHistory := entity.UserHistory{
		UserID:       userID,
		Description:  "Payment for booking ID " + strconv.Itoa(bookingID),
		ActivityType: "Payment",
		ReferenceID:  bookingID,
	}
	if err := config.DB.Create(&userHistory).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to record user activity"})
	}

	// Respond with success message
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     "Payment successful",
		"booking_id":  booking.ID,
		"room_name":   booking.Room.Name,
		"total_price": booking.TotalPrice,
		"balance":     userAcc.Balance,
		"is_paid":     payment.IsPaid,
		"paid_at":     payment.PaidAt,
	})
}

func timePtr(t time.Time) *time.Time {
	return &t
}
