package models

import "time"

type User struct {
	ID             int
	Name           string
	Phone          string
	Email          *string
	PasswordHash   string
	DefaultAddress *string
	Role           string
	CreatedAt      time.Time
}

type Category struct {
	ID   int
	Name string
	Slug string
}

type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
	CategoryID  *int
	ImageURL    string
	Weight      int
	Calories    int
	IsAvailable bool
}

type Order struct {
	ID            int
	UserID        *int
	CustomerName  string
	Phone         string
	Address       string
	TotalPrice    float64
	PaymentStatus string
	PaymentID     *string
	CreatedAt     time.Time
}

type OrderItem struct {
	ID        int
	OrderID   int
	ProductID int
	Quantity  int
	Price     float64
}

type Reservation struct {
	ID           int
	UserID       *int
	CustomerName string
	Phone        string
	ReserveDate  time.Time
	ReserveTime  string
	GuestsCount  int
	Comment      string
	Status       string
}
