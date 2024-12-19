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

type UserProfileResponse struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Balance float64 `json:"balance"`
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

type GetUserRoomsResponse struct {
	BookingID  int       `json:"booking_id"`
	RoomID     int       `json:"room_id"`
	RoomName   string    `json:"room_name"`
	Category   string    `json:"category"`
	Price      float64   `json:"price"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	TotalPrice float64   `json:"total_price"`
	IsPaid     bool      `json:"is_paid"`
}

type TopUpRequest struct {
	Amount float64 `json:"amount" validate:"required"`
}

type UpdateBookingRequest struct {
	NewRoomID int       `json:"new_room_id,omitempty"`
	NewDays   int       `json:"new_days,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
}

type ProductRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type CustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
