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
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTopUpBalanceAmountValidation(t *testing.T) {
	// Mock the database connection
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	config.DB, err = gorm.Open(postgres.New(postgres.Config{
		Conn: mockDB,
	}), &gorm.Config{})
	assert.NoError(t, err)

	// Define test payload
	e := echo.New()
	reqBody := entity.TopUpRequest{
		Amount: -500,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users/topup", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	// Step 3: Create a new recorder and context
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Generate JWT token (mock)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": 1,
	})
	// Sign token with a mock secret key
	tokenString, _ := token.SignedString([]byte("12345"))
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Manually add the user to the context
	c.Set("user", jwt.MapClaims{
		"id": float64(1),
	})

	// Call the handler
	err = handler.TopUpBalance(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// Check if the response message indicates invalid amount
	var resp map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Amount must be greater than zero", resp["message"])

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
