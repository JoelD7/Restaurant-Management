package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	c "module/constants"
	f "module/utils"
	"strings"
	"time"
)

func fetchBuyersFromDB(page int, pageSize int) ([]byte, error) {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)
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
		return nil, err
	}

	qRes, err := txn.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	type Buyers struct{ Buyers []Buyer }
	var result Buyers
	err = json.Unmarshal(qRes.Json, &result)
	if err != nil {
		return nil, err
	}

	buyersCollection := &BuyerCollection{
		Buyers: result.Buyers,
		Count:  totalBuyers,
	}

	jsonRes, err := json.Marshal(buyersCollection)
	if err != nil {
		return nil, err
	}

	return jsonRes, nil

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

func fetchProductsFromDB(productIds string) ([]byte, error) {
	productList := strings.Split(productIds, ",")

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
		return nil, err
	}

	return res.Json, nil
}

func fetchTransactionHistoryFromDB(buyerId string) (TransactionCollection, error) {
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

func fetchTransactionsForIpsFromDB(ips []string) ([]Transaction, error) {
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

func fetchBuyersByIdFromDB(buyerIds []string, buyerId string) (BuyerCollection, error) {
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

func fetchProductRecommendationsFromDB(buyerTransactions []Transaction) ([]Product, error) {
	var boughtProducts []string

	for _, transaction := range buyerTransactions {
		boughtProducts = append(boughtProducts, transaction.Products...)
	}

	similarProductTransactions, err := fetchSimilarProductTransactionsFromDB(boughtProducts)
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
func fetchSimilarProductTransactionsFromDB(boughtProducts []string) ([]Transaction, error) {
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

func fetchBuyerNameFromDB(buyerId string) (string, error) {
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
