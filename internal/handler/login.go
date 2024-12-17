package handler

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/entity"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// LoginUser godoc
// @Summary Login a user
// @Description This endpoint allows users to login by providing email and password.
// @Tags Users
// @Accept json
// @Produce json
// @Param login body entity.LoginRequest true "Login User"
// @Success 200 {object} entity.LoginResponse
// @Failure 400 {object} map[string]interface{} "Invalid Request Parameters"
// @Failure 401 {object} map[string]interface{} "Invalid Email"
// @Failure 401 {object} map[string]interface{} "Invalid Password"
// @Failure 500 {object} map[string]interface{} "Invalid Generate Token"
// @Router /users/login [post]
func LoginUser(c echo.Context) error {
	var req entity.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid Request Parameters"})
	}
	var user entity.User
	err := config.DB.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid Email"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid Password"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte("12345"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Invalid Generate Token"})
	}

	return c.JSON(http.StatusOK, entity.LoginResponse{Token: tokenString})
}
