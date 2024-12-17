package entity

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
