package handler_test

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"MiniProjectPhase2/internal/handler"
	"bytes"
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

func TestLoginRequestPayload(t *testing.T) {
	// Mock the database connection
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	config.DB, err = gorm.Open(postgres.New(postgres.Config{
		Conn: mockDB,
	}), &gorm.Config{})
	assert.NoError(t, err)

	mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE email = \\$1 ORDER BY \"users\".\"id\" LIMIT 1").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).
			AddRow(1, "test@example.com", "$2a$10$hashedpassword123"))

	// Define a mock login request payload
	e := echo.New()
	reqBody := entity.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	// Create a new HTTP POST request
	req := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call the handler with the context
	err = handler.LoginUser(c)

	// Assert the payload handling
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

}
