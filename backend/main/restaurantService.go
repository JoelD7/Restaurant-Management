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
	"unicode"

	"github.com/go-chi/chi/v5"
)

const (
	buyerIdKey   key = "buyerId"
	dateKey      key = "date"
	productsKey  key = "products"
	pageKey      key = "page"
	pageSizeKey  key = "pageSize"
	pageBKey     key = "pageB"
	pageSizeBKey key = "pageSizeB"
	pageTKey     key = "pageT"
	pageSizeTKey key = "pageSizeT"
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
			http.Error(writter, err.Error(), http.StatusBadRequest)
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
	date, okParam := requestContext.Value(dateKey).(string)
	err := isDateParamValid(date)

	if err != nil {
		http.Error(writter, "Invalid date", http.StatusBadRequest)
		return
	}

	if !okParam {
		http.Error(writter, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	dataLoader := &DataLoader{
		dateStr: date,
		txn:     txn,
	}

	validDate, err := dataLoader.isDateRequestable()
	if err != nil {
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if !validDate {
		http.Error(writter, fmt.Sprintf("date already synchronized: '%s'", dataLoader.dateStr), http.StatusBadRequest)
		return
	}

	res, err := dataLoader.loadRestaurantData()
	if err != nil {
		err = fmt.Errorf("error while loading restaurant data: %w", err)

		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	writter.WriteHeader(http.StatusCreated)
	writter.Write(res)

}

/*
	Validates the date parameter by checking if it matches
	the layout used by Dgraph to store dates: yyyy-MM-DD.
	Returns an error if it doesn't and nil if it matches.
*/
func isDateParamValid(date string) error {
	_, err := time.Parse(c.DateLayout, date)
	if err != nil {
		return err
	}

	return nil
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

	totalBuyers, err := countEntities(countQuery)
	if err != nil {
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	qRes, err := txn.Query(ctx, query)
	if err != nil {
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	type Buyers struct{ Buyers []Buyer }
	var result Buyers
	err = json.Unmarshal(qRes.Json, &result)
	if err != nil {
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	response := &BuyerCollection{
		Buyers: result.Buyers,
		Count:  totalBuyers,
	}

	jsonRes, err := json.Marshal(response)
	if err != nil {
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	writter.Write(jsonRes)
}

func countEntities(countQuery string) (int, error) {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	cRes, err := txn.Query(ctx, countQuery)
	if err != nil {
		return 0, err
	}

	var collectionCount CollectionCount
	err = json.Unmarshal(cRes.Json, &collectionCount)
	if err != nil {
		return 0, err
	}

	return collectionCount.CountArray[0].Total, nil
}

func productsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		products := request.URL.Query().Get(string(productsKey))

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

func buyerCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		buyerId := chi.URLParam(request, string(buyerIdKey))

		if !isBuyerIdParamValid(buyerId) {
			http.Error(writter, "Invalid buyerId", http.StatusBadRequest)
			return
		}

		buyerReqParams, err := getBuyerRequestParams(writter, request)
		if err != nil {
			http.Error(writter, "Invalid request parameters", http.StatusBadRequest)
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

/*
	Validates de "buyerId" parameter by determining if it
	is an alphanumeric string and if it has the expected length.
*/
func isBuyerIdParamValid(buyerId string) bool {
	if len(buyerId) > 8 {
		return false
	}

	var digitCounter int
	var letterCounter int
	for _, char := range buyerId {
		if unicode.IsDigit(char) {
			digitCounter++
		}

		if unicode.IsLetter(char) {
			letterCounter++
		}
	}

	return (digitCounter + letterCounter) == len(buyerId)
}

func getBuyerRequestParams(writter http.ResponseWriter, request *http.Request) (BuyerRequestParams, error) {
	pageBParam := request.URL.Query().Get(string(pageBKey))
	pageSizeBParam := request.URL.Query().Get(string(pageSizeBKey))
	pageTParam := request.URL.Query().Get(string(pageTKey))
	pageSizeTParam := request.URL.Query().Get(string(pageSizeTKey))

	if pageBParam != "" && pageSizeBParam != "" && pageTParam != "" && pageSizeTParam != "" {
		pageB, err := strconv.Atoi(pageBParam)
		if err != nil {
			return BuyerRequestParams{}, err
		}
		pageSizeB, err := strconv.Atoi(pageSizeBParam)
		if err != nil {
			return BuyerRequestParams{}, err
		}
		pageT, err := strconv.Atoi(pageTParam)
		if err != nil {
			return BuyerRequestParams{}, err
		}
		pageSizeT, err := strconv.Atoi(pageSizeTParam)
		if err != nil {
			return BuyerRequestParams{}, err
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

func buyersCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		pageParam := request.URL.Query().Get(string(pageKey))
		pageSizeParam := request.URL.Query().Get(string(pageSizeKey))

		if pageParam != "" && pageSizeParam != "" {
			page, err := strconv.Atoi(pageParam)
			if err != nil {
				http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
				return
			}

			pageSize, err := strconv.Atoi(pageSizeParam)
			if err != nil {
				http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
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

	buyerTransactions, err := getTransactionHistory(buyerId)
	if err != nil {
		err := fmt.Errorf("error while fetching buyer | %w", err)
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var buyerIps []string

	for _, transaction := range buyerTransactions.Transactions {
		buyerIps = append(buyerIps, transaction.Ip)
	}

	transactionsForIps, err := getTransactionsForIps(buyerIps)
	if err != nil {
		err := fmt.Errorf("error while fetching buyer | %w", err)
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var buyerIds []string

	for _, transaction := range transactionsForIps {
		buyerIds = append(buyerIds, transaction.BuyerId)
	}

	buyersById, err := getBuyersById(buyerIds, buyerId)
	if err != nil {
		err = fmt.Errorf("error while fetching buyer | %w", err)
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	recommendedProducts, err := getProductRecommendations(buyerTransactions.Transactions)
	if err != nil {
		err := fmt.Errorf("error while fetching buyer | %w", err)
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	transactionHistory, buyersWithSameIp := getPagedCollections(request, buyerTransactions, buyersById)

	buyerName, err := fetchBuyerName(buyerId)
	if err != nil {
		http.Error(writter, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	dataToReturn := &BuyerIdEndpoint{
		Name:                buyerName,
		TransactionHistory:  transactionHistory,
		BuyersWithSameIp:    buyersWithSameIp,
		RecommendedProducts: recommendedProducts,
	}

	dataToReturnAsJson, err := json.Marshal(dataToReturn)

	if err != nil {
		err = fmt.Errorf("error while marshalling dataToReturn: %v", err)
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

	totalTransactions, err := countEntities(countQuery)
	if err != nil {
		return TransactionCollection{}, err
	}

	res, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while retrieving transaction history for buyer %s: %v\n", buyerId, err)
		return TransactionCollection{}, err
	}

	var transactionHistory TransactionHolder
	err = json.Unmarshal(res.Json, &transactionHistory)

	if err != nil {
		fmt.Printf("Error while unmarshalling transactions from database | %v", err)
		return TransactionCollection{}, err
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
	err = json.Unmarshal(res.Json, &transactionsForIps)
	if err != nil {
		fmt.Printf("Error while unmarshalling transactions for the specified ip addresses | %v\n", err)
		return nil, err
	}

	return transactionsForIps.Transactions, nil

}

func getBuyersById(buyerIds []string, buyerId string) (BuyerCollection, error) {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	query := fmt.Sprintf(`{
		buyersById(func: type(Buyer))
			@filter(anyofterms(BuyerId, "%s") and not anyofterms(BuyerId, "%s")) {
			  BuyerId
			  Age
			  Name
			  Date
		}
	}`, fmt.Sprint(buyerIds), buyerId)

	countQuery := fmt.Sprintf(`{
	CountArray(func: type(Buyer))
			@filter(anyofterms(BuyerId, "%s") and not anyofterms(BuyerId, "%s")) {
				total: count(uid)
		}
	}`, fmt.Sprint(buyerIds), buyerId)

	totalBuyers, err := countEntities(countQuery)
	if err != nil {
		return BuyerCollection{}, err
	}

	res, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while retrieving buyers: %v\n", err)
		return BuyerCollection{}, err
	}

	var buyersById BuyersById
	err = json.Unmarshal(res.Json, &buyersById)
	if err != nil {
		fmt.Printf("Error while unmarshalling buyersById | %v", err)
		return BuyerCollection{}, err
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

	similarProductTransactions, err := getSimilarProductTransactions(boughtProducts)
	if err != nil {
		fmt.Println(err)
		return nil, err
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

	productsRes, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while fetching products: %v\n", err)
		return nil, err
	}

	var productHolder ProductHolder
	err = json.Unmarshal(productsRes.Json, &productHolder)
	if err != nil {
		fmt.Printf("Error while unmarshaling products | %v\n", err)
		return nil, err
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

	transactionsRes, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while fetching transactions with products bought by this buyer: %v\n", err)
		return nil, err
	}

	var transactionsForBuyerProductsRes TransactionHolder
	err = json.Unmarshal(transactionsRes.Json, &transactionsForBuyerProductsRes)

	if err != nil {
		fmt.Printf("Error while unmarshalling transactions | %v\n", err)
		return nil, err
	}

	return transactionsForBuyerProductsRes.Transactions, nil
}

/*
	Applies pagination to the buyer's transactions and to the list of
	buyers using the same IP.
*/
func getPagedCollections(request *http.Request, buyerTransactions TransactionCollection,
	buyersById BuyerCollection) (TransactionCollection, BuyerCollection) {
	ctx := request.Context()
	pageB := ctx.Value(pageBKey).(int)
	pageSizeB := ctx.Value(pageSizeBKey).(int)
	pageT := ctx.Value(pageTKey).(int)
	pageSizeT := ctx.Value(pageSizeTKey).(int)

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

	return transactionHistory, buyersWithSameIp
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

func fetchBuyerName(buyerId string) (string, error) {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	query := fmt.Sprintf(`{
		buyerName(func: type(Buyer))
			@filter(eq(BuyerId, "%s")) {
			  Name
		}
	}`, buyerId)

	type BuyerName struct {
		BuyerName []struct {
			Name string
		}
	}
	var bn BuyerName

	res, err := txn.Query(ctx, query)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(res.Json, &bn)
	if err != nil {
		return "", err
	}

	return bn.BuyerName[0].Name, nil
}
