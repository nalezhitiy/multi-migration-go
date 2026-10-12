-- +goose Up
CREATE TABLE products (
      id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
      name VARCHAR(255) NOT NULL,
      description TEXT NOT NULL,
      img VARCHAR(2048) NULL,
      price DECIMAL(12, 2) NOT NULL,
      rating DECIMAL(3, 2) NULL,

      created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
          ON UPDATE CURRENT_TIMESTAMP,

      PRIMARY KEY (id),

      CONSTRAINT chk_products_price
          CHECK (price >= 0),

      CONSTRAINT chk_products_rating
          CHECK (
              rating IS NULL OR
              (rating >= 0 AND rating <= 5)
              ),

      INDEX idx_products_name (name)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS products;