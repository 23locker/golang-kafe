package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"backend/internal/config"
	"backend/internal/handlers"
	"backend/internal/models"
	"backend/internal/repositories"
	"backend/internal/services"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("ошибка загрузки конфигурации: %v", err)
	}

	var db *sql.DB
	for i := 0; i < 15; i++ {
		db, err = sql.Open("pgx", cfg.DatabaseURL)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("ожидание подключения к базе данных... попытка %d", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("ошибка подключения к базе данных после нескольких попыток: %v", err)
	}
	defer db.Close()

	migrationData, err := os.ReadFile("migrations/schema.sql")
	if err != nil {
		log.Fatalf("ошибка чтения файла миграции: %v", err)
	}
	if _, err = db.Exec(string(migrationData)); err != nil {
		log.Fatalf("ошибка применения миграции базы данных: %v", err)
	}
	log.Println("база данных успешно инициализирована")

	userRepo := repositories.NewPostgresUserRepository(db)
	productRepo := repositories.NewPostgresProductRepository(db)
	orderRepo := repositories.NewPostgresOrderRepository(db)
	resRepo := repositories.NewPostgresReservationRepository(db)
	blogRepo := repositories.NewPostgresBlogPostRepository(db)
	auditRepo := repositories.NewPostgresAuditLogRepository(db)

	authServ := services.NewAuthService(userRepo, cfg.JWTSecret)
	prodServ := services.NewProductService(productRepo)
	orderServ := services.NewOrderService(orderRepo, productRepo)
	resServ := services.NewReservationService(resRepo)
	blogServ := services.NewBlogService(blogRepo)
	adminServ := services.NewAdminService(orderRepo, resRepo, productRepo, blogRepo, userRepo, auditRepo)

	ctx := context.Background()

	defaultAccounts := []struct {
		name  string
		phone string
		role  string
	}{
		{"Суперадминистратор", "+79991234567", "super_admin"},
		{"Главный Администратор", "+7988548955", "super_admin"},
	}
	for _, a := range defaultAccounts {
		if _, err := userRepo.GetByPhone(ctx, a.phone); err != nil {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
			admin := &models.User{
				Name:         a.name,
				Phone:        a.phone,
				PasswordHash: string(hashedPassword),
				Role:         a.role,
			}
			if err := userRepo.Create(ctx, admin); err != nil {
				log.Printf("ошибка создания аккаунта %s: %v", a.phone, err)
			} else {
				log.Printf("создан аккаунт: %s (%s)", a.name, a.phone)
			}
		}
	}

	h := handlers.NewHandler(authServ, prodServ, orderServ, resServ, adminServ, blogServ, cfg.JWTSecret)

	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatalf("ошибка создания директории uploads: %v", err)
	}

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: h.InitRoutes(),
	}

	log.Printf("сервер успешно запущен на порту %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ошибка запуска сервера: %v", err)
	}
}
