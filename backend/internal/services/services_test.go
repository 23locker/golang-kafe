package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"backend/internal/dto"
	"backend/internal/models"
)

// --- MOCK REPOSITORIES ---

type MockUserRepository struct {
	Users     map[string]*models.User
	UsersByID map[int]*models.User
	NextID    int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users:     make(map[string]*models.User),
		UsersByID: make(map[int]*models.User),
		NextID:    1,
	}
}

func (m *MockUserRepository) Create(ctx context.Context, u *models.User) error {
	if _, exists := m.Users[u.Phone]; exists {
		return errors.New("user already exists")
	}
	u.ID = m.NextID
	m.NextID++
	m.Users[u.Phone] = u
	m.UsersByID[u.ID] = u
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	u, exists := m.UsersByID[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (m *MockUserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	u, exists := m.Users[phone]
	if !exists {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (m *MockUserRepository) UpdateEmail(ctx context.Context, userID int, email string) error {
	return nil
}

func (m *MockUserRepository) UpdateAddress(ctx context.Context, userID int, address string) error {
	u, exists := m.UsersByID[userID]
	if !exists {
		return errors.New("user not found")
	}
	u.DefaultAddress = &address
	return nil
}

func (m *MockUserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	var list []models.User
	for _, u := range m.UsersByID {
		list = append(list, *u)
	}
	return list, nil
}

func (m *MockUserRepository) UpdateRole(ctx context.Context, userID int, role string) error {
	u, exists := m.UsersByID[userID]
	if !exists {
		return errors.New("user not found")
	}
	u.Role = role
	return nil
}

func (m *MockUserRepository) CountByRole(ctx context.Context, role string) (int, error) {
	count := 0
	for _, u := range m.UsersByID {
		if u.Role == role {
			count++
		}
	}
	return count, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	u, exists := m.UsersByID[id]
	if !exists {
		return errors.New("user not found")
	}
	delete(m.Users, u.Phone)
	delete(m.UsersByID, id)
	return nil
}

type MockProductRepository struct {
	Products   map[int]*models.Product
	Categories map[int]*models.Category
	NextProdID int
	NextCatID  int
}

func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{
		Products:   make(map[int]*models.Product),
		Categories: make(map[int]*models.Category),
		NextProdID: 1,
		NextCatID:  1,
	}
}

func (m *MockProductRepository) GetCategories(ctx context.Context) ([]models.Category, error) {
	var list []models.Category
	for _, c := range m.Categories {
		list = append(list, *c)
	}
	return list, nil
}

func (m *MockProductRepository) GetProducts(ctx context.Context, categoryID *int, includeDeleted bool) ([]models.Product, error) {
	var list []models.Product
	for _, p := range m.Products {
		if !includeDeleted && p.IsDeleted {
			continue
		}
		if categoryID == nil || (p.CategoryID != nil && *p.CategoryID == *categoryID) {
			list = append(list, *p)
		}
	}
	return list, nil
}

func (m *MockProductRepository) GetProductByID(ctx context.Context, id int) (*models.Product, error) {
	p, exists := m.Products[id]
	if !exists {
		return nil, errors.New("product not found")
	}
	return p, nil
}

func (m *MockProductRepository) Create(ctx context.Context, p *models.Product) error {
	p.ID = m.NextProdID
	m.NextProdID++
	m.Products[p.ID] = p
	return nil
}

func (m *MockProductRepository) Update(ctx context.Context, p *models.Product) error {
	if _, exists := m.Products[p.ID]; !exists {
		return errors.New("product not found")
	}
	m.Products[p.ID] = p
	return nil
}

func (m *MockProductRepository) Delete(ctx context.Context, id int) error {
	if _, exists := m.Products[id]; !exists {
		return errors.New("product not found")
	}
	delete(m.Products, id)
	return nil
}

func (m *MockProductRepository) SoftDelete(ctx context.Context, id int) error {
	p, exists := m.Products[id]
	if !exists {
		return errors.New("product not found")
	}
	p.IsDeleted = true
	p.IsAvailable = false
	return nil
}

func (m *MockProductRepository) HasOrderHistory(ctx context.Context, id int) (bool, error) {
	return false, nil
}

func (m *MockProductRepository) HasProductsInCategory(ctx context.Context, categoryID int) (bool, error) {
	for _, p := range m.Products {
		if !p.IsDeleted && p.CategoryID != nil && *p.CategoryID == categoryID {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockProductRepository) CreateCategory(ctx context.Context, c *models.Category) error {
	c.ID = m.NextCatID
	m.NextCatID++
	m.Categories[c.ID] = c
	return nil
}

func (m *MockProductRepository) UpdateCategory(ctx context.Context, c *models.Category) error {
	if _, exists := m.Categories[c.ID]; !exists {
		return errors.New("category not found")
	}
	m.Categories[c.ID] = c
	return nil
}

func (m *MockProductRepository) DeleteCategory(ctx context.Context, id int) error {
	if _, exists := m.Categories[id]; !exists {
		return errors.New("category not found")
	}
	delete(m.Categories, id)
	return nil
}

type MockOrderRepository struct {
	Orders     map[int]*models.Order
	OrderItems map[int][]models.OrderItem
	NextID     int
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		Orders:     make(map[int]*models.Order),
		OrderItems: make(map[int][]models.OrderItem),
		NextID:     1,
	}
}

func (m *MockOrderRepository) Create(ctx context.Context, o *models.Order, items []models.OrderItem) error {
	o.ID = m.NextID
	m.NextID++
	o.CreatedAt = time.Now()
	m.Orders[o.ID] = o
	m.OrderItems[o.ID] = items
	return nil
}

func (m *MockOrderRepository) GetByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	var list []models.Order
	for _, o := range m.Orders {
		if o.UserID != nil && *o.UserID == userID {
			list = append(list, *o)
		}
	}
	return list, nil
}

func (m *MockOrderRepository) GetOrderItems(ctx context.Context, orderID int) ([]dto.OrderItemResponse, error) {
	var list []dto.OrderItemResponse
	for _, it := range m.OrderItems[orderID] {
		list = append(list, dto.OrderItemResponse{
			ProductID:       it.ProductID,
			ProductName:     it.ProductName,
			ProductImageURL: it.ProductImageURL,
			Quantity:        it.Quantity,
			Price:           it.Price,
		})
	}
	return list, nil
}

func (m *MockOrderRepository) UpdatePaymentStatus(ctx context.Context, paymentID string, status string) error {
	for _, o := range m.Orders {
		if o.PaymentID != nil && *o.PaymentID == paymentID {
			o.PaymentStatus = status
			return nil
		}
	}
	return errors.New("order not found")
}

func (m *MockOrderRepository) GetAll(ctx context.Context) ([]models.Order, error) {
	var list []models.Order
	for _, o := range m.Orders {
		list = append(list, *o)
	}
	return list, nil
}

func (m *MockOrderRepository) GetFiltered(ctx context.Context, phone string, orderID *int) ([]models.Order, error) {
	var list []models.Order
	for _, o := range m.Orders {
		if phone != "" && !strings.Contains(o.Phone, phone) {
			continue
		}
		if orderID != nil && o.ID != *orderID {
			continue
		}
		list = append(list, *o)
	}
	return list, nil
}

func (m *MockOrderRepository) UpdateStatusByID(ctx context.Context, orderID int, status string) error {
	o, exists := m.Orders[orderID]
	if !exists {
		return errors.New("order not found")
	}
	o.PaymentStatus = status
	return nil
}

func (m *MockOrderRepository) Delete(ctx context.Context, id int) error {
	delete(m.Orders, id)
	delete(m.OrderItems, id)
	return nil
}

type MockReservationRepository struct {
	Reservations map[int]*models.Reservation
	NextID       int
}

func NewMockReservationRepository() *MockReservationRepository {
	return &MockReservationRepository{
		Reservations: make(map[int]*models.Reservation),
		NextID:       1,
	}
}

func (m *MockReservationRepository) Create(ctx context.Context, r *models.Reservation) error {
	r.ID = m.NextID
	m.NextID++
	m.Reservations[r.ID] = r
	return nil
}

func (m *MockReservationRepository) GetByUserID(ctx context.Context, userID int) ([]models.Reservation, error) {
	var list []models.Reservation
	for _, r := range m.Reservations {
		if r.UserID != nil && *r.UserID == userID {
			list = append(list, *r)
		}
	}
	return list, nil
}

func (m *MockReservationRepository) GetAll(ctx context.Context) ([]models.Reservation, error) {
	var list []models.Reservation
	for _, r := range m.Reservations {
		list = append(list, *r)
	}
	return list, nil
}

func (m *MockReservationRepository) UpdateStatusByID(ctx context.Context, id int, status string) error {
	r, exists := m.Reservations[id]
	if !exists {
		return errors.New("reservation not found")
	}
	r.Status = status
	return nil
}

func (m *MockReservationRepository) Delete(ctx context.Context, id int) error {
	delete(m.Reservations, id)
	return nil
}

type MockBlogPostRepository struct{}

func (m *MockBlogPostRepository) GetAll(ctx context.Context, publishedOnly bool) ([]models.BlogPost, error) {
	return nil, nil
}
func (m *MockBlogPostRepository) GetByID(ctx context.Context, id int) (*models.BlogPost, error) {
	return nil, errors.New("not found")
}
func (m *MockBlogPostRepository) Create(ctx context.Context, p *models.BlogPost) error { return nil }
func (m *MockBlogPostRepository) Update(ctx context.Context, p *models.BlogPost) error { return nil }
func (m *MockBlogPostRepository) Delete(ctx context.Context, id int) error              { return nil }

type MockAuditLogRepository struct{}

func (m *MockAuditLogRepository) Log(ctx context.Context, adminID *int, action, entityType string, entityID *int, details string) error {
	return nil
}
func (m *MockAuditLogRepository) GetAll(ctx context.Context) ([]models.AuditLog, error) {
	return nil, nil
}

// --- TESTS ---

func TestAuthService_RegisterAndLogin(t *testing.T) {
	userRepo := NewMockUserRepository()
	authService := NewAuthService(userRepo, "my-secret-key")
	ctx := context.Background()

	regReq := dto.RegisterRequest{Name: "Test User", Phone: "+79998887766", Password: "password123"}

	userResp, err := authService.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("Expected no error on Register, got: %v", err)
	}
	if userResp.Name != regReq.Name || userResp.Phone != regReq.Phone {
		t.Errorf("Registered user name/phone mismatch. Got name: %s, phone: %s", userResp.Name, userResp.Phone)
	}

	// Duplicate phone should fail
	if _, err = authService.Register(ctx, regReq); err == nil {
		t.Error("Expected error when registering user with duplicate phone, got nil")
	}

	// Successful login
	token, err := authService.Login(ctx, dto.LoginRequest{Phone: "+79998887766", Password: "password123"})
	if err != nil {
		t.Fatalf("Expected no error on Login, got: %v", err)
	}
	if token == "" {
		t.Error("Expected non-empty JWT token on successful login")
	}

	// Wrong password
	if _, err = authService.Login(ctx, dto.LoginRequest{Phone: "+79998887766", Password: "wrongpassword"}); err == nil {
		t.Error("Expected error on login with incorrect password, got nil")
	}
}

func TestAdminService_GetStats(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	resRepo := NewMockReservationRepository()
	productRepo := NewMockProductRepository()

	adminService := NewAdminService(orderRepo, resRepo, productRepo, &MockBlogPostRepository{}, NewMockUserRepository(), &MockAuditLogRepository{})
	ctx := context.Background()

	orderRepo.Create(ctx, &models.Order{CustomerName: "Customer 1", Phone: "111", Address: "Addr 1", TotalPrice: 150.00, PaymentStatus: "paid"}, nil)
	orderRepo.Create(ctx, &models.Order{CustomerName: "Customer 2", Phone: "222", Address: "Addr 2", TotalPrice: 250.00, PaymentStatus: "paid"}, nil)
	orderRepo.Create(ctx, &models.Order{CustomerName: "Customer 3", Phone: "333", Address: "Addr 3", TotalPrice: 500.00, PaymentStatus: "cancelled"}, nil)

	resRepo.Create(ctx, &models.Reservation{CustomerName: "Reserver 1", Phone: "111", ReserveDate: time.Now(), GuestsCount: 4})

	productRepo.Create(ctx, &models.Product{Name: "Buuzy", Price: 90.00})
	productRepo.Create(ctx, &models.Product{Name: "Shulen", Price: 250.00})

	stats, err := adminService.GetStats(ctx, "", "")
	if err != nil {
		t.Fatalf("Expected no error on GetStats, got: %v", err)
	}

	if stats["total_orders"].(int) != 3 {
		t.Errorf("Expected 3 orders, got: %v", stats["total_orders"])
	}
	if stats["total_reservations"].(int) != 1 {
		t.Errorf("Expected 1 reservation, got: %v", stats["total_reservations"])
	}
	if stats["total_products"].(int) != 2 {
		t.Errorf("Expected 2 products, got: %v", stats["total_products"])
	}
	if stats["total_revenue"].(float64) != 400.00 {
		t.Errorf("Expected total revenue of 400.00, got: %f", stats["total_revenue"])
	}
}
