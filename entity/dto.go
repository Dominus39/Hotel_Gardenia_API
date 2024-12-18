package entity

import "time"

type RegisterUser struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required, email"`
	Password string `json:"password" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required, email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RoomResponse struct {
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	Stock    int     `json:"stock"`
}

type BookingRequest struct {
	RoomID    int       `json:"room_id" validate:"required"`
	StartDate time.Time `json:"start_date" validate:"required"`
	Days      int       `json:"days" validate:"required"`
}

type BookingResponse struct {
	Message    string  `json:"message"`
	RoomName   string  `json:"room_name"`
	Category   string  `json:"category"`
	TotalPrice float64 `json:"total_price"`
}

type TopUpRequest struct {
	Amount float64 `json:"amount" validate:"required"`
}
