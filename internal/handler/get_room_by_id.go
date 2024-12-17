package handler

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// GetRoomByID godoc
// @Summary Get a room by ID
// @Description Fetch room details by ID, including its category information.
// @Tags Rooms
// @Accept json
// @Produce json
// @Param id path int true "Room ID"
// @Success 200 {object} map[string]interface{} "Room details with category"
// @Failure 404 {object} map[string]string "Room not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /rooms/{id} [get]
func GetRoomByID(c echo.Context) error {
	// Parse the room ID from the path
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid room ID"})
	}

	// Create a variable to hold the room data
	var room entity.Room

	// Query the room with the given ID and preload the Category data
	if err := config.DB.Preload("Category").First(&room, roomID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Room not found"})
	}

	// Build the response
	response := map[string]interface{}{
		"id":       room.ID,
		"name":     room.Name,
		"category": room.Category.Name,
		"price":    room.Category.Price,
		"stock":    room.Stock,
	}

	// Return the room details
	return c.JSON(http.StatusOK, response)
}
