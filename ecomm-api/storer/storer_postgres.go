package storer

import (
	"context"
	"ecomm/domain"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type PostgresStorer struct {
	db *sqlx.DB
}

const (
	queryToInsertProduct = "INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES (:name, :image, :category, :description, :rating, :num_reviews, :price, :count_in_stock) RETURNING *"
	queryToSelectProduct = "SELECT * FROM products WHERE id=:id"

	queryToUpdateProduct = "UPDATE products SET name=:name, image=:image, category=:category, description=:description, rating=:rating, num_reviews=:num_reviews, price=:price, count_in_stock=:count_in_stock, updated_at=NOW() WHERE id=:id RETURNING *"

	queryToInsertOrder     = "INSERT INTO orders (payment_method, tax_price, shipping_price, total_price, user_id, status) VALUES (:payment_method, :tax_price, :shipping_price, :total_price, :user_id, :status) RETURNING *"
	queryToUpdateOrder     = "UPDATE orders SET payment_method=:payment_method, tax_price=:tax_price, shipping_price=:shipping_price, total_price=:total_price, user_id=:user_id, status=:status, updated_at=NOW() WHERE id=:id RETURNING *"
	queryToUpdateOrderItem = "UPDATE order_items SET name=:name, quantity=:quantity, image=:image, price=:price, product_id=:product_id, order_id=:order_id  WHERE order_id=:order_id RETURNING *"
	queryToInsertOrderItem = "INSERT INTO order_items (name, quantity, image, price, product_id, order_id) VALUES (:name, :quantity, :image, :price, :product_id, :order_id) RETURNING *"

	queryToGetOrder = "SELECT * FROM orders WHERE id=:id"

	queryToCancelOrder = "UPDATE orders SET status=:status WHERE id=:id"

	queryToInsertUser = "INSERT INTO users (name, email, password, role) VALUES (:name, :email, :password, :role) RETURNING *"

	queryToFindUserByEmail = "SELECT * FROM users WHERE email=:email"
)

func NewPostgresStorer(db *sqlx.DB) *PostgresStorer {
	return &PostgresStorer{db: db}
}

func (postgres *PostgresStorer) CreateProduct(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	rows, err := postgres.db.NamedQueryContext(ctx, queryToInsertProduct, p)
	if err != nil {
		return nil, fmt.Errorf("Error inserting product: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(p); err != nil {
			return nil, fmt.Errorf("Error scanning rows: %w", err)
		}
	} else {
		return nil, errors.New("product not created")
	}

	return p, nil
}

func (postgres *PostgresStorer) GetProduct(ctx context.Context, id int64) (*domain.Product, error) {
	op := "storer.GetProduct"
	product := domain.Product{}
	arg := map[string]interface{}{
		"id": id,
	}
	rows, err := postgres.db.NamedQueryContext(ctx, queryToSelectProduct, arg)
	if err != nil {
		return nil, fmt.Errorf("Error getting product: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(&product); err != nil {
			return nil, fmt.Errorf("Error scanning rows: %w", err)
		}
	} else {
		// sql.ErrNoRows здесь не будет, так как Next() просто вернет false
		return nil, NewNotFoundError(op, "product", id, nil)
	}

	return &product, nil
}

func (postgres *PostgresStorer) GetProductsByIDs(ctx context.Context, ids []int64) ([]*domain.Product, error) {
	if len(ids) == 0 {
		return []*domain.Product{}, nil
	}
	query := "SELECT * FROM products WHERE id IN (?)"
	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, fmt.Errorf("Error building query: %w", err)
	}
	query = postgres.db.Rebind(query)
	var products []*domain.Product
	if err := postgres.db.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, fmt.Errorf("Error getting products: %w", err)
	}

	return products, nil
}

func (postgres *PostgresStorer) GetProducts(ctx context.Context) ([]*domain.Product, error) {
	products := []*domain.Product{}
	err := postgres.db.SelectContext(ctx, &products, "SELECT * FROM products")
	if err != nil {
		return nil, fmt.Errorf("Error getting products: %w", err)
	}
	if len(products) == 0 {
		return []*domain.Product{}, nil
	}

	return products, nil
}

func (postgres *PostgresStorer) UpdateProduct(ctx context.Context, p *domain.Product) error {
	op := "storer.UpdateProduct"
	rows, err := postgres.db.NamedQueryContext(ctx, queryToUpdateProduct, p)
	if err != nil {
		return fmt.Errorf("Error updating product with id %d: %w", p.ID, err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(p); err != nil {
			return fmt.Errorf("Error scanning updated product: %w", err)
		}
	} else {
		return NewNotFoundError(op, "product", p.ID, nil)
	}

	return nil
}

func (postgres *PostgresStorer) DeleteProduct(ctx context.Context, id int64) error {
	op := "storer.DeleteProduct"
	res, err := postgres.db.ExecContext(ctx, "DELETE FROM products WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("failed delete product with id %d: %w", id, err)
	}
	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return fmt.Errorf("cannot get affected rows for product with id %d: %w", id, err)
	}

	if rowsAffected == 0 {
		return NewNotFoundError(op, "product", id, nil)
	}
	return nil
}

func (postgres *PostgresStorer) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	err := postgres.execTx(ctx, func(tx *sqlx.Tx) error {
		var txErr error
		_, txErr = createOrder(ctx, tx, order)
		if txErr != nil {
			return fmt.Errorf("error creating order row: %w", txErr)
		}

		for i := range order.Items {
			order.Items[i].OrderID = order.ID

			txErr = createOrderItem(ctx, tx, &order.Items[i])
			if txErr != nil {
				return fmt.Errorf("error creating order item row: %w", txErr)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err // Возвращаем ошибку, если транзакция не удалась
	}
	return order, nil
}

func createOrder(ctx context.Context, tx *sqlx.Tx, order *domain.Order) (*domain.Order, error) {
	stmt, err := tx.PrepareNamedContext(ctx, queryToInsertOrder)
	if err != nil {
		return nil, fmt.Errorf("Error creating statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx, order)

	if err != nil {
		return nil, fmt.Errorf("Error creating order: %w", err)
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("No rows selected: %w", err)
	}
	if err := rows.StructScan(order); err != nil {
		return nil, fmt.Errorf("Error scanning rows: %w", err)
	}

	return order, nil
}

func createOrderItem(ctx context.Context, tx *sqlx.Tx, orderItem *domain.OrderItem) error {
	stmt, err := tx.PrepareNamedContext(ctx, queryToInsertOrderItem)

	if err != nil {
		return fmt.Errorf("Error creating statement: %w", err)
	}

	defer stmt.Close()
	rows, err := stmt.QueryxContext(ctx, orderItem)

	if err != nil {
		return fmt.Errorf("Error creating orderItem: %w", err)
	}

	defer rows.Close()

	if !rows.Next() {
		return fmt.Errorf("No rows selected: %w", err)
	}
	if err := rows.StructScan(orderItem); err != nil {
		return fmt.Errorf("Error scanning rows: %w", err)
	}

	return nil
}

func (postgres *PostgresStorer) GetOrders(ctx context.Context) ([]*domain.Order, error) {
	orders := []*domain.Order{}
	err := postgres.db.SelectContext(ctx, &orders, "SELECT * FROM orders")
	if err != nil {
		return nil, fmt.Errorf("error getting orders: %w", err)
	}

	for i := range orders {
		var items []domain.OrderItem
		err = postgres.db.SelectContext(ctx, &items, "SELECT * FROM order_items WHERE order_id=$1", orders[i].ID)

		if err != nil {
			return nil, fmt.Errorf("error getting orderItems: %w", err)
		}

		orders[i].Items = items
	}

	return orders, nil
}

func (postgres *PostgresStorer) UpdateOrder(ctx context.Context, updatedOrder *domain.Order) (*domain.Order, error) {
	err := postgres.execTx(ctx, func(tx *sqlx.Tx) error {
		var txErr error
		_, txErr = postgres.updateOrder(ctx, tx, updatedOrder)
		if txErr != nil {
			return fmt.Errorf("error updating order row: %w", txErr)
		}

		for i := range updatedOrder.Items {
			updatedOrder.Items[i].OrderID = updatedOrder.ID

			txErr = updateOrderItem(ctx, tx, &updatedOrder.Items[i])
			if txErr != nil {
				return fmt.Errorf("error updating order item row: %w", txErr)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err // Возвращаем ошибку, если транзакция не удалась
	}
	return updatedOrder, nil
}

func (postgres *PostgresStorer) updateOrder(ctx context.Context, tx *sqlx.Tx, updatedOrder *domain.Order) (*domain.Order, error) {
	stmt, err := tx.PrepareNamedContext(ctx, queryToUpdateOrder)
	if err != nil {
		return nil, fmt.Errorf("Error creating statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx, updatedOrder)

	if err != nil {
		return nil, fmt.Errorf("Error updating order: %w", err)
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("No rows selected: %w", err)
	}
	if err = rows.StructScan(updatedOrder); err != nil {
		return nil, fmt.Errorf("Error scanning rows: %w", err)
	}

	return updatedOrder, nil
}

func updateOrderItem(ctx context.Context, tx *sqlx.Tx, orderItem *domain.OrderItem) error {
	stmt, err := tx.PrepareNamedContext(ctx, queryToUpdateOrderItem)

	if err != nil {
		return fmt.Errorf("Error creating statement: %w", err)
	}

	defer stmt.Close()
	rows, err := stmt.QueryxContext(ctx, orderItem)

	if err != nil {
		return fmt.Errorf("Error updating orderItem: %w", err)
	}

	defer rows.Close()

	if !rows.Next() {
		return fmt.Errorf("No rows selected: %w", err)
	}
	if err = rows.StructScan(orderItem); err != nil {
		return fmt.Errorf("Error scanning rows: %w", err)
	}

	return nil
}

func (postgres *PostgresStorer) GetOrder(ctx context.Context, id int64) (*domain.Order, error) {
	op := "GetOrder"
	order := &domain.Order{}
	arg := map[string]interface{}{
		"id": id,
	}
	rows, err := postgres.db.NamedQueryContext(ctx, queryToGetOrder, arg)
	if err != nil {
		return nil, fmt.Errorf("Error getting order: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.StructScan(order); err != nil {
			return nil, fmt.Errorf("Error scanning rows: %w", err)
		}
	} else {
		return nil, NewNotFoundError(op, "order", id, nil)
	}

	var items []domain.OrderItem
	err = postgres.db.SelectContext(ctx, &items, "SELECT * FROM order_items WHERE order_id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("error getting orderItems: %w", err)
	}

	order.Items = items

	return order, nil

}
func (postgres *PostgresStorer) CancelOrder(ctx context.Context, id int64) (*domain.Order, error) {
	op := "CancelOrder"
	order := &domain.Order{}
	args := map[string]interface{}{
		"id":     id,
		"status": domain.OrderStatusCancelled,
	}
	rows, err := postgres.db.NamedQueryContext(ctx, queryToCancelOrder, args)
	if err != nil {
		return nil, fmt.Errorf("Error cancelling order: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err = rows.StructScan(&order); err != nil {
			return nil, fmt.Errorf("Error scanning rows: %w", err)
		}

	} else {
		return nil, NewNotFoundError(op, "order", id, nil)
	}

	var items []domain.OrderItem
	err = postgres.db.SelectContext(ctx, &items, "SELECT * FROM order_items WHERE order_id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("error getting orderItems: %w", err)
	}

	order.Items = items

	return order, nil
}

func (postgres *PostgresStorer) execTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := postgres.db.BeginTxx(ctx, nil)

	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	err = fn(tx)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return fmt.Errorf("error in transaction: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (postgres *PostgresStorer) SaveUser(ctx context.Context, user *domain.User) error {
	rows, err := postgres.db.NamedQueryContext(ctx, queryToInsertUser, user)
	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err = rows.StructScan(user); err != nil {
			return fmt.Errorf("error scanning rows: %w", err)
		}

	} else {
		return fmt.Errorf("user not created")
	}
	return nil
}

func (postgres *PostgresStorer) ExistsByEmail(ctx context.Context, email string) bool {
	args := map[string]any{
		"email": email,
	}
	rows, err := postgres.db.NamedQueryContext(ctx, queryToFindUserByEmail, args)

	if err != nil {
		return false
	}
	defer rows.Close()
	if rows.Next() {
		return true
	}
	return false
}

func (postgres *PostgresStorer) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	op := "storer.FindByEmail"
	user := &domain.User{}
	args := map[string]any{
		"email": email,
	}
	rows, err := postgres.db.NamedQueryContext(ctx, queryToFindUserByEmail, args)

	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err = rows.StructScan(user); err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}
		return user, nil
	}
	return nil, NewNotFoundError(op, "user", email, err)
}
