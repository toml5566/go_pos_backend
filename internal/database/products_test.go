package database

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/toml5566/go_pos_backend/utils"
)

func createRandomProduct(t *testing.T, user1 User) Product {
	arg := CreateProductParams{
		ID:          uuid.New(),
		UserID:      user1.ID,
		Name:        utils.RandString(4),
		Price:       strconv.FormatFloat(utils.RandomFloat(1, 100), 'f', 2, 64),
		Description: utils.RandString(10),
	}

	product, err := testQueries.CreateProduct(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, product)

	require.Equal(t, arg.ID, product.ID)
	require.Equal(t, arg.UserID, product.UserID)
	require.Equal(t, arg.Name, product.Name)
	require.Equal(t, arg.Price, product.Price)
	require.Equal(t, arg.Description, product.Description)

	require.NotZero(t, product.CreatedAt)

	return product
}

func createProductWithSameName(t *testing.T, user User, productName string) []Product {
	// create 2 products with same name
	arg1 := CreateProductParams{
		ID:          uuid.New(),
		UserID:      user.ID,
		Name:        productName,
		Price:       "12.50",
		Description: "it is an apple",
	}
	arg2 := CreateProductParams{
		ID:          uuid.New(),
		UserID:      user.ID,
		Name:        productName,
		Price:       "20.00",
		Description: "it is an orange",
	}

	product1, err := testQueries.CreateProduct(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, product1)

	product2, err := testQueries.CreateProduct(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, product2)

	return []Product{product1, product2}
}

func TestCreateProduct(t *testing.T) {
	user1 := createRandomUser(t)
	createRandomProduct(t, user1)
}

func TestGetProduct(t *testing.T) {
	user1 := createRandomUser(t)

	product1 := createRandomProduct(t, user1)

	arg := GetProductParams{
		UserID: user1.ID,
		ID:     product1.ID,
	}

	productFromDB, err := testQueries.GetProduct(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productFromDB)

	require.Equal(t, productFromDB.ID, product1.ID)
	require.Equal(t, productFromDB.UserID, product1.UserID)
	require.Equal(t, productFromDB.Name, product1.Name)
	require.Equal(t, productFromDB.Price, product1.Price)
	require.Equal(t, productFromDB.Description, product1.Description)

	require.WithinDuration(t, productFromDB.CreatedAt, product1.CreatedAt, time.Second)

}

func TestGetProductsByName(t *testing.T) {
	user1 := createRandomUser(t)
	productName := utils.RandString(5)

	createProductWithSameName(t, user1, productName)

	// get products by name
	arg := GetProductsByNameParams{
		UserID: user1.ID,
		Name:   productName,
	}

	products, err := testQueries.GetProductsByName(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, products)

	for _, product := range products {
		require.NotEmpty(t, product)

		require.Equal(t, product.Name, productName)
		require.Equal(t, product.UserID, user1.ID)
	}

}

func TestGetAllProducts(t *testing.T) {
	user := createRandomUser(t)

	product1 := createRandomProduct(t, user)
	product2 := createRandomProduct(t, user)

	products, err := testQueries.GetAllProducts(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, products)

	var productIDList []uuid.UUID
	for _, p := range products {
		require.NotEmpty(t, p)
		require.Equal(t, p.UserID, user.ID)

		productIDList = append(productIDList, p.ID)
	}
	require.Contains(t, productIDList, product1.ID)
	require.Contains(t, productIDList, product2.ID)
}

func TestUpdateProduct(t *testing.T) {
	user := createRandomUser(t)
	product := createRandomProduct(t, user)

	arg := UpdateProductParams{
		UserID:      user.ID,
		ID:          product.ID,
		Name:        "Updated name",
		Price:       "1000.00",
		Description: "Updated description",
	}

	updatedProduct, err := testQueries.UpdateProduct(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedProduct)

	require.Equal(t, updatedProduct.ID, product.ID)
	require.Equal(t, updatedProduct.UserID, product.UserID)
	require.Equal(t, updatedProduct.Name, arg.Name)
	require.Equal(t, updatedProduct.Price, arg.Price)
	require.Equal(t, updatedProduct.Description, arg.Description)

}

func TestDeleteProduct(t *testing.T) {
	user := createRandomUser(t)
	product := createRandomProduct(t, user)

	arg1 := DeleteProductParams{
		UserID: user.ID,
		ID:     product.ID,
	}

	err := testQueries.DeleteProduct(context.Background(), arg1)
	require.NoError(t, err)

	arg2 := GetProductParams{
		UserID: user.ID,
		ID:     product.ID,
	}

	deletedProduct, err := testQueries.GetProduct(context.Background(), arg2)
	require.Error(t, err)
	require.EqualError(t, err, ErrRecordNotFound.Error())
	require.Empty(t, deletedProduct)
}
