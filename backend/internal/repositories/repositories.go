package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"backend/internal/dto"
	"backend/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, u *models.User) error
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByPhone(ctx context.Context, phone string) (*models.User, error)
	UpdateAddress(ctx context.Context, userID int, address string) error
	UpdateEmail(ctx context.Context, userID int, email string) error
}

type ProductRepository interface {
	GetCategories(ctx context.Context) ([]models.Category, error)
	GetProducts(ctx context.Context, categoryID *int) ([]models.Product, error)
	GetProductByID(ctx context.Context, id int) (*models.Product, error)
	Create(ctx context.Context, p *models.Product) error
	Update(ctx context.Context, p *models.Product) error
	Delete(ctx context.Context, id int) error
	CreateCategory(ctx context.Context, c *models.Category) error
	UpdateCategory(ctx context.Context, c *models.Category) error
	DeleteCategory(ctx context.Context, id int) error
}

type OrderRepository interface {
	Create(ctx context.Context, o *models.Order, items []models.OrderItem) error
	GetByUserID(ctx context.Context, userID int) ([]models.Order, error)
	GetOrderItems(ctx context.Context, orderID int) ([]dto.OrderItemResponse, error)
	UpdatePaymentStatus(ctx context.Context, paymentID string, status string) error
	GetAll(ctx context.Context) ([]models.Order, error)
	UpdateStatusByID(ctx context.Context, orderID int, status string) error
}

type BlogPostRepository interface {
	GetAll(ctx context.Context, publishedOnly bool) ([]models.BlogPost, error)
	GetByID(ctx context.Context, id int) (*models.BlogPost, error)
	Create(ctx context.Context, p *models.BlogPost) error
	Update(ctx context.Context, p *models.BlogPost) error
	Delete(ctx context.Context, id int) error
}

type ReservationRepository interface {
	Create(ctx context.Context, r *models.Reservation) error
	GetByUserID(ctx context.Context, userID int) ([]models.Reservation, error)
	GetAll(ctx context.Context) ([]models.Reservation, error)
	UpdateStatusByID(ctx context.Context, id int, status string) error
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *models.User) error {
	if u.Role == "" {
		u.Role = "user"
	}
	query := `INSERT INTO users (name, phone, email, password_hash, default_address, role)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	err := r.db.QueryRowContext(ctx, query, u.Name, u.Phone, u.Email, u.PasswordHash, u.DefaultAddress, u.Role).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка создания пользователя: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `SELECT id, name, phone, email, password_hash, default_address, role, created_at FROM users WHERE id = $1`
	var u models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Name, &u.Phone, &u.Email, &u.PasswordHash, &u.DefaultAddress, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка получения пользователя по идентификатору: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	query := `SELECT id, name, phone, email, password_hash, default_address, role, created_at FROM users WHERE phone = $1`
	var u models.User
	err := r.db.QueryRowContext(ctx, query, phone).Scan(&u.ID, &u.Name, &u.Phone, &u.Email, &u.PasswordHash, &u.DefaultAddress, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка получения пользователя по телефону: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepository) UpdateAddress(ctx context.Context, userID int, address string) error {
	query := `UPDATE users SET default_address = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, address, userID)
	if err != nil {
		return fmt.Errorf("ошибка обновления адреса пользователя: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) UpdateEmail(ctx context.Context, userID int, email string) error {
	query := `UPDATE users SET email = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, email, userID)
	if err != nil {
		return fmt.Errorf("ошибка обновления email пользователя: %w", err)
	}
	return nil
}

type PostgresProductRepository struct {
	db *sql.DB
}

func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) GetCategories(ctx context.Context) ([]models.Category, error) {
	query := `SELECT id, name, slug FROM categories`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения категорий: %w", err)
	}
	defer rows.Close()

	var list []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug); err != nil {
			return nil, fmt.Errorf("ошибка сканирования категории: %w", err)
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *PostgresProductRepository) GetProducts(ctx context.Context, categoryID *int) ([]models.Product, error) {
	var query string
	var rows *sql.Rows
	var err error

	if categoryID != nil {
		query = `SELECT id, name, description, price, category_id, image_url, weight, calories, is_available FROM products WHERE category_id = $1`
		rows, err = r.db.QueryContext(ctx, query, *categoryID)
	} else {
		query = `SELECT id, name, description, price, category_id, image_url, weight, calories, is_available FROM products`
		rows, err = r.db.QueryContext(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка товаров: %w", err)
	}
	defer rows.Close()

	var list []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CategoryID, &p.ImageURL, &p.Weight, &p.Calories, &p.IsAvailable); err != nil {
			return nil, fmt.Errorf("ошибка сканирования товара: %w", err)
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *PostgresProductRepository) GetProductByID(ctx context.Context, id int) (*models.Product, error) {
	query := `SELECT id, name, description, price, category_id, image_url, weight, calories, is_available FROM products WHERE id = $1`
	var p models.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CategoryID, &p.ImageURL, &p.Weight, &p.Calories, &p.IsAvailable)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("товар не найден")
		}
		return nil, fmt.Errorf("ошибка получения товара: %w", err)
	}
	return &p, nil
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *models.Product) error {
	query := `INSERT INTO products (name, description, price, category_id, image_url, weight, calories, is_available) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, p.Name, p.Description, p.Price, p.CategoryID, p.ImageURL, p.Weight, p.Calories, p.IsAvailable).Scan(&p.ID)
	if err != nil {
		return fmt.Errorf("ошибка добавления товара в базу: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *models.Product) error {
	query := `UPDATE products SET name = $1, description = $2, price = $3, category_id = $4, image_url = $5, weight = $6, calories = $7, is_available = $8 
	          WHERE id = $9`
	_, err := r.db.ExecContext(ctx, query, p.Name, p.Description, p.Price, p.CategoryID, p.ImageURL, p.Weight, p.Calories, p.IsAvailable, p.ID)
	if err != nil {
		return fmt.Errorf("ошибка обновления товара в базе: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления товара из базы: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) CreateCategory(ctx context.Context, c *models.Category) error {
	query := `INSERT INTO categories (name, slug) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, c.Name, c.Slug).Scan(&c.ID)
	if err != nil {
		return fmt.Errorf("ошибка добавления категории: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) UpdateCategory(ctx context.Context, c *models.Category) error {
	query := `UPDATE categories SET name = $1, slug = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, c.Name, c.Slug, c.ID)
	if err != nil {
		return fmt.Errorf("ошибка обновления категории: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) DeleteCategory(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления категории: %w", err)
	}
	return nil
}

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, o *models.Order, items []models.OrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback()

	queryOrder := `INSERT INTO orders (user_id, customer_name, phone, address, total_price, payment_status, payment_id) 
	               VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`
	err = tx.QueryRowContext(ctx, queryOrder, o.UserID, o.CustomerName, o.Phone, o.Address, o.TotalPrice, o.PaymentStatus, o.PaymentID).Scan(&o.ID, &o.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка записи заказа: %w", err)
	}

	queryItem := `INSERT INTO order_items (order_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)`
	for i := range items {
		_, err := tx.ExecContext(ctx, queryItem, o.ID, items[i].ProductID, items[i].Quantity, items[i].Price)
		if err != nil {
			return fmt.Errorf("ошибка записи элемента заказа: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}
	return nil
}

func (r *PostgresOrderRepository) GetByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	query := `SELECT id, user_id, customer_name, phone, address, total_price, payment_status, payment_id, created_at 
	          FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заказов пользователя: %w", err)
	}
	defer rows.Close()

	var list []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.UserID, &o.CustomerName, &o.Phone, &o.Address, &o.TotalPrice, &o.PaymentStatus, &o.PaymentID, &o.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования заказа: %w", err)
		}
		list = append(list, o)
	}
	return list, nil
}

func (r *PostgresOrderRepository) GetOrderItems(ctx context.Context, orderID int) ([]dto.OrderItemResponse, error) {
	query := `SELECT oi.id, oi.product_id, p.name, oi.quantity, oi.price 
	          FROM order_items oi
	          JOIN products p ON oi.product_id = p.id
	          WHERE oi.order_id = $1`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения позиций заказа: %w", err)
	}
	defer rows.Close()

	var list []dto.OrderItemResponse
	for rows.Next() {
		var item dto.OrderItemResponse
		err := rows.Scan(&item.ID, &item.ProductID, &item.ProductName, &item.Quantity, &item.Price)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования позиции заказа: %w", err)
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *PostgresOrderRepository) UpdatePaymentStatus(ctx context.Context, paymentID string, status string) error {
	query := `UPDATE orders SET payment_status = $1 WHERE payment_id = $2`
	_, err := r.db.ExecContext(ctx, query, status, paymentID)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса оплаты заказа: %w", err)
	}
	return nil
}

func (r *PostgresOrderRepository) GetAll(ctx context.Context) ([]models.Order, error) {
	query := `SELECT id, user_id, customer_name, phone, address, total_price, payment_status, payment_id, created_at 
	          FROM orders ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения всех заказов: %w", err)
	}
	defer rows.Close()

	var list []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.UserID, &o.CustomerName, &o.Phone, &o.Address, &o.TotalPrice, &o.PaymentStatus, &o.PaymentID, &o.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования заказа: %w", err)
		}
		list = append(list, o)
	}
	return list, nil
}

func (r *PostgresOrderRepository) UpdateStatusByID(ctx context.Context, orderID int, status string) error {
	query := `UPDATE orders SET payment_status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, orderID)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса заказа: %w", err)
	}
	return nil
}

type PostgresBlogPostRepository struct {
	db *sql.DB
}

func NewPostgresBlogPostRepository(db *sql.DB) *PostgresBlogPostRepository {
	return &PostgresBlogPostRepository{db: db}
}

func (r *PostgresBlogPostRepository) GetAll(ctx context.Context, publishedOnly bool) ([]models.BlogPost, error) {
	query := `SELECT id, title, subtitle, content, image_url, tag, read_time, is_published, created_at, updated_at FROM blog_posts`
	if publishedOnly {
		query += ` WHERE is_published = TRUE`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения постов блога: %w", err)
	}
	defer rows.Close()

	var list []models.BlogPost
	for rows.Next() {
		var p models.BlogPost
		if err := rows.Scan(&p.ID, &p.Title, &p.Subtitle, &p.Content, &p.ImageURL, &p.Tag, &p.ReadTime, &p.IsPublished, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("ошибка сканирования поста: %w", err)
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *PostgresBlogPostRepository) GetByID(ctx context.Context, id int) (*models.BlogPost, error) {
	query := `SELECT id, title, subtitle, content, image_url, tag, read_time, is_published, created_at, updated_at FROM blog_posts WHERE id = $1`
	var p models.BlogPost
	err := r.db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Title, &p.Subtitle, &p.Content, &p.ImageURL, &p.Tag, &p.ReadTime, &p.IsPublished, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("пост не найден")
		}
		return nil, fmt.Errorf("ошибка получения поста: %w", err)
	}
	return &p, nil
}

func (r *PostgresBlogPostRepository) Create(ctx context.Context, p *models.BlogPost) error {
	query := `INSERT INTO blog_posts (title, subtitle, content, image_url, tag, read_time, is_published)
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, p.Title, p.Subtitle, p.Content, p.ImageURL, p.Tag, p.ReadTime, p.IsPublished).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("ошибка создания поста: %w", err)
	}
	return nil
}

func (r *PostgresBlogPostRepository) Update(ctx context.Context, p *models.BlogPost) error {
	query := `UPDATE blog_posts SET title=$1, subtitle=$2, content=$3, image_url=$4, tag=$5, read_time=$6, is_published=$7, updated_at=NOW() WHERE id=$8`
	_, err := r.db.ExecContext(ctx, query, p.Title, p.Subtitle, p.Content, p.ImageURL, p.Tag, p.ReadTime, p.IsPublished, p.ID)
	if err != nil {
		return fmt.Errorf("ошибка обновления поста: %w", err)
	}
	return nil
}

func (r *PostgresBlogPostRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM blog_posts WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления поста: %w", err)
	}
	return nil
}

type PostgresReservationRepository struct {
	db *sql.DB
}

func NewPostgresReservationRepository(db *sql.DB) *PostgresReservationRepository {
	return &PostgresReservationRepository{db: db}
}

func (r *PostgresReservationRepository) Create(ctx context.Context, res *models.Reservation) error {
	if res.Status == "" {
		res.Status = "new"
	}
	query := `INSERT INTO reservations (user_id, customer_name, phone, reserve_date, reserve_time, guests_count, comment, status) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, res.UserID, res.CustomerName, res.Phone, res.ReserveDate, res.ReserveTime, res.GuestsCount, res.Comment, res.Status).Scan(&res.ID)
	if err != nil {
		return fmt.Errorf("ошибка создания бронирования: %w", err)
	}
	return nil
}

func (r *PostgresReservationRepository) GetByUserID(ctx context.Context, userID int) ([]models.Reservation, error) {
	query := `SELECT id, user_id, customer_name, phone, reserve_date, reserve_time, guests_count, comment, status 
	          FROM reservations WHERE user_id = $1 ORDER BY reserve_date DESC, reserve_time DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения бронирований: %w", err)
	}
	defer rows.Close()

	var list []models.Reservation
	for rows.Next() {
		var res models.Reservation
		var dbDate string
		var dbTime string
		err := rows.Scan(&res.ID, &res.UserID, &res.CustomerName, &res.Phone, &dbDate, &dbTime, &res.GuestsCount, &res.Comment, &res.Status)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования бронирования: %w", err)
		}
		parsedDate, parseErr := time.Parse("2006-01-02", dbDate[:10])
		if parseErr == nil {
			res.ReserveDate = parsedDate
		}
		res.ReserveTime = dbTime
		list = append(list, res)
	}
	return list, nil
}

func (r *PostgresReservationRepository) GetAll(ctx context.Context) ([]models.Reservation, error) {
	query := `SELECT id, user_id, customer_name, phone, reserve_date, reserve_time, guests_count, comment, status 
	          FROM reservations ORDER BY reserve_date DESC, reserve_time DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения всех бронирований: %w", err)
	}
	defer rows.Close()

	var list []models.Reservation
	for rows.Next() {
		var res models.Reservation
		var dbDate string
		var dbTime string
		err := rows.Scan(&res.ID, &res.UserID, &res.CustomerName, &res.Phone, &dbDate, &dbTime, &res.GuestsCount, &res.Comment, &res.Status)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования бронирования: %w", err)
		}
		parsedDate, parseErr := time.Parse("2006-01-02", dbDate[:10])
		if parseErr == nil {
			res.ReserveDate = parsedDate
		}
		res.ReserveTime = dbTime
		list = append(list, res)
	}
	return list, nil
}

func (r *PostgresReservationRepository) UpdateStatusByID(ctx context.Context, id int, status string) error {
	query := `UPDATE reservations SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса бронирования: %w", err)
	}
	return nil
}
