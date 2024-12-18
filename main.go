package main

import (
	config "MiniProjectPhase2/config/database"
	"MiniProjectPhase2/internal/handler"
	internal "MiniProjectPhase2/internal/middleware"

	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/joho/godotenv"

	"fmt"
	"os"

	echoswagger "github.com/swaggo/echo-swagger"
)

// @title GC 3 API
// @version 1.0
// @description social media hactivgram
// @host localhost:8080
// @BasePath /
func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize the database
	config.InitDB()

	e := echo.New()
	e.GET("/swagger/*", echoswagger.WrapHandler)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	public := e.Group("")
	public.POST("/users/register", handler.Register)
	public.POST("/users/login", handler.LoginUser)
	public.GET("/rooms", handler.GetRooms)

	private := e.Group("")
	private.Use(internal.CustomJwtMiddleware)
	private.GET("/rooms/users", handler.GetUserRooms)
	private.POST("/rooms/booking", handler.BookRoom)
	private.POST("/rooms/update/:id", handler.UpdateBooking)
	private.POST("/rooms/payment/:id", handler.PayBooking)
	private.DELETE("/rooms/booking/:id", handler.CancelBooking)
	private.POST("/users/topup", handler.TopUpBalance)

	private.GET("/users/history", handler.GetHistory)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on port:", port)
	e.Logger.Fatal(e.Start(":" + port))
}
