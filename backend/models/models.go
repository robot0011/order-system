package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username    string       `gorm:"size:255;not null;unique"`
	Password    string       `gorm:"size:255;not null"`
	Role        string       `gorm:"size:50;default:'owner'"` // You can add more roles (e.g., 'admin', 'staff')
	Restaurants []Restaurant `gorm:"foreignKey:UserID"`
	Email       string       `gorm:"size:255;not null;unique"`
}

type Restaurant struct {
	gorm.Model
	UserID      uint       `gorm:"not null"` // Link restaurant to a user (owner)
	Name        string     `gorm:"size:255;not null"`
	Address     string     `gorm:"size:255"`
	PhoneNumber string     `gorm:"size:50"`
	LogoURL     string     `gorm:"size:255"`
	Tables      []Table    `gorm:"foreignKey:RestaurantID"`
	MenuItems   []MenuItem `gorm:"foreignKey:RestaurantID"`
}

type Table struct {
	gorm.Model
	RestaurantID uint    `gorm:"not null"`
	TableNumber  int     `gorm:"not null"`
	QRCodeURL    string  `gorm:"size:255"`
	Orders       []Order `gorm:"foreignKey:TableID"`
}

type MenuItem struct {
	gorm.Model
	RestaurantID uint        `gorm:"not null"`
	Name         string      `gorm:"size:255;not null"`
	Description  string      `gorm:"type:text"`
	Price        float64     `gorm:"not null"`
	Category     string      `gorm:"size:50"` // starter, main, dessert, drink
	ImageURL     string      `gorm:"size:255"`
	Quantity     int         `gorm:"default:0"` // available quantity of the menu item
	OrderItems   []OrderItem `gorm:"foreignKey:MenuItemID"`
}

type Order struct {
	gorm.Model
	TableID      uint        `gorm:"not null"`
	CustomerName string      `gorm:"size:255"`                  // Name of the customer who placed the order
	Status       string      `gorm:"size:50;default:'pending'"` // pending, preparing, served, completed, cancelled
	TotalAmount  float64     `gorm:"not null"`
	CreatedAt    time.Time   `gorm:"autoCreateTime"`
	UpdatedAt    time.Time   `gorm:"autoUpdateTime"`
	OrderItems   []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	Payments     []Payment   `gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
	gorm.Model
	OrderID             uint      `gorm:"not null"`
	MenuItemID          uint      `gorm:"not null"`
	Quantity            int       `gorm:"default:1"`
	SpecialInstructions string    `gorm:"type:text"`
	MenuItem            MenuItem  `gorm:"foreignKey:MenuItemID;references:ID"`
}

type Payment struct {
	gorm.Model
	OrderID       uint      `gorm:"not null"`
	PaymentMethod string    `gorm:"size:50"`                   // credit_card, mobile_wallet, paypal, cash
	PaymentStatus string    `gorm:"size:50;default:'pending'"` // pending, completed, failed
	Amount        float64   `gorm:"not null"`
	PaymentDate   time.Time `gorm:"autoCreateTime"`
}

// swagger:model LoginRequest
type LoginRequest struct {
	// required: true
	Username string `json:"username" example:"john_doe"`
	// required: true
	Password string `json:"password" example:"password123"`
}

// swagger:model TokenResponse
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh..."`
}

// swagger:model RefreshRequest
type RefreshRequest struct {
	// required: true
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh..."`
}
