package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	db "github.com/toml5566/go_pos_backend/internal/database"
	mockdb "github.com/toml5566/go_pos_backend/internal/database/mock"
	"github.com/toml5566/go_pos_backend/token"
	"github.com/toml5566/go_pos_backend/utils"
	"go.uber.org/mock/gomock"
)

func addOrderItem(menuItem db.Menu, orderID uuid.UUID) db.Order {
	return db.Order{
		ID:           uuid.New(),
		ShopName:     menuItem.ShopName,
		OrderID:      orderID,
		OrderDay:     "2022-01-01",
		ProductName:  menuItem.ProductName,
		ProductPrice: menuItem.ProductPrice,
		Amount:       1,
		Status:       "Pending",
		CreatedAt:    time.Date(2022, time.January, 1, 12, 0, 0, 0, time.UTC),
	}
}

type eqCreateOrderItemParamsMatcher struct {
	arg db.CreateOrderItemParams
}

func (e eqCreateOrderItemParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateOrderItemParams)
	if !ok {
		return false
	}

	e.arg.ID = arg.ID // match the random generated uuid

	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateOrderItemParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v", e.arg)
}

func eqCreateOrderItemParams(arg db.CreateOrderItemParams) gomock.Matcher {
	return eqCreateOrderItemParamsMatcher{arg}
}

func requireBodyMatchOrder(t *testing.T, body *bytes.Buffer, orders []db.Order) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var ordersRes []db.Order
	err = json.Unmarshal(data, &ordersRes)
	require.NoError(t, err)

	for i, orderItem := range ordersRes {
		require.Equal(t, orderItem.ID, orders[i].ID)
		require.Equal(t, orderItem.ShopName, orders[i].ShopName)
		require.Equal(t, orderItem.OrderID, orders[i].OrderID)
		require.Equal(t, orderItem.OrderDay, orders[i].OrderDay)
		require.Equal(t, orderItem.ProductName, orders[i].ProductName)
		require.Equal(t, orderItem.ProductPrice, orders[i].ProductPrice)
		require.Equal(t, orderItem.Amount, orders[i].Amount)
		require.Equal(t, orderItem.Status, orders[i].Status)
	}
}

func requireBodyMatchOrderItem(t *testing.T, body *bytes.Buffer, orderItem db.Order) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var orderItemRes db.Order
	err = json.Unmarshal(data, &orderItemRes)
	require.NoError(t, err)

	require.Equal(t, orderItem.ID, orderItemRes.ID)
	require.Equal(t, orderItem.ShopName, orderItemRes.ShopName)
	require.Equal(t, orderItem.OrderID, orderItemRes.OrderID)
	require.Equal(t, orderItem.OrderDay, orderItemRes.OrderDay)
	require.Equal(t, orderItem.ProductName, orderItemRes.ProductName)
	require.Equal(t, orderItem.ProductPrice, orderItemRes.ProductPrice)
	require.Equal(t, orderItem.Amount, orderItemRes.Amount)
	require.Equal(t, orderItem.Status, orderItemRes.Status)
}

func TestCreateOrderItem(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)
	menuItem := createMenuItem(user, product, "breakfast")

	orderID := uuid.New()
	orderItem := addOrderItem(menuItem, orderID)
	orders := []db.Order{orderItem}

	orderItemFloatPrice, err := strconv.ParseFloat(orderItem.ProductPrice, 64)
	require.NoError(t, err)

	orderItemReq := createOrderItemRequest{
		orderItem.ShopName,
		orderItem.OrderDay,
		orderItem.ProductName,
		orderItemFloatPrice,
		orderItem.Amount,
		orderItem.Status,
	}

	testCases := []struct {
		name          string
		shopName      string
		body          gin.H
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			shopName: orderItem.ShopName,
			body: gin.H{
				"order_id": orderID,
				"orders": []createOrderItemRequest{
					orderItemReq,
				},
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.CreateOrderItemParams{
					ID:           orderItem.ID,
					ShopName:     orderItem.ShopName,
					OrderID:      orderID,
					OrderDay:     orderItem.OrderDay,
					ProductName:  orderItem.ProductName,
					ProductPrice: orderItem.ProductPrice,
					Amount:       orderItem.Amount,
					Status:       orderItem.Status,
				}
				store.EXPECT().
					CreateOrderItem(gomock.Any(), eqCreateOrderItemParams(arg)).
					Times(1).
					Return(orderItem, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchOrder(t, recorder.Body, orders)
			},
		},
		{
			name:     "InternalError",
			shopName: orderItem.ShopName,
			body: gin.H{
				"order_id": orderID,
				"orders": []createOrderItemRequest{
					orderItemReq,
				},
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateOrderItem(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Order{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:     "IncorrectJSONFormat",
			shopName: orderItem.ShopName,
			body: gin.H{
				"order_id": orderID,
				"orders":   "",
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateOrderItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStub(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			jsonData, err := json.Marshal(tc.body)
			require.NoError(t, err)
			jsonReader := bytes.NewReader(jsonData)

			url := fmt.Sprintf("/%v/order", tc.shopName)
			req, err := http.NewRequest(http.MethodPost, url, jsonReader)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestUpdateOrderItem(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)
	menuItem := createMenuItem(user, product, "breakfast")

	orderID := uuid.New()
	orderItem := addOrderItem(menuItem, orderID)

	updatedPrice := 1000.00
	updatedOrderItem := db.Order{
		ID:           orderItem.ID,
		ShopName:     orderItem.ShopName,
		OrderID:      orderItem.OrderID,
		OrderDay:     orderItem.OrderDay,
		ProductName:  "updated",
		ProductPrice: utils.FormottedDecimalToString(updatedPrice),
		Amount:       1000,
		Status:       "accepted",
		CreatedAt:    orderItem.CreatedAt,
	}

	testCases := []struct {
		name          string
		user          db.User
		orderId       uuid.UUID
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			user: user,
			body: gin.H{
				"id":        updatedOrderItem.ID,
				"shop_name": updatedOrderItem.ShopName,
				"amount":    updatedPrice,
				"status":    updatedOrderItem.Status,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.UpdateOrderItemParams{
					ID:       updatedOrderItem.ID,
					ShopName: updatedOrderItem.ShopName,
					Amount:   updatedOrderItem.Amount,
					Status:   updatedOrderItem.Status,
				}
				store.EXPECT().
					UpdateOrderItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updatedOrderItem, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchOrderItem(t, recorder.Body, updatedOrderItem)
			},
		},
		{
			name: "InternalError",
			user: user,
			body: gin.H{
				"id":        updatedOrderItem.ID,
				"shop_name": updatedOrderItem.ShopName,
				"amount":    updatedPrice,
				"status":    updatedOrderItem.Status,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateOrderItem(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Order{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "MissingJSONData",
			user: user,
			body: gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateOrderItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UnauthorizatedUser",
			user: user,
			body: gin.H{
				"id":        updatedOrderItem.ID,
				"shop_name": updatedOrderItem.ShopName,
				"amount":    updatedPrice,
				"status":    updatedOrderItem.Status,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "UnauthorizatedUser", time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateOrderItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStub(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			jsonData, err := json.Marshal(tc.body)
			require.NoError(t, err)
			jsonReader := bytes.NewReader(jsonData)

			url := fmt.Sprintf("/users/%v/orders/%v", tc.user.Username, tc.orderId)
			req, err := http.NewRequest(http.MethodPatch, url, jsonReader)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteOrderItem(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)
	menuItem := createMenuItem(user, product, "breakfast")

	orderID := uuid.New()
	orderItem := addOrderItem(menuItem, orderID)

	testCases := []struct {
		name          string
		user          db.User
		orderId       uuid.UUID
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			user: user,
			body: gin.H{
				"id":        orderItem.ID,
				"shop_name": orderItem.ShopName,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteOrderItemParams{
					ID:       orderItem.ID,
					ShopName: orderItem.ShopName,
				}
				store.EXPECT().
					DeleteOrderItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			user: user,
			body: gin.H{
				"id":        orderItem.ID,
				"shop_name": orderItem.ShopName,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteOrderItem(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "UnauthorizatedUser",
			user: user,
			body: gin.H{
				"id":        orderItem.ID,
				"shop_name": orderItem.ShopName,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "UnauthorizatedUser", time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteOrderItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "MissingJSONData",
			user: user,
			body: gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteOrderItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStub(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			jsonData, err := json.Marshal(tc.body)
			require.NoError(t, err)
			jsonReader := bytes.NewReader(jsonData)

			url := fmt.Sprintf("/users/%v/orders/%v", tc.user.Username, tc.orderId)
			req, err := http.NewRequest(http.MethodDelete, url, jsonReader)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetOrdersByDay(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)
	menuItem := createMenuItem(user, product, "breakfast")

	orderID := uuid.New()
	orderItem := addOrderItem(menuItem, orderID)

	orders := []db.Order{orderItem}

	testCases := []struct {
		name          string
		user          db.User
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			user: user,
			body: gin.H{
				"shop_name": orderItem.ShopName,
				"order_day": orderItem.OrderDay,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetOrdersByDayParams{
					OrderDay: orderItem.OrderDay,
					ShopName: orderItem.ShopName,
				}
				store.EXPECT().
					GetOrdersByDay(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(orders, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchOrder(t, recorder.Body, orders)
			},
		},
		{
			name: "InternalError",
			user: user,
			body: gin.H{
				"shop_name": orderItem.ShopName,
				"order_day": orderItem.OrderDay,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetOrdersByDay(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Order{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "MissingJSONData",
			user: user,
			body: gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetOrdersByDay(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UnauthorizatedUser",
			user: user,
			body: gin.H{
				"shop_name": orderItem.ShopName,
				"order_day": orderItem.OrderDay,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorizatedUser", time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetOrdersByDay(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStub(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			jsonData, err := json.Marshal(tc.body)
			require.NoError(t, err)
			jsonReader := bytes.NewReader(jsonData)

			url := fmt.Sprintf("/users/%v/orders", tc.user.Username)
			req, err := http.NewRequest(http.MethodGet, url, jsonReader)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetOrdersByOrderID(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)
	menuItem := createMenuItem(user, product, "breakfast")

	orderID := uuid.New()
	orderItem := addOrderItem(menuItem, orderID)

	orders := []db.Order{}

	testCases := []struct {
		name          string
		shopName      string
		orderId       uuid.UUID
		body          gin.H
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			shopName: orderItem.ShopName,
			orderId:  orderID,
			body: gin.H{
				"order_id":  orderItem.OrderID,
				"shop_name": orderItem.ShopName,
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetOrdersByOrderIDParams{
					OrderID:  orderItem.OrderID,
					ShopName: orderItem.ShopName,
				}
				store.EXPECT().
					GetOrdersByOrderID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(orders, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchOrder(t, recorder.Body, orders)
			},
		},
		{
			name:     "InternalError",
			shopName: orderItem.ShopName,
			orderId:  orderID,
			body: gin.H{
				"order_id":  orderItem.OrderID,
				"shop_name": orderItem.ShopName,
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetOrdersByOrderID(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Order{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:     "MissingJSONData",
			shopName: orderItem.ShopName,
			orderId:  orderID,
			body:     gin.H{},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetOrdersByOrderID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStub(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			jsonData, err := json.Marshal(tc.body)
			require.NoError(t, err)
			jsonReader := bytes.NewReader(jsonData)

			url := fmt.Sprintf("/%v/order/%v", tc.shopName, tc.orderId)
			req, err := http.NewRequest(http.MethodGet, url, jsonReader)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}
