package entity

import "time"

type User struct {
	ID       int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string    `json:"name" gorm:"not null;size:100"`
	Email    string    `json:"email" gorm:"unique;not null;size:100"`
	Password string    `json:"-" gorm:"not null;size:100"`
	Balance  float64   `json:"balance" gorm:"not null;default:0"`
	Bookings []Booking `json:"bookings" gorm:"foreignKey:UserID"`
}

type Category struct {
	ID          int     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string  `json:"name" gorm:"not null;unique;size:100"` // Standard, Deluxe, etc.
	Description string  `json:"description" gorm:"size:255"`
	Price       float64 `json:"price" gorm:"not null"` // Price per night
	Rooms       []Room  `json:"rooms" gorm:"foreignKey:CategoryID"`
}

type Room struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name       string    `json:"name" gorm:"not null;size:255"` // Room name or identifier
	CategoryID int       `json:"category_id" gorm:"not null"`   // FK to Category
	Category   Category  `json:"category" gorm:"foreignKey:CategoryID"`
	Stock      int       `json:"stock" gorm:"not null"`             // Available rooms
	Bookings   []Booking `json:"bookings" gorm:"foreignKey:RoomID"` // Related bookings
}

type Booking struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     int       `json:"user_id" gorm:"not null"` // FK to User
	User       User      `json:"user" gorm:"foreignKey:UserID"`
	RoomID     int       `json:"room_id" gorm:"not null"` // FK to Room
	Room       Room      `json:"room" gorm:"foreignKey:RoomID"`
	StartDate  time.Time `json:"start_date" gorm:"not null"`       // Start of booking
	EndDate    time.Time `json:"end_date" gorm:"not null"`         // End of booking
	TotalPrice float64   `json:"total_price" gorm:"not null"`      // Total cost
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"` // Booking timestamp
}

type Payment struct {
	ID        int        `json:"id" gorm:"primaryKey;autoIncrement"`
	BookingID int        `json:"booking_id" gorm:"not null;unique"` // FK to Booking
	Booking   Booking    `json:"-" gorm:"foreignKey:BookingID"`     // Avoid recursion by omitting JSON
	Amount    float64    `json:"amount" gorm:"not null"`            // Payment amount
	IsPaid    bool       `json:"is_paid" gorm:"not null;default:false"`
	PaidAt    *time.Time `json:"paid_at"` // Time of payment (nullable if unpaid)
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

type UserActivityLog struct {
	ID           int       `json:"id" gorm:"not null;primaryKey"`
	UserID       int       `json:"user_id" gorm:"column:user_id;not null"`      // FK to User
	User         User      `json:"user" gorm:"foreignKey:UserID;references:ID"` // Correct FK
	Description  string    `json:"description" gorm:"column:description;not null;size:255"`
	ActivityType string    `json:"activity_type" gorm:"column:activity_type;size:50"` // Type of activity
	ReferenceID  int       `json:"reference_id" gorm:"column:reference_id"`           // Optional ID for related actions
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}
