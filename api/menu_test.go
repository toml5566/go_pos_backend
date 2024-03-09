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

func createMenuItem(user db.User, product db.Product, catalog string) db.Menu {
	return db.Menu{
		ID:           uuid.New(),
		UserID:       user.ID,
		ShopName:     user.Username,
		ProductID:    product.ID,
		ProductName:  product.Name,
		ProductPrice: product.Price,
		Catalog:      catalog,
		Description:  product.Description,
		CreatedAt:    time.Date(2022, time.January, 1, 12, 0, 0, 0, time.UTC),
	}
}

type eqAddMenuItemParamsMatcher struct {
	arg db.AddMenuItemParams
}

func (e eqAddMenuItemParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.AddMenuItemParams)
	if !ok {
		return false
	}

	e.arg.ID = arg.ID // match the random generated uuid

	return reflect.DeepEqual(e.arg, arg)
}

func (e eqAddMenuItemParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v", e.arg)
}

func eqAddMenuItemParams(arg db.AddMenuItemParams) gomock.Matcher {
	return eqAddMenuItemParamsMatcher{arg}
}

func requireBodyMatchMenuItem(t *testing.T, body *bytes.Buffer, menuItem db.Menu) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var resMenuItem db.Menu
	err = json.Unmarshal(data, &resMenuItem)
	require.NoError(t, err)
	require.Equal(t, resMenuItem.ID, menuItem.ID)
	require.Equal(t, resMenuItem.UserID, menuItem.UserID)
	require.Equal(t, resMenuItem.ShopName, menuItem.ShopName)
	require.Equal(t, resMenuItem.ProductID, menuItem.ProductID)
	require.Equal(t, resMenuItem.ProductName, menuItem.ProductName)
	require.Equal(t, resMenuItem.ProductPrice, menuItem.ProductPrice)
	require.Equal(t, resMenuItem.Catalog, menuItem.Catalog)
	require.Equal(t, resMenuItem.Description, menuItem.Description)
}

func TestAddMenuItem(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)
	catalog := "breakfast"
	menuItem := createMenuItem(user, product, catalog)

	floatPrice, err := strconv.ParseFloat(product.Price, 64)
	require.NoError(t, err)

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
				"user_id":       user.ID,
				"shop_name":     user.Username,
				"product_id":    product.ID,
				"product_name":  product.Name,
				"product_price": floatPrice,
				"catalog":       catalog,
				"description":   product.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.AddMenuItemParams{
					ID:           menuItem.ID,
					UserID:       menuItem.UserID,
					ShopName:     menuItem.ShopName,
					ProductID:    menuItem.ProductID,
					ProductName:  menuItem.ProductName,
					ProductPrice: menuItem.ProductPrice,
					Catalog:      menuItem.Catalog,
					Description:  menuItem.Description,
				}
				store.EXPECT().
					AddMenuItem(gomock.Any(), eqAddMenuItemParams(arg)).
					Times(1).
					Return(menuItem, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchMenuItem(t, recorder.Body, menuItem)
			},
		},
		{
			name: "InternalServerError",
			user: user,
			body: gin.H{
				"user_id":       user.ID,
				"shop_name":     user.Username,
				"product_id":    product.ID,
				"product_name":  product.Name,
				"product_price": floatPrice,
				"catalog":       catalog,
				"description":   product.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					AddMenuItem(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Menu{}, sql.ErrConnDone)
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
					AddMenuItem(gomock.Any(), gomock.Any()).
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
				"user_id":       user.ID,
				"shop_name":     user.Username,
				"product_id":    product.ID,
				"product_name":  product.Name,
				"product_price": floatPrice,
				"catalog":       catalog,
				"description":   product.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "UnauthorizatedUser", time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					AddMenuItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/%v/menus", tc.user.Username)
			req, err := http.NewRequest(http.MethodPost, url, jsonReader)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestUpdateMenuItem(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)
	catalog := "breakfast"
	menuItem := createMenuItem(user, product, catalog)

	updatedPrice := 100000.00
	updatedMenuItem := db.Menu{
		ID:           menuItem.ID,
		UserID:       menuItem.UserID,
		ShopName:     menuItem.ShopName,
		ProductID:    menuItem.ProductID,
		ProductName:  "updated",
		ProductPrice: utils.FormottedDecimalToString(updatedPrice),
		Catalog:      "lunch",
		Description:  "updated",
	}

	testCases := []struct {
		name          string
		user          db.User
		menuItem      db.Menu
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			user:     user,
			menuItem: menuItem,
			body: gin.H{
				"id":            menuItem.ID,
				"user_id":       menuItem.UserID,
				"shop_name":     menuItem.ShopName,
				"product_name":  updatedMenuItem.ProductName,
				"product_price": updatedPrice,
				"catalog":       updatedMenuItem.Catalog,
				"description":   updatedMenuItem.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.UpdateMenuItemParams{
					ID:           menuItem.ID,
					UserID:       menuItem.UserID,
					ProductName:  updatedMenuItem.ProductName,
					ProductPrice: updatedMenuItem.ProductPrice,
					Catalog:      updatedMenuItem.Catalog,
					Description:  updatedMenuItem.Description,
				}
				store.EXPECT().
					UpdateMenuItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updatedMenuItem, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchMenuItem(t, recorder.Body, updatedMenuItem)
			},
		},
		{
			name:     "InternalError",
			user:     user,
			menuItem: menuItem,
			body: gin.H{
				"id":            menuItem.ID,
				"user_id":       menuItem.UserID,
				"shop_name":     menuItem.ShopName,
				"product_name":  updatedMenuItem.ProductName,
				"product_price": updatedPrice,
				"catalog":       updatedMenuItem.Catalog,
				"description":   updatedMenuItem.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateMenuItem(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Menu{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:     "MissingJSONData",
			user:     user,
			menuItem: menuItem,
			body:     gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateMenuItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "UnauthorizatedUser",
			user:     user,
			menuItem: menuItem,
			body: gin.H{
				"id":            menuItem.ID,
				"user_id":       menuItem.UserID,
				"shop_name":     menuItem.ShopName,
				"product_name":  updatedMenuItem.ProductName,
				"product_price": updatedPrice,
				"catalog":       updatedMenuItem.Catalog,
				"description":   updatedMenuItem.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "UnauthorizatedUser", time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateMenuItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/%v/menus/%v", tc.user.Username, tc.menuItem.ID)
			req, err := http.NewRequest(http.MethodPatch, url, jsonReader)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteMenuItem(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)
	catalog := "breakfast"
	menuItem := createMenuItem(user, product, catalog)

	testCases := []struct {
		name          string
		user          db.User
		menuItem      db.Menu
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			user:     user,
			menuItem: menuItem,
			body: gin.H{
				"id":        menuItem.ID,
				"user_id":   menuItem.UserID,
				"shop_name": menuItem.ShopName,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteMenuItemParams{
					ID:     menuItem.ID,
					UserID: menuItem.UserID,
				}
				store.EXPECT().
					DeleteMenuItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:     "InternalError",
			user:     user,
			menuItem: menuItem,
			body: gin.H{
				"id":        menuItem.ID,
				"user_id":   menuItem.UserID,
				"shop_name": menuItem.ShopName,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteMenuItem(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:     "MissingJSONData",
			user:     user,
			menuItem: menuItem,
			body:     gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteMenuItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "UnauthorizatedUser",
			user:     user,
			menuItem: menuItem,
			body: gin.H{
				"id":        menuItem.ID,
				"user_id":   menuItem.UserID,
				"shop_name": menuItem.ShopName,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "UnauthorizatedUser", time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteMenuItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/%v/menus/%v", tc.user.Username, tc.menuItem.ID)
			req, err := http.NewRequest(http.MethodDelete, url, jsonReader)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetAllMenuItems(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)
	catalog := "breakfast"
	menuItem := createMenuItem(user, product, catalog)

	testCases := []struct {
		name          string
		shopName      string
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			shopName: menuItem.ShopName,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAllMenuItems(gomock.Any(), gomock.Eq(menuItem.ShopName)).
					Times(1).
					Return([]db.Menu{menuItem}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:     "InternalError",
			shopName: menuItem.ShopName,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAllMenuItems(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Menu{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:     "WrongShopUri",
			shopName: "NotExisted",
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAllMenuItems(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Menu{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := fmt.Sprintf("/%v/menus", tc.shopName)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}
