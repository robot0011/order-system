package constants

// Order statuses for internal use
const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusPreparing = "preparing"
	OrderStatusReady     = "ready"
	OrderStatusDelivered = "delivered"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
)

// Simplified frontend order statuses
const (
	FrontendOrderStatusActive   = "active"
	FrontendOrderStatusDelivered = "delivered"
	FrontendOrderStatusPaid     = "paid"
)