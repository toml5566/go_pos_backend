package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/toml5566/go_pos_backend/utils"
)

func createRandomOrderItem(t *testing.T, user User, orderID uuid.UUID, orderDay string) Order {
	product := createRandomProduct(t, user)

	arg := CreateOrderItemParams{
		ID:           uuid.New(),
		ShopName:     user.Username,
		OrderID:      orderID,
		OrderDay:     orderDay,
		ProductName:  product.Name,
		ProductPrice: product.Price,
		Amount:       utils.RandomInt32(1, 10),
		Status:       "pending",
	}

	orderItem, err := testQueries.CreateOrderItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, orderItem)

	require.Equal(t, orderItem.ID, arg.ID)
	require.Equal(t, orderItem.ShopName, arg.ShopName)
	require.Equal(t, orderItem.OrderID, arg.OrderID)
	require.Equal(t, orderItem.OrderDay, arg.OrderDay)
	require.Equal(t, orderItem.ProductName, arg.ProductName)
	require.Equal(t, orderItem.ProductPrice, arg.ProductPrice)
	require.Equal(t, orderItem.Amount, arg.Amount)
	require.Equal(t, orderItem.Status, arg.Status)

	require.NotZero(t, orderItem.CreatedAt)

	return orderItem
}

func TestCreateOrderItem(t *testing.T) {
	user := createRandomUser(t)
	orderID := utils.RandOrderID()
	orderDay := utils.FormattedDateNow()

	createRandomOrderItem(t, user, orderID, orderDay)
}

func TestUpdateOrderItem(t *testing.T) {
	user := createRandomUser(t)
	orderID := utils.RandOrderID()
	orderDay := utils.FormattedDateNow()

	orderItem := createRandomOrderItem(t, user, orderID, orderDay)

	arg := UpdateOrderItemParams{
		ShopName: user.Username,
		ID:       orderItem.ID,
		Amount:   int32(1000),
		Status:   "accepted",
	}

	updatedOrderItem, err := testQueries.UpdateOrderItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedOrderItem)

	require.Equal(t, updatedOrderItem.ID, arg.ID)
	require.Equal(t, updatedOrderItem.ShopName, arg.ShopName)
	require.Equal(t, updatedOrderItem.Amount, arg.Amount)
	require.Equal(t, updatedOrderItem.Status, arg.Status)

	require.Equal(t, updatedOrderItem.ShopName, orderItem.ShopName)
	require.Equal(t, updatedOrderItem.OrderID, orderItem.OrderID)
	require.Equal(t, updatedOrderItem.OrderDay, orderItem.OrderDay)
	require.Equal(t, updatedOrderItem.ProductName, orderItem.ProductName)
	require.Equal(t, updatedOrderItem.ProductPrice, orderItem.ProductPrice)

	require.NotZero(t, orderItem.CreatedAt)
}

func TestDeleteOrderitem(t *testing.T) {
	user := createRandomUser(t)
	orderID := utils.RandOrderID()
	orderDay := utils.FormattedDateNow()

	orderItem := createRandomOrderItem(t, user, orderID, orderDay)

	arg := DeleteOrderItemParams{
		ShopName: user.Username,
		ID:       orderItem.ID,
	}

	err := testQueries.DeleteOrderItem(context.Background(), arg)
	require.NoError(t, err)

	arg1 := GetOrdersByOrderIDParams{
		ShopName: user.Username,
		OrderID:  orderItem.OrderID,
	}

	orders, err := testQueries.GetOrdersByOrderID(context.Background(), arg1)
	require.NoError(t, err)
	require.Empty(t, orders)
}

func TestGetOrderByDay(t *testing.T) {
	user := createRandomUser(t)
	orderID := utils.RandOrderID()
	orderDay := utils.FormattedDateNow()

	createRandomOrderItem(t, user, orderID, orderDay)
	createRandomOrderItem(t, user, uuid.New(), orderDay)
	createRandomOrderItem(t, user, uuid.New(), "2006-01-02")

	arg := GetOrdersByDayParams{
		ShopName: user.Username,
		OrderDay: orderDay,
	}

	orders, err := testQueries.GetOrdersByDay(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, orders)

	for _, o := range orders {
		require.NotEmpty(t, o)
		require.Equal(t, o.OrderDay, orderDay)
		require.Equal(t, o.ShopName, user.Username)
	}
}

func TestGetOrdersByOrderID(t *testing.T) {
	user := createRandomUser(t)
	orderID := utils.RandOrderID()
	orderDay := utils.FormattedDateNow()

	createRandomOrderItem(t, user, orderID, orderDay)
	createRandomOrderItem(t, user, orderID, orderDay)
	createRandomOrderItem(t, user, uuid.New(), "2006-01-02")

	arg := GetOrdersByOrderIDParams{
		ShopName: user.Username,
		OrderID:  orderID,
	}

	orders, err := testQueries.GetOrdersByOrderID(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, orders)

	for _, o := range orders {
		require.NotEmpty(t, o)
		require.Equal(t, o.ShopName, user.Username)
		require.Equal(t, o.OrderID, orderID)
	}
}
