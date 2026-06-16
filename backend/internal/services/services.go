package services

import (
	"context"
	"fmt"
	"time"

	"backend/internal/dto"
	"backend/internal/mappers"
	"backend/internal/models"
	"backend/internal/repositories"

	"regexp"

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

type BlogService interface {
	GetPosts(ctx context.Context) ([]dto.BlogPostResponse, error)
	GetPostByID(ctx context.Context, id int) (dto.BlogPostResponse, error)
}

type AdminService interface {
	// Orders & reservations
	GetOrders(ctx context.Context) ([]dto.OrderResponse, error)
	GetOrdersFiltered(ctx context.Context, phone string, orderID *int) ([]dto.OrderResponse, error)
	UpdateOrderStatus(ctx context.Context, orderID int, status string) error
	DeleteOrder(ctx context.Context, id int) error
	GetReservations(ctx context.Context) ([]dto.ReservationResponse, error)
	UpdateReservationStatus(ctx context.Context, id int, status string) error
	DeleteReservation(ctx context.Context, id int) error
	// Products (admin view includes deleted)
	GetAdminProducts(ctx context.Context, categoryID *int) ([]dto.ProductResponse, error)
	GetAdminProductByID(ctx context.Context, id int) (dto.ProductResponse, error)
	CreateProduct(ctx context.Context, req dto.ProductResponse) (dto.ProductResponse, error)
	UpdateProduct(ctx context.Context, id int, req dto.ProductResponse) (dto.ProductResponse, error)
	// DeleteProduct returns softDeleted=true when the product had order history and was soft-deleted
	DeleteProduct(ctx context.Context, id int) (softDeleted bool, err error)
	// Categories
	CreateCategory(ctx context.Context, req dto.CategoryResponse) (dto.CategoryResponse, error)
	UpdateCategory(ctx context.Context, id int, req dto.CategoryResponse) (dto.CategoryResponse, error)
	DeleteCategory(ctx context.Context, id int) error
	// Blog
	GetAllBlogPosts(ctx context.Context) ([]dto.BlogPostResponse, error)
	CreateBlogPost(ctx context.Context, req dto.BlogPostResponse) (dto.BlogPostResponse, error)
	UpdateBlogPost(ctx context.Context, id int, req dto.BlogPostResponse) (dto.BlogPostResponse, error)
	DeleteBlogPost(ctx context.Context, id int) error
	// Stats
	GetStats(ctx context.Context, startDate, endDate string) (map[string]interface{}, error)
	// User management (super_admin only)
	GetUsers(ctx context.Context) ([]dto.AdminUserResponse, error)
	UpdateUserRole(ctx context.Context, targetUserID, currentAdminID int, newRole string) error
	DeleteUser(ctx context.Context, targetUserID, currentAdminID int) error
	// Audit log
	GetAuditLog(ctx context.Context) ([]dto.AuditLogResponse, error)
	LogAudit(ctx context.Context, adminID *int, action, entityType string, entityID *int, details string) error
}

// AuthServiceImpl

type AuthServiceImpl struct {
	userRepo  repositories.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo repositories.UserRepository, jwtSecret string) *AuthServiceImpl {
	return &AuthServiceImpl{userRepo: userRepo, jwtSecret: jwtSecret}
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
	if _, err := s.userRepo.GetByPhone(ctx, req.Phone); err == nil {
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

	if err := s.userRepo.Create(ctx, u); err != nil {
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
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
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

// ProductServiceImpl

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
	products, err := s.productRepo.GetProducts(ctx, categoryID, false)
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
	if p.IsDeleted {
		return dto.ProductResponse{}, fmt.Errorf("товар не найден")
	}
	return mappers.ToProductResponse(p), nil
}

// OrderServiceImpl

type OrderServiceImpl struct {
	orderRepo   repositories.OrderRepository
	productRepo repositories.ProductRepository
}

func NewOrderService(orderRepo repositories.OrderRepository, productRepo repositories.ProductRepository) *OrderServiceImpl {
	return &OrderServiceImpl{orderRepo: orderRepo, productRepo: productRepo}
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
		if prod.IsDeleted {
			return nil, fmt.Errorf("товар '%s' недоступен", prod.Name)
		}
		if !prod.IsAvailable {
			return nil, fmt.Errorf("товар '%s' временно недоступен", prod.Name)
		}

		totalPrice += prod.Price * float64(itemReq.Quantity)
		orderItems = append(orderItems, models.OrderItem{
			ProductID:       itemReq.ProductID,
			ProductName:     prod.Name,
			ProductImageURL: prod.ImageURL,
			Quantity:        itemReq.Quantity,
			Price:           prod.Price,
		})
	}

	paymentStatus := "awaiting_delivery"
	paymentID := ""
	if req.PaymentMethod == "online" {
		paymentStatus = "pending"
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

	if err := s.orderRepo.Create(ctx, o, orderItems); err != nil {
		return nil, err
	}

	var itemResponses []dto.OrderItemResponse
	for _, oi := range orderItems {
		itemResponses = append(itemResponses, dto.OrderItemResponse{
			ProductID:       oi.ProductID,
			ProductName:     oi.ProductName,
			ProductImageURL: oi.ProductImageURL,
			Quantity:        oi.Quantity,
			Price:           oi.Price,
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

// ReservationServiceImpl

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

	var hour, minute int
	if n, _ := fmt.Sscanf(req.ReserveTime, "%d:%d", &hour, &minute); n != 2 || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return nil, fmt.Errorf("неверный формат времени, ожидается ЧЧ:ММ")
	}
	if hour < 10 || hour > 21 || (hour == 21 && minute > 30) {
		return nil, fmt.Errorf("бронирование доступно только в рабочее время кафе: с 10:00 до 21:30")
	}
	moscow := time.FixedZone("MSK", 3*60*60)
	reserveDateTime := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), hour, minute, 0, 0, moscow)
	if !reserveDateTime.After(time.Now().In(moscow)) {
		return nil, fmt.Errorf("выбранное время бронирования уже прошло. Пожалуйста, выберите более позднее время")
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
	if err := s.reservationRepo.Create(ctx, res); err != nil {
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

// BlogServiceImpl

type BlogServiceImpl struct {
	blogRepo repositories.BlogPostRepository
}

func NewBlogService(blogRepo repositories.BlogPostRepository) *BlogServiceImpl {
	return &BlogServiceImpl{blogRepo: blogRepo}
}

func (s *BlogServiceImpl) GetPosts(ctx context.Context) ([]dto.BlogPostResponse, error) {
	posts, err := s.blogRepo.GetAll(ctx, true)
	if err != nil {
		return nil, err
	}
	return mappers.ToBlogPostResponseList(posts), nil
}

func (s *BlogServiceImpl) GetPostByID(ctx context.Context, id int) (dto.BlogPostResponse, error) {
	p, err := s.blogRepo.GetByID(ctx, id)
	if err != nil {
		return dto.BlogPostResponse{}, err
	}
	return mappers.ToBlogPostResponse(p), nil
}

// AdminServiceImpl

type AdminServiceImpl struct {
	orderRepo   repositories.OrderRepository
	resRepo     repositories.ReservationRepository
	productRepo repositories.ProductRepository
	blogRepo    repositories.BlogPostRepository
	userRepo    repositories.UserRepository
	auditRepo   repositories.AuditLogRepository
}

func NewAdminService(
	orderRepo repositories.OrderRepository,
	resRepo repositories.ReservationRepository,
	productRepo repositories.ProductRepository,
	blogRepo repositories.BlogPostRepository,
	userRepo repositories.UserRepository,
	auditRepo repositories.AuditLogRepository,
) *AdminServiceImpl {
	return &AdminServiceImpl{
		orderRepo:   orderRepo,
		resRepo:     resRepo,
		productRepo: productRepo,
		blogRepo:    blogRepo,
		userRepo:    userRepo,
		auditRepo:   auditRepo,
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

func (s *AdminServiceImpl) GetOrdersFiltered(ctx context.Context, phone string, orderID *int) ([]dto.OrderResponse, error) {
	orders, err := s.orderRepo.GetFiltered(ctx, phone, orderID)
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

func (s *AdminServiceImpl) DeleteOrder(ctx context.Context, id int) error {
	return s.orderRepo.Delete(ctx, id)
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

func (s *AdminServiceImpl) DeleteReservation(ctx context.Context, id int) error {
	return s.resRepo.Delete(ctx, id)
}

func (s *AdminServiceImpl) GetAdminProducts(ctx context.Context, categoryID *int) ([]dto.ProductResponse, error) {
	products, err := s.productRepo.GetProducts(ctx, categoryID, true)
	if err != nil {
		return nil, err
	}
	return mappers.ToProductResponseList(products), nil
}

func (s *AdminServiceImpl) GetAdminProductByID(ctx context.Context, id int) (dto.ProductResponse, error) {
	p, err := s.productRepo.GetProductByID(ctx, id)
	if err != nil {
		return dto.ProductResponse{}, err
	}
	return mappers.ToProductResponse(p), nil
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
		IsAvailable: req.IsAvailable,
	}
	if err := s.productRepo.Create(ctx, p); err != nil {
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
		IsAvailable: req.IsAvailable,
	}
	if err := s.productRepo.Update(ctx, p); err != nil {
		return dto.ProductResponse{}, err
	}
	return mappers.ToProductResponse(p), nil
}

func (s *AdminServiceImpl) DeleteProduct(ctx context.Context, id int) (bool, error) {
	hasOrders, err := s.productRepo.HasOrderHistory(ctx, id)
	if err != nil {
		return false, err
	}
	if hasOrders {
		if err := s.productRepo.SoftDelete(ctx, id); err != nil {
			return false, err
		}
		return true, nil
	}
	if err := s.productRepo.Delete(ctx, id); err != nil {
		return false, err
	}
	return false, nil
}

func (s *AdminServiceImpl) CreateCategory(ctx context.Context, req dto.CategoryResponse) (dto.CategoryResponse, error) {
	if req.Name == "" || req.Slug == "" {
		return dto.CategoryResponse{}, fmt.Errorf("название категории и слаг обязательны")
	}
	c := &models.Category{Name: req.Name, Slug: req.Slug}
	if err := s.productRepo.CreateCategory(ctx, c); err != nil {
		return dto.CategoryResponse{}, err
	}
	return mappers.ToCategoryResponse(c), nil
}

func (s *AdminServiceImpl) UpdateCategory(ctx context.Context, id int, req dto.CategoryResponse) (dto.CategoryResponse, error) {
	if req.Name == "" || req.Slug == "" {
		return dto.CategoryResponse{}, fmt.Errorf("название категории и слаг обязательны")
	}
	c := &models.Category{ID: id, Name: req.Name, Slug: req.Slug}
	if err := s.productRepo.UpdateCategory(ctx, c); err != nil {
		return dto.CategoryResponse{}, err
	}
	return mappers.ToCategoryResponse(c), nil
}

func (s *AdminServiceImpl) DeleteCategory(ctx context.Context, id int) error {
	hasProducts, err := s.productRepo.HasProductsInCategory(ctx, id)
	if err != nil {
		return err
	}
	if hasProducts {
		return fmt.Errorf("невозможно удалить категорию: сначала удалите или перенесите все блюда из этой категории")
	}
	return s.productRepo.DeleteCategory(ctx, id)
}

func (s *AdminServiceImpl) GetAllBlogPosts(ctx context.Context) ([]dto.BlogPostResponse, error) {
	posts, err := s.blogRepo.GetAll(ctx, false)
	if err != nil {
		return nil, err
	}
	return mappers.ToBlogPostResponseList(posts), nil
}

func (s *AdminServiceImpl) CreateBlogPost(ctx context.Context, req dto.BlogPostResponse) (dto.BlogPostResponse, error) {
	if req.Title == "" {
		return dto.BlogPostResponse{}, fmt.Errorf("заголовок статьи обязателен")
	}
	p := &models.BlogPost{
		Title:       req.Title,
		Subtitle:    req.Subtitle,
		Content:     req.Content,
		ImageURL:    req.ImageURL,
		Tag:         req.Tag,
		ReadTime:    req.ReadTime,
		IsPublished: req.IsPublished,
	}
	if err := s.blogRepo.Create(ctx, p); err != nil {
		return dto.BlogPostResponse{}, err
	}
	return mappers.ToBlogPostResponse(p), nil
}

func (s *AdminServiceImpl) UpdateBlogPost(ctx context.Context, id int, req dto.BlogPostResponse) (dto.BlogPostResponse, error) {
	if req.Title == "" {
		return dto.BlogPostResponse{}, fmt.Errorf("заголовок статьи обязателен")
	}
	p := &models.BlogPost{
		ID:          id,
		Title:       req.Title,
		Subtitle:    req.Subtitle,
		Content:     req.Content,
		ImageURL:    req.ImageURL,
		Tag:         req.Tag,
		ReadTime:    req.ReadTime,
		IsPublished: req.IsPublished,
	}
	if err := s.blogRepo.Update(ctx, p); err != nil {
		return dto.BlogPostResponse{}, err
	}
	return mappers.ToBlogPostResponse(p), nil
}

func (s *AdminServiceImpl) DeleteBlogPost(ctx context.Context, id int) error {
	return s.blogRepo.Delete(ctx, id)
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
	products, err := s.productRepo.GetProducts(ctx, nil, false)
	if err != nil {
		return nil, err
	}

	var totalRevenue float64
	var filteredOrders, filteredReservations int

	var start, end time.Time
	hasDates := startDate != "" && endDate != ""
	if hasDates {
		start, _ = time.Parse("2006-01-02", startDate)
		end, _ = time.Parse("2006-01-02", endDate)
		end = end.Add(24*time.Hour - time.Second)
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

	return map[string]interface{}{
		"total_orders":       filteredOrders,
		"total_reservations": filteredReservations,
		"total_products":     len(products),
		"total_revenue":      totalRevenue,
	}, nil
}

func (s *AdminServiceImpl) GetUsers(ctx context.Context) ([]dto.AdminUserResponse, error) {
	users, err := s.userRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	var res []dto.AdminUserResponse
	for _, u := range users {
		res = append(res, dto.AdminUserResponse{
			ID:             u.ID,
			Name:           u.Name,
			Phone:          u.Phone,
			Email:          u.Email,
			DefaultAddress: u.DefaultAddress,
			Role:           u.Role,
			CreatedAt:      u.CreatedAt,
		})
	}
	return res, nil
}

func (s *AdminServiceImpl) UpdateUserRole(ctx context.Context, targetUserID, currentAdminID int, newRole string) error {
	validRoles := map[string]bool{"user": true, "admin": true, "super_admin": true}
	if !validRoles[newRole] {
		return fmt.Errorf("недопустимая роль: %s", newRole)
	}
	if targetUserID == currentAdminID {
		return fmt.Errorf("нельзя изменить собственную роль")
	}

	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("пользователь не найден")
	}

	if targetUser.Role == "super_admin" && newRole != "super_admin" {
		count, err := s.userRepo.CountByRole(ctx, "super_admin")
		if err != nil {
			return err
		}
		if count <= 1 {
			return fmt.Errorf("нельзя понизить роль последнего суперадминистратора")
		}
	}

	return s.userRepo.UpdateRole(ctx, targetUserID, newRole)
}

func (s *AdminServiceImpl) DeleteUser(ctx context.Context, targetUserID, currentAdminID int) error {
	if targetUserID == currentAdminID {
		return fmt.Errorf("нельзя удалить собственный аккаунт")
	}

	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("пользователь не найден")
	}

	if targetUser.Role == "super_admin" {
		count, err := s.userRepo.CountByRole(ctx, "super_admin")
		if err != nil {
			return err
		}
		if count <= 1 {
			return fmt.Errorf("нельзя удалить последнего суперадминистратора")
		}
	}

	return s.userRepo.Delete(ctx, targetUserID)
}

func (s *AdminServiceImpl) GetAuditLog(ctx context.Context) ([]dto.AuditLogResponse, error) {
	logs, err := s.auditRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	var res []dto.AuditLogResponse
	for _, l := range logs {
		res = append(res, dto.AuditLogResponse{
			ID:         l.ID,
			AdminID:    l.AdminID,
			Action:     l.Action,
			EntityType: l.EntityType,
			EntityID:   l.EntityID,
			Details:    l.Details,
			CreatedAt:  l.CreatedAt,
		})
	}
	return res, nil
}

func (s *AdminServiceImpl) LogAudit(ctx context.Context, adminID *int, action, entityType string, entityID *int, details string) error {
	return s.auditRepo.Log(ctx, adminID, action, entityType, entityID, details)
}
