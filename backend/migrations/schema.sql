CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    default_address TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);

CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    category_id INT REFERENCES categories(id) ON DELETE SET NULL,
    image_url VARCHAR(255),
    weight INT,
    calories INT,
    is_available BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    customer_name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    address TEXT NOT NULL,
    total_price NUMERIC(10, 2) NOT NULL,
    payment_status VARCHAR(50) NOT NULL,
    payment_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(id) ON DELETE CASCADE,
    product_id INT REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL,
    price NUMERIC(10, 2) NOT NULL
);

CREATE TABLE IF NOT EXISTS reservations (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    customer_name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    reserve_date DATE NOT NULL,
    reserve_time TIME NOT NULL,
    guests_count INT NOT NULL,
    comment TEXT,
    status VARCHAR(50) DEFAULT 'new'
);

INSERT INTO categories (id, name, slug) VALUES
(1, 'Буузы', 'buuzy'),
(2, 'Супы', 'soups'),
(3, 'Салаты', 'salads'),
(4, 'Десерты', 'desserts'),
(5, 'Напитки', 'drinks')
ON CONFLICT (id) DO NOTHING;

INSERT INTO products (id, name, description, price, category_id, image_url, weight, calories) VALUES
(1, 'Буузы Классические', 'Традиционные бурятские буузы из смеси говядины и свинины, с ароматным бульоном внутри.', 90.00, 1, '/images/buuzy-classic.jpg', 75, 180),
(2, 'Буузы с бараниной', 'Сочные буузы со стопроцентной рубленой бараниной и луком.', 100.00, 1, '/images/buuzy-lamb.jpg', 75, 190),
(3, 'Шулэн', 'Домашний суп-лапша с говядиной и свежей зеленью.', 250.00, 2, '/images/shulen.jpg', 350, 310),
(4, 'Бухлёр', 'Наваристый традиционный бульон с крупным куском нежной говядины и картофелем.', 290.00, 2, '/images/buhler.jpg', 400, 390),
(5, 'Салат Азиатский', 'Свежий салат с битыми огурцами, кунжутом и легкой заправкой.', 180.00, 3, '/images/asian-salad.jpg', 150, 120),
(6, 'Черемуховый пирог', 'Нежный пирог из черемуховой муки со сметанным кремом.', 220.00, 4, '/images/bird-cherry-cake.jpg', 120, 280),
(7, 'Чай с молоком', 'Традиционный бурятский чай с добавлением молока и щепотки соли по вкусу.', 70.00, 5, '/images/milky-tea.jpg', 200, 80),
(8, 'Облепиховый морс', 'Освежающий витаминный морс из натуральной облепихи.', 90.00, 5, '/images/sea-buckthorn-drink.jpg', 250, 95)
ON CONFLICT (id) DO NOTHING;

ALTER TABLE reservations ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'new';
ALTER TABLE products ADD COLUMN IF NOT EXISTS is_available BOOLEAN DEFAULT TRUE;
