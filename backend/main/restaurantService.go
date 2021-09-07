package main

import (
	"encoding/json"
	"fmt"
	c "module/constants"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	DateError  string = "DateError"
	OtherError string = "OtherError"
)

func startDataLoading(dataLoader *DataLoader) ([]byte, string, error) {
	err := isDateParamValid(dataLoader.dateStr)

	if err != nil {
		return nil, DateError, fmt.Errorf("invalid date")
	}

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	validDate, err := dataLoader.isDateRequestable()
	if err != nil {
		return nil, OtherError, err
	}

	if !validDate {
		return nil, DateError, fmt.Errorf("date already synchronized: '%s'", dataLoader.dateStr)
	}

	res, err := dataLoader.loadRestaurantData()
	if err != nil {
		return nil, OtherError, err
	}

	return res, "", nil
}

/*
	Validates the date parameter by checking if it matches
	the layout used by Dgraph to store dates: yyyy-MM-DD.
	Returns an error if it doesn't and nil if it matches.
*/
func isDateParamValid(date string) error {
	//date = ""
	_, err := time.Parse(c.DateLayout, date)
	if err != nil {
		return err
	}

	return nil
}

func validatePageParams(pageParam string, pageSizeParam string) (int, int, error) {
	if pageParam == "" && pageSizeParam == "" {
		return 0, 0, fmt.Errorf("invalid page parameters")
	}

	page, err := strconv.Atoi(pageParam)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page parameters")
	}

	pageSize, err := strconv.Atoi(pageSizeParam)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page parameters")
	}

	if page < 0 || pageSize < 0 {
		return 0, 0, fmt.Errorf("invalid page parameters")
	}

	return page, pageSize, nil

}

func fetchBuyers(page int, pageSize int) ([]byte, error) {
	res, err := fetchBuyersFromDB(page, pageSize)
	if err != nil {
		return nil, err
	}

	return res, nil
}

/*
	Validates that "products" is a comma separated string
	of valid productIds.
*/
func isProductParamValid(products string) bool {
	productArr := strings.Split(products, ",")
	var digitCounter int
	var letterCounter int

	for _, productId := range productArr {
		fmt.Println(productId)

		if len(productId) > 8 || len(productId) == 0 {
			return false
		}

		for _, char := range productId {
			if unicode.IsDigit(char) {
				digitCounter++
			}

			if unicode.IsLetter(char) {
				letterCounter++
			}
		}

		if (digitCounter + letterCounter) != len(productId) {
			return false
		}

		digitCounter = 0
		letterCounter = 0
	}

	return true
}

func fetchProducts(productIds string) ([]byte, error) {
	productsJson, err := fetchProductsFromDB(productIds)
	if err != nil {
		return nil, err
	}

	return productsJson, nil
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

func getBuyerRequestParams(pageBParam string, pageSizeBParam string, pageTParam string, pageSizeTParam string) (BuyerRequestParams, error) {
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

		if pageB <= 0 ||
			pageSizeB <= 0 ||
			pageT <= 0 ||
			pageSizeT <= 0 {
			return BuyerRequestParams{}, fmt.Errorf("invalid page parameters")
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

func fetchBuyer(buyerId string, buyerReqParams BuyerRequestParams) ([]byte, error) {
	buyerTransactions, err := fetchTransactionHistoryFromDB(buyerId)
	if err != nil {
		return nil, err
	}

	var buyerIps []string

	for _, transaction := range buyerTransactions.Transactions {
		buyerIps = append(buyerIps, transaction.Ip)
	}

	transactionsForIps, err := fetchTransactionsForIpsFromDB(buyerIps)
	if err != nil {
		return nil, err
	}

	var buyerIds []string

	for _, transaction := range transactionsForIps {
		buyerIds = append(buyerIds, transaction.BuyerId)
	}

	buyersById, err := fetchBuyersByIdFromDB(buyerIds, buyerId)
	if err != nil {
		fmt.Printf("error while fetching buyer | %v\n", err)
		return nil, err
	}

	recommendedProducts, err := fetchProductRecommendationsFromDB(buyerTransactions.Transactions)
	if err != nil {
		fmt.Printf("error while fetching buyer | %v\n", err)
		return nil, err
	}

	transactionHistory, buyersWithSameIp := getPagedCollections(buyerReqParams, buyerTransactions, buyersById)

	buyerName, err := fetchBuyerNameFromDB(buyerId)
	if err != nil {
		fmt.Printf("error while fetching buyer | %v\n", err)
		return nil, err
	}

	dataToReturn := &BuyerIdEndpoint{
		Name:                buyerName,
		TransactionHistory:  transactionHistory,
		BuyersWithSameIp:    buyersWithSameIp,
		RecommendedProducts: recommendedProducts,
	}

	dataToReturnAsJson, err := json.Marshal(dataToReturn)

	if err != nil {
		fmt.Printf("Error while marshalling dataToReturn | %v\n", err)
		return nil, err
	}

	return dataToReturnAsJson, nil
}

/*
	Applies pagination to the buyer's transactions and to the list of
	buyers using the same IP.
*/
func getPagedCollections(buyerReqParams BuyerRequestParams, buyerTransactions TransactionCollection,
	buyersById BuyerCollection) (TransactionCollection, BuyerCollection) {
	pageB := buyerReqParams.PageBParam
	pageSizeB := buyerReqParams.PageSizeBParam
	pageT := buyerReqParams.PageTParam
	pageSizeT := buyerReqParams.PageSizeTParam

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
