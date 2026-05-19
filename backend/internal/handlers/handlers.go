package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"backend/internal/dto"
	"backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type contextKey string

const userIDKey contextKey = "userID"

type Handler struct {
	authService        services.AuthService
	productService     services.ProductService
	orderService       services.OrderService
	reservationService services.ReservationService
	adminService       services.AdminService
	jwtSecret          string
}

func NewHandler(
	authService services.AuthService,
	productService services.ProductService,
	orderService services.OrderService,
	reservationService services.ReservationService,
	adminService services.AdminService,
	jwtSecret string,
) *Handler {
	return &Handler{
		authService:        authService,
		productService:     productService,
		orderService:       orderService,
		reservationService: reservationService,
		adminService:       adminService,
		jwtSecret:          jwtSecret,
	}
}

func (h *Handler) InitRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(h.corsMiddleware)

	// Статика для загруженных файлов
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "uploads"))
	FileServer(r, "/uploads", filesDir)

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.register)
			r.Post("/login", h.login)
			r.Post("/logout", h.logout)

			r.Group(func(r chi.Router) {
				r.Use(h.authRequiredMiddleware)
				r.Get("/profile", h.profile)
				r.Put("/profile", h.updateProfile)
			})
		})

		r.Get("/categories", h.getCategories)
		r.Get("/products", h.getProducts)
		r.Get("/products/{id}", h.getProductByID)

		r.Group(func(r chi.Router) {
			r.Use(h.authOptionalMiddleware)
			r.Post("/orders", h.createOrder)
			r.Post("/reservations", h.createReservation)
		})

		r.Group(func(r chi.Router) {
			r.Use(h.authRequiredMiddleware)
			r.Get("/orders", h.getMyOrders)
			r.Get("/reservations", h.getMyReservations)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Get("/orders", h.adminGetOrders)
			r.Put("/orders/{id}/status", h.adminUpdateOrderStatus)
			r.Get("/reservations", h.adminGetReservations)
			r.Put("/reservations/{id}/status", h.adminUpdateReservationStatus)
			r.Post("/products", h.adminCreateProduct)
			r.Put("/products/{id}", h.adminUpdateProduct)
			r.Delete("/products/{id}", h.adminDeleteProduct)
			r.Post("/categories", h.adminCreateCategory)
			r.Put("/categories/{id}", h.adminUpdateCategory)
			r.Delete("/categories/{id}", h.adminDeleteCategory)
			r.Get("/stats", h.adminGetStats)
		})
	})

	return r
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"
	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func (h *Handler) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) authOptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неверный метод подписи токена")
			}
			return []byte(h.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			next.ServeHTTP(w, r)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, int(userIDFloat))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) authRequiredMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "ошибка: неавторизован", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неверный метод подписи токена")
			}
			return []byte(h.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "ошибка: недействительный токен сессии", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "ошибка: недействительные данные сессии", http.StatusUnauthorized)
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "ошибка: недействительный идентификатор пользователя в токене", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, int(userIDFloat))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка декодирования тела запроса", http.StatusBadRequest)
		return
	}

	resp, err := h.authService.Register(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка декодирования тела запроса", http.StatusBadRequest)
		return
	}

	tokenStr, err := h.authService.Login(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    tokenStr,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "успешный вход"})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Secure:   false,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "успешный выход"})
}

func (h *Handler) profile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		http.Error(w, "ошибка: неавторизован", http.StatusUnauthorized)
		return
	}

	resp, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		http.Error(w, "ошибка: неавторизован", http.StatusUnauthorized)
		return
	}

	var req struct {
		DefaultAddress string `json:"default_address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка декодирования тела запроса", http.StatusBadRequest)
		return
	}

	err := h.authService.UpdateAddress(r.Context(), userID, req.DefaultAddress)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "адрес обновлен"})
}

func (h *Handler) getCategories(w http.ResponseWriter, r *http.Request) {
	resp, err := h.productService.GetCategories(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) getProducts(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := r.URL.Query().Get("category_id")
	var categoryID *int

	if categoryIDStr != "" {
		id, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			http.Error(w, "неверный формат category_id", http.StatusBadRequest)
			return
		}
		categoryID = &id
	}

	resp, err := h.productService.GetProducts(r.Context(), categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) getProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "неверный формат идентификатора товара", http.StatusBadRequest)
		return
	}

	resp, err := h.productService.GetProductByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка декодирования тела запроса", http.StatusBadRequest)
		return
	}

	var userID *int
	if id, ok := r.Context().Value(userIDKey).(int); ok {
		userID = &id
	}

	resp, err := h.orderService.CreateOrder(r.Context(), userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) getMyOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		http.Error(w, "ошибка: неавторизован", http.StatusUnauthorized)
		return
	}

	resp, err := h.orderService.GetOrdersByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) createReservation(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка декодирования тела запроса", http.StatusBadRequest)
		return
	}

	var userID *int
	if id, ok := r.Context().Value(userIDKey).(int); ok {
		userID = &id
	}

	resp, err := h.reservationService.CreateReservation(r.Context(), userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) getMyReservations(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		http.Error(w, "ошибка: неавторизован", http.StatusUnauthorized)
		return
	}

	resp, err := h.reservationService.GetReservationsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminGetOrders(w http.ResponseWriter, r *http.Request) {
	resp, err := h.adminService.GetOrders(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "неверный формат идентификатора заказа", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка декодирования тела запроса", http.StatusBadRequest)
		return
	}

	err = h.adminService.UpdateOrderStatus(r.Context(), id, req.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "статус заказа обновлен"})
}

func (h *Handler) adminGetReservations(w http.ResponseWriter, r *http.Request) {
	resp, err := h.adminService.GetReservations(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) parseProductForm(r *http.Request) (dto.ProductResponse, error) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		return dto.ProductResponse{}, fmt.Errorf("ошибка парсинга формы: %w", err)
	}

	var req dto.ProductResponse
	req.Name = r.FormValue("name")
	req.Description = r.FormValue("description")
	req.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
	
	if catIDStr := r.FormValue("category_id"); catIDStr != "" {
		if catID, err := strconv.Atoi(catIDStr); err == nil {
			req.CategoryID = &catID
		}
	}
	
	req.Weight, _ = strconv.Atoi(r.FormValue("weight"))
	req.Calories, _ = strconv.Atoi(r.FormValue("calories"))
	req.IsAvailable = r.FormValue("is_available") == "true"

	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		ext := filepath.Ext(header.Filename)
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		
		workDir, _ := os.Getwd()
		uploadsDir := filepath.Join(workDir, "uploads")
		os.MkdirAll(uploadsDir, 0755)
		
		dstPath := filepath.Join(uploadsDir, filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			return req, fmt.Errorf("ошибка сохранения файла: %w", err)
		}
		defer dst.Close()
		
		if _, err := io.Copy(dst, file); err != nil {
			return req, fmt.Errorf("ошибка записи файла: %w", err)
		}
		
		req.ImageURL = "/uploads/" + filename
	}

	return req, nil
}

func (h *Handler) adminCreateProduct(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseProductForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.adminService.CreateProduct(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminUpdateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "неверный формат идентификатора товара", http.StatusBadRequest)
		return
	}

	req, err := h.parseProductForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Если картинка не была загружена, мы должны сохранить старую
	if req.ImageURL == "" {
		oldProduct, err := h.productService.GetProductByID(r.Context(), id)
		if err == nil {
			req.ImageURL = oldProduct.ImageURL
		}
	} else {
		// Удаляем старую картинку, если была загружена новая
		oldProduct, err := h.productService.GetProductByID(r.Context(), id)
		if err == nil && oldProduct.ImageURL != "" && strings.HasPrefix(oldProduct.ImageURL, "/uploads/") {
			workDir, _ := os.Getwd()
			oldFilePath := filepath.Join(workDir, "uploads", strings.TrimPrefix(oldProduct.ImageURL, "/uploads/"))
			os.Remove(oldFilePath)
		}
	}

	resp, err := h.adminService.UpdateProduct(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminDeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "неверный формат идентификатора товара", http.StatusBadRequest)
		return
	}

	// Удаляем картинку
	oldProduct, err := h.productService.GetProductByID(r.Context(), id)
	if err == nil && oldProduct.ImageURL != "" && strings.HasPrefix(oldProduct.ImageURL, "/uploads/") {
		workDir, _ := os.Getwd()
		oldFilePath := filepath.Join(workDir, "uploads", strings.TrimPrefix(oldProduct.ImageURL, "/uploads/"))
		os.Remove(oldFilePath)
	}

	err = h.adminService.DeleteProduct(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "товар успешно удален"})
}

func (h *Handler) adminGetStats(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	resp, err := h.adminService.GetStats(r.Context(), startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminUpdateReservationStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "неверный формат идентификатора бронирования", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка декодирования тела запроса", http.StatusBadRequest)
		return
	}

	err = h.adminService.UpdateReservationStatus(r.Context(), id, req.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "статус бронирования обновлен"})
}

func (h *Handler) adminCreateCategory(w http.ResponseWriter, r *http.Request) {
	var req dto.CategoryResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка декодирования тела запроса", http.StatusBadRequest)
		return
	}

	resp, err := h.adminService.CreateCategory(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminUpdateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "неверный формат идентификатора категории", http.StatusBadRequest)
		return
	}

	var req dto.CategoryResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка декодирования тела запроса", http.StatusBadRequest)
		return
	}

	resp, err := h.adminService.UpdateCategory(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminDeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "неверный формат идентификатора категории", http.StatusBadRequest)
		return
	}

	err = h.adminService.DeleteCategory(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "категория успешно удалена"})
}
