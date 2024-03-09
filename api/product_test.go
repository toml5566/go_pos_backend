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

func randomProduct(user db.User) db.Product {
	return db.Product{
		ID:          uuid.New(),
		UserID:      user.ID,
		Name:        utils.RandString(6),
		Price:       fmt.Sprintf("%.2f", utils.RandomFloat(1, 100)),
		Description: utils.RandString(10),
		CreatedAt:   time.Date(2022, time.January, 1, 12, 0, 0, 0, time.UTC),
	}
}

type eqCreateProductParamsMatcher struct {
	arg db.CreateProductParams
}

func (e eqCreateProductParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateProductParams)
	if !ok {
		return false
	}

	e.arg.ID = arg.ID // match the random generated uuid

	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateProductParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v", e.arg)
}

func eqCreateProductParams(arg db.CreateProductParams) gomock.Matcher {
	return eqCreateProductParamsMatcher{arg}
}

func requireBodyMatchProduct(t *testing.T, body *bytes.Buffer, product db.Product) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var resProduct db.Product
	err = json.Unmarshal(data, &resProduct)
	require.NoError(t, err)
	require.Equal(t, resProduct.ID, product.ID)
	require.Equal(t, resProduct.UserID, product.UserID)
	require.Equal(t, resProduct.Name, product.Name)
	require.Equal(t, resProduct.Price, product.Price)
	require.Equal(t, resProduct.Description, product.Description)
}

func TestCreateProduct(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)
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
				"user_id":     product.UserID,
				"username":    user.Username,
				"name":        product.Name,
				"price":       floatPrice,
				"description": product.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.CreateProductParams{
					ID:          product.ID,
					UserID:      product.UserID,
					Name:        product.Name,
					Price:       product.Price,
					Description: product.Description,
				}
				store.EXPECT().
					CreateProduct(gomock.Any(), eqCreateProductParams(arg)).
					Times(1).
					Return(product, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchProduct(t, recorder.Body, product)
			},
		},
		{
			name: "InternalError",
			user: user,
			body: gin.H{
				"user_id":     product.UserID,
				"username":    user.Username,
				"name":        product.Name,
				"price":       floatPrice,
				"description": product.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateProduct(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Product{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "MissingJSONField",
			user: user,
			body: gin.H{
				"user_id":  product.UserID,
				"username": user.Username,
				"name":     product.Name,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateProduct(gomock.Any(), gomock.Any()).
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
				"user_id":     product.UserID,
				"username":    user.Username,
				"name":        product.Name,
				"price":       floatPrice,
				"description": product.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "UnauthorizatedUser", time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateProduct(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/%v/products", tc.user.Username)
			req, err := http.NewRequest(http.MethodPost, url, jsonReader)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetAllProducts(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)

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
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAllProducts(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.Product{product}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			user: user,
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAllProducts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Product{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "MissingJSONField",
			user: user,
			body: gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAllProducts(gomock.Any(), gomock.Any()).
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
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "UnauthorizatedUser", time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAllProducts(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/%v/products", tc.user.Username)
			req, err := http.NewRequest(http.MethodGet, url, jsonReader)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestUpdateProduct(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)
	updatedPrice := 10000.00
	updatedProduct := db.Product{
		ID:          product.ID,
		UserID:      user.ID,
		Name:        "updated",
		Price:       utils.FormottedDecimalToString(updatedPrice),
		Description: "updated",
		CreatedAt:   product.CreatedAt,
	}

	testCases := []struct {
		name          string
		user          db.User
		product       db.Product
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			user:    user,
			product: product,
			body: gin.H{
				"user_id":     user.ID,
				"id":          product.ID,
				"name":        updatedProduct.Name,
				"price":       updatedPrice,
				"description": updatedProduct.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.UpdateProductParams{
					UserID:      user.ID,
					ID:          product.ID,
					Name:        updatedProduct.Name,
					Price:       updatedProduct.Price,
					Description: updatedProduct.Description,
				}

				store.EXPECT().
					UpdateProduct(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updatedProduct, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchProduct(t, recorder.Body, updatedProduct)
			},
		},
		{
			name:    "InternalError",
			user:    user,
			product: product,
			body: gin.H{
				"user_id":     user.ID,
				"id":          product.ID,
				"name":        updatedProduct.Name,
				"price":       updatedPrice,
				"description": updatedProduct.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateProduct(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Product{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:    "MissingJSONField",
			user:    user,
			product: product,
			body:    gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateProduct(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:    "UnauthorizatedUser",
			user:    user,
			product: product,
			body: gin.H{
				"user_id":     user.ID,
				"id":          product.ID,
				"name":        updatedProduct.Name,
				"price":       updatedPrice,
				"description": updatedProduct.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "UnauthorizatedUser", time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					UpdateProduct(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/%v/products/%v", tc.user.Username, tc.product.ID)
			req, err := http.NewRequest(http.MethodPatch, url, jsonReader)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	user, _ := randomUser(t)
	product := randomProduct(user)

	testCases := []struct {
		name          string
		user          db.User
		product       db.Product
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			user:    user,
			product: product,
			body: gin.H{
				"user_id":    user.ID,
				"product_id": product.ID,
				"username":   user.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductParams{
					UserID: user.ID,
					ID:     product.ID,
				}
				store.EXPECT().
					DeleteProduct(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:    "InternalError",
			user:    user,
			product: product,
			body: gin.H{
				"user_id":    user.ID,
				"product_id": product.ID,
				"username":   user.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProduct(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:    "MissingJSONField",
			user:    user,
			product: product,
			body:    gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProduct(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:    "UnauthorizatedUser",
			user:    user,
			product: product,
			body: gin.H{
				"user_id":    user.ID,
				"product_id": product.ID,
				"username":   user.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "UnauthorizatedUser", time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProduct(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/%v/products/%v", tc.user.Username, tc.product.ID)
			req, err := http.NewRequest(http.MethodDelete, url, jsonReader)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}
