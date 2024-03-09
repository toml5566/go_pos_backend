package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/toml5566/go_pos_backend/utils"
)

func addRandomMenuItem(t *testing.T, user User) Menu {
	product1 := createRandomProduct(t, user)

	arg := AddMenuItemParams{
		ID:           uuid.New(),
		UserID:       user.ID,
		ShopName:     user.Username,
		ProductID:    product1.ID,
		ProductName:  product1.Name,
		ProductPrice: product1.Price,
		Catalog:      "breakfast",
		Description:  utils.RandString(10),
	}

	menuItem, err := testQueries.AddMenuItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, menuItem)

	require.Equal(t, menuItem.ID, arg.ID)
	require.Equal(t, menuItem.UserID, arg.UserID)
	require.Equal(t, menuItem.ShopName, arg.ShopName)
	require.Equal(t, menuItem.ProductID, arg.ProductID)
	require.Equal(t, menuItem.ProductName, arg.ProductName)
	require.Equal(t, menuItem.ProductPrice, arg.ProductPrice)
	require.Equal(t, menuItem.Catalog, arg.Catalog)
	require.Equal(t, menuItem.Description, arg.Description)

	require.NotZero(t, menuItem.CreatedAt)

	return menuItem
}

func TestAddMenuItem(t *testing.T) {
	user := createRandomUser(t)

	addRandomMenuItem(t, user)
}

func TestUpdateMenuItem(t *testing.T) {
	user := createRandomUser(t)
	menuItem := addRandomMenuItem(t, user)

	arg := UpdateMenuItemParams{
		UserID:       user.ID,
		ID:           menuItem.ID,
		ProductName:  "updated",
		ProductPrice: "1000.00",
		Catalog:      "updated",
		Description:  "updated",
	}

	updatedMenuItem, err := testQueries.UpdateMenuItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedMenuItem)

	require.Equal(t, updatedMenuItem.ID, menuItem.ID)
	require.Equal(t, updatedMenuItem.UserID, menuItem.UserID)
	require.Equal(t, updatedMenuItem.ShopName, menuItem.ShopName)
	require.Equal(t, updatedMenuItem.ProductID, menuItem.ProductID)
	require.Equal(t, updatedMenuItem.ProductName, arg.ProductName)
	require.Equal(t, updatedMenuItem.ProductPrice, arg.ProductPrice)
	require.Equal(t, updatedMenuItem.Catalog, arg.Catalog)
	require.Equal(t, updatedMenuItem.Description, arg.Description)

	require.NotZero(t, updatedMenuItem.CreatedAt)
}

func TestDeleteMenuItem(t *testing.T) {
	user := createRandomUser(t)
	menuItem := addRandomMenuItem(t, user)

	arg := DeleteMenuItemParams{
		UserID: user.ID,
		ID:     menuItem.ID,
	}

	err := testQueries.DeleteMenuItem(context.Background(), arg)
	require.NoError(t, err)

	allMenuItems, err := testQueries.GetAllMenuItems(context.Background(), menuItem.ShopName)
	require.NoError(t, err)
	require.Empty(t, allMenuItems)
}

func TestGetAllMenuItems(t *testing.T) {
	user := createRandomUser(t)
	menuItem1 := addRandomMenuItem(t, user)
	menuItem2 := addRandomMenuItem(t, user)

	allMenuItems, err := testQueries.GetAllMenuItems(context.Background(), menuItem1.ShopName)
	require.NoError(t, err)
	require.NotEmpty(t, allMenuItems)

	var menuItemIDList []uuid.UUID
	for _, i := range allMenuItems {
		require.NotEmpty(t, i)
		require.Equal(t, i.UserID, user.ID)

		menuItemIDList = append(menuItemIDList, i.ID)
	}

	require.Contains(t, menuItemIDList, menuItem1.ID)
	require.Contains(t, menuItemIDList, menuItem2.ID)
}
