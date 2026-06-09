package dto

import "time"

type RegisterRequest struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Phone          string  `json:"phone"`
	DefaultAddress *string `json:"default_address"`
	Role           string  `json:"role"`
}

type CategoryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProductResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	CategoryID  *int    `json:"category_id"`
	ImageURL    string  `json:"image_url"`
	Weight      int     `json:"weight"`
	Calories    int     `json:"calories"`
	IsAvailable bool    `json:"is_available"`
}

type OrderItemRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type CreateOrderRequest struct {
	CustomerName  string             `json:"customer_name"`
	Phone         string             `json:"phone"`
	Address       string             `json:"address"`
	PaymentMethod string             `json:"payment_method"`
	Items         []OrderItemRequest `json:"items"`
}

type OrderItemResponse struct {
	ID          int     `json:"id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type OrderResponse struct {
	ID            int                 `json:"id"`
	UserID        *int                `json:"user_id"`
	CustomerName  string              `json:"customer_name"`
	Phone         string              `json:"phone"`
	Address       string              `json:"address"`
	TotalPrice    float64             `json:"total_price"`
	PaymentStatus string              `json:"payment_status"`
	PaymentID     *string             `json:"payment_id"`
	CreatedAt     time.Time           `json:"created_at"`
	Items         []OrderItemResponse `json:"items"`
}

type CreateReservationRequest struct {
	CustomerName string `json:"customer_name"`
	Phone        string `json:"phone"`
	ReserveDate  string `json:"reserve_date"`
	ReserveTime  string `json:"reserve_time"`
	GuestsCount  int    `json:"guests_count"`
	Comment      string `json:"comment"`
}

type ReservationResponse struct {
	ID           int       `json:"id"`
	UserID       *int      `json:"user_id"`
	CustomerName string    `json:"customer_name"`
	Phone        string    `json:"phone"`
	ReserveDate  time.Time `json:"reserve_date"`
	ReserveTime  string    `json:"reserve_time"`
	GuestsCount  int       `json:"guests_count"`
	Comment      string    `json:"comment"`
	Status       string    `json:"status"`
}
