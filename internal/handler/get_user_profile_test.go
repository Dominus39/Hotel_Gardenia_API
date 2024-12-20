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
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestUserProfile(t *testing.T) {
	// Mock the database connection
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	config.DB, err = gorm.Open(postgres.New(postgres.Config{
		Conn: mockDB,
	}), &gorm.Config{})
	assert.NoError(t, err)

	// Define the mock user profile
	mockUser := entity.User{
		ID:      1,
		Name:    "Peter Parker",
		Email:   "peter@example.com",
		Balance: 500000,
	}

	// Mock the query for user profile
	mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"."id" = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "balance"}).
			AddRow(mockUser.ID, mockUser.Name, mockUser.Email, mockUser.Balance))

	// Setup Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/profile", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Add JWT claims to context
	claims := jwt.MapClaims{
		"id": float64(1),
	}
	c.Set("user", claims)

	// Call the handler
	err = handler.UserProfile(c)

	// Assert the response
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response body
	var response entity.UserProfileResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify the response content
	assert.Equal(t, 1, response.ID)
	assert.Equal(t, "Peter Parker", response.Name)
	assert.Equal(t, "peter@example.com", response.Email)
	assert.Equal(t, float64(500000), response.Balance)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
