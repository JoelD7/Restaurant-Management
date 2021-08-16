package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	c "module/constants"
	f "module/utils"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type TransactionHolder struct {
	Transactions []Transaction
}

type ProductHolder struct {
	Products []Product
}

type BuyersById struct {
	Buyers []Buyer `json:"buyersById"`
}

type BuyerIdEndpoint struct {
	TransactionHistory  []Transaction
	BuyersWithSameIp    []Buyer
	RecommendedProducts []Product
}

type key string

const buyerIdKey key = "buyerId"
const dateKey key = "date"
const productsKey key = "products"

func CorsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		writter.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		writter.Header().Set("Access-Control-Allow-Credentials", "true")
		writter.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		writter.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		next.ServeHTTP(writter, request)
	})
}

/*
	Extracts the url parameter from the request and adds it to
	the context so that the handlers have can use it.
*/
func RestaurantCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {

		if request.Method == "OPTIONS" {
			writter.WriteHeader(http.StatusOK)
			return
		}

		body, bodyReadErr := io.ReadAll(request.Body)

		if bodyReadErr != nil {
			http.Error(writter, bodyReadErr.Error(), http.StatusUnprocessableEntity)
			return
		}

		var requestBody RequestBody
		json.Unmarshal(body, &requestBody)

		ctx := context.WithValue(request.Context(), dateKey, requestBody.Date)
		next.ServeHTTP(writter, request.WithContext(ctx))
	})
}

func loadRestaurantData(writter http.ResponseWriter, request *http.Request) {
	requestContext := request.Context()
	date, ok := requestContext.Value(dateKey).(string)

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	dataLoader := &DataLoader{
		dateStr: date,
		txn:     txn,
	}

	res, loadErr := dataLoader.loadRestaurantData()
	if loadErr != nil {
		err := fmt.Errorf("error while loading restaurant data: %w", loadErr)

		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	writter.Write([]byte(res))
	writter.Header().Set("Content-Type", "text/plain")

	if !ok {
		http.Error(writter, http.StatusText(http.StatusUnprocessableEntity),
			http.StatusUnprocessableEntity)
		return
	}
}

func getBuyers(writter http.ResponseWriter, request *http.Request) {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	query := `
	{
		buyers(func: type(Buyer)){
			  expand(_all_){}
		}
	  }
	`

	res, err := txn.Query(ctx, query)

	if err != nil {
		http.Error(writter, err.Error(), http.StatusNotFound)
		return
	}

	writter.Write(res.Json)
}

func ProductsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		writter.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		writter.Header().Set("Access-Control-Allow-Credentials", "true")
		writter.Header().Set("Content-Type", "application/json")

		products := request.URL.Query().Get("products")

		ctx := context.WithValue(request.Context(), productsKey, products)
		next.ServeHTTP(writter, request.WithContext(ctx))
	})
}

func getProducts(writter http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	products := ctx.Value(productsKey).(string)
	productList := strings.Split(products, ",")

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	query := fmt.Sprintf(`{
		products(func: type(Product)) 
			@filter(anyofterms(ProductId, "%s")) {
			  expand(_all_){}
		}
	  }`, productList)

	res, err := txn.Query(ctx, query)
	if err != nil {
		http.Error(writter, err.Error(), http.StatusNotFound)
		return
	}

	writter.Write(res.Json)
}

func BuyerCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		buyerId := chi.URLParam(request, "buyerId")

		ctx := context.WithValue(request.Context(), buyerIdKey, buyerId)
		next.ServeHTTP(writter, request.WithContext(ctx))
	})
}

func getBuyer(writter http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	buyerId := ctx.Value(buyerIdKey).(string)

	buyerTransactions, transErr := getTransactionHistory(buyerId)
	if transErr != nil {
		err := fmt.Errorf("error while fetching buyer | %w", transErr)
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var buyerIps []string

	for _, transaction := range buyerTransactions {
		buyerIps = append(buyerIps, transaction.Ip)
	}

	transactionsForIps, transForIpErr := getTransactionsForIps(buyerIps)
	if transForIpErr != nil {
		err := fmt.Errorf("error while fetching buyer | %w", transForIpErr)
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var buyerIds []string

	for _, transaction := range transactionsForIps {
		buyerIds = append(buyerIds, transaction.BuyerId)
	}

	buyersById, buyerErr := getBuyersById(buyerIds)
	if buyerErr != nil {
		err := fmt.Errorf("error while fetching buyer | %w", buyerErr)
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	recommendedProducts, productErr := getProductRecommendations(buyerTransactions)
	if productErr != nil {
		err := fmt.Errorf("error while fetching buyer | %w", productErr)
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	dataToReturn := &BuyerIdEndpoint{
		TransactionHistory:  buyerTransactions,
		BuyersWithSameIp:    buyersById,
		RecommendedProducts: recommendedProducts,
	}

	dataToReturnAsJson, mErr := json.Marshal(dataToReturn)

	if mErr != nil {
		err := fmt.Errorf("error while marshalling dataToReturn: %v", mErr)
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	writter.Write(dataToReturnAsJson)
}

func getTransactionHistory(buyerId string) ([]Transaction, error) {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	query := fmt.Sprintf(`{
		transactions(func: type(Transaction)) 
			@filter(eq(BuyerId, "%s")) {
			  expand(_all_){}
		}
	  }`, buyerId)

	res, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while retrieving transaction history for buyer %s: %v\n", buyerId, err)
		return nil, err
	}

	var transactionHistory TransactionHolder
	uErr := json.Unmarshal(res.Json, &transactionHistory)

	if uErr != nil {
		fmt.Printf("Error while unmarshalling transactions from database | %v", uErr)
		return nil, uErr
	}

	return transactionHistory.Transactions, nil
}

func getTransactionsForIps(ips []string) ([]Transaction, error) {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	query := fmt.Sprintf(`{
		transactions(func: type(Transaction))
			@filter(anyofterms(Ip, "%s")) {
			  expand(_all_){}
		}
	  }`, fmt.Sprint(ips))

	res, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while retrieving transaction for the specified ip addresses: %v\n", err)
		return nil, err
	}

	var transactionsForIps TransactionHolder
	uErr := json.Unmarshal(res.Json, &transactionsForIps)
	if uErr != nil {
		fmt.Printf("Error while unmarshalling transactions for the specified ip addresses | %v\n", uErr)
		return nil, uErr
	}

	return transactionsForIps.Transactions, nil

}

func getBuyersById(buyerIds []string) ([]Buyer, error) {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	query := fmt.Sprintf(`{
		buyersById(func: type(Buyer))
			@filter(anyofterms(BuyerId, "%s")) {
			  BuyerId
			  Age
			  Name
			  Date
		}
	  }`, fmt.Sprint(buyerIds))

	res, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while retrieving buyers: %v\n", err)
		return nil, err
	}

	var buyersById BuyersById
	uErr := json.Unmarshal(res.Json, &buyersById)
	if uErr != nil {
		fmt.Printf("Error while unmarshalling buyersById | %v", uErr)
		return nil, uErr
	}

	return buyersById.Buyers, nil
}

func getProductRecommendations(buyerTransactions []Transaction) ([]Product, error) {
	var boughtProducts []string

	for _, transaction := range buyerTransactions {
		boughtProducts = append(boughtProducts, transaction.Products...)
	}

	similarProductTransactions, productErr := getSimilarProductTransactions(boughtProducts)
	if productErr != nil {
		fmt.Println(productErr)
		return nil, productErr
	}

	var productIdsBuffer []string
	for _, transaction := range similarProductTransactions {
		productIdsBuffer = append(productIdsBuffer, transaction.Products...)
	}

	//Filter out bought products
	productIds := filterBoughtProductIds(boughtProducts, productIdsBuffer)

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	query := fmt.Sprintf(`{
		products(func: type(Product)) 
			@filter(anyofterms(ProductId, "%s")) {
			  expand(_all_){}
		}
	  }`, productIds)

	productsRes, queryErr := txn.Query(ctx, query)
	if queryErr != nil {
		fmt.Printf("Error while fetching products: %v\n", queryErr)
		return nil, queryErr
	}

	var productHolder ProductHolder
	uErr := json.Unmarshal(productsRes.Json, &productHolder)
	if uErr != nil {
		fmt.Printf("Error while unmarshaling products | %v\n", uErr)
		return nil, uErr
	}

	recommendedProducts := filterRepeatedProducts(productHolder.Products)

	return recommendedProducts, nil
}

/*
	Returns transactions that contain products specified in @boughtProducts
*/
func getSimilarProductTransactions(boughtProducts []string) ([]Transaction, error) {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	query := fmt.Sprintf(`{
		transactions(func: type(Transaction), first: 10) 
			@filter(anyofterms(Products, "%s")) {
			  expand(_all_){}
		}
	  }`, boughtProducts)

	transactionsRes, queryErr := txn.Query(ctx, query)
	if queryErr != nil {
		fmt.Printf("Error while fetching transactions with products bought by this buyer: %v\n", queryErr)
		return nil, queryErr
	}

	var transactionsForBuyerProductsRes TransactionHolder
	uErr := json.Unmarshal(transactionsRes.Json, &transactionsForBuyerProductsRes)

	if uErr != nil {
		fmt.Printf("Error while unmarshalling transactions | %v\n", uErr)
		return nil, uErr
	}

	return transactionsForBuyerProductsRes.Transactions, nil
}

func filterBoughtProductIds(boughtProducts []string, productIdsBuffer []string) []string {
	var filteredProductIds []string

	for _, id := range productIdsBuffer {
		if !f.ArrayContains(boughtProducts, id) && !f.ArrayContains(filteredProductIds, id) {
			filteredProductIds = append(filteredProductIds, id)
		}
	}

	return filteredProductIds
}

func filterRepeatedProducts(products []Product) []Product {
	//Shuffle the array so that each time, new recommendations
	//are generated.
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(products), func(i, j int) {
		products[i], products[j] = products[j], products[i]
	})

	var addedIds []string
	var result []Product

	var max int
	if len(products) < c.MaxProductRecommendations {
		max = len(products)
	} else {
		max = c.MaxProductRecommendations
	}

	for _, product := range products {
		if len(addedIds) >= max {
			break
		}

		if !f.ArrayContains(addedIds, product.ProductId) {
			addedIds = append(addedIds, product.ProductId)
			result = append(result, product)
		}
	}

	return result
}
