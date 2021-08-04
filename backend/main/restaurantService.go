package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
