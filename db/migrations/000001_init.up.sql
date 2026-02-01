CREATE TABLE products (
                          id SERIAL PRIMARY KEY,
                          name VARCHAR(255) NOT NULL,
                          image VARCHAR(255) NOT NULL,
                          category VARCHAR(255) NOT NULL,
                          description TEXT,
                          rating INTEGER NOT NULL,
                          num_reviews INTEGER NOT NULL DEFAULT 0,
                          price NUMERIC(10,2) NOT NULL,
                          count_in_stock INTEGER NOT NULL,
                          created_at TIMESTAMP DEFAULT now(),
                          updated_at TIMESTAMP
);

CREATE TABLE orders (
                        id SERIAL PRIMARY KEY,
                        payment_method VARCHAR(255) NOT NULL,
                        tax_price NUMERIC(10,2) NOT NULL,
                        shipping_price NUMERIC(10,2) NOT NULL,
                        total_price NUMERIC(10,2) NOT NULL,
                        created_at TIMESTAMP DEFAULT now(),
                        updated_at TIMESTAMP
);

CREATE TABLE order_items (
                             id SERIAL PRIMARY KEY,
                             order_id INTEGER NOT NULL,
                             product_id INTEGER NOT NULL,
                             name VARCHAR(255) NOT NULL,
                             quantity INTEGER NOT NULL,
                             image VARCHAR(255) NOT NULL,
                             price NUMERIC(10,2) NOT NULL,
                             CONSTRAINT fk_order
                                 FOREIGN KEY(order_id) REFERENCES orders(id)
                                     ON DELETE CASCADE,
                             CONSTRAINT fk_product
                                 FOREIGN KEY(product_id) REFERENCES products(id)
                                     ON DELETE CASCADE
);
