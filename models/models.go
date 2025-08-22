package models

import (
	"gorm.io/gorm"
)

// type Note struct {
// 	gorm.Model
// 	Title   string `json:"title"`
// 	Content string `json:"content"`
// }

type User struct {
	gorm.Model
	Username string   `gorm:"unique" json:"username"`
	Password string   `json:"password"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Tenants  []Tenant `gorm:"many2many:user_tenants;"`
}

type Task struct {
	gorm.Model
	Name   string `json:"name"`
	RoleID uint   `json:"role_id"`
	Status string `json:"status"`
}

type Role struct {
	gorm.Model
	Name  string `json:"name"`
	Tasks []Task `gorm:"many2many:role_tasks;"`
}

type Tenant struct {
	gorm.Model
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Location    string     `json:"location"`
	Logo        string     `json:"logo"`
	Status      string     `json:"status"`
	OwnerID     uint       `json:"owner_id"`
	MenuItems   []MenuItem `gorm:"foreignkey:TenantID"` // Link to menu items
}

type UserTenantRole struct {
	gorm.Model
	UserID   uint `json:"user_id"`
	TenantID uint `json:"tenant_id"`
	RoleID   uint `json:"role_id"`
	Active   bool `json:"active"`
}

type MenuItem struct {
	gorm.Model
	Name        string  `json:"name"`        // Name of the food item (e.g., "Pizza Margherita")
	Description string  `json:"description"` // Description of the food item (e.g., "Classic Italian pizza with mozzarella and basil")
	Price       float64 `json:"price"`       // Price of the food item
	Category    string  `json:"category"`    // Category of the item (e.g., "Main Course", "Appetizer")
	TenantID    uint    `json:"tenant_id"`   // The tenant that sells this item
}
