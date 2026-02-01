package storer

import (
	"context"
	"ecomm/ecomm-api/storer/types"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func withTestDB(t *testing.T, fn func(*sqlx.DB, sqlmock.Sqlmock)) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")
	fn(db, mock)
}

func TestCreateProduct(t *testing.T) {
	p := &types.Product{
		Name:         "test product",
		Image:        "test.jpg",
		Category:     "test category",
		Description:  "test description",
		Rating:       5,
		NumReviews:   10,
		Price:        100.0,
		CountInStock: 100,
	}

	tcs := []struct {
		name string
		test func(*testing.T, *PostgresStorer, sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				// Для именованных запросов sqlx сначала преобразует плейсхолдеры в '?'
				// sqlmock перехватывает запрос уже в этом виде.
				expectedQuery := regexp.QuoteMeta(`INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *`)

				rows := sqlmock.NewRows([]string{"id", "name", "image", "category", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at"}).
					AddRow(1, p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock, time.Now(), nil)

				mock.ExpectQuery(expectedQuery).
					WithArgs(p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock).
					WillReturnRows(rows)

				createdProduct, err := postgresTest.CreateProduct(context.Background(), p)
				require.NoError(t, err)
				require.NotNil(t, createdProduct)
				require.Equal(t, int64(1), createdProduct.ID)
				require.Equal(t, p.Name, createdProduct.Name)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed inserting product",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(`INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *`)
				mock.ExpectQuery(expectedQuery).
					WithArgs(p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock).
					WillReturnError(fmt.Errorf("Error inserting product"))
				_, err := postgresTest.CreateProduct(context.Background(), p)
				require.Error(t, err)
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to scan rows",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "this_is_a_bad_column", "category", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at"}).
					AddRow(1, p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock, time.Now(), nil)
				expectedQuery := regexp.QuoteMeta(`INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *`)
				mock.ExpectQuery(expectedQuery).
					WithArgs(p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock).
					WillReturnRows(rows)
				_, err := postgresTest.CreateProduct(context.Background(), p)
				require.Error(t, err)
				require.ErrorContains(t, err, "Error scanning rows")
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withTestDB(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
				postgresTest := NewPostgresStorer(db)
				tc.test(t, postgresTest, mock)
			})
		})
	}
}

func TestGetProduct(t *testing.T) {
	p := &types.Product{
		ID:           1,
		Name:         "test product",
		Image:        "test.jpg",
		Category:     "test category",
		Description:  "test description",
		Rating:       5,
		NumReviews:   10,
		Price:        100.0,
		CountInStock: 100,
	}
	tcs := []struct {
		name string
		test func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(`SELECT * FROM products WHERE id=?`)

				rows := sqlmock.NewRows([]string{
					"id", "name", "image", "category", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at",
				}).
					AddRow(p.ID, p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock, time.Now(), nil)

				mock.ExpectQuery(expectedQuery).WithArgs(p.ID).WillReturnRows(rows)

				foundProduct, err := postgresTest.GetProduct(context.Background(), p.ID)

				require.NoError(t, err)
				require.NotNil(t, foundProduct)
				require.Equal(t, p.ID, foundProduct.ID)
				require.Equal(t, p.Name, foundProduct.Name)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed getting product",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(`SELECT * FROM products WHERE id=?`)
				mock.ExpectQuery(expectedQuery).WithArgs(p.ID).WillReturnError(fmt.Errorf("Error getting product"))
				_, err := postgresTest.GetProduct(context.Background(), p.ID)
				require.Error(t, err)
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed scanning rows",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(`SELECT * FROM products WHERE id=?`)
				rows := sqlmock.NewRows([]string{"id", "name", "this_is_a_bad_column", "category", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at"}).
					AddRow(1, p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock, time.Now(), nil)

				mock.ExpectQuery(expectedQuery).WithArgs(p.ID).WillReturnRows(rows)

				_, err := postgresTest.GetProduct(context.Background(), p.ID)
				require.Error(t, err)
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "product not found",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(`SELECT * FROM products WHERE id=?`)
				columns := []string{"id", "name", "image", "category", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at"}
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(expectedQuery).WithArgs(p.ID).WillReturnRows(rows)

				_, err := postgresTest.GetProduct(context.Background(), p.ID)
				require.Error(t, err)
				require.ErrorContains(t, err, fmt.Sprintf("Product with id %d not found", p.ID))
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withTestDB(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
				postgresTest := NewPostgresStorer(db)
				tc.test(t, postgresTest, mock)
			})
		})
	}
}
