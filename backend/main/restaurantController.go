package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	f "module/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	buyerIdKey     key = "buyerId"
	dateKey        key = "date"
	productsKey    key = "products"
	pageKey        key = "page"
	pageSizeKey    key = "pageSize"
	pageBKey       key = "pageB"
	pageSizeBKey   key = "pageSizeB"
	pageTKey       key = "pageT"
	pageSizeTKey   key = "pageSizeT"
	buyerParamsKey key = "buyerParams"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		writter.Header().Set("Access-Control-Allow-Origin", f.GoDotEnvVariable("ALLOWED_ORIGIN"))
		writter.Header().Set("Access-Control-Allow-Credentials", "true")
		writter.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		writter.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		writter.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(writter, request)
	})
}

/*
	Extracts the request body and adds it to
	the context so that the handlers can use it.
*/
func restaurantCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		//To solve CORS preflight invalid status error
		if request.Method == http.MethodOptions {
			writter.WriteHeader(http.StatusOK)
			return
		}

		body, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writter, "Error while processing request body", http.StatusInternalServerError)
			return
		}

		var requestBody RequestBody
		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			http.Error(writter, "Error while processing request", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(request.Context(), dateKey, requestBody.Date)
		next.ServeHTTP(writter, request.WithContext(ctx))
	})
}

func loadRestaurantData(writter http.ResponseWriter, request *http.Request) {
	requestContext := request.Context()
	date := requestContext.Value(dateKey).(string)

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	dataLoader := &DataLoader{
		dateStr: date,
		txn:     txn,
	}

	res, errorType, err := startDataLoading(dataLoader)
	if err != nil {
		if errorType == DateError {
			http.Error(writter, "invalid date", http.StatusBadRequest)
		} else {
			http.Error(writter, "error while loading restaurant data", http.StatusInternalServerError)
		}
		return
	}

	writter.WriteHeader(http.StatusCreated)
	writter.Write(res)
}

func buyersCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/buyer/all" {
			pageParam := request.URL.Query().Get(string(pageKey))
			pageSizeParam := request.URL.Query().Get(string(pageSizeKey))

			page, pageSize, err := validatePageParams(pageParam, pageSizeParam)
			if err != nil {
				http.Error(writter, err.Error(), http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(request.Context(), pageKey, page)
			ctx = context.WithValue(ctx, pageSizeKey, pageSize)
			next.ServeHTTP(writter, request.WithContext(ctx))
		} else {
			next.ServeHTTP(writter, request)
		}

	})
}

func getBuyers(writter http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	page := ctx.Value(pageKey).(int)
	pageSize := ctx.Value(pageSizeKey).(int)

	res, err := fetchBuyers(page, pageSize)
	if err != nil {
		http.Error(writter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writter.Write(res)
}

func productsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		products := request.URL.Query().Get(string(productsKey))
		if !isProductParamValid(products) {
			http.Error(writter, "Invalid products", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(request.Context(), productsKey, products)
		next.ServeHTTP(writter, request.WithContext(ctx))
	})
}

func getProducts(writter http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	productIds := ctx.Value(productsKey).(string)

	products, err := fetchProducts(productIds)
	if err != nil {
		http.Error(writter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writter.Write(products)
}

func buyerCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		buyerId := chi.URLParam(request, string(buyerIdKey))

		if !isBuyerIdParamValid(buyerId) {
			http.Error(writter, "Invalid buyerId", http.StatusBadRequest)
			return
		}

		pageBParam := request.URL.Query().Get(string(pageBKey))
		pageSizeBParam := request.URL.Query().Get(string(pageSizeBKey))
		pageTParam := request.URL.Query().Get(string(pageTKey))
		pageSizeTParam := request.URL.Query().Get(string(pageSizeTKey))

		buyerReqParams, err := getBuyerRequestParams(pageBParam,
			pageSizeBParam,
			pageTParam,
			pageSizeTParam)
		if err != nil {
			http.Error(writter, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(request.Context(), buyerIdKey, buyerId)
		ctx = context.WithValue(ctx, buyerParamsKey, buyerReqParams)
		next.ServeHTTP(writter, request.WithContext(ctx))
	})
}

func getBuyer(writter http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	buyerId := ctx.Value(buyerIdKey).(string)
	buyerReqParams := ctx.Value(buyerParamsKey).(BuyerRequestParams)

	buyer, err := fetchBuyer(buyerId, buyerReqParams)
	if err != nil {
		fmt.Printf("error while fetching buyer | %v\n", err)
		http.Error(writter, "Error while fetching buyer", http.StatusInternalServerError)
		return
	}

	writter.Write(buyer)
}
