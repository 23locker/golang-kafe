package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"backend/internal/dto"
	"backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userID"
const userRoleKey contextKey = "userRole"

type Handler struct {
	authService        services.AuthService
	productService     services.ProductService
	orderService       services.OrderService
	reservationService services.ReservationService
	adminService       services.AdminService
	blogService        services.BlogService
	jwtSecret          string
}

func NewHandler(
	authService services.AuthService,
	productService services.ProductService,
	orderService services.OrderService,
	reservationService services.ReservationService,
	adminService services.AdminService,
	blogService services.BlogService,
	jwtSecret string,
) *Handler {
	return &Handler{
		authService:        authService,
		productService:     productService,
		orderService:       orderService,
		reservationService: reservationService,
		adminService:       adminService,
		blogService:        blogService,
		jwtSecret:          jwtSecret,
	}
}

func (h *Handler) InitRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(h.corsMiddleware)

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
		r.Get("/blog", h.getBlogPosts)
		r.Get("/blog/{id}", h.getBlogPostByID)

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
			// admin OR super_admin
			r.Group(func(r chi.Router) {
				r.Use(h.adminRequiredMiddleware)
				r.Get("/orders", h.adminGetOrders)
				r.Put("/orders/{id}/status", h.adminUpdateOrderStatus)
				r.Get("/reservations", h.adminGetReservations)
				r.Put("/reservations/{id}/status", h.adminUpdateReservationStatus)
				r.Get("/products", h.adminGetProducts)
				r.Post("/products", h.adminCreateProduct)
				r.Put("/products/{id}", h.adminUpdateProduct)
				r.Delete("/products/{id}", h.adminDeleteProduct)
				r.Post("/categories", h.adminCreateCategory)
				r.Put("/categories/{id}", h.adminUpdateCategory)
				r.Delete("/categories/{id}", h.adminDeleteCategory)
				r.Get("/stats", h.adminGetStats)
				r.Get("/blog", h.adminGetBlogPosts)
				r.Post("/blog", h.adminCreateBlogPost)
				r.Put("/blog/{id}", h.adminUpdateBlogPost)
				r.Delete("/blog/{id}", h.adminDeleteBlogPost)
				// Read-only user list available to both admin and super_admin
				r.Get("/users", h.adminGetUsers)
			})

			// super_admin only: role/user mutations and audit log
			r.Group(func(r chi.Router) {
				r.Use(h.superAdminRequiredMiddleware)
				r.Put("/users/{id}/role", h.adminUpdateUserRole)
				r.Delete("/users/{id}", h.adminDeleteUser)
				r.Get("/audit-log", h.adminGetAuditLog)
			})
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

func (h *Handler) parseJWT(tokenStr string) (jwt.MapClaims, bool) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неверный метод подписи токена")
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	return claims, ok
}

func (h *Handler) authOptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		claims, ok := h.parseJWT(cookie.Value)
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
		claims, ok := h.parseJWT(cookie.Value)
		if !ok {
			http.Error(w, "ошибка: недействительный токен сессии", http.StatusUnauthorized)
			return
		}
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "ошибка: недействительный идентификатор пользователя в токене", http.StatusUnauthorized)
			return
		}
		userRole, _ := claims["role"].(string)
		ctx := context.WithValue(r.Context(), userIDKey, int(userIDFloat))
		ctx = context.WithValue(ctx, userRoleKey, userRole)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// adminRequiredMiddleware allows access for 'admin' and 'super_admin' roles.
func (h *Handler) adminRequiredMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.authRequiredMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(userRoleKey).(string)
			if role != "admin" && role != "super_admin" {
				http.Error(w, "ошибка: доступ запрещен", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})).ServeHTTP(w, r)
	})
}

// superAdminRequiredMiddleware allows access only for 'super_admin' role.
func (h *Handler) superAdminRequiredMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.authRequiredMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(userRoleKey).(string)
			if role != "super_admin" {
				http.Error(w, "ошибка: доступ запрещен", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})).ServeHTTP(w, r)
	})
}

func (h *Handler) getAdminID(r *http.Request) *int {
	if id, ok := r.Context().Value(userIDKey).(int); ok {
		return &id
	}
	return nil
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
		Email          string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка декодирования тела запроса", http.StatusBadRequest)
		return
	}
	if err := h.authService.UpdateAddress(r.Context(), userID, req.DefaultAddress); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if req.Email != "" {
		if err := h.authService.UpdateEmail(r.Context(), userID, req.Email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "профиль обновлен"})
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
	var categoryID *int
	if s := r.URL.Query().Get("category_id"); s != "" {
		id, err := strconv.Atoi(s)
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
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
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
	phone := strings.TrimSpace(r.URL.Query().Get("phone"))
	var orderID *int
	if s := r.URL.Query().Get("order_id"); s != "" {
		if id, err := strconv.Atoi(s); err == nil {
			orderID = &id
		}
	}

	var resp []dto.OrderResponse
	var err error
	if phone != "" || orderID != nil {
		resp, err = h.adminService.GetOrdersFiltered(r.Context(), phone, orderID)
	} else {
		resp, err = h.adminService.GetOrders(r.Context())
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
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
	if err := h.adminService.UpdateOrderStatus(r.Context(), id, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	adminID := h.getAdminID(r)
	_ = h.adminService.LogAudit(r.Context(), adminID, "update_order_status", "order", &id, req.Status)

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

func (h *Handler) adminUpdateReservationStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
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
	if err := h.adminService.UpdateReservationStatus(r.Context(), id, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	adminID := h.getAdminID(r)
	_ = h.adminService.LogAudit(r.Context(), adminID, "update_reservation_status", "reservation", &id, req.Status)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "статус бронирования обновлен"})
}

func (h *Handler) adminGetProducts(w http.ResponseWriter, r *http.Request) {
	var categoryID *int
	if s := r.URL.Query().Get("category_id"); s != "" {
		id, err := strconv.Atoi(s)
		if err == nil {
			categoryID = &id
		}
	}
	resp, err := h.adminService.GetAdminProducts(r.Context(), categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) parseProductForm(r *http.Request) (dto.ProductResponse, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
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
	adminID := h.getAdminID(r)
	_ = h.adminService.LogAudit(r.Context(), adminID, "create_product", "product", &resp.ID, resp.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminUpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "неверный формат идентификатора товара", http.StatusBadRequest)
		return
	}
	req, err := h.parseProductForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.ImageURL == "" {
		oldProduct, err := h.adminService.GetAdminProductByID(r.Context(), id)
		if err == nil {
			req.ImageURL = oldProduct.ImageURL
		}
	} else {
		oldProduct, err := h.adminService.GetAdminProductByID(r.Context(), id)
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
	adminID := h.getAdminID(r)
	_ = h.adminService.LogAudit(r.Context(), adminID, "update_product", "product", &id, resp.Name)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminDeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "неверный формат идентификатора товара", http.StatusBadRequest)
		return
	}

	oldProduct, _ := h.adminService.GetAdminProductByID(r.Context(), id)

	softDeleted, err := h.adminService.DeleteProduct(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !softDeleted && oldProduct.ImageURL != "" && strings.HasPrefix(oldProduct.ImageURL, "/uploads/") {
		workDir, _ := os.Getwd()
		oldFilePath := filepath.Join(workDir, "uploads", strings.TrimPrefix(oldProduct.ImageURL, "/uploads/"))
		os.Remove(oldFilePath)
	}

	action := "delete_product"
	msg := "товар успешно удален"
	if softDeleted {
		action = "soft_delete_product"
		msg = "товар скрыт из меню (мягкое удаление)"
	}
	adminID := h.getAdminID(r)
	_ = h.adminService.LogAudit(r.Context(), adminID, action, "product", &id, msg)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       msg,
		"soft_deleted": softDeleted,
	})
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
	adminID := h.getAdminID(r)
	_ = h.adminService.LogAudit(r.Context(), adminID, "create_category", "category", &resp.ID, resp.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
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
	adminID := h.getAdminID(r)
	_ = h.adminService.LogAudit(r.Context(), adminID, "update_category", "category", &id, resp.Name)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "неверный формат идентификатора категории", http.StatusBadRequest)
		return
	}
	if err := h.adminService.DeleteCategory(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	adminID := h.getAdminID(r)
	_ = h.adminService.LogAudit(r.Context(), adminID, "delete_category", "category", &id, "")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "категория успешно удалена"})
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

func (h *Handler) getBlogPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.blogService.GetPosts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(posts)
}

func (h *Handler) getBlogPostByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "неверный формат идентификатора", http.StatusBadRequest)
		return
	}
	post, err := h.blogService.GetPostByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(post)
}

func (h *Handler) parseBlogPostForm(r *http.Request) (dto.BlogPostResponse, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return dto.BlogPostResponse{}, fmt.Errorf("ошибка парсинга формы: %w", err)
	}
	var req dto.BlogPostResponse
	req.Title = r.FormValue("title")
	req.Subtitle = r.FormValue("subtitle")
	req.Content = r.FormValue("content")
	req.Tag = r.FormValue("tag")
	req.ReadTime = r.FormValue("read_time")
	req.IsPublished = r.FormValue("is_published") == "true"

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

func (h *Handler) adminGetBlogPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.adminService.GetAllBlogPosts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(posts)
}

func (h *Handler) adminCreateBlogPost(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseBlogPostForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := h.adminService.CreateBlogPost(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	adminID := h.getAdminID(r)
	_ = h.adminService.LogAudit(r.Context(), adminID, "create_blog_post", "blog_post", &resp.ID, resp.Title)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminUpdateBlogPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "неверный формат идентификатора", http.StatusBadRequest)
		return
	}
	req, err := h.parseBlogPostForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.ImageURL == "" {
		existing, err := h.blogService.GetPostByID(r.Context(), id)
		if err == nil {
			req.ImageURL = existing.ImageURL
		}
	} else {
		existing, err := h.blogService.GetPostByID(r.Context(), id)
		if err == nil && existing.ImageURL != "" && strings.HasPrefix(existing.ImageURL, "/uploads/") {
			workDir, _ := os.Getwd()
			oldPath := filepath.Join(workDir, "uploads", strings.TrimPrefix(existing.ImageURL, "/uploads/"))
			os.Remove(oldPath)
		}
	}
	resp, err := h.adminService.UpdateBlogPost(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	adminID := h.getAdminID(r)
	_ = h.adminService.LogAudit(r.Context(), adminID, "update_blog_post", "blog_post", &id, resp.Title)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminDeleteBlogPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "неверный формат идентификатора", http.StatusBadRequest)
		return
	}
	existing, err := h.blogService.GetPostByID(r.Context(), id)
	if err == nil && existing.ImageURL != "" && strings.HasPrefix(existing.ImageURL, "/uploads/") {
		workDir, _ := os.Getwd()
		oldPath := filepath.Join(workDir, "uploads", strings.TrimPrefix(existing.ImageURL, "/uploads/"))
		os.Remove(oldPath)
	}
	if err := h.adminService.DeleteBlogPost(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	adminID := h.getAdminID(r)
	_ = h.adminService.LogAudit(r.Context(), adminID, "delete_blog_post", "blog_post", &id, "")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "статья успешно удалена"})
}

// User management handlers (super_admin only)

func (h *Handler) adminGetUsers(w http.ResponseWriter, r *http.Request) {
	resp, err := h.adminService.GetUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) adminUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	targetID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "неверный формат идентификатора пользователя", http.StatusBadRequest)
		return
	}
	var req dto.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка декодирования тела запроса", http.StatusBadRequest)
		return
	}
	currentAdminID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		http.Error(w, "ошибка: неавторизован", http.StatusUnauthorized)
		return
	}
	callerRole, _ := r.Context().Value(userRoleKey).(string)
	// ADMIN role cannot grant or set super_admin
	if callerRole == "admin" && req.Role == "super_admin" {
		http.Error(w, "доступ запрещен: нельзя назначить роль суперадминистратора", http.StatusForbidden)
		return
	}
	if err := h.adminService.UpdateUserRole(r.Context(), targetID, currentAdminID, req.Role); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	adminID := h.getAdminID(r)
	details := fmt.Sprintf("new_role=%s", req.Role)
	_ = h.adminService.LogAudit(r.Context(), adminID, "update_user_role", "user", &targetID, details)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "роль пользователя обновлена"})
}

func (h *Handler) adminDeleteUser(w http.ResponseWriter, r *http.Request) {
	targetID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "неверный формат идентификатора пользователя", http.StatusBadRequest)
		return
	}
	currentAdminID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		http.Error(w, "ошибка: неавторизован", http.StatusUnauthorized)
		return
	}
	if err := h.adminService.DeleteUser(r.Context(), targetID, currentAdminID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	adminID := h.getAdminID(r)
	_ = h.adminService.LogAudit(r.Context(), adminID, "delete_user", "user", &targetID, "")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "пользователь удален"})
}

func (h *Handler) adminGetAuditLog(w http.ResponseWriter, r *http.Request) {
	resp, err := h.adminService.GetAuditLog(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
