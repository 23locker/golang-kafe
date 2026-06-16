package mappers

import (
	"backend/internal/dto"
	"backend/internal/models"
)

func ToUserResponse(u *models.User) *dto.UserResponse {
	if u == nil {
		return nil
	}
	return &dto.UserResponse{
		ID:             u.ID,
		Name:           u.Name,
		Phone:          u.Phone,
		Email:          u.Email,
		DefaultAddress: u.DefaultAddress,
		Role:           u.Role,
	}
}

func ToCategoryResponse(c *models.Category) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:   c.ID,
		Name: c.Name,
		Slug: c.Slug,
	}
}

func ToCategoryResponseList(categories []models.Category) []dto.CategoryResponse {
	res := make([]dto.CategoryResponse, len(categories))
	for i, c := range categories {
		res[i] = ToCategoryResponse(&c)
	}
	return res
}

func ToProductResponse(p *models.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		CategoryID:  p.CategoryID,
		ImageURL:    p.ImageURL,
		Weight:      p.Weight,
		Calories:    p.Calories,
		IsAvailable: p.IsAvailable,
	}
}

func ToProductResponseList(products []models.Product) []dto.ProductResponse {
	res := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		res[i] = ToProductResponse(&p)
	}
	return res
}

func ToOrderResponse(o *models.Order, items []dto.OrderItemResponse) dto.OrderResponse {
	return dto.OrderResponse{
		ID:            o.ID,
		UserID:        o.UserID,
		CustomerName:  o.CustomerName,
		Phone:         o.Phone,
		Address:       o.Address,
		TotalPrice:    o.TotalPrice,
		PaymentStatus: o.PaymentStatus,
		PaymentID:     o.PaymentID,
		CreatedAt:     o.CreatedAt,
		Items:         items,
	}
}

func ToReservationResponse(r *models.Reservation) dto.ReservationResponse {
	return dto.ReservationResponse{
		ID:           r.ID,
		UserID:       r.UserID,
		CustomerName: r.CustomerName,
		Phone:        r.Phone,
		ReserveDate:  r.ReserveDate,
		ReserveTime:  r.ReserveTime,
		GuestsCount:  r.GuestsCount,
		Comment:      r.Comment,
		Status:       r.Status,
	}
}

func ToReservationResponseList(reservations []models.Reservation) []dto.ReservationResponse {
	res := make([]dto.ReservationResponse, len(reservations))
	for i, r := range reservations {
		res[i] = ToReservationResponse(&r)
	}
	return res
}
