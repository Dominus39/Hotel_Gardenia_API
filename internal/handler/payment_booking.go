package handler

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"

	"MiniProjectPhase2/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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

	// Begin GORM transaction
	tx := config.DB.Begin()

	// Find the booking by ID and ensure it belongs to the authenticated user
	var booking entity.Booking
	if err := tx.Where("id = ? AND user_id = ?", bookingID, userID).Preload("Room").First(&booking).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Booking not found"})
	}

	// Check if a payment already exists for this booking
	var payment entity.PaymentForBooking
	paymentExists := true
	if err := tx.Where("booking_id = ?", bookingID).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			paymentExists = false
		} else {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to query payment record"})
		}
	}

	// If payment exists and is already paid, return an error
	if paymentExists && booking.IsPaid {
		tx.Rollback()
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Booking is already paid"})
	}

	// Find the user's account
	var userAcc entity.User
	if err := tx.First(&userAcc, userID).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusNotFound, map[string]string{"message": "User not found"})
	}

	// Check if the user has enough balance
	if userAcc.Balance < booking.TotalPrice {
		tx.Rollback()
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Insufficient balance"})
	}

	// Deduct the total price from the user's balance
	userAcc.Balance -= booking.TotalPrice
	if err := tx.Save(&userAcc).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update user balance"})
	}

	// Create or update the payment record
	if paymentExists {
		booking.IsPaid = true
		payment.CreatedAt = time.Now()
		if err := tx.Save(&payment).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update payment record"})
		}
	} else {
		booking.IsPaid = true
		payment = entity.PaymentForBooking{
			BookingID: booking.ID,
			Amount:    booking.TotalPrice,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&payment).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to create payment record"})
		}
	}
	// Save the booking updates
	if err := tx.Save(&booking).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update booking status"})
	}

	// Add a record to the user's history
	userHistory := entity.UserHistory{
		UserID:       userID,
		Description:  "Payment for booking ID " + strconv.Itoa(bookingID),
		ActivityType: "Payment",
		ReferenceID:  bookingID,
	}
	if err := tx.Create(&userHistory).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to record user activity"})
	}

	// Create the invoice via Xendit API
	_, err = utils.CreateInvoice(booking, userAcc)
	if err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to create invoice", "error": err.Error()})
	}

	// Commit transaction
	tx.Commit()

	// Respond with success message
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     "Payment successful",
		"booking_id":  booking.ID,
		"room_name":   booking.Room.Name,
		"total_price": booking.TotalPrice,
		"balance":     userAcc.Balance,
		"is_paid":     booking.IsPaid,
		"paid_at":     payment.CreatedAt,
	})
}
