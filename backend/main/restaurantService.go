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

/*
	Extracts the url parameter from the request and adds it to
	the context so that the handlers have can use it.
*/
func RestaurantCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		writter.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		writter.Header().Set("Access-Control-Allow-Credentials", "true")

		body, err := io.ReadAll(request.Body)

		var requestBody RequestBody
		json.Unmarshal(body, &requestBody)

		if err != nil {
			fmt.Println(err)
		}

		if err != nil {
			http.Error(writter, http.StatusText(404), 404)
			return
		}

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

	writter.Write([]byte(dataLoader.loadRestaurantData()))
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
		fmt.Printf("An error has ocurred while trying to fetch all the buyers: %v\n", err)
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
		fmt.Printf("Error while retrieving products: %v\n", err)
	}

	writter.Write(res.Json)
}

func BuyerCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		writter.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		writter.Header().Set("Access-Control-Allow-Credentials", "true")
		writter.Header().Set("Content-Type", "application/json")

		buyerId := chi.URLParam(request, "buyerId")

		ctx := context.WithValue(request.Context(), buyerIdKey, buyerId)
		next.ServeHTTP(writter, request.WithContext(ctx))
	})
}

func getBuyer(writter http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	buyerId := ctx.Value(buyerIdKey).(string)

	buyerTransactions := getTransactionHistory(buyerId)
	var buyerIps []string

	for _, transaction := range buyerTransactions {
		buyerIps = append(buyerIps, transaction.Ip)
	}

	transactionsForIps := getTransactionsForIps(buyerIps)
	var buyerIds []string

	for _, transaction := range transactionsForIps {
		buyerIds = append(buyerIds, transaction.BuyerId)
	}

	buyersById := getBuyersById(buyerIds)

	recommendedProducts := getProductRecommendations(buyerTransactions)

	dataToReturn := &BuyerIdEndpoint{
		TransactionHistory:  buyerTransactions,
		BuyersWithSameIp:    buyersById,
		RecommendedProducts: recommendedProducts,
	}

	dataToReturnAsJson, err := json.Marshal(dataToReturn)

	if err != nil {
		fmt.Printf("Error while marshalling dataToReturn: %v\n", err)
	}

	writter.Write(dataToReturnAsJson)
}

func getTransactionHistory(buyerId string) []Transaction {
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
	}

	var transactionHistory TransactionHolder
	json.Unmarshal(res.Json, &transactionHistory)

	return transactionHistory.Transactions
}

func getTransactionsForIps(ips []string) []Transaction {
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
	}

	var transactionsForIps TransactionHolder
	json.Unmarshal(res.Json, &transactionsForIps)

	return transactionsForIps.Transactions

}

func getBuyersById(buyerIds []string) []Buyer {
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
	}

	var buyersById BuyersById
	json.Unmarshal(res.Json, &buyersById)

	return buyersById.Buyers
}

func getProductRecommendations(buyerTransactions []Transaction) []Product {
	var boughtProducts []string

	for _, transaction := range buyerTransactions {
		boughtProducts = append(boughtProducts, transaction.Products...)
	}

	similarProductTransactions := getSimilarProductTransactions(boughtProducts)

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

	productsRes, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while fetching products: %v\n", err)
	}

	var productHolder ProductHolder
	json.Unmarshal(productsRes.Json, &productHolder)

	recommendedProducts := filterRepeatedProducts(productHolder.Products)

	return recommendedProducts
}

/*
	Returns transactions that contain products specified in @boughtProducts
*/
func getSimilarProductTransactions(boughtProducts []string) []Transaction {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	query := fmt.Sprintf(`{
		transactions(func: type(Transaction)) 
			@filter(anyofterms(Products, "%s")) {
			  expand(_all_){}
		}
	  }`, boughtProducts)

	transactionsRes, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while fetching transactions with products bought by this buyer: %v\n", err)
	}

	var transactionsForBuyerProductsRes TransactionHolder
	json.Unmarshal(transactionsRes.Json, &transactionsForBuyerProductsRes)

	return transactionsForBuyerProductsRes.Transactions
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
