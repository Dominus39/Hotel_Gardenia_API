package handler

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// BookRoom godoc
// @Summary Book a room
// @Description Book a room for a given number of days and start date.
// @Tags Rooms
// @Accept json
// @Produce json
// @Param booking body BookingRequest true "Booking Request"
// @Success 200 {object} BookingResponse "Booking Successful"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 404 {object} map[string]string "Room not found"
// @Failure 500 {object} map[string]string "Booking failed"
// @Router /rooms/booking [post]
func BookRoom(c echo.Context) error {
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

	var req entity.BookingRequest

	// Bind and validate the request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request parameters"})
	}

	if req.RoomID == 0 || req.Days <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Room ID and number of days are required"})
	}

	// Find the room by ID with its category
	var room entity.Room
	if err := config.DB.Preload("Category").First(&room, req.RoomID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Room not found"})
	}

	// Check if the room has enough stock
	if room.Stock <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Room is fully booked"})
	}

	// Calculate end date and total price
	endDate := req.StartDate.AddDate(0, 0, req.Days)
	totalPrice := float64(req.Days) * room.Category.Price

	// Create a new booking record
	newBooking := entity.Booking{
		UserID:     userID,
		RoomID:     req.RoomID,
		StartDate:  req.StartDate,
		EndDate:    endDate,
		TotalPrice: totalPrice,
	}

	if err := config.DB.Create(&newBooking).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Booking failed"})
	}

	// Update room stock
	room.Stock -= 1
	if err := config.DB.Save(&room).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update room stock"})
	}

	log := entity.UserHistory{
		UserID:       userID,
		Description:  "Booked room " + room.Name,
		ActivityType: "Booking Room",
		ReferenceID:  newBooking.ID,
	}

	if err := config.DB.Create(&log).Error; err != nil {
		// Log the error internally without breaking the response
		c.Logger().Error("Failed to log user activity: ", err)
	}

	// Build response
	response := entity.BookingResponse{
		Message:    "Order successful",
		RoomName:   room.Name,
		Category:   room.Category.Name,
		TotalPrice: totalPrice,
	}

	return c.JSON(http.StatusOK, response)
}
