package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dgraph-io/dgo/v2"
	"github.com/go-chi/chi/v5"
)

type TransactionHistory struct {
	Transactions []Transaction
}

type TransactionsForIps struct {
	TransactionsForIps []Transaction
}

type RestaurantService struct {
	txn *dgo.Txn
}

/*
	Extracts the url parameter from the request and adds it to
	the context so that the handlers have can use it.
*/
func RestaurantCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
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

		ctx := context.WithValue(request.Context(), "date", requestBody.Date)
		next.ServeHTTP(writter, request.WithContext(ctx))
	})
}

func loadRestaurantData(writter http.ResponseWriter, request *http.Request) {
	requestContext := request.Context()
	date, ok := requestContext.Value("date").(string)

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
	writter.Header().Set("Content-Type", "application/json")

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

func BuyerCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		writter.Header().Set("Content-Type", "application/json")

		buyerId := chi.URLParam(request, "buyerId")

		ctx := context.WithValue(request.Context(), "buyerId", buyerId)
		next.ServeHTTP(writter, request.WithContext(ctx))
	})
}

func getBuyer(writter http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	buyerId := ctx.Value("buyerId").(string)

	buyerTransactions := getTransactionHistory(buyerId)
	var buyerIps []string

	for _, transaction := range buyerTransactions {
		buyerIps = append(buyerIps, transaction.Ip)
	}

	transactionsForIps := getTransactionsForIps(buyerIps)
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

	var transactionHistory TransactionHistory

	json.Unmarshal(res.Json, &transactionHistory)

	return transactionHistory.Transactions
}

func getTransactionsForIps(ips []string) []Transaction {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	query := fmt.Sprintf(`{
		transactionsForIps(func: type(Transaction))
			@filter(anyofterms(Ip, "%s")) {
			  expand(_all_){}
		}
	  }`, fmt.Sprint(ips))

	res, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while retrieving transaction for the specified ip addresses: %v\n", err)
	}

	var transactionsForIps TransactionsForIps
	json.Unmarshal(res.Json, &transactionsForIps)

	return transactionsForIps.TransactionsForIps

}
