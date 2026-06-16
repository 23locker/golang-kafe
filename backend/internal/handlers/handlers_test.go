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

func (m *MockAuthService) UpdateEmail(ctx context.Context, userID int, email string) error {
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

type MockBlogService struct {
	GetPostsFunc    func(ctx context.Context) ([]dto.BlogPostResponse, error)
	GetPostByIDFunc func(ctx context.Context, id int) (dto.BlogPostResponse, error)
}

func (m *MockBlogService) GetPosts(ctx context.Context) ([]dto.BlogPostResponse, error) {
	if m.GetPostsFunc != nil {
		return m.GetPostsFunc(ctx)
	}
	return nil, nil
}

func (m *MockBlogService) GetPostByID(ctx context.Context, id int) (dto.BlogPostResponse, error) {
	if m.GetPostByIDFunc != nil {
		return m.GetPostByIDFunc(ctx, id)
	}
	return dto.BlogPostResponse{}, errors.New("not found")
}

type MockAdminService struct {
	GetOrdersFunc               func(ctx context.Context) ([]dto.OrderResponse, error)
	GetOrdersFilteredFunc       func(ctx context.Context, phone string, orderID *int) ([]dto.OrderResponse, error)
	UpdateOrderStatusFunc       func(ctx context.Context, orderID int, status string) error
	DeleteOrderFunc             func(ctx context.Context, id int) error
	GetReservationsFunc         func(ctx context.Context) ([]dto.ReservationResponse, error)
	UpdateReservationStatusFunc func(ctx context.Context, id int, status string) error
	DeleteReservationFunc       func(ctx context.Context, id int) error
	GetAdminProductsFunc        func(ctx context.Context, categoryID *int) ([]dto.ProductResponse, error)
	GetAdminProductByIDFunc     func(ctx context.Context, id int) (dto.ProductResponse, error)
	CreateProductFunc           func(ctx context.Context, req dto.ProductResponse) (dto.ProductResponse, error)
	UpdateProductFunc           func(ctx context.Context, id int, req dto.ProductResponse) (dto.ProductResponse, error)
	DeleteProductFunc           func(ctx context.Context, id int) (bool, error)
	CreateCategoryFunc          func(ctx context.Context, req dto.CategoryResponse) (dto.CategoryResponse, error)
	UpdateCategoryFunc          func(ctx context.Context, id int, req dto.CategoryResponse) (dto.CategoryResponse, error)
	DeleteCategoryFunc          func(ctx context.Context, id int) error
	GetStatsFunc                func(ctx context.Context, startDate, endDate string) (map[string]interface{}, error)
	GetAllBlogPostsFunc         func(ctx context.Context) ([]dto.BlogPostResponse, error)
	CreateBlogPostFunc          func(ctx context.Context, req dto.BlogPostResponse) (dto.BlogPostResponse, error)
	UpdateBlogPostFunc          func(ctx context.Context, id int, req dto.BlogPostResponse) (dto.BlogPostResponse, error)
	DeleteBlogPostFunc          func(ctx context.Context, id int) error
	GetUsersFunc                func(ctx context.Context) ([]dto.AdminUserResponse, error)
	UpdateUserRoleFunc          func(ctx context.Context, targetUserID, currentAdminID int, newRole string) error
	DeleteUserFunc              func(ctx context.Context, targetUserID, currentAdminID int) error
	GetAuditLogFunc             func(ctx context.Context) ([]dto.AuditLogResponse, error)
}

func (m *MockAdminService) GetOrders(ctx context.Context) ([]dto.OrderResponse, error) {
	if m.GetOrdersFunc != nil {
		return m.GetOrdersFunc(ctx)
	}
	return nil, nil
}

func (m *MockAdminService) GetOrdersFiltered(ctx context.Context, phone string, orderID *int) ([]dto.OrderResponse, error) {
	if m.GetOrdersFilteredFunc != nil {
		return m.GetOrdersFilteredFunc(ctx, phone, orderID)
	}
	return nil, nil
}

func (m *MockAdminService) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	if m.UpdateOrderStatusFunc != nil {
		return m.UpdateOrderStatusFunc(ctx, orderID, status)
	}
	return nil
}

func (m *MockAdminService) DeleteOrder(ctx context.Context, id int) error {
	if m.DeleteOrderFunc != nil {
		return m.DeleteOrderFunc(ctx, id)
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

func (m *MockAdminService) DeleteReservation(ctx context.Context, id int) error {
	if m.DeleteReservationFunc != nil {
		return m.DeleteReservationFunc(ctx, id)
	}
	return nil
}

func (m *MockAdminService) GetAdminProducts(ctx context.Context, categoryID *int) ([]dto.ProductResponse, error) {
	if m.GetAdminProductsFunc != nil {
		return m.GetAdminProductsFunc(ctx, categoryID)
	}
	return nil, nil
}

func (m *MockAdminService) GetAdminProductByID(ctx context.Context, id int) (dto.ProductResponse, error) {
	if m.GetAdminProductByIDFunc != nil {
		return m.GetAdminProductByIDFunc(ctx, id)
	}
	return dto.ProductResponse{}, nil
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

func (m *MockAdminService) DeleteProduct(ctx context.Context, id int) (bool, error) {
	if m.DeleteProductFunc != nil {
		return m.DeleteProductFunc(ctx, id)
	}
	return false, nil
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

func (m *MockAdminService) GetAllBlogPosts(ctx context.Context) ([]dto.BlogPostResponse, error) {
	if m.GetAllBlogPostsFunc != nil {
		return m.GetAllBlogPostsFunc(ctx)
	}
	return nil, nil
}

func (m *MockAdminService) CreateBlogPost(ctx context.Context, req dto.BlogPostResponse) (dto.BlogPostResponse, error) {
	if m.CreateBlogPostFunc != nil {
		return m.CreateBlogPostFunc(ctx, req)
	}
	return dto.BlogPostResponse{}, nil
}

func (m *MockAdminService) UpdateBlogPost(ctx context.Context, id int, req dto.BlogPostResponse) (dto.BlogPostResponse, error) {
	if m.UpdateBlogPostFunc != nil {
		return m.UpdateBlogPostFunc(ctx, id, req)
	}
	return dto.BlogPostResponse{}, nil
}

func (m *MockAdminService) DeleteBlogPost(ctx context.Context, id int) error {
	if m.DeleteBlogPostFunc != nil {
		return m.DeleteBlogPostFunc(ctx, id)
	}
	return nil
}

func (m *MockAdminService) GetUsers(ctx context.Context) ([]dto.AdminUserResponse, error) {
	if m.GetUsersFunc != nil {
		return m.GetUsersFunc(ctx)
	}
	return nil, nil
}

func (m *MockAdminService) UpdateUserRole(ctx context.Context, targetUserID, currentAdminID int, newRole string) error {
	if m.UpdateUserRoleFunc != nil {
		return m.UpdateUserRoleFunc(ctx, targetUserID, currentAdminID, newRole)
	}
	return nil
}

func (m *MockAdminService) DeleteUser(ctx context.Context, targetUserID, currentAdminID int) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, targetUserID, currentAdminID)
	}
	return nil
}

func (m *MockAdminService) GetAuditLog(ctx context.Context) ([]dto.AuditLogResponse, error) {
	if m.GetAuditLogFunc != nil {
		return m.GetAuditLogFunc(ctx)
	}
	return nil, nil
}

func (m *MockAdminService) LogAudit(ctx context.Context, adminID *int, action, entityType string, entityID *int, details string) error {
	return nil
}

// --- HANDLER TESTS ---

func TestHandler_Register(t *testing.T) {
	authService := &MockAuthService{}
	handler := NewHandler(authService, &MockProductService{}, &MockOrderService{}, &MockReservationService{}, &MockAdminService{}, &MockBlogService{}, "secret")
	router := handler.InitRoutes()

	authService.RegisterFunc = func(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
		if req.Phone == "+79998887766" {
			return &dto.UserResponse{ID: 1, Name: req.Name, Phone: req.Phone, Role: "user"}, nil
		}
		return nil, errors.New("registration failed")
	}

	regReq := dto.RegisterRequest{Name: "Alice", Phone: "+79998887766", Password: "password123"}
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
	handler := NewHandler(&MockAuthService{}, prodService, &MockOrderService{}, &MockReservationService{}, &MockAdminService{}, &MockBlogService{}, "secret")
	router := handler.InitRoutes()

	prodService.GetProductsFunc = func(ctx context.Context, categoryID *int) ([]dto.ProductResponse, error) {
		return []dto.ProductResponse{
			{ID: 1, Name: "Classic Buuzy", Price: 90.00, Description: "Juicy beef & pork buuzy", Weight: 75, Calories: 180, IsAvailable: true},
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
		t.Errorf("Products response mismatch. Got items count: %d", len(resp))
	}
}
