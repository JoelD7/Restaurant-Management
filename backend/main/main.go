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
	d "github.com/shopspring/decimal"
	"google.golang.org/grpc"
)

type Buyer struct {
	BuyerId      string
	Name         string
	Transactions []Transaction
	Type         string `json:"dgraph.type,omitempty"`
}

type Product struct {
	ProductId string
	Name      string
	Date      string
	Price     d.Decimal
}

type Transaction struct {
	TransactionId string
	BuyerId       string
	Ip            string
	Device        string
	Products      []string
	Date          string
	Type          string `json:"dgraph.type,omitempty"`
}

var ctx context.Context = context.Background()
var dgraphClient *dgo.Dgraph = newClient()

func main() {
	txn := dgraphClient.NewTxn()

	defer txn.Discard(ctx)

}

func newClient() *dgo.Dgraph {
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	dc := api.NewDgraphClient(d)
	return dgo.NewDgraphClient(dc)
}

func persistBuyers(txn *dgo.Txn, dateStr string) {

}

func fetchBuyers(txn *dgo.Txn, dateStr string) {
	req, err := http.NewRequest("GET", c.BuyersURL, nil)

	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	var dateAsTimestamp string = fmt.Sprint(f.DateStringToTimestamp(dateStr))
	q.Add("date", dateAsTimestamp)
}

func persistTransactions(txn *dgo.Txn, dateStr string) {

	if !isDateRequestable(dateStr, txn, c.TransactionType) {
		fmt.Println("No need to fetch transactions.")
		return
	}

	jsonTransactions := fetchTransactions(dateStr)

	mutation := &api.Mutation{
		SetJson:   jsonTransactions,
		CommitNow: true,
	}

	fmt.Println("Saving transactions data...")
	_, err := txn.Mutate(context.Background(), mutation)

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
func isDateRequestable(date string, txn *dgo.Txn, nodeType string) bool {
	query := fmt.Sprintf(`{
		q(func: type(%s)) @filter(eq(Date, "%s")){
			uid
		}
	}`, nodeType, date)

	res, err := txn.Query(ctx, query)

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

func fetchTransactions(dateStr string) []byte {

	req, err := http.NewRequest("GET", c.TransactionsURL, nil)

	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	var dateAsTimestamp string = fmt.Sprint(f.DateStringToTimestamp(dateStr))
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
	var transactions []Transaction = parseTransactions(rawTransactions, dateStr)
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

func parseTransactions(rawTransactions []string, dateStr string) []Transaction {
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
			Date:          dateStr,
			Type:          c.TransactionType,
		}

		transactions = append(transactions, newTransaction)

	}

	return transactions
}

func fetchTransactionsFromDB(txn *dgo.Txn, date string) []byte {
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
	  `, date)

	res, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while fetching transactions from the DB: %v\n", err)
	}

	return res.Json
}
