package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/dto"
)

// --- SERVICE MOCKS ---

type MockAuthService struct {
	RegisterFunc      func(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error)
	LoginFunc         func(ctx context.Context, req dto.LoginRequest) (string, error)
	GetUserByIDFunc   func(ctx context.Context, id int) (*dto.UserResponse, error)
	UpdateAddressFunc func(ctx context.Context, userID int, address string) error
}

func (m *MockAuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockAuthService) Login(ctx context.Context, req dto.LoginRequest) (string, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, req)
	}
	return "", nil
}

func (m *MockAuthService) GetUserByID(ctx context.Context, id int) (*dto.UserResponse, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockAuthService) UpdateAddress(ctx context.Context, userID int, address string) error {
	if m.UpdateAddressFunc != nil {
		return m.UpdateAddressFunc(ctx, userID, address)
	}
	return nil
}

type MockProductService struct {
	GetCategoriesFunc  func(ctx context.Context) ([]dto.CategoryResponse, error)
	GetProductsFunc    func(ctx context.Context, categoryID *int) ([]dto.ProductResponse, error)
	GetProductByIDFunc func(ctx context.Context, id int) (dto.ProductResponse, error)
}

func (m *MockProductService) GetCategories(ctx context.Context) ([]dto.CategoryResponse, error) {
	if m.GetCategoriesFunc != nil {
		return m.GetCategoriesFunc(ctx)
	}
	return nil, nil
}

func (m *MockProductService) GetProducts(ctx context.Context, categoryID *int) ([]dto.ProductResponse, error) {
	if m.GetProductsFunc != nil {
		return m.GetProductsFunc(ctx, categoryID)
	}
	return nil, nil
}

func (m *MockProductService) GetProductByID(ctx context.Context, id int) (dto.ProductResponse, error) {
	if m.GetProductByIDFunc != nil {
		return m.GetProductByIDFunc(ctx, id)
	}
	return dto.ProductResponse{}, nil
}

type MockOrderService struct {
	CreateOrderFunc       func(ctx context.Context, userID *int, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetOrdersByUserIDFunc func(ctx context.Context, userID int) ([]dto.OrderResponse, error)
}

func (m *MockOrderService) CreateOrder(ctx context.Context, userID *int, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	if m.CreateOrderFunc != nil {
		return m.CreateOrderFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockOrderService) GetOrdersByUserID(ctx context.Context, userID int) ([]dto.OrderResponse, error) {
	if m.GetOrdersByUserIDFunc != nil {
		return m.GetOrdersByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

type MockReservationService struct {
	CreateReservationFunc       func(ctx context.Context, userID *int, req dto.CreateReservationRequest) (*dto.ReservationResponse, error)
	GetReservationsByUserIDFunc func(ctx context.Context, userID int) ([]dto.ReservationResponse, error)
}

func (m *MockReservationService) CreateReservation(ctx context.Context, userID *int, req dto.CreateReservationRequest) (*dto.ReservationResponse, error) {
	if m.CreateReservationFunc != nil {
		return m.CreateReservationFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockReservationService) GetReservationsByUserID(ctx context.Context, userID int) ([]dto.ReservationResponse, error) {
	if m.GetReservationsByUserIDFunc != nil {
		return m.GetReservationsByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

type MockAdminService struct {
	GetOrdersFunc               func(ctx context.Context) ([]dto.OrderResponse, error)
	UpdateOrderStatusFunc       func(ctx context.Context, orderID int, status string) error
	GetReservationsFunc         func(ctx context.Context) ([]dto.ReservationResponse, error)
	UpdateReservationStatusFunc func(ctx context.Context, id int, status string) error
	CreateProductFunc           func(ctx context.Context, req dto.ProductResponse) (dto.ProductResponse, error)
	UpdateProductFunc           func(ctx context.Context, id int, req dto.ProductResponse) (dto.ProductResponse, error)
	DeleteProductFunc           func(ctx context.Context, id int) error
	CreateCategoryFunc          func(ctx context.Context, req dto.CategoryResponse) (dto.CategoryResponse, error)
	UpdateCategoryFunc          func(ctx context.Context, id int, req dto.CategoryResponse) (dto.CategoryResponse, error)
	DeleteCategoryFunc          func(ctx context.Context, id int) error
	GetStatsFunc                func(ctx context.Context, startDate, endDate string) (map[string]interface{}, error)
}

func (m *MockAdminService) GetOrders(ctx context.Context) ([]dto.OrderResponse, error) {
	if m.GetOrdersFunc != nil {
		return m.GetOrdersFunc(ctx)
	}
	return nil, nil
}

func (m *MockAdminService) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	if m.UpdateOrderStatusFunc != nil {
		return m.UpdateOrderStatusFunc(ctx, orderID, status)
	}
	return nil
}

func (m *MockAdminService) GetReservations(ctx context.Context) ([]dto.ReservationResponse, error) {
	if m.GetReservationsFunc != nil {
		return m.GetReservationsFunc(ctx)
	}
	return nil, nil
}

func (m *MockAdminService) UpdateReservationStatus(ctx context.Context, id int, status string) error {
	if m.UpdateReservationStatusFunc != nil {
		return m.UpdateReservationStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *MockAdminService) CreateProduct(ctx context.Context, req dto.ProductResponse) (dto.ProductResponse, error) {
	if m.CreateProductFunc != nil {
		return m.CreateProductFunc(ctx, req)
	}
	return dto.ProductResponse{}, nil
}

func (m *MockAdminService) UpdateProduct(ctx context.Context, id int, req dto.ProductResponse) (dto.ProductResponse, error) {
	if m.UpdateProductFunc != nil {
		return m.UpdateProductFunc(ctx, id, req)
	}
	return dto.ProductResponse{}, nil
}

func (m *MockAdminService) DeleteProduct(ctx context.Context, id int) error {
	if m.DeleteProductFunc != nil {
		return m.DeleteProductFunc(ctx, id)
	}
	return nil
}

func (m *MockAdminService) CreateCategory(ctx context.Context, req dto.CategoryResponse) (dto.CategoryResponse, error) {
	if m.CreateCategoryFunc != nil {
		return m.CreateCategoryFunc(ctx, req)
	}
	return dto.CategoryResponse{}, nil
}

func (m *MockAdminService) UpdateCategory(ctx context.Context, id int, req dto.CategoryResponse) (dto.CategoryResponse, error) {
	if m.UpdateCategoryFunc != nil {
		return m.UpdateCategoryFunc(ctx, id, req)
	}
	return dto.CategoryResponse{}, nil
}

func (m *MockAdminService) DeleteCategory(ctx context.Context, id int) error {
	if m.DeleteCategoryFunc != nil {
		return m.DeleteCategoryFunc(ctx, id)
	}
	return nil
}

func (m *MockAdminService) GetStats(ctx context.Context, startDate, endDate string) (map[string]interface{}, error) {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc(ctx, startDate, endDate)
	}
	return nil, nil
}

// --- HANDLER TESTS ---

func TestHandler_Register(t *testing.T) {
	authService := &MockAuthService{}
	handler := NewHandler(authService, &MockProductService{}, &MockOrderService{}, &MockReservationService{}, &MockAdminService{}, "secret")
	router := handler.InitRoutes()

	// Mock register success
	authService.RegisterFunc = func(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
		if req.Phone == "+79998887766" {
			return &dto.UserResponse{
				ID:    1,
				Name:  req.Name,
				Phone: req.Phone,
				Role:  "user",
			}, nil
		}
		return nil, errors.New("registration failed")
	}

	// 1. Success case
	regReq := dto.RegisterRequest{
		Name:     "Alice",
		Phone:    "+79998887766",
		Password: "password123",
	}
	body, _ := json.Marshal(regReq)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code 201 Created, got: %d", w.Code)
	}

	var resp dto.UserResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if resp.Name != "Alice" || resp.Phone != "+79998887766" {
		t.Errorf("Register response mismatch. Got name: %s, phone: %s", resp.Name, resp.Phone)
	}

	// 2. Failure case (invalid request payload)
	reqFail := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString("{invalid-json}"))
	reqFail.Header.Set("Content-Type", "application/json")
	wFail := httptest.NewRecorder()

	router.ServeHTTP(wFail, reqFail)

	if wFail.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400 Bad Request, got: %d", wFail.Code)
	}
}

func TestHandler_GetProducts(t *testing.T) {
	prodService := &MockProductService{}
	handler := NewHandler(&MockAuthService{}, prodService, &MockOrderService{}, &MockReservationService{}, &MockAdminService{}, "secret")
	router := handler.InitRoutes()

	prodService.GetProductsFunc = func(ctx context.Context, categoryID *int) ([]dto.ProductResponse, error) {
		return []dto.ProductResponse{
			{
				ID:          1,
				Name:        "Classic Buuzy",
				Price:       90.00,
				Description: "Juicy beef & pork buuzy",
				Weight:      75,
				Calories:    180,
				IsAvailable: true,
			},
		}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200 OK, got: %d", w.Code)
	}

	var resp []dto.ProductResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if len(resp) != 1 || resp[0].Name != "Classic Buuzy" {
		t.Errorf("Products response mismatch. Got items count: %d, first item: %+v", len(resp), resp)
	}
}
