package storer

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func newSingleProductForUpdateTesting() Product {
	return Product{
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
}

func withTestDB(t *testing.T, fn func(*sqlx.DB, sqlmock.Sqlmock)) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "postgres")
	fn(db, mock)
}

func TestCreateProduct(t *testing.T) {
	p := &Product{
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
				expectedQuery := regexp.QuoteMeta(`INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *`)

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
				expectedQuery := regexp.QuoteMeta(`INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *`)
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
				expectedQuery := regexp.QuoteMeta(`INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *`)
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
	p := &Product{
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
				expectedQuery := regexp.QuoteMeta(`SELECT * FROM products WHERE id=$1`)

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
				expectedQuery := regexp.QuoteMeta(`SELECT * FROM products WHERE id=$1`)
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
				expectedQuery := regexp.QuoteMeta(`SELECT * FROM products WHERE id=$1`)
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
				expectedQuery := regexp.QuoteMeta(`SELECT * FROM products WHERE id=$1`)
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

func TestGetProducts(t *testing.T) {
	product1 := Product{
		ID:           1,
		Name:         "test product1",
		Image:        "test.jpg",
		Category:     "test category1",
		Description:  "test description1",
		Rating:       5,
		NumReviews:   10,
		Price:        100.0,
		CountInStock: 100,
	}
	product2 := Product{
		ID:           1,
		Name:         "test product1",
		Image:        "test.jpg",
		Category:     "test category1",
		Description:  "test description1",
		Rating:       5,
		NumReviews:   10,
		Price:        100.0,
		CountInStock: 100,
	}
	product3 := Product{
		ID:           1,
		Name:         "test product1",
		Image:        "test.jpg",
		Category:     "test category1",
		Description:  "test description1",
		Rating:       5,
		NumReviews:   10,
		Price:        100.0,
		CountInStock: 100,
	}

	products := []Product{product1, product2, product3}
	tcs := []struct {
		name string
		test func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(`SELECT * FROM products`)
				columns := []string{"id", "name", "image", "category", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at"}
				rows := sqlmock.NewRows(columns).
					AddRow(product1.ID, product1.Name, product1.Image, product1.Category, product1.Description, product1.Rating, product1.NumReviews, product1.Price, product1.CountInStock, product1.CreatedAt, nil).
					AddRow(product2.ID, product2.Name, product2.Image, product2.Category, product2.Description, product2.Rating, product2.NumReviews, product2.Price, product2.CountInStock, product2.CreatedAt, nil).
					AddRow(product3.ID, product3.Name, product3.Image, product3.Category, product3.Description, product3.Rating, product3.NumReviews, product3.Price, product3.CountInStock, product3.CreatedAt, nil)
				mock.ExpectQuery(expectedQuery).WithArgs().WillReturnRows(rows)

				foundProducts, err := postgresTest.GetProducts(context.Background())
				require.NoError(t, err)
				require.NotNil(t, foundProducts)
				require.Equal(t, len(products), len(foundProducts))
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "error getting products",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(`SELECT * FROM products`)
				mock.ExpectQuery(expectedQuery).WillReturnError(fmt.Errorf("Error getting products"))
				_, err := postgresTest.GetProducts(context.Background())
				require.Error(t, err)
				require.ErrorContains(t, err, fmt.Sprintf("Error getting products"))
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "products not found",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				columns := []string{"id", "name", "image", "category", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at"}
				expectedQuery := regexp.QuoteMeta(`SELECT * FROM products`)
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
				foundProducts, err := postgresTest.GetProducts(context.Background())
				require.Error(t, err)
				require.Nil(t, foundProducts)
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

func TestUpdateProduct(t *testing.T) {

	tcs := []struct {
		name string
		test func(*testing.T, *PostgresStorer, sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				product := newSingleProductForUpdateTesting()
				columns := []string{"name", "image", "category", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at"}
				expectedQuery := regexp.QuoteMeta("UPDATE products SET name=$1, image=$2, category=$3, description=$4, rating=$5, num_reviews=$6, price=$7, count_in_stock=$8, updated_at=NOW() WHERE id=$9 RETURNING *")
				rows := sqlmock.NewRows(columns).
					AddRow("updated test product", "updated test.jpg", "updated test category", "updated test description", 1, 1, 10.0, 10, time.Now(), time.Now())
				mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

				err := postgresTest.UpdateProduct(context.Background(), &product)

				require.NoError(t, err)
				require.NotNil(t, product)
				require.Equal(t, int64(1), product.ID)
				require.Equal(t, "updated test product", product.Name)
				require.Equal(t, "updated test.jpg", product.Image)
				require.Equal(t, "updated test category", product.Category)
				require.Equal(t, "updated test description", product.Description)
				require.Equal(t, int64(1), product.Rating)
				require.Equal(t, int64(1), product.NumReviews)
				require.InDelta(t, 10.0, product.Price, 0.001)
				require.Equal(t, int64(10), product.CountInStock)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "error updating product",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				product := newSingleProductForUpdateTesting()
				expectedQuery := regexp.QuoteMeta("UPDATE products SET name=$1, image=$2, category=$3, description=$4, rating=$5, num_reviews=$6, price=$7, count_in_stock=$8, updated_at=NOW() WHERE id=$9 RETURNING *")
				mock.ExpectQuery(expectedQuery).WillReturnError(fmt.Errorf("Error updating product with id 1"))

				err := postgresTest.UpdateProduct(context.Background(), &product)
				require.Error(t, err)
				require.ErrorContains(t, err, "Error updating product with id 1")
				require.NotNil(t, product)
				require.Equal(t, int64(1), product.ID)
				require.Equal(t, "test product", product.Name)
				require.Equal(t, "test.jpg", product.Image)
				require.Equal(t, "test category", product.Category)
				require.Equal(t, "test description", product.Description)
				require.Equal(t, int64(5), product.Rating)
				require.Equal(t, int64(10), product.NumReviews)
				require.InDelta(t, 100.0, product.Price, 0.001)
				require.Equal(t, int64(100), product.CountInStock)

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)

			},
		},
		{
			name: "error scanning updated product",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				product := newSingleProductForUpdateTesting()
				columns := []string{"name", "image", "category", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at"}
				expectedQuery := regexp.QuoteMeta("UPDATE products SET name=$1, image=$2, category=$3, description=$4, rating=$5, num_reviews=$6, price=$7, count_in_stock=$8, updated_at=NOW() WHERE id=$9 RETURNING *")

				rows := sqlmock.NewRows(columns).
					AddRow("name", "image", "this_is_a_bad_column", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at")

				mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

				err := postgresTest.UpdateProduct(context.Background(), &product)

				require.Error(t, err)
				require.ErrorContains(t, err, "Error scanning updated product")

				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "product for update not found",
			test: func(t *testing.T, postgresTest *PostgresStorer, mock sqlmock.Sqlmock) {
				product := newSingleProductForUpdateTesting()
				columns := []string{"name", "image", "category", "description", "rating", "num_reviews", "price", "count_in_stock", "created_at", "updated_at"}
				expectedQuery := regexp.QuoteMeta("UPDATE products SET name=$1, image=$2, category=$3, description=$4, rating=$5, num_reviews=$6, price=$7, count_in_stock=$8, updated_at=NOW() WHERE id=$9 RETURNING *")
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
				err := postgresTest.UpdateProduct(context.Background(), &product)
				require.Error(t, err)
				require.ErrorContains(t, err, "Product with id 1 not found for update")
				require.NotNil(t, product)
				require.Equal(t, int64(1), product.ID)
				require.Equal(t, "test product", product.Name)
				require.Equal(t, "test.jpg", product.Image)
				require.Equal(t, "test category", product.Category)
				require.Equal(t, "test description", product.Description)
				require.Equal(t, int64(5), product.Rating)
				require.Equal(t, int64(10), product.NumReviews)
				require.InDelta(t, 100.0, product.Price, 0.001)
				require.Equal(t, int64(100), product.CountInStock)

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

func TestDeleteProduct(t *testing.T) {
	product := Product{
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
		test func(*testing.T, *PostgresStorer, sqlmock.Sqlmock)
	}{
		{
			name: "success",
			test: func(t *testing.T, postgres *PostgresStorer, mock sqlmock.Sqlmock) {

				expectedQuery := regexp.QuoteMeta("DELETE FROM products WHERE id=$1")

				mock.ExpectExec(expectedQuery).WithArgs(product.ID).WillReturnResult(sqlmock.NewResult(1, 1))

				err := postgres.DeleteProduct(context.Background(), product.ID)
				require.NoError(t, err)
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to delete product",
			test: func(t *testing.T, postgres *PostgresStorer, mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta("DELETE FROM products WHERE id=$1")
				mock.ExpectExec(expectedQuery).WithArgs(product.ID).WillReturnError(fmt.Errorf("failed delete product with id 1"))
				err := postgres.DeleteProduct(context.Background(), product.ID)
				require.Error(t, err)
				require.ErrorContains(t, err, "failed delete product with id 1")
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "failed to get affected rows",
			test: func(t *testing.T, postgres *PostgresStorer, mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta("DELETE FROM products WHERE id=$1")
				mock.ExpectExec(expectedQuery).WithArgs(product.ID).WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))
				err := postgres.DeleteProduct(context.Background(), product.ID)
				require.Error(t, err)
				require.ErrorContains(t, err, "cannot get affected rows for product with id 1")
				err = mock.ExpectationsWereMet()
				require.NoError(t, err)
			},
		},
		{
			name: "product for delete not found",
			test: func(t *testing.T, postgres *PostgresStorer, mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta("DELETE FROM products WHERE id=$1")
				mock.ExpectExec(expectedQuery).WithArgs(product.ID).WillReturnResult(sqlmock.NewResult(1, 0))
				err := postgres.DeleteProduct(context.Background(), product.ID)
				require.Error(t, err)
				require.ErrorContains(t, err, "product with id 1 not found")
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
