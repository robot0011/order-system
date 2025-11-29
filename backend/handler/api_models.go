package handler

// swagger:model RegisterRequest
type RegisterRequest struct {
	// required: true
	Username string `json:"username" example:"john_doe"`
	// required: true
	Password string `json:"password" example:"password123"`
	// required: true
	Email    string `json:"email" example:"john@example.com"`
	Role     string `json:"role" example:"owner"`
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

// swagger:model User
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// swagger:model Restaurant
type Restaurant struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
	LogoURL     string `json:"logo_url"`
}

// swagger:model Table
type Table struct {
	ID           uint   `json:"id"`
	RestaurantID uint   `json:"restaurant_id"`
	TableNumber  int    `json:"table_number"`
	QRCodeURL    string `json:"qr_code_url"`
}

// swagger:model MenuItem
type MenuItem struct {
	ID           uint    `json:"id"`
	RestaurantID uint    `json:"restaurant_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Category     string  `json:"category"`
	ImageURL     string  `json:"image_url"`
	Quantity     int     `json:"quantity"`
}

// swagger:model Order
type Order struct {
	ID           uint        `json:"id"`
	TableID      uint        `json:"table_id"`
	CustomerName string      `json:"customer_name"`
	Status       string      `json:"status"`
	TotalAmount  float64     `json:"total_amount"`
	OrderItems   []OrderItem `json:"order_items"`
}

// swagger:model OrderItem
type OrderItem struct {
	ID                  uint   `json:"id"`
	OrderID             uint   `json:"order_id"`
	MenuItemID          uint   `json:"menu_item_id"`
	Quantity            int    `json:"quantity"`
	SpecialInstructions string `json:"special_instructions"`
}

// swagger:model OrderStatusUpdate
type OrderStatusUpdate struct {
	Status string `json:"status" example:"completed"`
}

// swagger:model OrderResponse
type OrderResponse struct {
	Order
	RestaurantName string `json:"restaurant_name"`
	RestaurantID   uint   `json:"restaurant_id"`
}