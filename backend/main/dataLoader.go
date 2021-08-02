package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	c "module/constants"
	f "module/utils"
	"net/http"
	"regexp"
	"strings"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

type DataLoader struct {
	dateStr string
	txn     *dgo.Txn
}

func (dataLoader *DataLoader) persistBuyers() {

}

func (dataLoader *DataLoader) fetchBuyers() {
	req, err := http.NewRequest("GET", c.BuyersURL, nil)
	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	var dateAsTimestamp string = fmt.Sprint(f.DateStringToTimestamp(dataLoader.dateStr))
	q.Add("date", dateAsTimestamp)
}

func (dataLoader *DataLoader) fetchTransactions() []byte {

	req, err := http.NewRequest("GET", c.TransactionsURL, nil)

	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	var dateAsTimestamp string = fmt.Sprint(f.DateStringToTimestamp(dataLoader.dateStr))
	q.Add("date", dateAsTimestamp)

	req.URL.RawQuery = q.Encode()
	query := req.URL.String()

	resp, err := http.Get(query)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	rawTransactions := strings.Split(string(body), "#")
	var transactions []Transaction = dataLoader.parseTransactions(rawTransactions)
	rawJsonTransactions, err := json.Marshal(transactions)

	if err != nil {
		fmt.Println(err)
	}

	/*
		Replace all appearances of the unicode null character: \u0000 with an
		empty string.
	*/
	jsonTransactions := strings.Replace(string(rawJsonTransactions), "\\u0000", "", -1)
	return []byte(jsonTransactions)

}

func (dataLoader *DataLoader) fetchTransactionsFromDB() []byte {
	query := fmt.Sprintf(`{
		transactions(func: type(Transaction)) @filter(eq(Date, "%s")){
		  expand(_all_){
			   expand(_all_){
				expand(_all_){
				  }
			  } 
		  }
		}
	  }
	  `, dataLoader.dateStr)

	res, err := dataLoader.txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while fetching transactions from the DB: %v\n", err)
	}

	return res.Json
}

func (dataLoader *DataLoader) parseTransactions(rawTransactions []string) []Transaction {
	var transactions []Transaction
	transactionsQty := len(rawTransactions)

	for i := 0; i < transactionsQty; i++ {
		line := rawTransactions[i]
		size := len(line)

		if size == 0 {
			continue
		}

		transactionId := line[0:12]
		buyerId := line[12:21]
		deviceRgx := regexp.MustCompile(`[a-z]`)
		productRgx := regexp.MustCompile(`\(`)

		deviceIndex := deviceRgx.FindStringIndex(line[21 : size-1])[0]
		if deviceIndex == 0 {
			continue
		}

		productIndex := productRgx.FindStringIndex(line[21 : size-1])[0]

		ip := line[21 : deviceIndex+21]
		device := line[deviceIndex+21 : productIndex+21]
		products := line[productIndex+21:]
		products = products[1 : len(products)-3]

		newTransaction := Transaction{
			BuyerId:       buyerId,
			TransactionId: transactionId,
			Ip:            ip,
			Device:        device,
			Products:      strings.Split(products, ","),
			Date:          dataLoader.dateStr,
			Type:          c.TransactionType,
		}

		transactions = append(transactions, newTransaction)

	}

	return transactions
}

func (dataLoader *DataLoader) persistTransactions() {

	if !dataLoader.isDateRequestable(c.TransactionType) {
		fmt.Println("No need to fetch transactions.")
		return
	}

	jsonTransactions := dataLoader.fetchTransactions()

	mutation := &api.Mutation{
		SetJson:   jsonTransactions,
		CommitNow: true,
	}

	fmt.Println("Saving transactions data...")
	_, err := dataLoader.txn.Mutate(context.Background(), mutation)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("New transactions saved.")
}

/*
	Determines if the database has information about the
	requested node based on the date queried by the client.
	In that case, a request to AWS is not necessary.
*/
func (dataLoader *DataLoader) isDateRequestable(nodeType string) bool {
	query := fmt.Sprintf(`{
		q(func: type(%s)) @filter(eq(Date, "%s")){
			uid
		}
	}`, nodeType, dataLoader.dateStr)

	res, err := dataLoader.txn.Query(ctx, query)

	if err != nil {
		fmt.Println(err)
	}

	resultSize := res.Metrics.NumUids["uid"]

	if resultSize > 0 {
		return false
	} else {
		return true
	}
}
