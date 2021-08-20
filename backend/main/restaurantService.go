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
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

const buyerIdKey key = "buyerId"
const dateKey key = "date"
const productsKey key = "products"
const pageKey key = "page"
const pageSizeKey key = "pageSize"
const pageBKey key = "pageB"
const pageSizeBKey key = "pageSizeB"
const pageTKey key = "pageT"
const pageSizeTKey key = "pageSizeT"

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		writter.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
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
func RestaurantCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {

		//To solve CORS preflight invalid status error
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
		uErr := json.Unmarshal(body, &requestBody)
		if uErr != nil {
			http.Error(writter, uErr.Error(), http.StatusUnprocessableEntity)
		}

		ctx := context.WithValue(request.Context(), dateKey, requestBody.Date)
		next.ServeHTTP(writter, request.WithContext(ctx))
	})
}

func loadRestaurantData(writter http.ResponseWriter, request *http.Request) {
	requestContext := request.Context()
	date, okParam := requestContext.Value(dateKey).(string)

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	dataLoader := &DataLoader{
		dateStr: date,
		txn:     txn,
	}

	validDate, dateErr := dataLoader.isDateRequestable()
	if dateErr != nil {
		http.Error(writter, dateErr.Error(), http.StatusUnprocessableEntity)
		return
	}

	if !validDate {
		http.Error(writter, fmt.Sprintf("La fecha '%s' ya ha sido sincronizada", dataLoader.dateStr), http.StatusBadRequest)
		return
	}

	res, loadErr := dataLoader.loadRestaurantData()
	if loadErr != nil {
		err := fmt.Errorf("error while loading restaurant data: %w", loadErr)

		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	writter.WriteHeader(http.StatusCreated)
	writter.Write([]byte(res))

	if !okParam {
		http.Error(writter, http.StatusText(http.StatusUnprocessableEntity),
			http.StatusUnprocessableEntity)
		return
	}
}

func getBuyers(writter http.ResponseWriter, request *http.Request) {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	ctx := request.Context()
	page := ctx.Value(pageKey).(int)
	pageSize := ctx.Value(pageSizeKey).(int)
	offset := pageSize * page

	query := fmt.Sprintf(`
	{
		buyers(func: type(Buyer), offset: %v, first: %v){
			  expand(_all_){}
		}
	  }
	`, offset, pageSize)

	countQuery := `
	{
		CountArray(func: type(Buyer)){
			  total: count(uid)
		}
	  }
	`

	totalBuyers, countErr := countEntities(countQuery)
	if countErr != nil {
		http.Error(writter, countErr.Error(), http.StatusUnprocessableEntity)
		return
	}

	qRes, qErr := txn.Query(ctx, query)
	if qErr != nil {
		http.Error(writter, qErr.Error(), http.StatusUnprocessableEntity)
		return
	}

	type Buyers struct{ Buyers []Buyer }
	var result Buyers
	uErr := json.Unmarshal(qRes.Json, &result)
	if uErr != nil {
		http.Error(writter, uErr.Error(), http.StatusUnprocessableEntity)
		return
	}

	response := &BuyerCollection{
		Buyers: result.Buyers,
		Count:  totalBuyers,
	}

	jsonRes, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		http.Error(writter, jsonErr.Error(), http.StatusUnprocessableEntity)
		return
	}

	writter.Write(jsonRes)
}

func countEntities(countQuery string) (int, error) {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	cRes, cErr := txn.Query(ctx, countQuery)
	if cErr != nil {
		return 0, cErr
	}

	var collectionCount CollectionCount
	uErr := json.Unmarshal(cRes.Json, &collectionCount)
	if uErr != nil {
		return 0, uErr
	}

	return collectionCount.CountArray[0].Total, nil
}

func ProductsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
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

		buyerReqParams, err := getBuyerRequestParams(writter, request)
		if err != nil {
			http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		ctx := context.WithValue(request.Context(), buyerIdKey, buyerId)
		ctx = context.WithValue(ctx, pageBKey, buyerReqParams.PageBParam)
		ctx = context.WithValue(ctx, pageSizeBKey, buyerReqParams.PageSizeBParam)
		ctx = context.WithValue(ctx, pageTKey, buyerReqParams.PageTParam)
		ctx = context.WithValue(ctx, pageSizeTKey, buyerReqParams.PageSizeTParam)
		next.ServeHTTP(writter, request.WithContext(ctx))
	})
}

func getBuyerRequestParams(writter http.ResponseWriter, request *http.Request) (BuyerRequestParams, error) {
	pageBParam := request.URL.Query().Get("pageB")
	pageSizeBParam := request.URL.Query().Get("pageSizeB")
	pageTParam := request.URL.Query().Get("pageT")
	pageSizeTParam := request.URL.Query().Get("pageSizeT")

	if pageBParam != "" && pageSizeBParam != "" && pageTParam != "" && pageSizeTParam != "" {
		pageB, pageBErr := strconv.Atoi(pageBParam)
		if pageBErr != nil {
			return BuyerRequestParams{}, pageBErr
		}
		pageSizeB, pageSizeBErr := strconv.Atoi(pageSizeBParam)
		if pageSizeBErr != nil {
			return BuyerRequestParams{}, pageSizeBErr
		}
		pageT, pageTErr := strconv.Atoi(pageTParam)
		if pageTErr != nil {
			return BuyerRequestParams{}, pageTErr
		}
		pageSizeT, pageSizeTErr := strconv.Atoi(pageSizeTParam)
		if pageSizeTErr != nil {
			return BuyerRequestParams{}, pageSizeTErr
		}

		return BuyerRequestParams{
			PageBParam:     pageB,
			PageSizeBParam: pageSizeB,
			PageTParam:     pageT,
			PageSizeTParam: pageSizeT,
		}, nil
	}

	return BuyerRequestParams{}, fmt.Errorf("missing parameter")
}

func BuyersCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		pageParam := request.URL.Query().Get("page")
		pageSizeParam := request.URL.Query().Get("pageSize")

		if pageParam != "" && pageSizeParam != "" {
			page, e := strconv.Atoi(pageParam)
			if e != nil {
				http.Error(writter, e.Error(), http.StatusUnprocessableEntity)
				return
			}

			pageSize, er := strconv.Atoi(pageSizeParam)
			if er != nil {
				http.Error(writter, er.Error(), http.StatusUnprocessableEntity)
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

func getBuyer(writter http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	buyerId := ctx.Value(buyerIdKey).(string)
	pageB := ctx.Value(pageBKey).(int)
	pageSizeB := ctx.Value(pageSizeBKey).(int)
	pageT := ctx.Value(pageTKey).(int)
	pageSizeT := ctx.Value(pageSizeTKey).(int)

	buyerTransactions, transErr := getTransactionHistory(buyerId)
	if transErr != nil {
		err := fmt.Errorf("error while fetching buyer | %w", transErr)
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var buyerIps []string

	for _, transaction := range buyerTransactions.Transactions {
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

	recommendedProducts, productErr := getProductRecommendations(buyerTransactions.Transactions)
	if productErr != nil {
		err := fmt.Errorf("error while fetching buyer | %w", productErr)
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var transactionHistory TransactionCollection

	if ((pageT-1)*pageSizeT)+pageSizeT < buyerTransactions.Count {
		transactionHistory = TransactionCollection{
			Transactions: buyerTransactions.Transactions[(pageT-1)*pageSizeT : ((pageT-1)*pageSizeT)+pageSizeT],
			Count:        buyerTransactions.Count,
		}
	} else {
		transactionHistory = TransactionCollection{
			Transactions: buyerTransactions.Transactions[(pageT-1)*pageSizeT:],
			Count:        buyerTransactions.Count,
		}
	}

	var buyersWithSameIp BuyerCollection

	if ((pageB-1)*pageSizeB)+pageSizeB < buyersById.Count {
		buyersWithSameIp = BuyerCollection{
			Buyers: buyersById.Buyers[(pageB-1)*pageSizeB : ((pageB-1)*pageSizeB)+pageSizeB],
			Count:  buyersById.Count,
		}
	} else {
		buyersWithSameIp = BuyerCollection{
			Buyers: buyersById.Buyers[(pageB-1)*pageSizeB:],
			Count:  buyersById.Count,
		}
	}

	dataToReturn := &BuyerIdEndpoint{
		TransactionHistory:  transactionHistory,
		BuyersWithSameIp:    buyersWithSameIp,
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

func getTransactionHistory(buyerId string) (TransactionCollection, error) {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	query := fmt.Sprintf(`{
		transactions(func: type(Transaction)) 
			@filter(eq(BuyerId, "%s")) {
			  expand(_all_){}
		}
	  }`, buyerId)

	countQuery := fmt.Sprintf(`
	  {
		  CountArray(func: type(Transaction))
			  @filter(eq(BuyerId, "%s")){
				total: count(uid)
		  }
		}
	  `, buyerId)

	totalTransactions, countErr := countEntities(countQuery)
	if countErr != nil {
		return TransactionCollection{}, countErr
	}

	res, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while retrieving transaction history for buyer %s: %v\n", buyerId, err)
		return TransactionCollection{}, err
	}

	var transactionHistory TransactionHolder
	uErr := json.Unmarshal(res.Json, &transactionHistory)

	if uErr != nil {
		fmt.Printf("Error while unmarshalling transactions from database | %v", uErr)
		return TransactionCollection{}, uErr
	}

	return TransactionCollection{
		Transactions: transactionHistory.Transactions,
		Count:        totalTransactions,
	}, nil
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

func getBuyersById(buyerIds []string) (BuyerCollection, error) {
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

	countQuery := fmt.Sprintf(`
	  {
		  CountArray(func: type(Buyer))
		  	@filter(anyofterms(BuyerId, "%s")){
				total: count(uid)
		  }
		}
	  `, fmt.Sprint(buyerIds))

	totalBuyers, countErr := countEntities(countQuery)
	if countErr != nil {
		return BuyerCollection{}, countErr
	}

	res, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while retrieving buyers: %v\n", err)
		return BuyerCollection{}, err
	}

	var buyersById BuyersById
	uErr := json.Unmarshal(res.Json, &buyersById)
	if uErr != nil {
		fmt.Printf("Error while unmarshalling buyersById | %v", uErr)
		return BuyerCollection{}, uErr
	}

	return BuyerCollection{
		Buyers: buyersById.Buyers,
		Count:  totalBuyers,
	}, nil
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
	var result []Product = []Product{}

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
