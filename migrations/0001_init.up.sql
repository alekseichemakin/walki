-- ENUM TYPES
CREATE TYPE route_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE order_status AS ENUM ('pending', 'paid', 'cancelled', 'refunded');
CREATE TYPE progress_status AS ENUM ('in_progress', 'completed', 'abandoned');
CREATE TYPE payment_status AS ENUM ('pending', 'success', 'failed');
CREATE TYPE media_type AS ENUM ('image', 'video', 'audio', 'document');
CREATE TYPE point_status AS ENUM ('active', 'inactive');
CREATE TYPE review_status AS ENUM ('pending', 'published', 'rejected');

-- TABLES
CREATE TABLE users
(
    id          SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username    VARCHAR(255),
    full_name   VARCHAR(255),
    phone       VARCHAR(20),
    email       VARCHAR(255),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE routes
(
    id         SERIAL PRIMARY KEY,
    status     route_status DEFAULT 'draft',
    is_visible BOOLEAN      DEFAULT true,
    created_by INT,
    created_at TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE route_versions
(
    id               SERIAL PRIMARY KEY,
    route_id         INT,
    version_number   INT,
    title            VARCHAR(255),
    description      TEXT,
    duration_minutes INT,
    length_km        NUMERIC(5, 2),
    theme            VARCHAR(100),
    price            NUMERIC(10, 2) NOT NULL,
    city             VARCHAR(100)   NOT NULL,
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE route_points
(
    id                   SERIAL PRIMARY KEY,
    version_id           INT,
    title                VARCHAR(255) NOT NULL,
    description          TEXT,
    latitude             DECIMAL(10, 8),
    longitude            DECIMAL(11, 8),
    order_index          INT          NOT NULL,
    status               point_status DEFAULT 'active',
    arrival_instructions TEXT,
    created_at           TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE media
(
    id          SERIAL PRIMARY KEY,
    type        media_type NOT NULL,
    url         TEXT       NOT NULL,
    filename    TEXT,
    size_bytes  BIGINT,
    description TEXT,
    metadata    JSONB,
    uploaded_by INT,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_public   BOOLEAN   DEFAULT true
);

CREATE TABLE route_point_media
(
    route_point_id INT NOT NULL,
    media_id       INT NOT NULL,
    PRIMARY KEY (route_point_id, media_id)
);

CREATE TABLE orders
(
    id            SERIAL PRIMARY KEY,
    user_id       INT            NOT NULL,
    version_id    INT,
    route_id      INT            NOT NULL,
    status        order_status DEFAULT 'pending',
    amount        DECIMAL(10, 2) NOT NULL,
    created_at    TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    paid_at       TIMESTAMP,
    access_expiry TIMESTAMP
);

CREATE TABLE payments
(
    id               SERIAL PRIMARY KEY,
    order_id         INT            NOT NULL,
    payment_provider VARCHAR(50),
    transaction_id   VARCHAR(255),
    amount           DECIMAL(10, 2) NOT NULL,
    currency         VARCHAR(10)    DEFAULT 'RUB',
    status           payment_status DEFAULT 'pending',
    created_at       TIMESTAMP      DEFAULT CURRENT_TIMESTAMP,
    confirmed_at     TIMESTAMP
);

CREATE TABLE route_progress
(
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT    NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    route_id    BIGINT    NOT NULL REFERENCES routes (id) ON DELETE CASCADE, -- агрегат (удобно искать все прогрессы по маршруту)
    version_id  BIGINT    NOT NULL REFERENCES route_versions (id) ON DELETE CASCADE,
    current_idx INT       NOT NULL DEFAULT 0,                                -- индекс следующей точки к показу
    started_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMP NULL,
    content_msg_id BIGINT,
    voice_msg_id   BIGINT,
    UNIQUE (user_id, version_id)                                             -- один активный прогресс на версию
);

CREATE TABLE route_point_logs
(
    id          SERIAL PRIMARY KEY,
    progress_id INT NOT NULL,
    point_id    INT NOT NULL,
    visited_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE reviews
(
    id         SERIAL PRIMARY KEY,
    order_id   INT      NOT NULL,
    route_id   INT      NOT NULL,
    user_id    INT      NOT NULL,
    rating     SMALLINT NOT NULL,
    comment    TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE route_version_media
(
    route_version_id INT NOT NULL,
    media_id         INT NOT NULL,
    is_cover         BOOLEAN   DEFAULT FALSE,
    display_order    INT       DEFAULT 0,
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (route_version_id, media_id),
    FOREIGN KEY (route_version_id) REFERENCES route_versions (id) ON DELETE CASCADE,
    FOREIGN KEY (media_id) REFERENCES media (id) ON DELETE CASCADE
);

-- FOREIGN KEYS
ALTER TABLE route_versions
    ADD FOREIGN KEY (route_id) REFERENCES routes (id) ON DELETE CASCADE;
ALTER TABLE route_points
    ADD FOREIGN KEY (version_id) REFERENCES route_versions (id) ON DELETE CASCADE;
ALTER TABLE route_point_media
    ADD FOREIGN KEY (route_point_id) REFERENCES route_points (id) ON DELETE CASCADE;
ALTER TABLE route_point_media
    ADD FOREIGN KEY (media_id) REFERENCES media (id) ON DELETE CASCADE;
ALTER TABLE media
    ADD FOREIGN KEY (uploaded_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE orders
    ADD FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
ALTER TABLE orders
    ADD FOREIGN KEY (route_id) REFERENCES routes (id) ON DELETE RESTRICT;
ALTER TABLE orders
    ADD FOREIGN KEY (version_id) REFERENCES route_versions (id) ON DELETE RESTRICT;
ALTER TABLE payments
    ADD FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE;
ALTER TABLE route_point_logs
    ADD FOREIGN KEY (progress_id) REFERENCES route_progress (id) ON DELETE CASCADE;
ALTER TABLE route_point_logs
    ADD FOREIGN KEY (point_id) REFERENCES route_points (id) ON DELETE SET NULL;
ALTER TABLE reviews
    ADD FOREIGN KEY (route_id) REFERENCES routes (id) ON DELETE CASCADE;
ALTER TABLE reviews
    ADD FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
ALTER TABLE reviews
    ADD FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE;
