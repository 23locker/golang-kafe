package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"backend/internal/dto"
	"backend/internal/mappers"
	"backend/internal/models"
	"backend/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var nonDigitRe = regexp.MustCompile(`\D`)

func validatePhone(phone string) error {
	if phone == "" {
		return fmt.Errorf("номер телефона обязателен")
	}
	digits := nonDigitRe.ReplaceAllString(phone, "")
	if len(digits) < 10 || len(digits) > 12 {
		return fmt.Errorf("неверный формат номера телефона. Используйте формат +7XXXXXXXXXX")
	}
	return nil
}

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (string, error)
	GetUserByID(ctx context.Context, id int) (*dto.UserResponse, error)
	UpdateAddress(ctx context.Context, userID int, address string) error
	UpdateEmail(ctx context.Context, userID int, email string) error
}

type ProductService interface {
	GetCategories(ctx context.Context) ([]dto.CategoryResponse, error)
	GetProducts(ctx context.Context, categoryID *int) ([]dto.ProductResponse, error)
	GetProductByID(ctx context.Context, id int) (dto.ProductResponse, error)
}

type OrderService interface {
	CreateOrder(ctx context.Context, userID *int, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetOrdersByUserID(ctx context.Context, userID int) ([]dto.OrderResponse, error)
}

type ReservationService interface {
	CreateReservation(ctx context.Context, userID *int, req dto.CreateReservationRequest) (*dto.ReservationResponse, error)
	GetReservationsByUserID(ctx context.Context, userID int) ([]dto.ReservationResponse, error)
}

type AdminService interface {
	GetOrders(ctx context.Context) ([]dto.OrderResponse, error)
	UpdateOrderStatus(ctx context.Context, orderID int, status string) error
	GetReservations(ctx context.Context) ([]dto.ReservationResponse, error)
	UpdateReservationStatus(ctx context.Context, id int, status string) error
	CreateProduct(ctx context.Context, req dto.ProductResponse) (dto.ProductResponse, error)
	UpdateProduct(ctx context.Context, id int, req dto.ProductResponse) (dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, id int) error
	CreateCategory(ctx context.Context, req dto.CategoryResponse) (dto.CategoryResponse, error)
	UpdateCategory(ctx context.Context, id int, req dto.CategoryResponse) (dto.CategoryResponse, error)
	DeleteCategory(ctx context.Context, id int) error
	GetStats(ctx context.Context, startDate, endDate string) (map[string]interface{}, error)
}

type AuthServiceImpl struct {
	userRepo  repositories.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo repositories.UserRepository, jwtSecret string) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
	if req.Name == "" || req.Phone == "" || req.Password == "" {
		return nil, fmt.Errorf("имя, телефон и пароль обязательны для заполнения")
	}

	if err := validatePhone(req.Phone); err != nil {
		return nil, err
	}

	if len(req.Password) < 6 {
		return nil, fmt.Errorf("пароль должен содержать не менее 6 символов")
	}

	_, err := s.userRepo.GetByPhone(ctx, req.Phone)
	if err == nil {
		return nil, fmt.Errorf("пользователь с таким номером телефона уже существует")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка хэширования пароля: %w", err)
	}

	u := &models.User{
		Name:         req.Name,
		Phone:        req.Phone,
		PasswordHash: string(hashedPassword),
	}
	if req.Email != "" {
		u.Email = &req.Email
	}

	err = s.userRepo.Create(ctx, u)
	if err != nil {
		return nil, err
	}

	return mappers.ToUserResponse(u), nil
}

func (s *AuthServiceImpl) Login(ctx context.Context, req dto.LoginRequest) (string, error) {
	if req.Phone == "" || req.Password == "" {
		return "", fmt.Errorf("телефон и пароль обязательны")
	}

	if err := validatePhone(req.Phone); err != nil {
		return "", err
	}

	u, err := s.userRepo.GetByPhone(ctx, req.Phone)
	if err != nil {
		return "", fmt.Errorf("неверный номер телефона или пароль")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password))
	if err != nil {
		return "", fmt.Errorf("неверный номер телефона или пароль")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.ID,
		"role":    u.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("ошибка генерации токена сессии: %w", err)
	}

	return tokenStr, nil
}

func (s *AuthServiceImpl) GetUserByID(ctx context.Context, id int) (*dto.UserResponse, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mappers.ToUserResponse(u), nil
}

func (s *AuthServiceImpl) UpdateAddress(ctx context.Context, userID int, address string) error {
	return s.userRepo.UpdateAddress(ctx, userID, address)
}

func (s *AuthServiceImpl) UpdateEmail(ctx context.Context, userID int, email string) error {
	return s.userRepo.UpdateEmail(ctx, userID, email)
}

type ProductServiceImpl struct {
	productRepo repositories.ProductRepository
}

func NewProductService(productRepo repositories.ProductRepository) *ProductServiceImpl {
	return &ProductServiceImpl{productRepo: productRepo}
}

func (s *ProductServiceImpl) GetCategories(ctx context.Context) ([]dto.CategoryResponse, error) {
	categories, err := s.productRepo.GetCategories(ctx)
	if err != nil {
		return nil, err
	}
	return mappers.ToCategoryResponseList(categories), nil
}

func (s *ProductServiceImpl) GetProducts(ctx context.Context, categoryID *int) ([]dto.ProductResponse, error) {
	products, err := s.productRepo.GetProducts(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	return mappers.ToProductResponseList(products), nil
}

func (s *ProductServiceImpl) GetProductByID(ctx context.Context, id int) (dto.ProductResponse, error) {
	p, err := s.productRepo.GetProductByID(ctx, id)
	if err != nil {
		return dto.ProductResponse{}, err
	}
	return mappers.ToProductResponse(p), nil
}

type OrderServiceImpl struct {
	orderRepo   repositories.OrderRepository
	productRepo repositories.ProductRepository
}

func NewOrderService(orderRepo repositories.OrderRepository, productRepo repositories.ProductRepository) *OrderServiceImpl {
	return &OrderServiceImpl{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (s *OrderServiceImpl) CreateOrder(ctx context.Context, userID *int, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	if req.CustomerName == "" || req.Phone == "" || req.Address == "" || len(req.Items) == 0 {
		return nil, fmt.Errorf("все поля заказа обязательны для заполнения")
	}

	if err := validatePhone(req.Phone); err != nil {
		return nil, err
	}

	var orderItems []models.OrderItem
	var totalPrice float64

	for _, itemReq := range req.Items {
		prod, err := s.productRepo.GetProductByID(ctx, itemReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("товар с идентификатором %d не найден", itemReq.ProductID)
		}

		itemPrice := prod.Price
		totalPrice += itemPrice * float64(itemReq.Quantity)

		orderItems = append(orderItems, models.OrderItem{
			ProductID: itemReq.ProductID,
			Quantity:  itemReq.Quantity,
			Price:     itemPrice,
		})
	}

	var paymentStatus string
	if req.PaymentMethod == "online" {
		paymentStatus = "pending"
	} else {
		paymentStatus = "awaiting_delivery"
	}

	paymentID := ""
	if req.PaymentMethod == "online" {
		paymentID = fmt.Sprintf("pay_%d", time.Now().UnixNano())
	}

	o := &models.Order{
		UserID:        userID,
		CustomerName:  req.CustomerName,
		Phone:         req.Phone,
		Address:       req.Address,
		TotalPrice:    totalPrice,
		PaymentStatus: paymentStatus,
	}
	if paymentID != "" {
		o.PaymentID = &paymentID
	}

	err := s.orderRepo.Create(ctx, o, orderItems)
	if err != nil {
		return nil, err
	}

	var itemResponses []dto.OrderItemResponse
	for _, oi := range orderItems {
		pInfo, _ := s.productRepo.GetProductByID(ctx, oi.ProductID)
		itemResponses = append(itemResponses, dto.OrderItemResponse{
			ProductID:   oi.ProductID,
			ProductName: pInfo.Name,
			Quantity:    oi.Quantity,
			Price:       oi.Price,
		})
	}

	res := mappers.ToOrderResponse(o, itemResponses)
	return &res, nil
}

func (s *OrderServiceImpl) GetOrdersByUserID(ctx context.Context, userID int) ([]dto.OrderResponse, error) {
	orders, err := s.orderRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var res []dto.OrderResponse
	for i := range orders {
		items, err := s.orderRepo.GetOrderItems(ctx, orders[i].ID)
		if err != nil {
			return nil, err
		}
		res = append(res, mappers.ToOrderResponse(&orders[i], items))
	}
	return res, nil
}

type ReservationServiceImpl struct {
	reservationRepo repositories.ReservationRepository
}

func NewReservationService(reservationRepo repositories.ReservationRepository) *ReservationServiceImpl {
	return &ReservationServiceImpl{reservationRepo: reservationRepo}
}

func (s *ReservationServiceImpl) CreateReservation(ctx context.Context, userID *int, req dto.CreateReservationRequest) (*dto.ReservationResponse, error) {
	if req.CustomerName == "" || req.Phone == "" || req.ReserveDate == "" || req.ReserveTime == "" || req.GuestsCount <= 0 {
		return nil, fmt.Errorf("все поля бронирования обязательны")
	}

	if err := validatePhone(req.Phone); err != nil {
		return nil, err
	}

	parsedDate, err := time.Parse("2006-01-02", req.ReserveDate)
	if err != nil {
		return nil, fmt.Errorf("неверный формат даты, ожидается гггг-мм-дд: %w", err)
	}

	res := &models.Reservation{
		UserID:       userID,
		CustomerName: req.CustomerName,
		Phone:        req.Phone,
		ReserveDate:  parsedDate,
		ReserveTime:  req.ReserveTime,
		GuestsCount:  req.GuestsCount,
		Comment:      req.Comment,
	}

	err = s.reservationRepo.Create(ctx, res)
	if err != nil {
		return nil, err
	}

	resp := mappers.ToReservationResponse(res)
	return &resp, nil
}

func (s *ReservationServiceImpl) GetReservationsByUserID(ctx context.Context, userID int) ([]dto.ReservationResponse, error) {
	reservations, err := s.reservationRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return mappers.ToReservationResponseList(reservations), nil
}

type AdminServiceImpl struct {
	orderRepo   repositories.OrderRepository
	resRepo     repositories.ReservationRepository
	productRepo repositories.ProductRepository
}

func NewAdminService(
	orderRepo repositories.OrderRepository,
	resRepo repositories.ReservationRepository,
	productRepo repositories.ProductRepository,
) *AdminServiceImpl {
	return &AdminServiceImpl{
		orderRepo:   orderRepo,
		resRepo:     resRepo,
		productRepo: productRepo,
	}
}

func (s *AdminServiceImpl) GetOrders(ctx context.Context) ([]dto.OrderResponse, error) {
	orders, err := s.orderRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var res []dto.OrderResponse
	for i := range orders {
		items, err := s.orderRepo.GetOrderItems(ctx, orders[i].ID)
		if err != nil {
			return nil, err
		}
		res = append(res, mappers.ToOrderResponse(&orders[i], items))
	}
	return res, nil
}

func (s *AdminServiceImpl) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	if status == "" {
		return fmt.Errorf("статус не может быть пустым")
	}
	return s.orderRepo.UpdateStatusByID(ctx, orderID, status)
}

func (s *AdminServiceImpl) GetReservations(ctx context.Context) ([]dto.ReservationResponse, error) {
	reservations, err := s.resRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return mappers.ToReservationResponseList(reservations), nil
}

func (s *AdminServiceImpl) UpdateReservationStatus(ctx context.Context, id int, status string) error {
	if status == "" {
		return fmt.Errorf("статус не может быть пустым")
	}
	return s.resRepo.UpdateStatusByID(ctx, id, status)
}

func (s *AdminServiceImpl) CreateCategory(ctx context.Context, req dto.CategoryResponse) (dto.CategoryResponse, error) {
	if req.Name == "" || req.Slug == "" {
		return dto.CategoryResponse{}, fmt.Errorf("название категории и слаг обязательны")
	}
	c := &models.Category{
		Name: req.Name,
		Slug: req.Slug,
	}
	err := s.productRepo.CreateCategory(ctx, c)
	if err != nil {
		return dto.CategoryResponse{}, err
	}
	return mappers.ToCategoryResponse(c), nil
}

func (s *AdminServiceImpl) UpdateCategory(ctx context.Context, id int, req dto.CategoryResponse) (dto.CategoryResponse, error) {
	if req.Name == "" || req.Slug == "" {
		return dto.CategoryResponse{}, fmt.Errorf("название категории и слаг обязательны")
	}
	c := &models.Category{
		ID:   id,
		Name: req.Name,
		Slug: req.Slug,
	}
	err := s.productRepo.UpdateCategory(ctx, c)
	if err != nil {
		return dto.CategoryResponse{}, err
	}
	return mappers.ToCategoryResponse(c), nil
}

func (s *AdminServiceImpl) DeleteCategory(ctx context.Context, id int) error {
	// Сначала получаем все продукты этой категории
	products, err := s.productRepo.GetProducts(ctx, &id)
	if err == nil {
		for _, p := range products {
			// Удаляем картинку
			if p.ImageURL != "" && strings.HasPrefix(p.ImageURL, "/uploads/") {
				workDir, _ := os.Getwd()
				oldFilePath := filepath.Join(workDir, "uploads", strings.TrimPrefix(p.ImageURL, "/uploads/"))
				os.Remove(oldFilePath)
			}
			// Удаляем продукт
			s.productRepo.Delete(ctx, p.ID)
		}
	}
	return s.productRepo.DeleteCategory(ctx, id)
}

func (s *AdminServiceImpl) CreateProduct(ctx context.Context, req dto.ProductResponse) (dto.ProductResponse, error) {
	if req.Name == "" || req.Price <= 0 {
		return dto.ProductResponse{}, fmt.Errorf("название товара и цена обязательны")
	}

	p := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
		ImageURL:    req.ImageURL,
		Weight:      req.Weight,
		Calories:    req.Calories,
	}

	err := s.productRepo.Create(ctx, p)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	return mappers.ToProductResponse(p), nil
}

func (s *AdminServiceImpl) UpdateProduct(ctx context.Context, id int, req dto.ProductResponse) (dto.ProductResponse, error) {
	if req.Name == "" || req.Price <= 0 {
		return dto.ProductResponse{}, fmt.Errorf("название товара и цена обязательны")
	}

	p := &models.Product{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
		ImageURL:    req.ImageURL,
		Weight:      req.Weight,
		Calories:    req.Calories,
	}

	err := s.productRepo.Update(ctx, p)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	return mappers.ToProductResponse(p), nil
}

func (s *AdminServiceImpl) DeleteProduct(ctx context.Context, id int) error {
	return s.productRepo.Delete(ctx, id)
}

func (s *AdminServiceImpl) GetStats(ctx context.Context, startDate, endDate string) (map[string]interface{}, error) {
	orders, err := s.orderRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	reservations, err := s.resRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	products, err := s.productRepo.GetProducts(ctx, nil)
	if err != nil {
		return nil, err
	}

	var totalRevenue float64
	var filteredOrders int
	var filteredReservations int

	var start, end time.Time
	hasDates := startDate != "" && endDate != ""
	if hasDates {
		start, _ = time.Parse("2006-01-02", startDate)
		end, _ = time.Parse("2006-01-02", endDate)
		end = end.Add(24*time.Hour - time.Second) // До конца дня
	}

	for _, o := range orders {
		if hasDates && (o.CreatedAt.Before(start) || o.CreatedAt.After(end)) {
			continue
		}
		filteredOrders++
		if o.PaymentStatus != "cancelled" && o.PaymentStatus != "Отменен" {
			totalRevenue += o.TotalPrice
		}
	}

	for _, r := range reservations {
		if hasDates && (r.ReserveDate.Before(start) || r.ReserveDate.After(end)) {
			continue
		}
		filteredReservations++
	}

	stats := map[string]interface{}{
		"total_orders":       filteredOrders,
		"total_reservations": filteredReservations,
		"total_products":     len(products),
		"total_revenue":      totalRevenue,
	}

	return stats, nil
}
