<script lang="ts">
    import { fade, fly, scale } from "svelte/transition";
    import { onMount } from "svelte";
    import {
        MapPin,
        ShoppingCart,
        ChevronLeft,
        ChevronRight,
        Minus,
        Plus,
        X,
        Phone,
        Mail,
        Send,
        User,
        Calendar,
        Clock,
        Home,
        TrendingUp,
        Layers,
        Trash2,
        PlusCircle,
    } from "@lucide/svelte";
    import Instagram from "./lib/icons/Instagram.svelte";
    import Facebook from "./lib/icons/Facebook.svelte";
    import MenuCategory from "./lib/MenuCategory.svelte";
    import MenuCard from "./lib/MenuCard.svelte";
    import type { Dish } from "./lib/constants";

    interface Category {
        id: number;
        name: string;
        slug: string;
    }

    interface Product {
        id: number;
        name: string;
        description: string;
        price: number;
        category_id: number;
        image_url: string;
        weight: number;
        calories: number;
        is_available: boolean;
    }

    interface UserInfo {
        id: number;
        name: string;
        phone: string;
        email: string | null;
        default_address: string | null;
        role: string;
    }

    interface OrderItem {
        product_id: number;
        product_name?: string;
        quantity: number;
        price?: number;
    }

    interface Order {
        id: number;
        customer_name: string;
        phone: string;
        address: string;
        total_price: number;
        payment_status: string;
        created_at: string;
        items: OrderItem[];
    }

    interface Reservation {
        id: number;
        customer_name: string;
        phone: string;
        reserve_date: string;
        reserve_time: string;
        guests_count: number;
        comment: string;
        status: string;
    }

    // Navigation state: 'home' | 'admin' | 'menu' | 'blog' | 'about'
    let currentView = $state<"home" | "admin" | "menu" | "blog" | "about">(
        "home",
    );
    let adminTab = $state<
        | "stats"
        | "orders"
        | "reservations"
        | "menu"
        | "categories"
        | "blog"
        | "users"
        | "audit-log"
    >("stats");

    // Booking date boundaries
    let minDate = $state("");
    let maxDate = $state("");

    // Category Form states
    let isCatFormOpen = $state(false);
    let editingCatId = $state<number | null>(null);
    let catFormName = $state("");
    let catFormSlug = $state("");

    // Orders filtering and sorting states
    let orderFilterStatus = $state("all");
    let orderSortField = $state<"id" | "date" | "total">("date");
    let orderSortDirection = $state<"asc" | "desc">("desc");

    // API dynamic states
    let dishes = $state<Dish[]>([]);
    let categories = $state<Category[]>([]);
    let productsList = $state<Product[]>([]);

    // Auth state
    let currentUser = $state<UserInfo | null>(null);

    interface ApiBlogPost {
        id: number;
        title: string;
        subtitle: string;
        content: string;
        image_url: string;
        tag: string;
        read_time: string;
        is_published: boolean;
        created_at: string;
        updated_at: string;
    }

    let blogPosts = $state<ApiBlogPost[]>([]);
    let adminBlogPosts = $state<ApiBlogPost[]>([]);

    let selectedBlogPost = $state<ApiBlogPost | null>(null);

    // Blog admin form state
    let isBlogFormOpen = $state(false);
    let editingBlogId = $state<number | null>(null);
    let blogFormTitle = $state("");
    let blogFormSubtitle = $state("");
    let blogFormContent = $state("");
    let blogFormTag = $state("");
    let blogFormReadTime = $state("");
    let blogFormImage = $state("");
    let blogFormImageFile = $state<File | null>(null);
    let blogFormIsPublished = $state(true);

    function formatBlogDate(isoDate: string): string {
        return new Date(isoDate).toLocaleDateString("ru-RU", {
            day: "numeric",
            month: "long",
            year: "numeric",
        });
    }

    function localizeOrderStatus(status: string): string {
        const map: Record<string, string> = {
            pending: "Ожидает оплаты",
            awaiting_delivery: "Ожидает доставки",
            preparing: "Готовится",
            delivering: "В пути",
            delivered: "Доставлен",
            cancelled: "Отменён",
        };
        return map[status] ?? status;
    }

    function isAdminRole(user: UserInfo | null): boolean {
        return user?.role === "super_admin" || user?.role === "admin";
    }

    // Первые 4 доступных блюда для секции "Популярное"
    let popularDishes = $derived(dishes.slice(0, 4));

    function isValidPhone(phone: string): boolean {
        const digits = phone.replace(/\D/g, "");
        return digits.length >= 10 && digits.length <= 12;
    }

    let isAuthModalOpen = $state(false);
    let authMode = $state<"login" | "register">("login");
    let authName = $state("");
    let authPhone = $state("");
    let authEmail = $state("");
    let authPassword = $state("");
    let authError = $state("");

    // Profile and History Drawer
    let isProfileOpen = $state(false);
    let userOrders = $state<Order[]>([]);
    let userReservations = $state<Reservation[]>([]);
    let editAddress = $state("");
    let editEmail = $state("");
    let profileMessage = $state("");

    // Cart & Order state
    let cart = $state<{ dishId: string; quantity: number }[]>([]);
    let showAgeVerification = $state(true);
    let isCartOpen = $state(false);
    let showCheckoutForm = $state(false);
    let deliveryMode = $state<"delivery" | "bar">("delivery");
    let checkoutName = $state("");
    let checkoutPhone = $state("");
    let checkoutAddress = $state("");
    let checkoutPayment = $state<"cash" | "online">("cash");
    let orderSuccessMsg = $state("");

    // Reservation state
    let reserveName = $state("");
    let reservePhone = $state("");
    let reserveDate = $state("");
    let reserveTime = $state("");
    let reserveGuests = $state(2);
    let reserveComment = $state("");
    let reserveSuccessMsg = $state("");

    // Admin Data states
    let adminOrders = $state<Order[]>([]);
    let adminReservations = $state<Reservation[]>([]);
    let adminStats = $state<{
        total_orders: number;
        total_reservations: number;
        total_products: number;
        total_revenue: number;
    } | null>(null);
    let statsStartDate = $state("");
    let statsEndDate = $state("");

    interface AdminUser {
        id: number;
        name: string;
        phone: string;
        email: string | null;
        default_address: string | null;
        role: string;
        created_at: string;
    }

    interface AuditLogEntry {
        id: number;
        admin_id: number | null;
        action: string;
        entity_type: string;
        entity_id: number | null;
        details: string;
        created_at: string;
    }

    let adminUsers = $state<AdminUser[]>([]);
    let userSearchPhone = $state("");
    let auditLog = $state<AuditLogEntry[]>([]);
    let orderSearchPhone = $state("");
    let orderSearchId = $state("");
    let livePolling = $state(false);
    let selectedDish = $state<Dish | null>(null);
    // plain (non-reactive) ref — must not be proxied by Svelte
    let _pollingInterval: ReturnType<typeof setInterval> | null = null;

    // Admin Product Form states
    let isProdFormOpen = $state(false);
    let editingProdId = $state<number | null>(null);
    let prodFormName = $state("");
    let prodFormDesc = $state("");
    let prodFormPrice = $state(0);
    let prodFormCategory = $state(1);
    let prodFormImage = $state("");
    let prodFormImageFile = $state<File | null>(null);
    let prodFormWeight = $state(0);
    let prodFormCalories = $state(0);
    let prodFormIsAvailable = $state(true);

    // Global notification
    let globalAlert = $state("");

    let isMounted = $state(false);

    // Load menu products and categories
    async function fetchMenu() {
        try {
            const catRes = await fetch("/api/categories");
            if (catRes.ok) {
                categories = await catRes.json();
            }

            const prodRes = await fetch("/api/products");
            if (prodRes.ok) {
                productsList = await prodRes.json();
                dishes = productsList
                    .filter((p) => p.is_available)
                    .map((p) => {
                        const cat = categories.find((c) => c.id === p.category_id);
                        return {
                            id: String(p.id),
                            name: p.name,
                            description: p.description,
                            price: p.price,
                            image: p.image_url || "/images/placeholder.jpg",
                            category: (cat?.slug || "main") as any,
                            categoryName: cat?.name,
                            calories: p.calories || undefined,
                            weight: p.weight || undefined,
                        };
                    });
            }
        } catch (e) {
            console.error("Failed to load menu data:", e);
        }
    }

    async function fetchBlogPosts() {
        try {
            const res = await fetch("/api/blog");
            if (res.ok) {
                blogPosts = await res.json();
            }
        } catch (e) {
            console.error("Failed to load blog posts:", e);
        }
    }

    async function fetchAdminBlogPosts() {
        try {
            const res = await fetch("/api/admin/blog");
            if (res.ok) {
                adminBlogPosts = await res.json();
            }
        } catch (e) {
            console.error(e);
        }
    }

    function openCreateBlogForm() {
        editingBlogId = null;
        blogFormTitle = "";
        blogFormSubtitle = "";
        blogFormContent = "";
        blogFormTag = "";
        blogFormReadTime = "";
        blogFormImage = "";
        blogFormImageFile = null;
        blogFormIsPublished = true;
        isBlogFormOpen = true;
    }

    function openEditBlogForm(post: ApiBlogPost) {
        editingBlogId = post.id;
        blogFormTitle = post.title;
        blogFormSubtitle = post.subtitle;
        blogFormContent = post.content;
        blogFormTag = post.tag;
        blogFormReadTime = post.read_time;
        blogFormImage = post.image_url;
        blogFormImageFile = null;
        blogFormIsPublished = post.is_published;
        isBlogFormOpen = true;
    }

    async function handleSaveBlogPost() {
        if (!blogFormTitle) {
            alert("Укажите заголовок статьи");
            return;
        }
        const formData = new FormData();
        formData.append("title", blogFormTitle);
        formData.append("subtitle", blogFormSubtitle);
        formData.append("content", blogFormContent);
        formData.append("tag", blogFormTag);
        formData.append("read_time", blogFormReadTime);
        formData.append("is_published", blogFormIsPublished.toString());
        if (blogFormImageFile) {
            formData.append("image", blogFormImageFile);
        }
        try {
            const url = editingBlogId
                ? `/api/admin/blog/${editingBlogId}`
                : "/api/admin/blog";
            const method = editingBlogId ? "PUT" : "POST";
            const res = await fetch(url, { method, body: formData });
            if (res.ok) {
                isBlogFormOpen = false;
                fetchBlogPosts();
                fetchAdminBlogPosts();
            } else {
                const err = await res.text();
                alert("Ошибка при сохранении статьи: " + err);
            }
        } catch (e) {
            alert("Ошибка соединения с сервером");
        }
    }

    async function handleDeleteBlogPost(id: number) {
        if (!confirm("Вы действительно хотите удалить эту статью?")) return;
        try {
            const res = await fetch(`/api/admin/blog/${id}`, {
                method: "DELETE",
            });
            if (res.ok) {
                fetchBlogPosts();
                fetchAdminBlogPosts();
            } else {
                const err = await res.text();
                alert("Ошибка при удалении статьи: " + err);
            }
        } catch (e) {
            alert("Ошибка соединения");
        }
    }

    // Load current user profile.
    // Reads sessionStorage immediately (no latency), then verifies with the server.
    async function fetchProfile() {
        // Restore from cache first so the navbar renders instantly
        const cached = sessionStorage.getItem("currentUser");
        if (cached && !currentUser) {
            try {
                currentUser = JSON.parse(cached);
            } catch {}
        }

        try {
            const res = await fetch("/api/auth/profile");
            if (res.ok) {
                currentUser = await res.json();
                if (currentUser) {
                    sessionStorage.setItem("currentUser", JSON.stringify(currentUser));
                    checkoutName = currentUser.name;
                    checkoutPhone = currentUser.phone;
                    checkoutAddress = currentUser.default_address || "";
                    editAddress = currentUser.default_address || "";
                    editEmail = currentUser.email || "";
                }
            } else {
                currentUser = null;
                sessionStorage.removeItem("currentUser");
            }
        } catch (e) {
            // Keep cached value on network error so UI doesn't flicker
        }
    }

    // Автоматически подставляем данные пользователя в форму бронирования
    $effect(() => {
        if (currentUser) {
            if (!reserveName) reserveName = currentUser.name;
            if (!reservePhone) reservePhone = currentUser.phone;
        }
    });

    // Fetch orders and reservations for the current user
    async function fetchUserHistory() {
        if (!currentUser) return;
        try {
            const ordRes = await fetch("/api/orders");
            if (ordRes.ok) {
                userOrders = await ordRes.json();
            }
            const resRes = await fetch("/api/reservations");
            if (resRes.ok) {
                userReservations = await resRes.json();
            }
        } catch (e) {
            console.error(e);
        }
    }

    // Fetch admin dashboard details
    async function fetchAdminData() {
        try {
            const ordRes = await fetch("/api/admin/orders");
            if (ordRes.ok) {
                adminOrders = await ordRes.json();
            }
            const resRes = await fetch("/api/admin/reservations");
            if (resRes.ok) {
                adminReservations = await resRes.json();
            }
            if (isAdminRole(currentUser)) {
                let statsUrl = "/api/admin/stats";
                if (statsStartDate && statsEndDate) {
                    statsUrl += `?start_date=${statsStartDate}&end_date=${statsEndDate}`;
                }
                const statRes = await fetch(statsUrl);
                if (statRes.ok) {
                    adminStats = await statRes.json();
                }
            }
            // Both admin and super_admin can read users list
            const usersRes = await fetch("/api/admin/users");
            if (usersRes.ok) {
                adminUsers = await usersRes.json();
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function fetchAdminOrders() {
        try {
            const params = new URLSearchParams();
            if (orderSearchPhone.trim())
                params.set("phone", orderSearchPhone.trim());
            if (orderSearchId.trim())
                params.set("order_id", orderSearchId.trim());
            const url =
                "/api/admin/orders" +
                (params.toString() ? "?" + params.toString() : "");
            const res = await fetch(url);
            if (res.ok) {
                adminOrders = await res.json();
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function fetchAuditLog() {
        try {
            const res = await fetch("/api/admin/audit-log");
            if (res.ok) {
                auditLog = await res.json();
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function handleUpdateUserRole(userId: number, newRole: string) {
        try {
            const res = await fetch(`/api/admin/users/${userId}/role`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ role: newRole }),
            });
            if (res.ok) {
                const usersRes = await fetch("/api/admin/users");
                if (usersRes.ok) adminUsers = await usersRes.json();
            } else {
                const err = await res.text();
                alert("Ошибка: " + err);
            }
        } catch (e) {
            alert("Ошибка соединения");
        }
    }

    async function handleDeleteUser(userId: number) {
        if (!confirm("Вы уверены, что хотите удалить этого пользователя?"))
            return;
        try {
            const res = await fetch(`/api/admin/users/${userId}`, {
                method: "DELETE",
            });
            if (res.ok) {
                adminUsers = adminUsers.filter((u) => u.id !== userId);
            } else {
                const err = await res.text();
                alert("Ошибка: " + err);
            }
        } catch (e) {
            alert("Ошибка соединения");
        }
    }

    function startOrdersPolling() {
        if (_pollingInterval !== null) return;
        livePolling = true;
        _pollingInterval = setInterval(fetchAdminOrders, 5000);
    }

    function stopOrdersPolling() {
        if (_pollingInterval !== null) {
            clearInterval(_pollingInterval);
            _pollingInterval = null;
        }
        livePolling = false;
    }

    function openDishModal(dish: Dish) {
        selectedDish = dish;
    }

    function closeDishModal() {
        selectedDish = null;
    }

    let filteredAdminUsers = $derived(
        userSearchPhone.trim()
            ? adminUsers.filter((u) => u.phone.includes(userSearchPhone.trim()))
            : adminUsers,
    );

    onMount(async () => {
        isMounted = true;
        fetchMenu();
        fetchBlogPosts();
        await fetchProfile();

        // Set reservation date boundaries (today to today + 1 month)
        const todayObj = new Date();
        minDate = todayObj.toISOString().split("T")[0];

        const maxDateObj = new Date();
        maxDateObj.setMonth(maxDateObj.getMonth() + 1);
        maxDate = maxDateObj.toISOString().split("T")[0];

        // Check hash for routing
        const handleHash = () => {
            if (window.location.hash === "#admin") {
                if (!isAdminRole(currentUser)) {
                    window.location.hash = "";
                    currentView = "home";
                    stopOrdersPolling();
                    return;
                }
                currentView = "admin";
                fetchAdminData();
                startOrdersPolling();
            } else {
                if (currentView === "admin") stopOrdersPolling();
                if (window.location.hash === "#menu") {
                    currentView = "menu";
                } else if (window.location.hash === "#blog") {
                    currentView = "blog";
                } else if (window.location.hash === "#about") {
                    currentView = "about";
                } else {
                    currentView = "home";
                }
            }
        };
        window.addEventListener("hashchange", handleHash);
        handleHash();
    });

    // Reactive Derived Values
    let cartTotal = $derived(
        cart.reduce((acc, item) => {
            const dish = dishes.find((d) => d.id === item.dishId);
            return acc + (dish?.price || 0) * item.quantity;
        }, 0),
    );

    let cartItemsCount = $derived(
        cart.reduce((acc, item) => acc + item.quantity, 0),
    );

    let filteredOrders = $derived.by(() => {
        let list = [...adminOrders];
        if (orderFilterStatus !== "all") {
            list = list.filter((o) => o.payment_status === orderFilterStatus);
        }
        list.sort((a, b) => {
            let comparison = 0;
            if (orderSortField === "id") {
                comparison = a.id - b.id;
            } else if (orderSortField === "date") {
                comparison =
                    new Date(a.created_at).getTime() -
                    new Date(b.created_at).getTime();
            } else if (orderSortField === "total") {
                comparison = a.total_price - b.total_price;
            }
            return orderSortDirection === "desc" ? -comparison : comparison;
        });
        return list;
    });

    // Functions
    function addToCart(dishId: string) {
        const existing = cart.find((item) => item.dishId === dishId);
        if (existing) {
            existing.quantity += 1;
        } else {
            cart.push({ dishId, quantity: 1 });
        }
    }

    function removeFromCart(dishId: string) {
        const idx = cart.findIndex((item) => item.dishId === dishId);
        if (idx !== -1) {
            if (cart[idx].quantity === 1) {
                cart.splice(idx, 1);
            } else {
                cart[idx].quantity -= 1;
            }
        }
    }

    function getQuantity(dishId: string) {
        return cart.find((item) => item.dishId === dishId)?.quantity || 0;
    }

    // Order Placement
    async function handlePlaceOrder() {
        if (!checkoutName || !checkoutPhone || !checkoutAddress) {
            alert("Пожалуйста, заполните все обязательные поля");
            return;
        }
        if (!isValidPhone(checkoutPhone)) {
            alert(
                "Неверный формат номера телефона. Используйте формат +7XXXXXXXXXX",
            );
            return;
        }

        try {
            const items = cart.map((i) => ({
                product_id: parseInt(i.dishId),
                quantity: i.quantity,
            }));

            const res = await fetch("/api/orders", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    customer_name: checkoutName,
                    phone: checkoutPhone,
                    address: checkoutAddress,
                    payment_method: checkoutPayment,
                    items: items,
                }),
            });

            if (res.ok) {
                const orderData = await res.json();
                cart = [];
                showCheckoutForm = false;
                isCartOpen = false;

                if (checkoutPayment === "online") {
                    orderSuccessMsg = `Заказ успешно оформлен! Инициализирован онлайн-платеж ЮKassa. ID транзакции: ${orderData.payment_id}.`;
                } else {
                    orderSuccessMsg =
                        "Заказ успешно оформлен! Наш оператор свяжется с вами для подтверждения доставки.";
                }

                // Refresh local history if user is logged in
                if (currentUser) {
                    fetchUserHistory();
                }
            } else {
                const errMsg = await res.text();
                alert("Ошибка при оформлении заказа: " + errMsg);
            }
        } catch (err) {
            alert("Ошибка соединения с сервером");
        }
    }

    // Repeat Order from history
    function repeatOrder(order: Order) {
        cart = [];
        order.items.forEach((item) => {
            cart.push({
                dishId: String(item.product_id),
                quantity: item.quantity,
            });
        });
        isProfileOpen = false;
        isCartOpen = true;
        showCheckoutForm = false;
    }

    async function handleReservation() {
        if (
            !reserveName ||
            !reservePhone ||
            !reserveDate ||
            !reserveTime ||
            reserveGuests <= 0
        ) {
            alert("Пожалуйста, заполните все обязательные поля");
            return;
        }
        if (!isValidPhone(reservePhone)) {
            alert(
                "Неверный формат номера телефона. Используйте формат +7XXXXXXXXXX",
            );
            return;
        }

        // Validate booking date boundaries (no past dates, no more than 30 days ahead)
        const today = new Date();
        today.setHours(0, 0, 0, 0);

        const oneMonthFromNow = new Date();
        oneMonthFromNow.setMonth(oneMonthFromNow.getMonth() + 1);
        oneMonthFromNow.setHours(23, 59, 59, 999);

        const selectedDate = new Date(reserveDate);
        if (isNaN(selectedDate.getTime())) {
            alert("Пожалуйста, выберите корректную дату");
            return;
        }
        if (selectedDate < today) {
            alert("Нельзя забронировать столик на прошедшую дату");
            return;
        }
        if (selectedDate > oneMonthFromNow) {
            alert("Бронирование возможно не более чем на месяц вперед");
            return;
        }

        // Validate that reservation datetime has not yet passed in Moscow time (UTC+3)
        const moscowOffsetMs = 3 * 60 * 60 * 1000;
        const nowMoscowMs = Date.now() + moscowOffsetMs;
        const [rYear, rMonth, rDay] = reserveDate.split("-").map(Number);
        const [rHour, rMinute] = reserveTime.split(":").map(Number);
        const reserveMoscowMs = Date.UTC(
            rYear,
            rMonth - 1,
            rDay,
            rHour,
            rMinute,
            0,
        );
        if (reserveMoscowMs <= nowMoscowMs) {
            alert(
                "Выбранное время уже прошло. Пожалуйста, выберите более позднее время (по московскому времени)",
            );
            return;
        }

        try {
            const res = await fetch("/api/reservations", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    customer_name: reserveName,
                    phone: reservePhone,
                    reserve_date: reserveDate,
                    reserve_time: reserveTime,
                    guests_count: reserveGuests,
                    comment: reserveComment,
                }),
            });

            if (res.ok) {
                reserveSuccessMsg =
                    "Столик успешно забронирован! Будем ждать вас в указанное время.";
                // Сбрасываем только поля даты/времени, имя и телефон оставляем для удобства
                if (!currentUser) {
                    reserveName = "";
                    reservePhone = "";
                }
                reserveDate = "";
                reserveTime = "";
                reserveGuests = 2;
                reserveComment = "";
                if (currentUser) {
                    fetchUserHistory();
                }
            } else {
                const errMsg = await res.text();
                alert("Ошибка при бронировании столика: " + errMsg);
            }
        } catch (err) {
            alert("Ошибка соединения с сервером");
        }
    }

    // Authentication logic
    async function handleAuth() {
        authError = "";
        if (!authPhone || !authPassword) {
            authError = "Заполните номер телефона и пароль";
            return;
        }
        if (authMode === "register" && !authName) {
            authError = "Введите ваше имя";
            return;
        }
        if (!isValidPhone(authPhone)) {
            authError =
                "Неверный формат номера телефона. Используйте формат +7XXXXXXXXXX";
            return;
        }
        if (authMode === "register" && authPassword.length < 6) {
            authError = "Пароль должен содержать не менее 6 символов";
            return;
        }

        try {
            const url =
                authMode === "login" ? "/api/auth/login" : "/api/auth/register";
            const body =
                authMode === "login"
                    ? { phone: authPhone, password: authPassword }
                    : {
                          name: authName,
                          phone: authPhone,
                          email: authEmail,
                          password: authPassword,
                      };

            const res = await fetch(url, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });

            if (res.ok) {
                await fetchProfile();
                isAuthModalOpen = false;
                authName = "";
                authPhone = "";
                authEmail = "";
                authPassword = "";
                if (currentUser) {
                    fetchUserHistory();
                }
            } else {
                authError = await res.text();
            }
        } catch (e) {
            authError = "Ошибка соединения с сервером";
        }
    }

    async function handleLogout() {
        try {
            const res = await fetch("/api/auth/logout", { method: "POST" });
            if (res.ok) {
                currentUser = null;
                sessionStorage.removeItem("currentUser");
                userOrders = [];
                userReservations = [];
                isProfileOpen = false;
                stopOrdersPolling();
                if (currentView === "admin") {
                    currentView = "home";
                    window.location.hash = "";
                }
                // Сбрасываем данные форм
                checkoutName = "";
                checkoutPhone = "";
                checkoutAddress = "";
                reserveName = "";
                reservePhone = "";
                editAddress = "";
                editEmail = "";
            }
        } catch (e) {
            console.error(e);
        }
    }

    // Profile Update (address + email)
    async function updateProfileAddress() {
        profileMessage = "";
        try {
            const res = await fetch("/api/auth/profile", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    default_address: editAddress,
                    email: editEmail,
                }),
            });

            if (res.ok) {
                profileMessage = "Профиль успешно обновлен";
                if (currentUser) {
                    currentUser.default_address = editAddress;
                    currentUser.email = editEmail || null;
                    checkoutAddress = editAddress;
                }
            } else {
                profileMessage = "Ошибка при обновлении профиля";
            }
        } catch (e) {
            profileMessage = "Ошибка соединения с сервером";
        }
    }

    // Admin Actions
    async function updateOrderStatus(orderId: number, status: string) {
        try {
            const res = await fetch(`/api/admin/orders/${orderId}/status`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ status }),
            });
            if (res.ok) {
                fetchAdminData();
            }
        } catch (e) {
            console.error(e);
        }
    }

    function openCreateProductForm() {
        editingProdId = null;
        prodFormName = "";
        prodFormDesc = "";
        prodFormPrice = 0;
        prodFormCategory = categories[0]?.id || 1;
        prodFormImage = "";
        prodFormImageFile = null;
        prodFormWeight = 0;
        prodFormCalories = 0;
        prodFormIsAvailable = true;
        isProdFormOpen = true;
    }

    function openEditProductForm(prod: Product) {
        editingProdId = prod.id;
        prodFormName = prod.name;
        prodFormDesc = prod.description;
        prodFormPrice = prod.price;
        prodFormCategory = prod.category_id;
        prodFormImage = prod.image_url;
        prodFormImageFile = null;
        prodFormWeight = prod.weight;
        prodFormCalories = prod.calories;
        prodFormIsAvailable = prod.is_available;
        isProdFormOpen = true;
    }

    async function handleSaveProduct() {
        if (!prodFormName || prodFormPrice <= 0) {
            alert("Укажите корректное название и цену товара");
            return;
        }

        const formData = new FormData();
        formData.append("name", prodFormName);
        formData.append("description", prodFormDesc);
        formData.append("price", prodFormPrice.toString());
        formData.append("category_id", prodFormCategory.toString());
        if (prodFormImageFile) {
            formData.append("image", prodFormImageFile);
        }
        formData.append("weight", prodFormWeight.toString());
        formData.append("calories", prodFormCalories.toString());
        formData.append("is_available", prodFormIsAvailable.toString());

        try {
            const url = editingProdId
                ? `/api/admin/products/${editingProdId}`
                : "/api/admin/products";
            const method = editingProdId ? "PUT" : "POST";

            const res = await fetch(url, {
                method,
                body: formData,
            });

            if (res.ok) {
                isProdFormOpen = false;
                fetchMenu();
                fetchAdminData();
            } else {
                const err = await res.text();
                alert("Ошибка при сохранении товара: " + err);
            }
        } catch (e) {
            alert("Ошибка соединения с сервером");
        }
    }

    async function handleDeleteProduct(id: number) {
        if (!confirm("Вы действительно хотите удалить этот товар?")) return;
        try {
            const res = await fetch(`/api/admin/products/${id}`, {
                method: "DELETE",
            });
            if (res.ok) {
                fetchMenu();
                fetchAdminData();
            } else {
                const err = await res.text();
                alert("Ошибка при удалении товара: " + err);
            }
        } catch (e) {
            alert("Ошибка соединения");
        }
    }

    // Category management helper functions
    function openCreateCategoryForm() {
        editingCatId = null;
        catFormName = "";
        catFormSlug = "";
        isCatFormOpen = true;
    }

    function openEditCategoryForm(cat: Category) {
        editingCatId = cat.id;
        catFormName = cat.name;
        catFormSlug = cat.slug;
        isCatFormOpen = true;
    }

    async function handleSaveCategory() {
        if (!catFormName || !catFormSlug) {
            alert("Заполните название и слаг");
            return;
        }
        const body = { name: catFormName, slug: catFormSlug };
        const method = editingCatId ? "PUT" : "POST";
        const url = editingCatId
            ? `/api/admin/categories/${editingCatId}`
            : "/api/admin/categories";

        try {
            const res = await fetch(url, {
                method,
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });

            if (res.ok) {
                isCatFormOpen = false;
                editingCatId = null;
                catFormName = "";
                catFormSlug = "";
                fetchMenu(); // reload categories
            } else {
                const err = await res.text();
                alert("Ошибка при сохранении категории: " + err);
            }
        } catch (e) {
            alert("Ошибка соединения с сервером");
        }
    }

    async function handleDeleteCategory(id: number) {
        if (
            !confirm(
                "Вы действительно хотите удалить эту категорию? Все товары в ней потеряют привязку!",
            )
        )
            return;
        try {
            const res = await fetch(`/api/admin/categories/${id}`, {
                method: "DELETE",
            });
            if (res.ok) {
                fetchMenu();
            } else {
                const err = await res.text();
                alert("Ошибка при удалении категории: " + err);
            }
        } catch (e) {
            alert("Ошибка соединения");
        }
    }

    // Reservation management status update
    async function updateReservationStatus(id: number, status: string) {
        try {
            const res = await fetch(`/api/admin/reservations/${id}/status`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ status }),
            });
            if (res.ok) {
                fetchAdminData();
            } else {
                const err = await res.text();
                alert("Ошибка при обновлении статуса бронирования: " + err);
            }
        } catch (e) {
            alert("Ошибка соединения");
        }
    }
</script>

<div
    class="min-h-screen bg-brand-dark overflow-x-hidden selection:bg-brand-red selection:text-white"
>
    <!-- Global Alerts -->
    {#if orderSuccessMsg}
        <div
            transition:fade
            class="fixed bottom-6 right-6 z-[90] max-w-md bg-emerald-950/90 border border-emerald-500/30 p-6 rounded-sm text-white shadow-2xl"
        >
            <div class="flex items-start justify-between gap-4">
                <div>
                    <h4
                        class="font-bold text-xs uppercase tracking-widest text-emerald-400 mb-1"
                    >
                        Успешно
                    </h4>
                    <p class="text-xs text-white/80 leading-relaxed">
                        {orderSuccessMsg}
                    </p>
                </div>
                <button
                    onclick={() => (orderSuccessMsg = "")}
                    class="text-white/40 hover:text-white"
                    ><X class="w-4 h-4" /></button
                >
            </div>
        </div>
    {/if}

    {#if reserveSuccessMsg}
        <div
            transition:fade
            class="fixed bottom-6 right-6 z-[90] max-w-md bg-emerald-950/90 border border-emerald-500/30 p-6 rounded-sm text-white shadow-2xl"
        >
            <div class="flex items-start justify-between gap-4">
                <div>
                    <h4
                        class="font-bold text-xs uppercase tracking-widest text-emerald-400 mb-1"
                    >
                        Успешно
                    </h4>
                    <p class="text-xs text-white/80 leading-relaxed">
                        {reserveSuccessMsg}
                    </p>
                </div>
                <button
                    onclick={() => (reserveSuccessMsg = "")}
                    class="text-white/40 hover:text-white"
                    ><X class="w-4 h-4" /></button
                >
            </div>
        </div>
    {/if}

    <!-- Auth Modal -->
    {#if isAuthModalOpen}
        <div
            transition:fade
            onclick={() => (isAuthModalOpen = false)}
            class="fixed inset-0 z-50 bg-black/80 backdrop-blur-md cursor-pointer"
        ></div>
        <div
            transition:scale
            class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-md bg-[#0a0a0a] border border-white/10 p-10 rounded-sm shadow-2xl"
        >
            <div class="flex items-center justify-between mb-8">
                <h3
                    class="text-2xl font-display font-light uppercase tracking-tight text-white"
                >
                    {#if authMode === "login"}Войти в <span
                            class="font-serif italic text-white/50 lowercase"
                            >кабинет</span
                        >{:else}Регистрация{/if}
                </h3>
                <button
                    onclick={() => (isAuthModalOpen = false)}
                    class="text-white/40 hover:text-white"
                    ><X class="w-5 h-5" /></button
                >
            </div>

            {#if authError}
                <div
                    class="bg-brand-red/10 border border-brand-red/20 text-brand-red text-xs p-4 mb-6 uppercase tracking-wider"
                >
                    {authError}
                </div>
            {/if}

            <div class="space-y-6">
                {#if authMode === "register"}
                    <div class="space-y-2">
                        <label
                            class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                            >Ваше Имя</label
                        >
                        <input
                            type="text"
                            bind:value={authName}
                            placeholder="Алексей"
                            class="w-full bg-white/5 border border-white/10 px-4 py-4 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light"
                        />
                    </div>
                {/if}
                <div class="space-y-2">
                    <label
                        class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                        >Номер телефона</label
                    >
                    <input
                        type="tel"
                        bind:value={authPhone}
                        placeholder="+79991234567"
                        class="w-full bg-white/5 border border-white/10 px-4 py-4 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light"
                    />
                </div>
                {#if authMode === "register"}
                    <div class="space-y-2">
                        <label
                            class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                            >Email <span class="text-white/20"
                                >(необязательно)</span
                            ></label
                        >
                        <input
                            type="email"
                            bind:value={authEmail}
                            placeholder="your@email.com"
                            class="w-full bg-white/5 border border-white/10 px-4 py-4 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light"
                        />
                    </div>
                {/if}
                <div class="space-y-2">
                    <label
                        class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                        >Пароль</label
                    >
                    <input
                        type="password"
                        bind:value={authPassword}
                        placeholder="••••••••"
                        class="w-full bg-white/5 border border-white/10 px-4 py-4 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light"
                    />
                </div>

                <button
                    onclick={handleAuth}
                    class="w-full bg-white text-black py-4 font-bold uppercase tracking-widest text-xs hover:bg-brand-red hover:text-white transition-all cursor-pointer"
                >
                    {#if authMode === "login"}Войти{:else}Создать аккаунт{/if}
                </button>

                <div class="text-center pt-4">
                    <button
                        onclick={() => {
                            authMode =
                                authMode === "login" ? "register" : "login";
                            authError = "";
                            authEmail = "";
                        }}
                        class="text-[10px] uppercase tracking-widest text-white/40 hover:text-white transition-colors"
                    >
                        {#if authMode === "login"}Нет аккаунта?
                            Зарегистрироваться{:else}Уже зарегистрированы? Войти{/if}
                    </button>
                </div>
            </div>
        </div>
    {/if}

    <!-- Profile & History Drawer -->
    {#if isProfileOpen}
        <div
            transition:fade={{ duration: 200, delay: 50 }}
            onclick={() => (isProfileOpen = false)}
            class="fixed inset-0 z-50 bg-black/80 backdrop-blur-md cursor-pointer"
        ></div>
        <div
            transition:fly={{ x: 450, duration: 400 }}
            onclick={(e) => e.stopPropagation()}
            class="fixed inset-y-0 right-0 z-50 w-full max-w-2xl bg-[#050505] border-l border-white/5 shadow-2xl overflow-y-auto p-10 flex flex-col"
        >
            <div class="flex items-center justify-between mb-10">
                <h2
                    class="text-3xl font-display font-light uppercase tracking-tighter"
                >
                    Личный <span
                        class="font-serif italic font-medium text-white/60"
                        >Кабинет</span
                    >
                </h2>
                <button
                    onclick={() => (isProfileOpen = false)}
                    class="p-3 hover:bg-white/5 rounded-full transition-colors cursor-pointer text-white/40 hover:text-white"
                >
                    <X class="w-5 h-5" />
                </button>
            </div>

            {#if currentUser}
                <div class="space-y-8 flex-1">
                    <!-- Profile details -->
                    <div class="border border-white/10 p-6 bg-white/[0.01]">
                        <h3
                            class="text-xs uppercase tracking-widest text-white/40 font-mono mb-4"
                        >
                            Данные профиля
                        </h3>
                        <p class="text-lg font-light">{currentUser.name}</p>
                        <p class="text-sm font-mono text-white/40 mt-1">
                            {currentUser.phone}
                        </p>
                        {#if currentUser.email}
                            <p class="text-sm font-mono text-white/30 mt-1">
                                {currentUser.email}
                            </p>
                        {/if}

                        <div
                            class="mt-6 pt-6 border-t border-white/5 space-y-4"
                        >
                            <label
                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                >Email</label
                            >
                            <input
                                type="email"
                                bind:value={editEmail}
                                placeholder="your@email.com"
                                class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red font-light"
                            />
                            <label
                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                >Адрес доставки по умолчанию</label
                            >
                            <div class="flex gap-4">
                                <input
                                    type="text"
                                    bind:value={editAddress}
                                    placeholder="Улица, дом, квартира"
                                    class="flex-1 bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red font-light"
                                />
                                <button
                                    onclick={updateProfileAddress}
                                    class="px-6 bg-white text-black text-[10px] font-bold uppercase tracking-wider hover:bg-brand-red hover:text-white transition-colors cursor-pointer"
                                    >Сохранить</button
                                >
                            </div>
                            {#if profileMessage}
                                <p
                                    class="text-[10px] font-mono {profileMessage.includes(
                                        'Ошибка',
                                    )
                                        ? 'text-brand-red'
                                        : 'text-emerald-400'} mt-2"
                                >
                                    {profileMessage}
                                </p>
                            {/if}
                        </div>
                    </div>

                    <!-- Order history -->
                    <div class="space-y-4">
                        <h3
                            class="text-xs uppercase tracking-widest text-white/60 font-mono"
                        >
                            История заказов
                        </h3>
                        {#if userOrders.length === 0}
                            <p
                                class="text-xs text-white/20 italic font-mono uppercase tracking-wider py-8"
                            >
                                Вы еще не совершали заказов
                            </p>
                        {:else}
                            <div class="space-y-4">
                                {#each userOrders as order}
                                    <div
                                        class="border border-white/5 bg-white/[0.01] p-6 rounded-sm space-y-4"
                                    >
                                        <div
                                            class="flex items-center justify-between text-[10px] font-mono text-white/40"
                                        >
                                            <span
                                                >Заказ #{order.id} от {new Date(
                                                    order.created_at,
                                                ).toLocaleDateString()}</span
                                            >
                                            <span
                                                class="text-brand-red uppercase"
                                                >{localizeOrderStatus(
                                                    order.payment_status,
                                                )}</span
                                            >
                                        </div>
                                        <div class="space-y-2">
                                            {#each order.items as item}
                                                <div
                                                    class="flex items-center justify-between text-xs font-light"
                                                >
                                                    <span class="text-white/60"
                                                        >{item.product_name ||
                                                            `Товар #${item.product_id}`}
                                                        × {item.quantity}</span
                                                    >
                                                    <span class="font-mono"
                                                        >{item.price
                                                            ? `${item.price * item.quantity} ₽`
                                                            : ""}</span
                                                    >
                                                </div>
                                            {/each}
                                        </div>
                                        <div
                                            class="flex items-center justify-between pt-4 border-t border-white/5"
                                        >
                                            <span class="text-sm font-mono"
                                                >{order.total_price} ₽</span
                                            >
                                            <button
                                                onclick={() =>
                                                    repeatOrder(order)}
                                                class="text-[9px] uppercase tracking-widest font-bold text-white hover:text-brand-red transition-colors cursor-pointer"
                                                >Повторить заказ</button
                                            >
                                        </div>
                                    </div>
                                {/each}
                            </div>
                        {/if}
                    </div>

                    <!-- Reservations list -->
                    <div class="space-y-4">
                        <h3
                            class="text-xs uppercase tracking-widest text-white/60 font-mono"
                        >
                            Ваши бронирования
                        </h3>
                        {#if userReservations.length === 0}
                            <p
                                class="text-xs text-white/20 italic font-mono uppercase tracking-wider py-8"
                            >
                                История бронирований пуста
                            </p>
                        {:else}
                            <div class="space-y-4">
                                {#each userReservations as res}
                                    <div
                                        class="border border-white/5 bg-white/[0.01] p-6 rounded-sm flex justify-between items-start"
                                    >
                                        <div class="space-y-2">
                                            <p class="text-sm font-medium">
                                                Столик на {res.guests_count} персон
                                            </p>
                                            <p
                                                class="text-xs text-white/40 font-mono"
                                            >
                                                {res.reserve_date
                                                    ? new Date(
                                                          res.reserve_date,
                                                      ).toLocaleDateString()
                                                    : ""} в {res.reserve_time}
                                            </p>
                                            {#if res.comment}
                                                <p
                                                    class="text-xs text-white/60 italic"
                                                >
                                                    "{res.comment}"
                                                </p>
                                            {/if}
                                        </div>
                                        <span
                                            class="text-[9px] font-mono uppercase text-brand-red bg-brand-red/10 px-2 py-1"
                                            >Подтверждена</span
                                        >
                                    </div>
                                {/each}
                            </div>
                        {/if}
                    </div>
                </div>

                <div class="pt-8 border-t border-white/5">
                    <button
                        onclick={handleLogout}
                        class="w-full bg-white/5 border border-white/10 hover:bg-brand-red hover:text-white transition-colors text-white py-4 text-xs font-bold uppercase tracking-widest cursor-pointer"
                    >
                        Выйти из аккаунта
                    </button>
                </div>
            {/if}
        </div>
    {/if}

    <!-- Cart Drawer -->
    {#if isCartOpen}
        <div
            transition:fade={{ duration: 300 }}
            onclick={() => (isCartOpen = false)}
            class="fixed inset-0 z-50 bg-black/80 backdrop-blur-md cursor-pointer"
        ></div>
        <div
            transition:fly={{ x: 450, duration: 400 }}
            class="fixed inset-y-0 right-0 z-50 w-full max-w-md bg-[#050505] border-l border-white/5 shadow-2xl overflow-y-auto"
        >
            <div class="p-10 h-full flex flex-col">
                <div class="flex items-center justify-between mb-12">
                    <h2
                        class="text-3xl font-display font-light uppercase tracking-tighter"
                    >
                        Ваш <span
                            class="font-serif italic font-medium text-white/60"
                            >Заказ</span
                        >
                    </h2>
                    <button
                        onclick={() => (isCartOpen = false)}
                        class="p-3 hover:bg-white/5 rounded-full transition-colors cursor-pointer text-white/40 hover:text-white"
                    >
                        <X class="w-5 h-5" />
                    </button>
                </div>

                {#if !showCheckoutForm}
                    <div class="flex-1 space-y-8">
                        {#if cart.length === 0}
                            <div
                                class="flex flex-col items-center justify-center h-40 text-white/20"
                            >
                                <ShoppingCart class="w-10 h-10 mb-6 stroke-1" />
                                <p
                                    class="text-[10px] uppercase tracking-[0.3em]"
                                >
                                    Корзина пуста
                                </p>
                            </div>
                        {:else}
                            {#each cart as item (item.dishId)}
                                {@const dish = dishes.find(
                                    (d) => d.id === item.dishId,
                                )}
                                <div
                                    class="flex gap-6 pb-6 border-b border-white/5 group"
                                >
                                    <div
                                        class="w-24 h-24 rounded-sm overflow-hidden bg-white/5 border border-white/5"
                                    >
                                        <img
                                            src={dish?.image}
                                            alt={dish?.name}
                                            class="w-full h-full object-cover grayscale group-hover:grayscale-0 transition-all duration-700"
                                        />
                                    </div>
                                    <div class="flex-1">
                                        <h3
                                            class="font-medium text-[10px] leading-relaxed mb-1 uppercase tracking-widest text-white"
                                        >
                                            {dish?.name}
                                        </h3>
                                        <p
                                            class="text-white/30 font-mono text-[10px] mb-4"
                                        >
                                            {dish?.price} ₽ / ед.
                                        </p>
                                        <div
                                            class="flex items-center justify-between"
                                        >
                                            <div
                                                class="flex items-center gap-4 bg-white/5 border border-white/10 px-3 py-1 rounded-sm"
                                            >
                                                <button
                                                    onclick={() =>
                                                        removeFromCart(
                                                            item.dishId,
                                                        )}
                                                    class="text-white/30 hover:text-white transition-colors cursor-pointer"
                                                >
                                                    <Minus class="w-3 h-3" />
                                                </button>
                                                <span
                                                    class="font-mono text-[10px] text-white"
                                                    >{item.quantity}</span
                                                >
                                                <button
                                                    onclick={() =>
                                                        addToCart(item.dishId)}
                                                    class="text-brand-red hover:text-red-400 cursor-pointer"
                                                >
                                                    <Plus class="w-3 h-3" />
                                                </button>
                                            </div>
                                            <span
                                                class="font-mono text-xs text-white/80"
                                                >{(dish?.price || 0) *
                                                    item.quantity} ₽</span
                                            >
                                        </div>
                                    </div>
                                </div>
                            {/each}
                        {/if}
                    </div>

                    {#if cart.length > 0}
                        <div class="mt-12 pt-8 border-t border-white/5">
                            <div class="flex items-center justify-between mb-8">
                                <span
                                    class="text-white/20 uppercase text-[9px] tracking-[0.4em]"
                                    >Итого к оплате</span
                                >
                                <span
                                    class="text-4xl font-mono tracking-tighter text-white"
                                    >{cartTotal} ₽</span
                                >
                            </div>
                            <button
                                onclick={() => (showCheckoutForm = true)}
                                class="w-full bg-white text-black py-6 rounded-sm font-bold uppercase tracking-[0.25em] text-[10px] hover:bg-brand-red hover:text-white transition-all active:scale-95 cursor-pointer"
                                id="place-order-btn"
                            >
                                Оформить доставку
                            </button>
                        </div>
                    {/if}
                {:else}
                    <!-- Checkout Form inside Drawer -->
                    <div class="flex-1 space-y-6">
                        <h3
                            class="text-xs uppercase tracking-widest text-white/40 font-mono mb-4"
                        >
                            Оформление доставки
                        </h3>

                        <div class="space-y-2">
                            <label
                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                >Ваше Имя</label
                            >
                            <input
                                type="text"
                                bind:value={checkoutName}
                                placeholder="Алексей"
                                class="w-full bg-white/5 border border-white/10 px-4 py-3 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light"
                            />
                        </div>

                        <div class="space-y-2">
                            <label
                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                >Телефон</label
                            >
                            <input
                                type="tel"
                                bind:value={checkoutPhone}
                                placeholder="+79991234567"
                                class="w-full bg-white/5 border border-white/10 px-4 py-3 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light"
                            />
                        </div>

                        <div class="space-y-2">
                            <label
                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                >Адрес доставки</label
                            >
                            <input
                                type="text"
                                bind:value={checkoutAddress}
                                placeholder="Улица, дом, квартира"
                                class="w-full bg-white/5 border border-white/10 px-4 py-3 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light"
                            />
                        </div>

                        <div class="space-y-2">
                            <label
                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                >Способ оплаты</label
                            >
                            <div class="grid grid-cols-2 gap-4">
                                <button
                                    onclick={() => (checkoutPayment = "cash")}
                                    class="py-3 border text-[10px] font-bold uppercase tracking-widest transition-colors cursor-pointer {checkoutPayment ===
                                    'cash'
                                        ? 'border-white bg-white text-black'
                                        : 'border-white/10 text-white/60 hover:text-white bg-white/5'}"
                                >
                                    При получении
                                </button>
                                <button
                                    onclick={() => (checkoutPayment = "online")}
                                    class="py-3 border text-[10px] font-bold uppercase tracking-widest transition-colors cursor-pointer {checkoutPayment ===
                                    'online'
                                        ? 'border-white bg-white text-black'
                                        : 'border-white/10 text-white/60 hover:text-white bg-white/5'}"
                                >
                                    ЮKassa (Онлайн)
                                </button>
                            </div>
                        </div>
                    </div>

                    <div class="mt-12 pt-8 border-t border-white/5 space-y-4">
                        <div class="flex items-center justify-between text-xs">
                            <span class="text-white/40 uppercase font-mono"
                                >К оплате:</span
                            >
                            <span class="font-mono text-white text-lg"
                                >{cartTotal} ₽</span
                            >
                        </div>

                        <div class="flex gap-4">
                            <button
                                onclick={() => (showCheckoutForm = false)}
                                class="flex-1 bg-white/5 border border-white/10 text-white py-4 rounded-sm text-[10px] font-bold uppercase tracking-widest hover:bg-white/10 transition-colors cursor-pointer"
                            >
                                Назад
                            </button>
                            <button
                                onclick={handlePlaceOrder}
                                class="flex-1 bg-white text-black py-4 rounded-sm text-[10px] font-bold uppercase tracking-widest hover:bg-brand-red hover:text-white transition-all cursor-pointer"
                            >
                                Подтвердить
                            </button>
                        </div>
                    </div>
                {/if}
            </div>
        </div>
    {/if}

    {#if currentView === "home" || currentView === "menu" || currentView === "blog" || currentView === "about"}
        <!-- Navigation -->
        <nav
            class="fixed top-0 left-0 right-0 z-40 bg-brand-dark/80 bg-blur border-b border-white/5 py-6"
        >
            <div
                class="w-full px-6 md:px-12 lg:px-20 flex items-center justify-between"
            >
                <div class="flex items-center gap-16">
                    <a
                        href="#"
                        onclick={() => (currentView = "home")}
                        class="flex items-center gap-2 group transition-all"
                    >
                        <div
                            class="w-6 h-6 bg-white flex items-center justify-center rounded-sm group-hover:bg-brand-red transition-colors"
                        >
                            <div class="w-3 h-3 bg-black rotate-45" />
                        </div>
                        <div class="flex flex-col">
                            <span
                                class="text-white font-display text-xs font-black tracking-tighter leading-none"
                                >БАЙКАЛ</span
                            >
                            <span
                                class="text-brand-red font-display text-xs font-black tracking-widest leading-none"
                                >БУУЗЫ</span
                            >
                        </div>
                    </a>
                    <ul
                        class="hidden lg:flex items-center gap-10 uppercase text-[9px] font-bold tracking-[0.3em] text-white/40"
                    >
                        <li>
                            <a
                                href="#menu"
                                class="hover:text-white transition-colors {currentView ===
                                'menu'
                                    ? 'text-white border-b border-brand-red pb-1'
                                    : ''}">Меню</a
                            >
                        </li>
                        <li>
                            <a
                                href="#blog"
                                class="hover:text-white transition-colors {currentView ===
                                'blog'
                                    ? 'text-white border-b border-brand-red pb-1'
                                    : ''}">Блог</a
                            >
                        </li>
                        <li>
                            <a
                                href="#about"
                                class="hover:text-white transition-colors {currentView ===
                                'about'
                                    ? 'text-white border-b border-brand-red pb-1'
                                    : ''}">О нас</a
                            >
                        </li>
                        <li>
                            <a
                                href="#reviews"
                                onclick={() => (currentView = "home")}
                                class="hover:text-white transition-colors"
                                >Отзывы</a
                            >
                        </li>
                        <li>
                            <a
                                href="#tour"
                                onclick={() => (currentView = "home")}
                                class="hover:text-white transition-colors"
                                >Забронировать</a
                            >
                        </li>
                        <li>
                            <a
                                href="#contacts"
                                onclick={() => (currentView = "home")}
                                class="hover:text-white transition-colors"
                                >Контакты</a
                            >
                        </li>
                        {#if isAdminRole(currentUser)}
                            <li>
                                <a
                                    href="#admin"
                                    class="text-brand-red/60 hover:text-brand-red transition-colors"
                                    >{currentUser?.role === "super_admin"
                                        ? "Админ-Панель"
                                        : "Управление"}</a
                                >
                            </li>
                        {/if}
                    </ul>
                </div>
                <div class="flex items-center gap-8">
                    <a
                        href="tel:+79994562323"
                        class="hidden sm:block text-[10px] font-mono text-white/40 hover:text-white transition-colors tracking-tight"
                        >+7 999 456-23-23</a
                    >

                    {#if currentUser}
                        <button
                            onclick={(e) => {
                                e.stopPropagation();
                                if (!currentUser) return;
                                isProfileOpen = true;
                                editAddress = currentUser.default_address || "";
                                editEmail = currentUser.email || "";
                                fetchUserHistory();
                            }}
                            class="flex items-center gap-2 text-[10px] font-mono text-white/60 hover:text-white transition-colors cursor-pointer"
                        >
                            <User class="w-4 h-4 text-brand-red stroke-1" />
                            <span>{currentUser.name}</span>
                        </button>
                    {:else}
                        <button
                            onclick={(e) => {
                                e.stopPropagation();
                                isAuthModalOpen = true;
                            }}
                            class="flex items-center gap-2 text-[10px] font-mono text-white/40 hover:text-white transition-colors cursor-pointer"
                        >
                            <User class="w-4 h-4 stroke-1" />
                            <span>Войти</span>
                        </button>
                    {/if}

                    <!-- Cart Button -->
                    <button
                        onclick={() => (isCartOpen = true)}
                        class="flex items-center gap-4 group cursor-pointer"
                    >
                        <div class="flex flex-col items-end">
                            <span
                                class="text-white font-mono text-xs font-medium tracking-tight"
                                >{cartTotal} ₽</span
                            >
                        </div>
                        <div
                            class="relative w-10 h-10 border border-white/10 rounded-full flex items-center justify-center group-hover:bg-white transition-all duration-500"
                        >
                            <ShoppingCart
                                class="w-4 h-4 text-white group-hover:text-black transition-colors stroke-1"
                            />
                            {#if cart.length > 0}
                                <span
                                    class="absolute -top-1 -right-1 bg-brand-red text-white text-[8px] font-bold w-4 h-4 rounded-full flex items-center justify-center ring-2 ring-brand-dark"
                                >
                                    {cartItemsCount}
                                </span>
                            {/if}
                        </div>
                    </button>
                </div>
            </div>
        </nav>

        {#if currentView === "home"}
            <!-- Hero Section -->
            <section
                class="relative pt-32 pb-20 px-6 md:px-12 lg:px-20 overflow-hidden min-h-screen flex items-center"
            >
                <div
                    class="absolute top-0 right-0 w-[50%] h-[100%] bg-white/[0.02] -skew-x-12 translate-x-1/2 pointer-events-none"
                />
                <div
                    class="w-full grid grid-cols-1 lg:grid-cols-2 gap-20 items-center relative z-10"
                >
                    <div class="space-y-10">
                        <div
                            class="inline-flex items-center gap-2 px-3 py-1 bg-white/5 border border-white/10 rounded-full"
                        >
                            <span
                                class="w-2 h-2 bg-brand-red rounded-full animate-pulse"
                            ></span>
                            <span
                                class="text-[10px] uppercase tracking-widest text-white/60 font-bold font-mono"
                                >Традиционная бурятская кухня • Est. 2014</span
                            >
                        </div>

                        <div class="space-y-4">
                            <h1
                                class="text-6xl lg:text-8xl font-display font-extralight leading-[0.95] uppercase tracking-tighter text-white"
                            >
                                Кафе <br />
                                <span
                                    class="font-serif italic text-white/30 block mt-2"
                                    >Байкал</span
                                >
                                Буузы
                            </h1>
                            <p
                                class="text-white/40 text-lg max-w-md leading-relaxed font-light"
                            >
                                Сочные бурятские буузы из свежего рубленого мяса
                                и наваристые супы, приготовленные по
                                оригинальным рецептам.
                            </p>
                        </div>

                        <div class="flex flex-col sm:flex-row gap-4">
                            <a
                                href="#menu"
                                class="px-10 py-5 bg-white text-black text-center text-xs font-bold uppercase tracking-widest rounded-sm shadow-2xl hover:bg-brand-red hover:text-white transition-all duration-500 cursor-pointer"
                            >
                                Перейти к меню
                            </a>
                            <a
                                href="#tour"
                                class="px-10 py-5 bg-transparent border border-white/10 text-white text-center text-xs font-bold uppercase tracking-widest rounded-sm hover:bg-white/5 transition-all cursor-pointer"
                            >
                                Забронировать стол
                            </a>
                        </div>
                    </div>

                    <div class="relative">
                        <div
                            class="relative z-20 mix-blend-screen grayscale contrast-125 brightness-90 transition-all duration-[1500ms] ease-out {isMounted
                                ? 'opacity-100 scale-100'
                                : 'opacity-0 scale-105'}"
                        >
                            <img
                                src="/images/hero_buuzy_plate.png"
                                alt="Buryat Food"
                                class="w-full max-w-2xl mx-auto drop-shadow-[0_0_50px_rgba(255,255,255,0.05)]"
                            />
                        </div>
                        <div
                            class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[120%] h-[120%] border border-white/5 rounded-full pointer-events-none"
                        />
                        <div
                            class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[80%] h-[80%] border border-white/10 rounded-full pointer-events-none"
                        />
                    </div>
                </div>
            </section>

            <!-- Popular Products Section -->
            <section
                id="popular"
                class="py-32 px-6 md:px-12 lg:px-20 border-t border-white/5 bg-[#030303]"
            >
                <div class="w-full space-y-16">
                    <div class="text-center space-y-4">
                        <div
                            class="inline-flex items-center gap-2 px-3 py-1 bg-white/5 border border-white/10 rounded-full"
                        >
                            <span class="w-2 h-2 bg-brand-red rounded-full"
                            ></span>
                            <span
                                class="text-[9px] uppercase tracking-widest text-white/60 font-bold font-mono"
                                >Выбор гостей</span
                            >
                        </div>
                        <h2
                            class="text-5xl font-display font-extralight uppercase tracking-tight text-white"
                        >
                            Популярно <span
                                class="font-serif italic text-white/20"
                                >По статистике</span
                            >
                        </h2>
                        <p
                            class="max-w-md mx-auto text-white/40 text-sm font-light leading-relaxed"
                        >
                            Наши самые заказываемые традиционные блюда,
                            заслужившие признание сотен гостей.
                        </p>
                    </div>

                    {#if popularDishes.length === 0}
                        <p class="text-white/40 text-center font-mono">
                            Загрузка популярных блюд...
                        </p>
                    {:else}
                        <div
                            class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-8"
                        >
                            {#each popularDishes as item}
                                <MenuCard
                                    {item}
                                    qty={getQuantity(item.id)}
                                    onAdd={() => addToCart(item.id)}
                                    onRemove={() => removeFromCart(item.id)}
                                    onCardClick={() => openDishModal(item)}
                                />
                            {/each}
                        </div>
                    {/if}
                </div>
            </section>
        {:else if currentView === "menu"}
            <!-- Dedicated Menu View Header -->
            <section
                class="relative pt-48 pb-12 px-6 md:px-12 lg:px-20 overflow-hidden bg-[#030303]"
            >
                <div
                    class="absolute top-0 right-0 w-[50%] h-[100%] bg-white/[0.01] -skew-x-12 translate-x-1/2 pointer-events-none"
                />
                <div class="w-full text-center space-y-6 relative z-10">
                    <div
                        class="inline-flex items-center gap-2 px-3 py-1 bg-white/5 border border-white/10 rounded-full"
                    >
                        <span class="w-2 h-2 bg-brand-red rounded-full"></span>
                        <span
                            class="text-[9px] uppercase tracking-widest text-white/60 font-bold font-mono"
                            >Полный кулинарный каталог</span
                        >
                    </div>
                    <h1
                        class="text-5xl lg:text-7xl font-display font-extralight uppercase tracking-tight text-white leading-none"
                    >
                        Кулинарные <span class="font-serif italic text-white/20"
                            >Протоколы</span
                        >
                    </h1>
                    <p
                        class="max-w-xl mx-auto text-white/40 text-sm leading-relaxed font-light"
                    >
                        Каждое блюдо в нашем меню приготовлено по старинным
                        бурятским рецептам из свежих локальных продуктов с
                        особым вниманием к деталям.
                    </p>
                </div>
            </section>

            <!-- Catalog Section -->
            <section
                id="menu"
                class="py-40 px-6 md:px-12 lg:px-20 relative bg-[#030303]"
            >
                <div class="w-full space-y-32">
                    <div
                        class="grid grid-cols-1 md:grid-cols-2 gap-20 items-end"
                    >
                        <div class="space-y-6">
                            <h2
                                class="text-[10rem] font-display font-black uppercase tracking-tighter text-white/5 absolute -top-20 left-0 pointer-events-none overflow-hidden select-none"
                            >
                                МЕНЮ
                            </h2>
                            <h2
                                class="text-6xl font-display font-extralight uppercase tracking-tight relative z-10"
                            >
                                Кулинарные <span
                                    class="font-serif italic text-white/20"
                                    >Протоколы</span
                                >
                            </h2>
                            <p
                                class="max-w-md text-white/40 text-sm leading-relaxed font-light"
                            >
                                Наше меню основывается на традиционных вкусах
                                степной Азии. Каждое блюдо — это баланс специй и
                                свежего мяса.
                            </p>
                        </div>
                    </div>

                    <div class="space-y-40">
                        {#if categories.length === 0}
                            <p class="text-white/40 text-center font-mono">
                                Загрузка меню...
                            </p>
                        {:else}
                            {#each categories as cat}
                                {@const catDishes = dishes.filter(
                                    (d) => d.category === cat.slug,
                                )}
                                {#if catDishes.length > 0}
                                    <MenuCategory
                                        title={cat.name}
                                        subtitle={`${cat.name} Section`}
                                        dishes={catDishes}
                                        isExpanded={true}
                                        toggleExpand={() => {}}
                                        getQty={getQuantity}
                                        onAdd={addToCart}
                                        onRemove={removeFromCart}
                                        onCardClick={(id) => openDishModal(dishes.find((d) => d.id === id)!)}
                                    />
                                {/if}
                            {/each}
                        {/if}
                    </div>
                </div>
            </section>
        {:else if currentView === "blog"}
            <!-- Dedicated Blog Header -->
            <section
                class="relative pt-48 pb-12 px-6 md:px-12 lg:px-20 overflow-hidden bg-[#030303]"
            >
                <div
                    class="absolute top-0 right-0 w-[50%] h-[100%] bg-white/[0.01] -skew-x-12 translate-x-1/2 pointer-events-none"
                />
                <div class="w-full text-center space-y-6 relative z-10">
                    <div
                        class="inline-flex items-center gap-2 px-3 py-1 bg-white/5 border border-white/10 rounded-full"
                    >
                        <span class="w-2 h-2 bg-brand-red rounded-full"></span>
                        <span
                            class="text-[9px] uppercase tracking-widest text-white/60 font-bold font-mono"
                            >Кулинарные хроники</span
                        >
                    </div>
                    <h1
                        class="text-5xl lg:text-7xl font-display font-extralight uppercase tracking-tight text-white leading-none"
                    >
                        Наш <span class="font-serif italic text-white/20"
                            >Блог</span
                        >
                    </h1>
                    <p
                        class="max-w-xl mx-auto text-white/40 text-sm leading-relaxed font-light"
                    >
                        Истории о традициях бурятской кухни, секретах
                        приготовления идеальных блюд и новостях нашего
                        заведения.
                    </p>
                </div>
            </section>

            <!-- Blog Posts Grid -->
            <section class="py-20 px-6 md:px-12 lg:px-20 bg-[#030303]">
                {#if blogPosts.length === 0}
                    <p
                        class="text-white/30 text-center font-mono text-sm py-20"
                    >
                        Статьи пока не опубликованы
                    </p>
                {:else}
                    <div class="w-full grid grid-cols-1 md:grid-cols-2 gap-12">
                        {#each blogPosts as post}
                            <button
                                onclick={() => (selectedBlogPost = post)}
                                class="text-left bg-white/[0.01] border border-white/5 p-8 group transition-all duration-700 hover:border-white/20 hover:-translate-y-1 relative overflow-hidden flex flex-col justify-between h-[450px] cursor-pointer"
                            >
                                <div
                                    class="relative w-full h-48 overflow-hidden bg-[#0a0a0a] mb-6"
                                >
                                    <img
                                        src={post.image_url ||
                                            "/images/placeholder.jpg"}
                                        alt={post.title}
                                        class="w-full h-full object-cover grayscale opacity-60 group-hover:grayscale-0 group-hover:opacity-100 group-hover:scale-105 transition-all duration-1000 ease-out"
                                    />
                                    {#if post.tag}
                                        <span
                                            class="absolute top-4 left-4 bg-black/80 border border-white/10 px-3 py-1 text-[8px] font-mono uppercase tracking-widest text-white"
                                        >
                                            {post.tag}
                                        </span>
                                    {/if}
                                </div>

                                <div class="flex-1 space-y-4">
                                    <div
                                        class="flex items-center gap-4 text-[10px] font-mono text-white/40"
                                    >
                                        <span
                                            >{formatBlogDate(
                                                post.created_at,
                                            )}</span
                                        >
                                        {#if post.read_time}
                                            <span>•</span>
                                            <span>{post.read_time}</span>
                                        {/if}
                                    </div>
                                    <h3
                                        class="text-2xl font-display font-light uppercase tracking-tight text-white group-hover:text-brand-red transition-colors duration-500"
                                    >
                                        {post.title}
                                    </h3>
                                    <p
                                        class="text-xs text-white/50 font-light line-clamp-3 leading-relaxed"
                                    >
                                        {post.subtitle}
                                    </p>
                                </div>

                                <div
                                    class="mt-6 pt-4 border-t border-white/5 flex items-center justify-between text-[10px] font-bold uppercase tracking-[0.2em] text-white/40 group-hover:text-white transition-colors duration-500"
                                >
                                    <span>Читать статью</span>
                                    <ChevronRight
                                        class="w-3 h-3 text-white/40 group-hover:text-white transition-colors"
                                    />
                                </div>
                            </button>
                        {/each}
                    </div>
                {/if}
            </section>
        {:else if currentView === "about"}
            <!-- Dedicated About Header -->
            <section
                class="relative pt-48 pb-12 px-6 md:px-12 lg:px-20 overflow-hidden bg-[#030303]"
            >
                <div
                    class="absolute top-0 right-0 w-[50%] h-[100%] bg-white/[0.01] -skew-x-12 translate-x-1/2 pointer-events-none"
                />
                <div class="w-full text-center space-y-6 relative z-10">
                    <div
                        class="inline-flex items-center gap-2 px-3 py-1 bg-white/5 border border-white/10 rounded-full"
                    >
                        <span class="w-2 h-2 bg-brand-red rounded-full"></span>
                        <span
                            class="text-[9px] uppercase tracking-widest text-white/60 font-bold font-mono"
                            >О кафе</span
                        >
                    </div>
                    <h1
                        class="text-5xl lg:text-7xl font-display font-extralight uppercase tracking-tight text-white leading-none"
                    >
                        О <span class="font-serif italic text-white/20"
                            >Нас</span
                        >
                    </h1>
                    <p
                        class="max-w-xl mx-auto text-white/40 text-sm leading-relaxed font-light"
                    >
                        «Байкал Буузы» — новое кафе, которое только открыло свои
                        двери. Простое место для тех, кто хочет перекусить и
                        отдохнуть.
                    </p>
                </div>
            </section>

            <!-- About Content Sections -->
            <section
                class="py-20 px-6 md:px-12 lg:px-20 bg-[#030303] border-t border-white/5"
            >
                <div class="w-full space-y-32">
                    <!-- Section 1: What we are -->
                    <div
                        class="grid grid-cols-1 lg:grid-cols-2 gap-20 items-center"
                    >
                        <div class="space-y-8">
                            <span
                                class="text-[10px] uppercase tracking-[0.4em] text-brand-red font-bold font-mono"
                                >Что мы такое</span
                            >
                            <h2
                                class="text-5xl font-display font-extralight uppercase tracking-tight text-white"
                            >
                                Новое место <br />
                                <span
                                    class="font-serif italic text-white/30 lowercase font-medium"
                                    >для отдыха</span
                                >
                            </h2>
                            <div class="h-px w-20 bg-white/20" />
                            <p
                                class="text-white/50 text-sm leading-relaxed font-light"
                            >
                                «Байкал Буузы» — небольшое кафе с понятным меню:
                                кофе, закуски и блюда на каждый день. Подходит
                                для обеда, короткого перерыва или неспешной
                                встречи.
                            </p>
                            <p
                                class="text-white/50 text-sm leading-relaxed font-light"
                            >
                                Кафе находится на начальном этапе работы. Мы
                                последовательно расширяем меню и отлаживаем
                                сервис.
                            </p>
                        </div>
                        <div
                            class="relative border border-white/10 p-2 bg-white/[0.01]"
                        >
                            <img
                                src="/images/hero_buuzy_plate.png"
                                alt="Блюда кафе"
                                class="w-full object-cover grayscale contrast-125 brightness-90 hover:grayscale-0 transition-all duration-1000"
                            />
                        </div>
                    </div>

                    <!-- Section 2: What we offer (3 Columns) -->
                    <div class="border-t border-white/5 pt-32">
                        <div class="text-center space-y-4 mb-20">
                            <span
                                class="text-[10px] uppercase tracking-[0.4em] text-brand-red font-bold font-mono"
                                >Что есть в меню</span
                            >
                            <h2
                                class="text-4xl font-display font-extralight uppercase text-white"
                            >
                                Что мы предлагаем
                            </h2>
                        </div>

                        <div class="grid grid-cols-1 md:grid-cols-3 gap-12">
                            <div
                                class="border border-white/5 bg-white/[0.01] p-10 space-y-6"
                            >
                                <span
                                    class="text-xs font-mono text-white/20 uppercase tracking-widest"
                                    >01 / Напитки</span
                                >
                                <h3
                                    class="text-xl font-display font-light uppercase text-white"
                                >
                                    Кофе
                                </h3>
                                <p
                                    class="text-xs text-white/40 leading-relaxed font-light"
                                >
                                    Эспрессо, американо, капучино и другие
                                    кофейные напитки для любого времени дня.
                                </p>
                            </div>
                            <div
                                class="border border-white/5 bg-white/[0.01] p-10 space-y-6"
                            >
                                <span
                                    class="text-xs font-mono text-white/20 uppercase tracking-widest"
                                    >02 / Еда</span
                                >
                                <h3
                                    class="text-xl font-display font-light uppercase text-white"
                                >
                                    Закуски и блюда
                                </h3>
                                <p
                                    class="text-xs text-white/40 leading-relaxed font-light"
                                >
                                    Лёгкие перекусы и сытные позиции для обеда
                                    или ужина — без лишних сложностей.
                                </p>
                            </div>
                            <div
                                class="border border-white/5 bg-white/[0.01] p-10 space-y-6"
                            >
                                <span
                                    class="text-xs font-mono text-white/20 uppercase tracking-widest"
                                    >03 / Атмосфера</span
                                >
                                <h3
                                    class="text-xl font-display font-light uppercase text-white"
                                >
                                    Место для отдыха
                                </h3>
                                <p
                                    class="text-xs text-white/40 leading-relaxed font-light"
                                >
                                    Спокойное пространство, где можно
                                    расслабиться в одиночку или провести время в
                                    компании.
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </section>
        {/if}

        {#if currentView === "home"}
            <!-- Table Booking Section -->
            <section
                id="tour"
                class="py-40 px-6 md:px-12 lg:px-20 border-t border-white/5 bg-[#050505]"
            >
                <div
                    class="w-full grid grid-cols-1 lg:grid-cols-2 gap-20 items-center"
                >
                    <div class="space-y-8">
                        <span
                            class="text-[10px] uppercase tracking-[0.4em] text-brand-red font-bold font-mono"
                            >Reservation terminal</span
                        >
                        <h2
                            class="text-5xl font-display font-extralight uppercase leading-tight tracking-tight text-white"
                        >
                            Бронирование <br />
                            <span
                                class="font-serif italic text-white/30 lowercase font-medium"
                                >Столика</span
                            >
                        </h2>
                        <div class="h-px w-20 bg-white/20" />
                        <p
                            class="text-white/40 text-sm leading-relaxed max-w-md"
                        >
                            Забронируйте столик заранее, чтобы провести
                            незабываемый вечер в уютной атмосфере бурятского
                            кафе «Байкал Буузы». Мы сохраним для вас лучшее
                            место.
                        </p>
                    </div>

                    <!-- Booking Form -->
                    <div
                        class="bg-white/[0.01] border border-white/10 p-10 space-y-6"
                    >
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div class="space-y-2">
                                <label
                                    class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                    >Ваше имя</label
                                >
                                <input
                                    type="text"
                                    bind:value={reserveName}
                                    placeholder="Алексей"
                                    class="w-full bg-white/5 border border-white/10 px-4 py-3 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light"
                                />
                            </div>
                            <div class="space-y-2">
                                <label
                                    class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                    >Телефон</label
                                >
                                <input
                                    type="tel"
                                    bind:value={reservePhone}
                                    placeholder="+79991234567"
                                    class="w-full bg-white/5 border border-white/10 px-4 py-3 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light"
                                />
                            </div>
                        </div>

                        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                            <div class="space-y-2">
                                <label
                                    class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                    >Дата</label
                                >
                                <input
                                    type="date"
                                    bind:value={reserveDate}
                                    min={minDate}
                                    max={maxDate}
                                    class="w-full bg-white/5 border border-white/10 px-4 py-3 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light"
                                />
                            </div>
                            <div class="space-y-2">
                                <label
                                    class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                    >Время</label
                                >
                                <input
                                    type="time"
                                    bind:value={reserveTime}
                                    class="w-full bg-white/5 border border-white/10 px-4 py-3 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light"
                                />
                            </div>
                            <div class="space-y-2">
                                <label
                                    class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                    >Количество гостей</label
                                >
                                <input
                                    type="number"
                                    min="1"
                                    max="20"
                                    bind:value={reserveGuests}
                                    class="w-full bg-white/5 border border-white/10 px-4 py-3 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light font-mono"
                                />
                            </div>
                        </div>

                        <div class="space-y-2">
                            <label
                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                >Комментарий (Пожелания)</label
                            >
                            <textarea
                                bind:value={reserveComment}
                                placeholder="Например: столик у окна, детский стульчик..."
                                rows="3"
                                class="w-full bg-white/5 border border-white/10 px-4 py-3 rounded-sm text-sm text-white focus:outline-none focus:border-brand-red font-light resize-none"
                            ></textarea>
                        </div>

                        <button
                            onclick={handleReservation}
                            class="w-full bg-white text-black py-4 font-bold uppercase tracking-widest text-xs hover:bg-brand-red hover:text-white transition-all cursor-pointer"
                        >
                            Забронировать
                        </button>
                    </div>
                </div>
            </section>

            <!-- Reviews Section -->
            <section
                id="reviews"
                class="py-40 px-6 md:px-12 lg:px-20 border-t border-white/5 bg-[#030303]"
            >
                <div class="w-full">
                    <div class="grid grid-cols-1 lg:grid-cols-3 gap-20">
                        <div class="space-y-8">
                            <span
                                class="text-[10px] uppercase tracking-[0.4em] text-brand-red font-bold font-mono"
                                >Feedback Grid</span
                            >
                            <h2
                                class="text-5xl font-display font-extralight uppercase leading-tight tracking-tight"
                            >
                                Отзывы <br />
                                <span
                                    class="font-serif italic text-white/30 lowercase font-medium"
                                    >Посетителей</span
                                >
                            </h2>
                            <div class="h-px w-20 bg-white/20" />
                            <p
                                class="text-white/40 text-sm leading-relaxed max-w-xs font-light"
                            >
                                Архивы отзывов наших гостей из социальных сетей.
                                Подтвержденное качество и искренняя
                                благодарность.
                            </p>
                        </div>

                        <div
                            class="lg:col-span-2 grid grid-cols-1 md:grid-cols-2 gap-8"
                        >
                            {#each [{ name: "Андрей В.", date: "20.10.25", text: "Буузы невероятно сочные! Горячий бульон внутри — это настоящее искусство.", platform: "ВКонтакте" }, { name: "Юлия К.", date: "23.10.25", text: "Атмосферное и стильное место. Очень понравился облепиховый морс.", platform: "Telegram" }, { name: "Станислав А.", date: "01.11.25", text: "Прекрасный суп шулэн, согревает в любую непогоду. Обязательно вернусь.", platform: "Telegram" }] as review}
                                <div
                                    class="p-8 border border-white/10 bg-white/[0.01] flex flex-col justify-between group hover:border-white/30 transition-all duration-700"
                                >
                                    <div class="space-y-4">
                                        <div
                                            class="flex items-center justify-between"
                                        >
                                            <span
                                                class="text-[10px] font-mono text-white/40"
                                                >{review.date}</span
                                            >
                                            <span
                                                class="text-[10px] font-mono text-brand-red uppercase"
                                                >{review.platform}</span
                                            >
                                        </div>
                                        <p
                                            class="text-sm font-light text-white/60 leading-relaxed italic"
                                        >
                                            "{review.text}"
                                        </p>
                                    </div>
                                    <div
                                        class="mt-8 flex items-center justify-between"
                                    >
                                        <span
                                            class="text-xs font-bold uppercase tracking-widest text-white"
                                            >{review.name}</span
                                        >
                                        <ChevronRight
                                            class="w-4 h-4 text-white/20 group-hover:text-white transition-colors"
                                        />
                                    </div>
                                </div>
                            {/each}
                        </div>
                    </div>
                </div>
            </section>

            <!-- Contacts & Map -->
            <section
                id="contacts"
                class="py-40 px-6 md:px-12 lg:px-20 border-t border-white/5"
            >
                <div
                    class="w-full grid grid-cols-1 lg:grid-cols-2 gap-12 h-[600px]"
                >
                    <div
                        class="bg-white/[0.02] border border-white/10 p-16 flex flex-col justify-between relative overflow-hidden group"
                    >
                        <div class="space-y-16 relative z-10">
                            <h2
                                class="text-7xl font-display font-light uppercase tracking-tighter text-white"
                            >
                                Контакты <br />
                                <span
                                    class="font-serif italic text-white/30 lowercase font-medium"
                                    >кафе</span
                                >
                            </h2>

                            <div class="grid grid-cols-1 md:grid-cols-2 gap-12">
                                <div class="space-y-2">
                                    <p
                                        class="text-[10px] font-mono uppercase tracking-widest text-white/20"
                                    >
                                        Телефон бронирования
                                    </p>
                                    <a
                                        href="tel:+79994562323"
                                        class="text-xl font-light text-white hover:text-brand-red transition-colors"
                                        >+7 999 456-23-23</a
                                    >
                                </div>
                                <div class="space-y-2">
                                    <p
                                        class="text-[10px] font-mono uppercase tracking-widest text-white/20"
                                    >
                                        Адрес заведения
                                    </p>
                                    <p
                                        class="text-xl font-light italic text-white/60 leading-tight"
                                    >
                                        45-я Параллель 5/3А, Геленджик
                                    </p>
                                </div>
                            </div>
                        </div>

                        <div
                            class="flex flex-col sm:flex-row items-center gap-8 relative z-10"
                        >
                            <a
                                href="#tour"
                                class="px-12 py-5 bg-white text-black text-[10px] text-center font-bold uppercase tracking-[0.3em] hover:bg-brand-red hover:text-white transition-all duration-500 cursor-pointer"
                            >
                                Заказать столик
                            </a>
                            <div class="flex gap-6">
                                {#each [Instagram, Facebook, Send] as Icon}
                                    <button
                                        class="text-white/40 hover:text-white transition-colors cursor-pointer"
                                    >
                                        <Icon class="w-5 h-5 stroke-1" />
                                    </button>
                                {/each}
                            </div>
                        </div>
                    </div>

                    <div
                        class="border border-white/10 overflow-hidden relative grayscale brightness-75 contrast-125 opacity-40 hover:opacity-100 transition-opacity duration-1000"
                    >
                        <iframe
                            src="https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d23956!2d38.0769!3d44.5616!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x40f6050bde93ad69%3A0xfedf04cb7b8fbb39!2z0JPQtdC70LXQvdC00LbQuNC6!5e0!3m2!1sru!2sru!4v1716123456789!5m2!1sru!2sru"
                            width="100%"
                            height="100%"
                            style="border: 0;"
                            loading="lazy"
                            title="Google Map Location"
                        />
                    </div>
                </div>
            </section>
        {/if}

        <!-- Footer -->
        <footer
            class="py-20 px-6 md:px-12 lg:px-20 border-t border-white/5 bg-brand-dark"
        >
            <div
                class="w-full flex flex-col md:flex-row items-center justify-between gap-12 opacity-40 hover:opacity-100 transition-opacity"
            >
                <div class="flex flex-col items-center md:items-start gap-4">
                    <div class="flex items-center gap-2">
                        <div class="w-4 h-4 bg-white rotate-45" />
                        <span
                            class="text-sm font-black uppercase tracking-widest text-white"
                            >БАЙКАЛ.БУУЗЫ</span
                        >
                    </div>
                    <p
                        class="text-[10px] font-mono uppercase tracking-[0.2em] mt-2"
                    >
                        © 2026 БАЙКАЛ БУУЗЫ КАФЕ
                    </p>
                </div>

                <ul
                    class="flex flex-wrap items-center justify-center gap-10 uppercase text-[9px] font-bold tracking-[0.3em] text-white"
                >
                    <li>
                        <a
                            href="#menu"
                            class="hover:text-brand-red transition-colors"
                            >Меню</a
                        >
                    </li>
                    <li>
                        <a
                            href="#blog"
                            class="hover:text-brand-red transition-colors"
                            >Блог</a
                        >
                    </li>
                    <li>
                        <a
                            href="#about"
                            class="hover:text-brand-red transition-colors"
                            >О нас</a
                        >
                    </li>
                    <li>
                        <a
                            href="#reviews"
                            onclick={() => (currentView = "home")}
                            class="hover:text-brand-red transition-colors"
                            >Отзывы</a
                        >
                    </li>
                    <li>
                        <a
                            href="#tour"
                            onclick={() => (currentView = "home")}
                            class="hover:text-brand-red transition-colors"
                            >Столы</a
                        >
                    </li>
                    <li>
                        <a
                            href="#contacts"
                            onclick={() => (currentView = "home")}
                            class="hover:text-brand-red transition-colors font-mono tracking-normal"
                            >Контакты</a
                        >
                    </li>
                </ul>
            </div>
        </footer>

        <!-- Blog Article Detail Overlay -->
        {#if selectedBlogPost}
            <div
                transition:fade={{ duration: 400 }}
                class="fixed inset-0 z-50 bg-brand-dark/95 backdrop-blur-md flex items-center justify-center p-6 md:p-12 overflow-y-auto"
            >
                <div
                    transition:fly={{ y: 50, duration: 600 }}
                    class="w-full max-w-4xl bg-[#080808] border border-white/10 rounded-sm relative overflow-hidden"
                >
                    <!-- Close Button -->
                    <button
                        onclick={() => (selectedBlogPost = null)}
                        class="absolute top-6 right-6 w-10 h-10 border border-white/10 rounded-full flex items-center justify-center hover:bg-white hover:text-black transition-all duration-500 cursor-pointer z-50 text-white"
                    >
                        <X class="w-4 h-4" />
                    </button>

                    <!-- Large Header Image -->
                    <div
                        class="relative w-full h-[350px] overflow-hidden bg-black"
                    >
                        <img
                            src={selectedBlogPost.image_url ||
                                "/images/placeholder.jpg"}
                            alt={selectedBlogPost.title}
                            class="w-full h-full object-cover opacity-50"
                        />
                        <div
                            class="absolute inset-0 bg-gradient-to-t from-[#080808] via-transparent to-transparent"
                        ></div>
                        <div
                            class="absolute bottom-10 left-10 right-10 space-y-4 font-mono"
                        >
                            {#if selectedBlogPost.tag}
                                <span
                                    class="bg-brand-red px-3 py-1 text-[9px] uppercase tracking-widest text-white"
                                >
                                    {selectedBlogPost.tag}
                                </span>
                            {/if}
                            <h2
                                class="text-4xl lg:text-5xl font-display font-light uppercase tracking-tight text-white"
                            >
                                {selectedBlogPost.title}
                            </h2>
                        </div>
                    </div>

                    <!-- Content Area -->
                    <div
                        class="p-10 md:p-16 space-y-8 max-h-[50vh] overflow-y-auto font-light leading-relaxed text-white/70 text-sm"
                    >
                        <div
                            class="flex items-center gap-6 text-[10px] font-mono text-white/40 border-b border-white/5 pb-4"
                        >
                            <span
                                >Опубликовано: {formatBlogDate(
                                    selectedBlogPost.created_at,
                                )}</span
                            >
                            {#if selectedBlogPost.read_time}
                                <span>•</span>
                                <span
                                    >Время чтения: {selectedBlogPost.read_time}</span
                                >
                            {/if}
                        </div>

                        <p class="text-base italic text-white/90">
                            {selectedBlogPost.subtitle}
                        </p>

                        <p
                            class="whitespace-pre-line text-justify leading-loose"
                        >
                            {selectedBlogPost.content}
                        </p>
                    </div>
                </div>
            </div>
        {/if}
    {/if}

    <!-- Admin Panel View -->
    {#if currentView === "admin"}
        <!-- Admin Header -->
        <nav
            class="bg-brand-gray border-b border-white/10 py-6 sticky top-0 z-40"
        >
            <div
                class="w-full px-6 md:px-12 lg:px-20 flex items-center justify-between"
            >
                <div class="flex items-center gap-8">
                    <div class="flex items-center gap-2">
                        <div class="w-5 h-5 bg-brand-red rotate-45" />
                        <span
                            class="font-display text-sm font-bold uppercase tracking-widest text-white"
                            >Панель Управления</span
                        >
                    </div>
                    <span
                        class="text-[9px] font-mono text-emerald-400 border border-emerald-400/20 bg-emerald-400/10 px-2 py-0.5 uppercase"
                        >Developer Mode</span
                    >
                </div>
                <div class="flex items-center gap-6">
                    {#if currentUser}
                        <button
                            onclick={() => {
                                isProfileOpen = true;
                                editAddress =
                                    currentUser?.default_address || "";
                                editEmail = currentUser?.email || "";
                                fetchUserHistory();
                            }}
                            class="flex items-center gap-2 text-[10px] font-mono text-white/60 hover:text-white transition-colors cursor-pointer"
                        >
                            <User class="w-4 h-4 text-brand-red stroke-1" />
                            <span>{currentUser.name}</span>
                        </button>
                    {/if}
                    <a
                        href="#"
                        class="flex items-center gap-2 text-[10px] font-bold uppercase tracking-widest text-white hover:text-brand-red transition-colors"
                    >
                        <Home class="w-4 h-4" />
                        <span>Вернуться на сайт</span>
                    </a>
                </div>
            </div>
        </nav>

        <!-- Admin Body -->
        <div
            class="w-full px-6 md:px-12 lg:px-20 py-12 flex flex-col lg:flex-row gap-12"
        >
            <!-- Admin Navigation Sidebar -->
            <aside class="w-full lg:w-64 space-y-2 flex-shrink-0">
                {#if isAdminRole(currentUser)}
                    <button
                        onclick={() => (adminTab = "stats")}
                        class="w-full text-left px-6 py-4 text-[11px] font-mono uppercase tracking-widest border transition-all cursor-pointer flex items-center gap-4 {adminTab ===
                        'stats'
                            ? 'bg-white text-black border-white'
                            : 'bg-transparent text-white/60 hover:text-white border-white/5 hover:border-white/20'}"
                    >
                        <TrendingUp class="w-4 h-4" />
                        <span>Статистика</span>
                    </button>
                {/if}

                <button
                    onclick={() => (adminTab = "orders")}
                    class="w-full text-left px-6 py-4 text-[11px] font-mono uppercase tracking-widest border transition-all cursor-pointer flex items-center gap-4 {adminTab ===
                    'orders'
                        ? 'bg-white text-black border-white'
                        : 'bg-transparent text-white/60 hover:text-white border-white/5 hover:border-white/20'}"
                >
                    <ShoppingCart class="w-4 h-4" />
                    <span>Заказы ({adminOrders.length})</span>
                </button>

                <button
                    onclick={() => (adminTab = "reservations")}
                    class="w-full text-left px-6 py-4 text-[11px] font-mono uppercase tracking-widest border transition-all cursor-pointer flex items-center gap-4 {adminTab ===
                    'reservations'
                        ? 'bg-white text-black border-white'
                        : 'bg-transparent text-white/60 hover:text-white border-white/5 hover:border-white/20'}"
                >
                    <Calendar class="w-4 h-4" />
                    <span>Бронирования ({adminReservations.length})</span>
                </button>

                {#if isAdminRole(currentUser)}
                    <button
                        onclick={() => (adminTab = "menu")}
                        class="w-full text-left px-6 py-4 text-[11px] font-mono uppercase tracking-widest border transition-all cursor-pointer flex items-center gap-4 {adminTab ===
                        'menu'
                            ? 'bg-white text-black border-white'
                            : 'bg-transparent text-white/60 hover:text-white border-white/5 hover:border-white/20'}"
                    >
                        <Layers class="w-4 h-4" />
                        <span>Управление меню</span>
                    </button>

                    <button
                        onclick={() => (adminTab = "categories")}
                        class="w-full text-left px-6 py-4 text-[11px] font-mono uppercase tracking-widest border transition-all cursor-pointer flex items-center gap-4 {adminTab ===
                        'categories'
                            ? 'bg-white text-black border-white'
                            : 'bg-transparent text-white/60 hover:text-white border-white/5 hover:border-white/20'}"
                    >
                        <Layers class="w-4 h-4" />
                        <span>Категории ({categories.length})</span>
                    </button>

                    <button
                        onclick={() => {
                            adminTab = "blog";
                            fetchAdminBlogPosts();
                        }}
                        class="w-full text-left px-6 py-4 text-[11px] font-mono uppercase tracking-widest border transition-all cursor-pointer flex items-center gap-4 {adminTab ===
                        'blog'
                            ? 'bg-white text-black border-white'
                            : 'bg-transparent text-white/60 hover:text-white border-white/5 hover:border-white/20'}"
                    >
                        <Send class="w-4 h-4" />
                        <span>Блог ({adminBlogPosts.length})</span>
                    </button>
                {/if}

                <button
                    onclick={() => (adminTab = "users")}
                    class="w-full text-left px-6 py-4 text-[11px] font-mono uppercase tracking-widest border transition-all cursor-pointer flex items-center gap-4 {adminTab ===
                    'users'
                        ? 'bg-white text-black border-white'
                        : 'bg-transparent text-white/60 hover:text-white border-white/5 hover:border-white/20'}"
                >
                    <User class="w-4 h-4" />
                    <span>Пользователи ({adminUsers.length})</span>
                </button>

                {#if currentUser?.role === "super_admin"}
                    <button
                        onclick={() => {
                            adminTab = "audit-log";
                            fetchAuditLog();
                        }}
                        class="w-full text-left px-6 py-4 text-[11px] font-mono uppercase tracking-widest border transition-all cursor-pointer flex items-center gap-4 {adminTab ===
                        'audit-log'
                            ? 'bg-white text-black border-white'
                            : 'bg-transparent text-white/60 hover:text-white border-white/5 hover:border-white/20'}"
                    >
                        <TrendingUp class="w-4 h-4" />
                        <span>Журнал аудита</span>
                    </button>
                {/if}
            </aside>

            <!-- Admin Main Area -->
            <main class="flex-1 min-w-0">
                {#if adminTab === "stats"}
                    <div class="space-y-12">
                        <h2
                            class="text-3xl font-display font-light uppercase tracking-tight text-white"
                        >
                            Статистика <span
                                class="font-serif italic text-white/40 lowercase"
                                >кафе</span
                            >
                        </h2>

                        <div class="flex items-center gap-4 mb-8">
                            <input
                                type="date"
                                bind:value={statsStartDate}
                                class="bg-white/5 border border-white/10 px-4 py-2 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm font-mono"
                            />
                            <span class="text-white/40">-</span>
                            <input
                                type="date"
                                bind:value={statsEndDate}
                                class="bg-white/5 border border-white/10 px-4 py-2 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm font-mono"
                            />
                            <button
                                onclick={fetchAdminData}
                                class="px-6 py-2 bg-white text-black text-[10px] font-bold uppercase tracking-widest hover:bg-brand-red hover:text-white transition-colors cursor-pointer rounded-sm"
                                >Применить</button
                            >
                        </div>
                        <div class="grid grid-cols-1 md:grid-cols-4 gap-8">
                            <div
                                class="border border-white/10 p-8 bg-white/[0.01]"
                            >
                                <p
                                    class="text-[10px] uppercase tracking-widest font-mono text-white/40 mb-2"
                                >
                                    Общая выручка
                                </p>
                                <p
                                    class="text-4xl font-mono tracking-tight text-white"
                                >
                                    {adminStats?.total_revenue || 0} ₽
                                </p>
                            </div>
                            <div
                                class="border border-white/10 p-8 bg-white/[0.01]"
                            >
                                <p
                                    class="text-[10px] uppercase tracking-widest font-mono text-white/40 mb-2"
                                >
                                    Всего заказов
                                </p>
                                <p
                                    class="text-4xl font-mono tracking-tight text-white"
                                >
                                    {adminStats?.total_orders || 0}
                                </p>
                            </div>
                            <div
                                class="border border-white/10 p-8 bg-white/[0.01]"
                            >
                                <p
                                    class="text-[10px] uppercase tracking-widest font-mono text-white/40 mb-2"
                                >
                                    Забронировано столов
                                </p>
                                <p
                                    class="text-4xl font-mono tracking-tight text-white"
                                >
                                    {adminStats?.total_reservations || 0}
                                </p>
                            </div>
                            <div
                                class="border border-white/10 p-8 bg-white/[0.01]"
                            >
                                <p
                                    class="text-[10px] uppercase tracking-widest font-mono text-white/40 mb-2"
                                >
                                    Блюд в каталоге
                                </p>
                                <p
                                    class="text-4xl font-mono tracking-tight text-white"
                                >
                                    {adminStats?.total_products || 0}
                                </p>
                            </div>
                        </div>

                        <div
                            class="border border-white/10 p-8 bg-white/[0.01] space-y-4"
                        >
                            <h3
                                class="text-xs uppercase tracking-widest text-white/60 font-mono"
                            >
                                Проверка работоспособности системы
                            </h3>
                            <p
                                class="text-xs text-white/40 leading-relaxed font-light"
                            >
                                Все действия на главной странице (оформление
                                заказа, бронирование столиков, изменение блюд)
                                мгновенно синхронизируются с базой данных
                                PostgreSQL. Используйте меню управления слева
                                для проверки транзакционной целостности и
                                тестирования бэкенда на Go.
                            </p>
                        </div>
                    </div>
                {/if}

                {#if adminTab === "orders"}
                    <div class="space-y-8">
                        <div class="flex items-center gap-4">
                            <h2 class="text-3xl font-display font-light uppercase tracking-tight text-white">
                                Список <span class="font-serif italic text-white/40 lowercase">заказов</span>
                            </h2>
                            {#if livePolling}
                                <span class="flex items-center gap-1.5 text-[9px] font-mono uppercase tracking-widest text-emerald-400 border border-emerald-400/20 bg-emerald-400/10 px-2 py-0.5">
                                    <span class="w-1.5 h-1.5 bg-emerald-400 rounded-full animate-pulse"></span>
                                    Live
                                </span>
                            {/if}
                        </div>

                        <!-- Server-side search by phone / order ID -->
                        <div
                            class="flex flex-wrap items-end gap-4 border-b border-white/10 pb-6"
                        >
                            <div class="space-y-1">
                                <label
                                    class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                    >Поиск по телефону</label
                                >
                                <input
                                    type="text"
                                    bind:value={orderSearchPhone}
                                    placeholder="+7..."
                                    class="bg-white/5 border border-white/10 px-4 py-2 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm font-mono w-44"
                                />
                            </div>
                            <div class="space-y-1">
                                <label
                                    class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                    >Номер заказа (ID)</label
                                >
                                <input
                                    type="number"
                                    bind:value={orderSearchId}
                                    placeholder="ID"
                                    class="bg-white/5 border border-white/10 px-4 py-2 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm font-mono w-28"
                                />
                            </div>
                            <button
                                onclick={fetchAdminOrders}
                                class="px-6 py-2 bg-white text-black text-[10px] font-bold uppercase tracking-widest hover:bg-brand-red hover:text-white transition-colors cursor-pointer rounded-sm"
                                >Найти</button
                            >
                            <button
                                onclick={() => {
                                    orderSearchPhone = "";
                                    orderSearchId = "";
                                    fetchAdminOrders();
                                }}
                                class="px-4 py-2 border border-white/10 text-[10px] font-mono uppercase text-white/60 hover:text-white hover:border-white transition-colors cursor-pointer rounded-sm"
                                >Сбросить</button
                            >
                        </div>

                        <!-- Filter & Sort controls -->
                        <div
                            class="flex flex-wrap items-center justify-between gap-6 border-b border-white/10 pb-6"
                        >
                            <div class="flex flex-wrap items-center gap-6">
                                <!-- Status Filter -->
                                <div class="space-y-1">
                                    <label
                                        class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                        >Статус оплаты/доставки</label
                                    >
                                    <select
                                        bind:value={orderFilterStatus}
                                        class="bg-white/5 border border-white/10 px-4 py-2 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm"
                                    >
                                        <option
                                            value="all"
                                            class="text-black bg-white"
                                            >Все заказы</option
                                        >
                                        <option
                                            value="pending"
                                            class="text-black bg-white"
                                            >Ожидает оплаты (pending)</option
                                        >
                                        <option
                                            value="awaiting_delivery"
                                            class="text-black bg-white"
                                            >В очереди (awaiting_delivery)</option
                                        >
                                        <option
                                            value="preparing"
                                            class="text-black bg-white"
                                            >Готовится (preparing)</option
                                        >
                                        <option
                                            value="delivering"
                                            class="text-black bg-white"
                                            >В пути (delivering)</option
                                        >
                                        <option
                                            value="delivered"
                                            class="text-black bg-white"
                                            >Доставлен (delivered)</option
                                        >
                                        <option
                                            value="cancelled"
                                            class="text-black bg-white"
                                            >Отменен (cancelled)</option
                                        >
                                    </select>
                                </div>

                                <!-- Sort Field -->
                                <div class="space-y-1">
                                    <label
                                        class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                        >Сортировать по</label
                                    >
                                    <select
                                        bind:value={orderSortField}
                                        class="bg-white/5 border border-white/10 px-4 py-2 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm"
                                    >
                                        <option
                                            value="id"
                                            class="text-black bg-white"
                                            >Номер заказа (ID)</option
                                        >
                                        <option
                                            value="date"
                                            class="text-black bg-white"
                                            >Дата создания</option
                                        >
                                        <option
                                            value="total"
                                            class="text-black bg-white"
                                            >Итоговая стоимость</option
                                        >
                                    </select>
                                </div>
                            </div>

                            <!-- Sort Direction Toggle -->
                            <div class="space-y-1 self-end">
                                <button
                                    onclick={() =>
                                        (orderSortDirection =
                                            orderSortDirection === "asc"
                                                ? "desc"
                                                : "asc")}
                                    class="px-4 py-2 border border-white/10 hover:border-white text-xs font-mono uppercase text-white tracking-widest transition-all cursor-pointer"
                                >
                                    Направление: {orderSortDirection === "asc"
                                        ? "▲ Возрастание"
                                        : "▼ Убывание"}
                                </button>
                            </div>
                        </div>

                        {#if filteredOrders.length === 0}
                            <p class="text-sm font-mono text-white/30 italic">
                                Заказов с выбранными параметрами нет
                            </p>
                        {:else}
                            <div class="space-y-6">
                                {#each filteredOrders as order}
                                    <div
                                        class="border border-white/10 p-8 bg-white/[0.01] rounded-sm space-y-6"
                                    >
                                        <div
                                            class="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-4 border-b border-white/5"
                                        >
                                            <div>
                                                <h4
                                                    class="text-sm font-bold uppercase tracking-wider text-white"
                                                >
                                                    Заказ #{order.id}
                                                </h4>
                                                <p
                                                    class="text-[10px] font-mono text-white/40 mt-1"
                                                >
                                                    {new Date(
                                                        order.created_at,
                                                    ).toLocaleString()}
                                                </p>
                                            </div>
                                            <div
                                                class="flex items-center gap-4"
                                            >
                                                <span
                                                    class="text-[9px] uppercase tracking-widest font-mono text-white/40"
                                                    >Статус заказа:</span
                                                >
                                                <select
                                                    value={order.payment_status}
                                                    onchange={(e) =>
                                                        updateOrderStatus(
                                                            order.id,
                                                            (
                                                                e.target as HTMLSelectElement
                                                            ).value,
                                                        )}
                                                    class="bg-brand-gray border border-white/10 px-4 py-2 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm"
                                                >
                                                    <option
                                                        value="pending"
                                                        class="text-black bg-white"
                                                        >Ожидает оплаты
                                                        (pending)</option
                                                    >
                                                    <option
                                                        value="awaiting_delivery"
                                                        class="text-black bg-white"
                                                        >В очереди
                                                        (awaiting_delivery)</option
                                                    >
                                                    <option
                                                        value="preparing"
                                                        class="text-black bg-white"
                                                        >Готовится (preparing)</option
                                                    >
                                                    <option
                                                        value="delivering"
                                                        class="text-black bg-white"
                                                        >В пути (delivering)</option
                                                    >
                                                    <option
                                                        value="delivered"
                                                        class="text-black bg-white"
                                                        >Доставлен (delivered)</option
                                                    >
                                                    <option
                                                        value="cancelled"
                                                        class="text-black bg-white"
                                                        >Отменен (cancelled)</option
                                                    >
                                                </select>
                                            </div>
                                        </div>

                                        <div
                                            class="grid grid-cols-1 md:grid-cols-2 gap-8 text-xs font-light leading-relaxed"
                                        >
                                            <div class="space-y-2">
                                                <p
                                                    class="text-[9px] uppercase tracking-widest font-mono text-white/40"
                                                >
                                                    Клиент
                                                </p>
                                                <p class="text-white">
                                                    {order.customer_name}
                                                </p>
                                                <p class="text-white/60">
                                                    {order.phone}
                                                </p>
                                                <p class="text-white/60">
                                                    {order.address}
                                                </p>
                                            </div>
                                            <div class="space-y-4">
                                                <p
                                                    class="text-[9px] uppercase tracking-widest font-mono text-white/40"
                                                >
                                                    Позиции заказа
                                                </p>
                                                <div
                                                    class="space-y-2 border-l border-white/5 pl-4"
                                                >
                                                    {#each order.items as item}
                                                        <div
                                                            class="flex justify-between"
                                                        >
                                                            <span
                                                                class="text-white/80"
                                                                >{item.product_name ||
                                                                    `Товар #${item.product_id}`}
                                                                × {item.quantity}</span
                                                            >
                                                            <span
                                                                class="font-mono text-white/60"
                                                                >{item.price
                                                                    ? `${item.price * item.quantity} ₽`
                                                                    : ""}</span
                                                            >
                                                        </div>
                                                    {/each}
                                                    <div
                                                        class="flex justify-between pt-2 border-t border-white/5 font-bold font-mono text-white text-sm"
                                                    >
                                                        <span>Итого:</span>
                                                        <span
                                                            >{order.total_price} ₽</span
                                                        >
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                {/each}
                            </div>
                        {/if}
                    </div>
                {/if}

                {#if adminTab === "reservations"}
                    <div class="space-y-8">
                        <h2
                            class="text-3xl font-display font-light uppercase tracking-tight text-white"
                        >
                            Бронирование <span
                                class="font-serif italic text-white/40 lowercase"
                                >столов</span
                            >
                        </h2>

                        {#if adminReservations.length === 0}
                            <p class="text-sm font-mono text-white/30 italic">
                                Бронирований столов пока нет
                            </p>
                        {:else}
                            <div class="overflow-x-auto border border-white/10">
                                <table
                                    class="w-full text-left border-collapse text-xs font-light"
                                >
                                    <thead>
                                        <tr
                                            class="border-b border-white/10 bg-white/[0.02] text-[9px] uppercase tracking-widest font-mono text-white/40"
                                        >
                                            <th class="p-6">ID</th>
                                            <th class="p-6">Имя</th>
                                            <th class="p-6">Телефон</th>
                                            <th class="p-6">Дата</th>
                                            <th class="p-6">Время</th>
                                            <th class="p-6">Гостей</th>
                                            <th class="p-6">Комментарий</th>
                                            <th class="p-6">Статус</th>
                                            <th class="p-6 text-right"
                                                >Действия</th
                                            >
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {#each adminReservations as res}
                                            <tr
                                                class="border-b border-white/5 hover:bg-white/[0.01] transition-colors"
                                            >
                                                <td
                                                    class="p-6 font-mono text-white/40"
                                                    >{res.id}</td
                                                >
                                                <td
                                                    class="p-6 font-bold text-white"
                                                    >{res.customer_name}</td
                                                >
                                                <td class="p-6 text-white/60"
                                                    >{res.phone}</td
                                                >
                                                <td
                                                    class="p-6 text-white/80 font-mono"
                                                    >{res.reserve_date}</td
                                                >
                                                <td
                                                    class="p-6 text-white/80 font-mono"
                                                    >{res.reserve_time}</td
                                                >
                                                <td
                                                    class="p-6 text-white/80 font-mono"
                                                    >{res.guests_count}</td
                                                >
                                                <td
                                                    class="p-6 text-white/50 italic max-w-xs truncate"
                                                    >{res.comment || "-"}</td
                                                >
                                                <td class="p-6">
                                                    {#if res.status === "new"}
                                                        <span
                                                            class="px-2.5 py-1 text-[9px] font-bold uppercase tracking-widest font-mono rounded-sm bg-blue-500/10 text-blue-400 border border-blue-500/20"
                                                            >Новый</span
                                                        >
                                                    {:else if res.status === "confirmed"}
                                                        <span
                                                            class="px-2.5 py-1 text-[9px] font-bold uppercase tracking-widest font-mono rounded-sm bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                                                            >Подтвержден</span
                                                        >
                                                    {:else if res.status === "called"}
                                                        <span
                                                            class="px-2.5 py-1 text-[9px] font-bold uppercase tracking-widest font-mono rounded-sm bg-yellow-500/10 text-yellow-400 border border-yellow-500/20"
                                                            >Созвон</span
                                                        >
                                                    {:else if res.status === "completed"}
                                                        <span
                                                            class="px-2.5 py-1 text-[9px] font-bold uppercase tracking-widest font-mono rounded-sm bg-white/10 text-white/60 border border-white/10"
                                                            >Завершен</span
                                                        >
                                                    {:else if res.status === "cancelled"}
                                                        <span
                                                            class="px-2.5 py-1 text-[9px] font-bold uppercase tracking-widest font-mono rounded-sm bg-brand-red/10 text-brand-red border border-brand-red/20"
                                                            >Отменен</span
                                                        >
                                                    {/if}
                                                </td>
                                                <td
                                                    class="p-6 text-right font-mono"
                                                >
                                                    <select
                                                        value={res.status}
                                                        onchange={(e) =>
                                                            updateReservationStatus(
                                                                res.id,
                                                                (
                                                                    e.target as HTMLSelectElement
                                                                ).value,
                                                            )}
                                                        class="bg-brand-gray border border-white/10 px-3 py-1.5 text-[10px] text-white focus:outline-none focus:border-brand-red rounded-sm"
                                                    >
                                                        <option
                                                            value="new"
                                                            class="text-black bg-white"
                                                            >Новый</option
                                                        >
                                                        <option
                                                            value="confirmed"
                                                            class="text-black bg-white"
                                                            >Подтвердить</option
                                                        >
                                                        <option
                                                            value="called"
                                                            class="text-black bg-white"
                                                            >Позвонили</option
                                                        >
                                                        <option
                                                            value="completed"
                                                            class="text-black bg-white"
                                                            >Завершен</option
                                                        >
                                                        <option
                                                            value="cancelled"
                                                            class="text-black bg-white"
                                                            >Отменить</option
                                                        >
                                                    </select>
                                                </td>
                                            </tr>
                                        {/each}
                                    </tbody>
                                </table>
                            </div>
                        {/if}
                    </div>
                {/if}

                {#if adminTab === "menu"}
                    <div class="space-y-8">
                        <div class="flex items-center justify-between">
                            <h2
                                class="text-3xl font-display font-light uppercase tracking-tight text-white"
                            >
                                Каталог <span
                                    class="font-serif italic text-white/40 lowercase"
                                    >блюд</span
                                >
                            </h2>
                            <button
                                onclick={openCreateProductForm}
                                class="px-6 py-3 bg-white text-black text-[10px] font-bold uppercase tracking-widest rounded-sm hover:bg-brand-red hover:text-white transition-colors cursor-pointer flex items-center gap-2"
                            >
                                <PlusCircle class="w-4 h-4" />
                                <span>Добавить блюдо</span>
                            </button>
                        </div>

                        <!-- Product Form Modal -->
                        {#if isProdFormOpen}
                            <div
                                transition:fade
                                class="fixed inset-0 z-50 bg-black/80 backdrop-blur-md"
                            ></div>
                            <div
                                transition:scale
                                class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-lg bg-[#0a0a0a] border border-white/10 p-10 rounded-sm shadow-2xl overflow-y-auto max-h-[90vh]"
                            >
                                <div
                                    class="flex items-center justify-between mb-8"
                                >
                                    <h3
                                        class="text-xl font-display font-light uppercase tracking-tight text-white"
                                    >
                                        {#if editingProdId === null}Добавить
                                            товар{:else}Редактировать товар{/if}
                                    </h3>
                                    <button
                                        onclick={() => (isProdFormOpen = false)}
                                        class="text-white/40 hover:text-white"
                                        ><X class="w-5 h-5" /></button
                                    >
                                </div>

                                <div class="space-y-4">
                                    <div class="space-y-1">
                                        <label
                                            class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                            >Название блюда</label
                                        >
                                        <input
                                            type="text"
                                            bind:value={prodFormName}
                                            class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm"
                                        />
                                    </div>

                                    <div class="space-y-1">
                                        <label
                                            class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                            >Описание</label
                                        >
                                        <textarea
                                            bind:value={prodFormDesc}
                                            rows="3"
                                            class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm resize-none"
                                        ></textarea>
                                    </div>

                                    <div class="grid grid-cols-2 gap-4">
                                        <div class="space-y-1">
                                            <label
                                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                                >Цена (₽)</label
                                            >
                                            <input
                                                type="number"
                                                bind:value={prodFormPrice}
                                                class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm font-mono"
                                            />
                                        </div>
                                        <div class="space-y-1">
                                            <label
                                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                                >Категория</label
                                            >
                                            <select
                                                bind:value={prodFormCategory}
                                                class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm"
                                            >
                                                {#each categories as cat}
                                                    <option
                                                        value={cat.id}
                                                        class="text-black bg-white"
                                                        >{cat.name}</option
                                                    >
                                                {/each}
                                            </select>
                                        </div>
                                    </div>

                                    <div class="space-y-1">
                                        <label
                                            class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                            >Изображение товара</label
                                        >
                                        <input
                                            type="file"
                                            accept="image/*"
                                            onchange={(e) =>
                                                (prodFormImageFile =
                                                    e.currentTarget
                                                        .files?.[0] || null)}
                                            class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm font-mono"
                                        />
                                        {#if prodFormImage && !prodFormImageFile}
                                            <p
                                                class="text-[10px] text-white/40 mt-1"
                                            >
                                                Текущее изображение: {prodFormImage}
                                            </p>
                                        {/if}
                                    </div>

                                    <div class="grid grid-cols-2 gap-4">
                                        <div class="space-y-1">
                                            <label
                                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                                >Вес (г)</label
                                            >
                                            <input
                                                type="number"
                                                bind:value={prodFormWeight}
                                                class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm font-mono"
                                            />
                                        </div>
                                        <div class="space-y-1">
                                            <label
                                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                                >Калории (ккал)</label
                                            >
                                            <input
                                                type="number"
                                                bind:value={prodFormCalories}
                                                class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm font-mono"
                                            />
                                        </div>
                                    </div>

                                    <div class="space-y-1 mt-4">
                                        <label
                                            class="flex items-center gap-2 cursor-pointer"
                                        >
                                            <input
                                                type="checkbox"
                                                bind:checked={
                                                    prodFormIsAvailable
                                                }
                                                class="w-4 h-4 accent-brand-red"
                                            />
                                            <span
                                                class="text-[9px] uppercase tracking-widest text-white/80 font-mono"
                                                >В наличии</span
                                            >
                                        </label>
                                    </div>

                                    <div class="pt-4 flex gap-4">
                                        <button
                                            onclick={() =>
                                                (isProdFormOpen = false)}
                                            class="flex-1 bg-white/5 border border-white/10 hover:bg-white/10 text-white py-3 text-xs font-bold uppercase tracking-widest transition-colors cursor-pointer"
                                        >
                                            Отмена
                                        </button>
                                        <button
                                            onclick={handleSaveProduct}
                                            class="flex-1 bg-white text-black hover:bg-brand-red hover:text-white py-3 text-xs font-bold uppercase tracking-widest transition-colors cursor-pointer"
                                        >
                                            Сохранить
                                        </button>
                                    </div>
                                </div>
                            </div>
                        {/if}

                        <!-- Products List -->
                        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                            {#each productsList as prod}
                                <div
                                    class="border border-white/10 bg-white/[0.01] p-6 rounded-sm space-y-4 flex flex-col justify-between"
                                >
                                    <div class="space-y-2">
                                        <div
                                            class="w-full aspect-video bg-[#0a0a0a] rounded-sm overflow-hidden border border-white/5"
                                        >
                                            <img
                                                src={prod.image_url ||
                                                    "/images/placeholder.jpg"}
                                                alt={prod.name}
                                                class="w-full h-full object-contain"
                                            />
                                        </div>
                                        <h4
                                            class="text-xs font-bold uppercase tracking-wider text-white"
                                        >
                                            {prod.name}
                                        </h4>
                                        <p
                                            class="text-[10px] text-white/40 font-mono"
                                        >
                                            Вес: {prod.weight}г | Ккал: {prod.calories}
                                        </p>
                                        <p
                                            class="text-xs text-white/60 line-clamp-2 leading-relaxed"
                                        >
                                            {prod.description}
                                        </p>
                                    </div>

                                    <div
                                        class="flex items-center justify-between pt-4 border-t border-white/5 mt-4"
                                    >
                                        <span
                                            class="text-sm font-mono text-white"
                                            >{prod.price} ₽</span
                                        >
                                        <div class="flex gap-2">
                                            <button
                                                onclick={() =>
                                                    openEditProductForm(prod)}
                                                class="px-4 py-2 border border-white/10 hover:border-white text-[10px] font-bold uppercase tracking-wider text-white transition-all cursor-pointer"
                                            >
                                                Редактировать
                                            </button>
                                            <button
                                                onclick={() =>
                                                    handleDeleteProduct(
                                                        prod.id,
                                                    )}
                                                class="p-2 border border-white/10 hover:bg-brand-red hover:border-brand-red hover:text-white text-brand-red transition-all cursor-pointer"
                                            >
                                                <Trash2 class="w-4 h-4" />
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    </div>
                {/if}

                {#if adminTab === "blog"}
                    <div class="space-y-8">
                        <div class="flex items-center justify-between">
                            <h2
                                class="text-3xl font-display font-light uppercase tracking-tight text-white"
                            >
                                Управление <span
                                    class="font-serif italic text-white/40 lowercase"
                                    >блогом</span
                                >
                            </h2>
                            <button
                                onclick={openCreateBlogForm}
                                class="px-6 py-3 bg-white text-black text-[10px] font-bold uppercase tracking-widest rounded-sm hover:bg-brand-red hover:text-white transition-colors cursor-pointer flex items-center gap-2"
                            >
                                <PlusCircle class="w-4 h-4" />
                                <span>Новая статья</span>
                            </button>
                        </div>

                        <!-- Blog Post Form Modal -->
                        {#if isBlogFormOpen}
                            <div
                                transition:fade
                                class="fixed inset-0 z-50 bg-black/80 backdrop-blur-md"
                            ></div>
                            <div
                                transition:scale
                                class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-2xl bg-[#0a0a0a] border border-white/10 p-10 rounded-sm shadow-2xl overflow-y-auto max-h-[90vh]"
                            >
                                <div
                                    class="flex items-center justify-between mb-8"
                                >
                                    <h3
                                        class="text-xl font-display font-light uppercase tracking-tight text-white"
                                    >
                                        {editingBlogId === null
                                            ? "Новая статья"
                                            : "Редактировать статью"}
                                    </h3>
                                    <button
                                        onclick={() => (isBlogFormOpen = false)}
                                        class="text-white/40 hover:text-white"
                                        ><X class="w-5 h-5" /></button
                                    >
                                </div>

                                <div class="space-y-4">
                                    <div class="space-y-1">
                                        <label
                                            class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                            >Заголовок</label
                                        >
                                        <input
                                            type="text"
                                            bind:value={blogFormTitle}
                                            class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm"
                                        />
                                    </div>
                                    <div class="space-y-1">
                                        <label
                                            class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                            >Подзаголовок</label
                                        >
                                        <input
                                            type="text"
                                            bind:value={blogFormSubtitle}
                                            class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm"
                                        />
                                    </div>
                                    <div class="space-y-1">
                                        <label
                                            class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                            >Содержание</label
                                        >
                                        <textarea
                                            bind:value={blogFormContent}
                                            rows="8"
                                            class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm resize-none"
                                        ></textarea>
                                    </div>
                                    <div class="grid grid-cols-2 gap-4">
                                        <div class="space-y-1">
                                            <label
                                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                                >Тег</label
                                            >
                                            <input
                                                type="text"
                                                bind:value={blogFormTag}
                                                placeholder="Традиции"
                                                class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm"
                                            />
                                        </div>
                                        <div class="space-y-1">
                                            <label
                                                class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                                >Время чтения</label
                                            >
                                            <input
                                                type="text"
                                                bind:value={blogFormReadTime}
                                                placeholder="5 мин чтения"
                                                class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm"
                                            />
                                        </div>
                                    </div>
                                    <div class="space-y-1">
                                        <label
                                            class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                            >Изображение</label
                                        >
                                        <input
                                            type="file"
                                            accept="image/*"
                                            onchange={(e) =>
                                                (blogFormImageFile =
                                                    e.currentTarget
                                                        .files?.[0] || null)}
                                            class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm font-mono"
                                        />
                                        {#if blogFormImage && !blogFormImageFile}
                                            <p
                                                class="text-[10px] text-white/40 mt-1"
                                            >
                                                Текущее: {blogFormImage}
                                            </p>
                                        {/if}
                                    </div>
                                    <div class="space-y-1 mt-2">
                                        <label
                                            class="flex items-center gap-2 cursor-pointer"
                                        >
                                            <input
                                                type="checkbox"
                                                bind:checked={
                                                    blogFormIsPublished
                                                }
                                                class="w-4 h-4 accent-brand-red"
                                            />
                                            <span
                                                class="text-[9px] uppercase tracking-widest text-white/80 font-mono"
                                                >Опубликовано</span
                                            >
                                        </label>
                                    </div>
                                    <div class="pt-4 flex gap-4">
                                        <button
                                            onclick={() =>
                                                (isBlogFormOpen = false)}
                                            class="flex-1 bg-white/5 border border-white/10 hover:bg-white/10 text-white py-3 text-xs font-bold uppercase tracking-widest transition-colors cursor-pointer"
                                            >Отмена</button
                                        >
                                        <button
                                            onclick={handleSaveBlogPost}
                                            class="flex-1 bg-white text-black hover:bg-brand-red hover:text-white py-3 text-xs font-bold uppercase tracking-widest transition-colors cursor-pointer"
                                            >Сохранить</button
                                        >
                                    </div>
                                </div>
                            </div>
                        {/if}

                        <!-- Blog Posts List -->
                        {#if adminBlogPosts.length === 0}
                            <p class="text-sm font-mono text-white/30 italic">
                                Статей пока нет
                            </p>
                        {:else}
                            <div class="space-y-4">
                                {#each adminBlogPosts as post}
                                    <div
                                        class="border border-white/10 bg-white/[0.01] p-6 rounded-sm flex items-start gap-6"
                                    >
                                        <div
                                            class="w-24 h-16 flex-shrink-0 bg-[#0a0a0a] border border-white/5 rounded-sm overflow-hidden"
                                        >
                                            <img
                                                src={post.image_url ||
                                                    "/images/placeholder.jpg"}
                                                alt={post.title}
                                                class="w-full h-full object-cover"
                                            />
                                        </div>
                                        <div class="flex-1 min-w-0">
                                            <div
                                                class="flex items-center gap-3 mb-1"
                                            >
                                                <h4
                                                    class="text-sm font-bold uppercase tracking-wider text-white truncate"
                                                >
                                                    {post.title}
                                                </h4>
                                                {#if !post.is_published}
                                                    <span
                                                        class="text-[9px] font-mono uppercase px-2 py-0.5 bg-yellow-500/10 text-yellow-400 border border-yellow-500/20 flex-shrink-0"
                                                        >Черновик</span
                                                    >
                                                {/if}
                                            </div>
                                            <p
                                                class="text-[10px] text-white/40 font-mono mb-1"
                                            >
                                                {formatBlogDate(
                                                    post.created_at,
                                                )}{post.tag
                                                    ? ` · ${post.tag}`
                                                    : ""}
                                            </p>
                                            <p
                                                class="text-xs text-white/50 line-clamp-1"
                                            >
                                                {post.subtitle}
                                            </p>
                                        </div>
                                        <div class="flex gap-2 flex-shrink-0">
                                            <button
                                                onclick={() =>
                                                    openEditBlogForm(post)}
                                                class="px-4 py-2 border border-white/10 hover:border-white text-[10px] font-bold uppercase tracking-wider text-white transition-all cursor-pointer"
                                                >Редактировать</button
                                            >
                                            <button
                                                onclick={() =>
                                                    handleDeleteBlogPost(
                                                        post.id,
                                                    )}
                                                class="p-2 border border-white/10 hover:bg-brand-red hover:border-brand-red hover:text-white text-brand-red transition-all cursor-pointer"
                                                ><Trash2
                                                    class="w-4 h-4"
                                                /></button
                                            >
                                        </div>
                                    </div>
                                {/each}
                            </div>
                        {/if}
                    </div>
                {/if}

                {#if adminTab === "categories"}
                    <div class="space-y-8">
                        <div class="flex items-center justify-between">
                            <h2
                                class="text-3xl font-display font-light uppercase tracking-tight text-white"
                            >
                                Управление <span
                                    class="font-serif italic text-white/40 lowercase"
                                    >категориями</span
                                >
                            </h2>
                            <button
                                onclick={openCreateCategoryForm}
                                class="px-6 py-3 bg-white text-black text-[10px] font-bold uppercase tracking-widest rounded-sm hover:bg-brand-red hover:text-white transition-colors cursor-pointer flex items-center gap-2"
                            >
                                <PlusCircle class="w-4 h-4" />
                                <span>Добавить категорию</span>
                            </button>
                        </div>

                        <!-- Category Form Modal -->
                        {#if isCatFormOpen}
                            <div
                                transition:fade
                                class="fixed inset-0 z-50 bg-black/80 backdrop-blur-md"
                            ></div>
                            <div
                                transition:scale
                                class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-lg bg-[#0a0a0a] border border-white/10 p-10 rounded-sm shadow-2xl"
                            >
                                <div
                                    class="flex items-center justify-between mb-8"
                                >
                                    <h3
                                        class="text-xl font-display font-light uppercase tracking-tight text-white"
                                    >
                                        {#if editingCatId === null}Добавить
                                            категорию{:else}Редактировать
                                            категорию{/if}
                                    </h3>
                                    <button
                                        onclick={() => (isCatFormOpen = false)}
                                        class="text-white/40 hover:text-white"
                                        ><X class="w-5 h-5" /></button
                                    >
                                </div>

                                <div class="space-y-4">
                                    <div class="space-y-1">
                                        <label
                                            class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                            >Название категории</label
                                        >
                                        <input
                                            type="text"
                                            bind:value={catFormName}
                                            placeholder="Супы"
                                            class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm"
                                        />
                                    </div>

                                    <div class="space-y-1">
                                        <label
                                            class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                            >Слаг (Slug URL-friendly)</label
                                        >
                                        <input
                                            type="text"
                                            bind:value={catFormSlug}
                                            placeholder="soups"
                                            class="w-full bg-white/5 border border-white/10 px-4 py-3 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm font-mono"
                                        />
                                    </div>

                                    <div class="pt-4 flex gap-4">
                                        <button
                                            onclick={() =>
                                                (isCatFormOpen = false)}
                                            class="flex-1 bg-white/5 border border-white/10 hover:bg-white/10 text-white py-3 text-xs font-bold uppercase tracking-widest transition-colors cursor-pointer"
                                        >
                                            Отмена
                                        </button>
                                        <button
                                            onclick={handleSaveCategory}
                                            class="flex-1 bg-white text-black hover:bg-brand-red hover:text-white py-3 text-xs font-bold uppercase tracking-widest transition-colors cursor-pointer"
                                        >
                                            Сохранить
                                        </button>
                                    </div>
                                </div>
                            </div>
                        {/if}

                        <!-- Category List Table -->
                        {#if categories.length === 0}
                            <p class="text-sm font-mono text-white/30 italic">
                                Категорий пока нет
                            </p>
                        {:else}
                            <div class="overflow-x-auto border border-white/10">
                                <table
                                    class="w-full text-left border-collapse text-xs font-light"
                                >
                                    <thead>
                                        <tr
                                            class="border-b border-white/10 bg-white/[0.02] text-[9px] uppercase tracking-widest font-mono text-white/40"
                                        >
                                            <th class="p-6">ID</th>
                                            <th class="p-6">Название</th>
                                            <th class="p-6">Слаг</th>
                                            <th class="p-6 text-right"
                                                >Действия</th
                                            >
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {#each categories as cat}
                                            <tr
                                                class="border-b border-white/5 hover:bg-white/[0.01] transition-colors"
                                            >
                                                <td
                                                    class="p-6 font-mono text-white/40"
                                                    >{cat.id}</td
                                                >
                                                <td
                                                    class="p-6 font-bold text-white"
                                                    >{cat.name}</td
                                                >
                                                <td
                                                    class="p-6 text-white/60 font-mono"
                                                    >{cat.slug}</td
                                                >
                                                <td class="p-6 text-right">
                                                    <div
                                                        class="flex items-center justify-end gap-2"
                                                    >
                                                        <button
                                                            onclick={() =>
                                                                openEditCategoryForm(
                                                                    cat,
                                                                )}
                                                            class="px-4 py-2 border border-white/10 hover:border-white text-[10px] font-bold uppercase tracking-wider text-white transition-all cursor-pointer"
                                                        >
                                                            Редактировать
                                                        </button>
                                                        <button
                                                            onclick={() =>
                                                                handleDeleteCategory(
                                                                    cat.id,
                                                                )}
                                                            class="p-2 border border-white/10 hover:bg-brand-red hover:border-brand-red hover:text-white text-brand-red transition-all cursor-pointer"
                                                        >
                                                            <Trash2
                                                                class="w-4 h-4"
                                                            />
                                                        </button>
                                                    </div>
                                                </td>
                                            </tr>
                                        {/each}
                                    </tbody>
                                </table>
                            </div>
                        {/if}
                    </div>
                {/if}
                {#if adminTab === "users"}
                    <div class="space-y-8">
                        <h2
                            class="text-3xl font-display font-light uppercase tracking-tight text-white"
                        >
                            Пользователи <span
                                class="font-serif italic text-white/40 lowercase"
                                >системы</span
                            >
                        </h2>

                        <!-- Search -->
                        <div
                            class="flex items-end gap-4 border-b border-white/10 pb-6"
                        >
                            <div class="space-y-1">
                                <label
                                    class="text-[9px] uppercase tracking-widest text-white/40 block font-mono"
                                    >Поиск по телефону</label
                                >
                                <input
                                    type="text"
                                    bind:value={userSearchPhone}
                                    placeholder="+7..."
                                    class="bg-white/5 border border-white/10 px-4 py-2 text-xs text-white focus:outline-none focus:border-brand-red rounded-sm font-mono w-52"
                                />
                            </div>
                            {#if userSearchPhone}
                                <button
                                    onclick={() => (userSearchPhone = "")}
                                    class="px-4 py-2 border border-white/10 text-[10px] font-mono uppercase text-white/60 hover:text-white hover:border-white transition-colors cursor-pointer rounded-sm"
                                    >Сбросить</button
                                >
                            {/if}
                        </div>

                        {#if filteredAdminUsers.length === 0}
                            <p class="text-sm font-mono text-white/30 italic">
                                Пользователи не найдены
                            </p>
                        {:else}
                            <div class="overflow-x-auto border border-white/10">
                                <table
                                    class="w-full text-left border-collapse text-xs font-light"
                                >
                                    <thead>
                                        <tr
                                            class="border-b border-white/10 bg-white/[0.02] text-[9px] uppercase tracking-widest font-mono text-white/40"
                                        >
                                            <th class="p-4">ID</th>
                                            <th class="p-4">Имя</th>
                                            <th class="p-4">Телефон</th>
                                            <th class="p-4">Email</th>
                                            <th class="p-4">Адрес</th>
                                            <th class="p-4">Роль</th>
                                            <th class="p-4">Дата регистрации</th
                                            >
                                            {#if currentUser?.role === "super_admin"}
                                                <th class="p-4 text-right"
                                                    >Действия</th
                                                >
                                            {/if}
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {#each filteredAdminUsers as u}
                                            <tr
                                                class="border-b border-white/5 hover:bg-white/[0.01] transition-colors"
                                            >
                                                <td
                                                    class="p-4 font-mono text-white/40"
                                                    >{u.id}</td
                                                >
                                                <td
                                                    class="p-4 font-bold text-white"
                                                    >{u.name}</td
                                                >
                                                <td
                                                    class="p-4 text-white/80 font-mono"
                                                    >{u.phone}</td
                                                >
                                                <td class="p-4 text-white/50"
                                                    >{u.email ?? "—"}</td
                                                >
                                                <td class="p-4 text-white/50"
                                                    >{u.default_address ??
                                                        "—"}</td
                                                >
                                                <td class="p-4">
                                                    <span
                                                        class="px-2 py-1 text-[9px] font-mono uppercase tracking-widest rounded-sm {u.role ===
                                                        'super_admin'
                                                            ? 'bg-brand-red text-white'
                                                            : u.role === 'admin'
                                                              ? 'bg-white/20 text-white'
                                                              : 'bg-white/5 text-white/50'}"
                                                    >
                                                        {u.role}
                                                    </span>
                                                </td>
                                                <td
                                                    class="p-4 font-mono text-white/40 text-[10px]"
                                                    >{new Date(
                                                        u.created_at,
                                                    ).toLocaleDateString(
                                                        "ru-RU",
                                                    )}</td
                                                >
                                                {#if currentUser?.role === "super_admin"}
                                                    <td class="p-4 text-right">
                                                        <div
                                                            class="flex items-center justify-end gap-2"
                                                        >
                                                            {#if u.id !== currentUser?.id}
                                                                <select
                                                                    value={u.role}
                                                                    onchange={(
                                                                        e,
                                                                    ) =>
                                                                        handleUpdateUserRole(
                                                                            u.id,
                                                                            (
                                                                                e.target as HTMLSelectElement
                                                                            )
                                                                                .value,
                                                                        )}
                                                                    class="bg-brand-gray border border-white/10 px-2 py-1 text-[10px] text-white focus:outline-none focus:border-brand-red rounded-sm cursor-pointer"
                                                                >
                                                                    <option
                                                                        value="user"
                                                                        class="text-black bg-white"
                                                                        >user</option
                                                                    >
                                                                    <option
                                                                        value="admin"
                                                                        class="text-black bg-white"
                                                                        >admin</option
                                                                    >
                                                                    <option
                                                                        value="super_admin"
                                                                        class="text-black bg-white"
                                                                        >super_admin</option
                                                                    >
                                                                </select>
                                                                <button
                                                                    onclick={() =>
                                                                        handleDeleteUser(
                                                                            u.id,
                                                                        )}
                                                                    class="p-2 border border-white/10 hover:bg-brand-red hover:border-brand-red hover:text-white text-brand-red transition-all cursor-pointer"
                                                                    title="Удалить пользователя"
                                                                >
                                                                    <Trash2
                                                                        class="w-3 h-3"
                                                                    />
                                                                </button>
                                                            {:else}
                                                                <span
                                                                    class="text-[9px] font-mono text-white/30"
                                                                    >вы</span
                                                                >
                                                            {/if}
                                                        </div>
                                                    </td>
                                                {/if}
                                            </tr>
                                        {/each}
                                    </tbody>
                                </table>
                            </div>
                        {/if}
                    </div>
                {/if}

                {#if adminTab === "audit-log"}
                    <div class="space-y-8">
                        <div class="flex items-center justify-between">
                            <h2
                                class="text-3xl font-display font-light uppercase tracking-tight text-white"
                            >
                                Журнал <span
                                    class="font-serif italic text-white/40 lowercase"
                                    >аудита</span
                                >
                            </h2>
                            <button
                                onclick={fetchAuditLog}
                                class="px-4 py-2 border border-white/10 text-[10px] font-mono uppercase text-white/60 hover:text-white hover:border-white transition-colors cursor-pointer"
                                >Обновить</button
                            >
                        </div>

                        {#if auditLog.length === 0}
                            <p class="text-sm font-mono text-white/30 italic">
                                Записей в журнале пока нет
                            </p>
                        {:else}
                            <div class="overflow-x-auto border border-white/10">
                                <table
                                    class="w-full text-left border-collapse text-xs font-light"
                                >
                                    <thead>
                                        <tr
                                            class="border-b border-white/10 bg-white/[0.02] text-[9px] uppercase tracking-widest font-mono text-white/40"
                                        >
                                            <th class="p-4">ID</th>
                                            <th class="p-4">Время</th>
                                            <th class="p-4">Действие</th>
                                            <th class="p-4">Тип</th>
                                            <th class="p-4">ID объекта</th>
                                            <th class="p-4">Детали</th>
                                            <th class="p-4">Администратор</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {#each auditLog as entry}
                                            <tr
                                                class="border-b border-white/5 hover:bg-white/[0.01] transition-colors"
                                            >
                                                <td
                                                    class="p-4 font-mono text-white/30"
                                                    >{entry.id}</td
                                                >
                                                <td
                                                    class="p-4 font-mono text-white/50 text-[10px] whitespace-nowrap"
                                                    >{new Date(
                                                        entry.created_at,
                                                    ).toLocaleString(
                                                        "ru-RU",
                                                    )}</td
                                                >
                                                <td class="p-4">
                                                    <span
                                                        class="px-2 py-1 text-[9px] font-mono uppercase tracking-widest bg-white/5 text-white/80 rounded-sm"
                                                        >{entry.action}</span
                                                    >
                                                </td>
                                                <td
                                                    class="p-4 font-mono text-white/50"
                                                    >{entry.entity_type ||
                                                        "—"}</td
                                                >
                                                <td
                                                    class="p-4 font-mono text-white/40"
                                                    >{entry.entity_id ??
                                                        "—"}</td
                                                >
                                                <td
                                                    class="p-4 text-white/60 max-w-xs truncate"
                                                    >{entry.details || "—"}</td
                                                >
                                                <td
                                                    class="p-4 font-mono text-white/40"
                                                    >{entry.admin_id ?? "—"}</td
                                                >
                                            </tr>
                                        {/each}
                                    </tbody>
                                </table>
                            </div>
                        {/if}
                    </div>
                {/if}
            </main>
        </div>
    {/if}

    <!-- Dish Detail Modal -->
    {#if selectedDish}
        <div
            transition:fade={{ duration: 150 }}
            onclick={closeDishModal}
            class="fixed inset-0 z-[70] bg-black/85 backdrop-blur-sm cursor-pointer"
        ></div>
        <div
            transition:fade={{ duration: 150 }}
            onclick={closeDishModal}
            class="fixed inset-0 z-[71] flex items-center justify-center p-4 pointer-events-none"
        >
            <div
                transition:scale={{ start: 0.96, duration: 200 }}
                onclick={(e) => e.stopPropagation()}
                class="bg-[#0c0c0c] border border-white/10 w-full max-w-lg max-h-[90vh] overflow-y-auto pointer-events-auto shadow-2xl"
            >
                <!-- Image -->
                <div class="relative aspect-video overflow-hidden bg-[#080808]">
                    <img
                        src={selectedDish.image}
                        alt={selectedDish.name}
                        class="w-full h-full object-cover opacity-80"
                    />
                    <button
                        onclick={closeDishModal}
                        class="absolute top-4 right-4 p-2 bg-black/60 backdrop-blur-sm border border-white/10 text-white/60 hover:text-white hover:bg-black/80 transition-all cursor-pointer"
                        aria-label="Закрыть"
                    >
                        <X class="w-4 h-4" />
                    </button>
                    {#if selectedDish.categoryName}
                        <div class="absolute bottom-0 left-0 px-4 py-2 bg-gradient-to-t from-black/80 to-transparent">
                            <span class="text-[9px] font-mono uppercase tracking-widest text-white/50">{selectedDish.categoryName}</span>
                        </div>
                    {/if}
                </div>

                <!-- Content -->
                <div class="p-8 space-y-6">
                    <h2 class="text-2xl font-display font-light uppercase tracking-tight text-white leading-tight">
                        {selectedDish.name}
                    </h2>

                    {#if selectedDish.description}
                        <p class="text-sm text-white/60 leading-relaxed font-light">
                            {selectedDish.description}
                        </p>
                    {/if}

                    <!-- Stats grid -->
                    <div class="grid grid-cols-3 gap-4 py-5 border-t border-b border-white/5">
                        <div>
                            <p class="text-[9px] uppercase tracking-widest text-white/30 font-mono mb-1">Цена</p>
                            <p class="text-xl font-mono tracking-tight text-white">{selectedDish.price} ₽</p>
                        </div>
                        {#if selectedDish.calories}
                            <div>
                                <p class="text-[9px] uppercase tracking-widest text-white/30 font-mono mb-1">Калории</p>
                                <p class="text-xl font-mono tracking-tight text-white">{selectedDish.calories} <span class="text-xs text-white/40">ккал</span></p>
                            </div>
                        {/if}
                        {#if selectedDish.weight}
                            <div>
                                <p class="text-[9px] uppercase tracking-widest text-white/30 font-mono mb-1">Вес</p>
                                <p class="text-xl font-mono tracking-tight text-white">{selectedDish.weight} <span class="text-xs text-white/40">г</span></p>
                            </div>
                        {/if}
                    </div>

                    <!-- Cart action -->
                    <div class="flex items-center gap-4">
                        {#if getQuantity(selectedDish.id) > 0}
                            <div class="flex items-center gap-5 bg-white/5 border border-white/10 px-6 py-3">
                                <button
                                    onclick={() => removeFromCart(selectedDish!.id)}
                                    class="text-white/40 hover:text-white transition-colors cursor-pointer"
                                >
                                    <Minus class="w-4 h-4" />
                                </button>
                                <span class="font-mono text-sm text-white min-w-[1ch] text-center">{getQuantity(selectedDish.id)}</span>
                                <button
                                    onclick={() => addToCart(selectedDish!.id)}
                                    class="text-brand-red hover:text-red-400 transition-colors cursor-pointer"
                                >
                                    <Plus class="w-4 h-4" />
                                </button>
                            </div>
                            <button
                                onclick={closeDishModal}
                                class="flex-1 py-3 text-center text-[10px] font-bold uppercase tracking-[0.2em] bg-white text-black hover:bg-brand-red hover:text-white transition-colors cursor-pointer"
                            >
                                В корзину ({getQuantity(selectedDish.id)})
                            </button>
                        {:else}
                            <button
                                onclick={() => { addToCart(selectedDish!.id); }}
                                class="flex-1 py-3 text-[10px] font-bold uppercase tracking-[0.2em] bg-white text-black hover:bg-brand-red hover:text-white transition-all cursor-pointer"
                            >
                                Добавить в корзину
                            </button>
                        {/if}
                    </div>
                </div>
            </div>
        </div>
    {/if}
</div>
