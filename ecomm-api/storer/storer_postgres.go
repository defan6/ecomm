package storer

import (
	"context"
	"ecomm/ecomm-api/storer/types"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type PostgresStorer struct {
	db *sqlx.DB
}

const (
	queryToInsert = `
			INSERT INTO products
			(name, image, category, description, rating, num_reviews, price, count_in_stock)
			 VALUES 
			(:name, :image, :category, :description, :rating, :num_reviews, :price, :count_in_stock)
			RETURNING *
			`
	queryToSelect = `SELECT * FROM products WHERE id=:id`

	queryToUpdate = `UPDATE products SET
	name=:name, image=:image, category=:category, description=:description, rating=:rating, num_reviews=:num_reviews, price=:price, count_in_stock=:count_in_stock, updated_at=NOW()
	WHERE id=:id
	RETURNING *
	`
)

func NewPostgresStorer(db *sqlx.DB) *PostgresStorer {
	return &PostgresStorer{db: db}
}

func (postgres *PostgresStorer) CreateProduct(ctx context.Context, p *types.Product) (*types.Product, error) {
	rows, err := postgres.db.NamedQueryContext(ctx, queryToInsert, p)
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

func (postgres *PostgresStorer) GetProduct(ctx context.Context, id int64) (*types.Product, error) {
	product := types.Product{}
	arg := map[string]interface{}{
		"id": id,
	}
	rows, err := postgres.db.NamedQueryContext(ctx, queryToSelect, arg)
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
		return nil, fmt.Errorf("Product with id %d not found", id)
	}

	return &product, nil
}

func (postgres *PostgresStorer) GetProducts(ctx context.Context) ([]*types.Product, error) {
	products := []*types.Product{}
	err := postgres.db.SelectContext(ctx, &products, "SELECT * FROM products")
	if err != nil {
		return nil, fmt.Errorf("Error getting products: %w", err)
	}
	if len(products) == 0 {
		return nil, errors.New("No products found")
	}

	return products, nil
}

func (postgres *PostgresStorer) UpdateProduct(ctx context.Context, p *types.Product) (*types.Product, error) {
	updatedProduct := &types.Product{}
	rows, err := postgres.db.NamedQueryContext(ctx, queryToUpdate, p)
	if err != nil {
		return nil, fmt.Errorf("Error updating product with id %d: %w", p.ID, err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(updatedProduct); err != nil {
			return nil, fmt.Errorf("Error scanning updated product: %w", err)
		}
	} else {
		return nil, fmt.Errorf("Product with id %d not found for update", p.ID)
	}

	return updatedProduct, nil
}

func (postgres *PostgresStorer) DeleteProduct(ctx context.Context, id int64) error {
	res, err := postgres.db.ExecContext(ctx, "DELETE FROM products WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("failed delete product with id %d: %w", id, err)
	}
	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return fmt.Errorf("cannot get affected rows for product with id %d: %w", id, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product with id %d not found", id)
	}
	return nil
}
