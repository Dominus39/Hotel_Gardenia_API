package handler_test

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"MiniProjectPhase2/internal/handler"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGetRooms(t *testing.T) {
	// Step 1: Mock the database connection
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	config.DB, err = gorm.Open(postgres.New(postgres.Config{
		Conn: mockDB,
	}), &gorm.Config{})
	assert.NoError(t, err)

	// Step 2: Prepare mock data for rooms
	rows := sqlmock.NewRows([]string{"id", "name", "category_id", "stock"}).
		AddRow(1, "Room 101", 1, 10).
		AddRow(2, "Room 102", 2, 5)

	// Mock the query for rooms
	mock.ExpectQuery("^SELECT \\* FROM \"rooms\"").WillReturnRows(rows)

	// Step 3: Mock the query for categories
	mock.ExpectQuery("^SELECT \\* FROM \"categories\"").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "price"}).
			AddRow(1, "Deluxe", 100).
			AddRow(2, "Standard", 50),
	)

	// Step 4: Create a new HTTP request for GET /rooms
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Step 5: Call the handler function
	err = handler.GetRooms(c)

	// Step 6: Assert the response
	assert.NoError(t, err) // Ensure no error occurs during execution

	// Step 7: Check the status code and response body
	assert.Equal(t, http.StatusOK, rec.Code)

	var rooms []entity.RoomResponse
	err = json.Unmarshal(rec.Body.Bytes(), &rooms)
	assert.NoError(t, err)

	// Check if the rooms data is correctly returned
	assert.Equal(t, 2, len(rooms))
	assert.Equal(t, "Room 101", rooms[0].Name)
	assert.Equal(t, "Deluxe", rooms[0].Category)
	assert.Equal(t, float64(100), rooms[0].Price)
	assert.Equal(t, 10, rooms[0].Stock)

	assert.Equal(t, "Room 102", rooms[1].Name)
	assert.Equal(t, "Standard", rooms[1].Category)
	assert.Equal(t, float64(50), rooms[1].Price)
	assert.Equal(t, 5, rooms[1].Stock)

	// Ensure all mock expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
