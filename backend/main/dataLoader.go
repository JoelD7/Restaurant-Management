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
	"sync"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	d "github.com/shopspring/decimal"
)

type Buyer struct {
	BuyerId string `json:"id,omitempty"`
	Age     int
	Name    string
	Date    string `json:"Date,omitempty"`
	Type    string `json:"dgraph.type,omitempty"`
}

type Product struct {
	ProductId string
	Name      string
	Date      string
	Price     d.Decimal
	Type      string `json:"dgraph.type,omitempty"`
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

type DataLoader struct {
	dateStr string
	txn     *dgo.Txn
}

func (dataLoader *DataLoader) loadRestaurantData() string {
	if !dataLoader.isDateRequestable() {
		fmt.Printf("The restaurant data for date %s has already loaded.\n", dataLoader.dateStr)
		return fmt.Sprintf("The restaurant data for date %s has already loaded.\n", dataLoader.dateStr)
	}

	functions := make([]func(), 0)
	functions = append(functions, dataLoader.loadBuyers)
	functions = append(functions, dataLoader.loadTransactions)
	functions = append(functions, dataLoader.loadProducts)

	waitGroup := sync.WaitGroup{}

	for i := range functions {
		waitGroup.Add(1)

		go func(function func()) {
			function()

			waitGroup.Done()
		}(functions[i])
	}

	waitGroup.Wait()

	return "All data succesfully loaded"
}

func (dataLoader *DataLoader) loadProducts() {
	fmt.Println("Loading products...")
	req, err := http.NewRequest("GET", c.ProductURL, nil)
	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	var dateAsTimestamp string = fmt.Sprint(f.DateStringToTimestamp(dataLoader.dateStr))
	q.Add("date", dateAsTimestamp)

	req.URL.RawQuery = q.Encode()
	requestUrl := req.URL.String()

	resp, err := http.Get(requestUrl)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	rawProductsLines := strings.Split(string(body), "\n")
	var products []Product = dataLoader.parseProducts(rawProductsLines)
	jsonProducts, _ := json.Marshal(products)

	dataLoader.persistProducts(jsonProducts)
}

func (dataLoader *DataLoader) parseProducts(rawProductsLines []string) []Product {
	var products []Product

	for _, line := range rawProductsLines {
		if len(strings.Split(line, "'")) < 3 {
			continue
		}

		id := strings.Split(line, "'")[0]
		name := strings.Split(line, "'")[1]
		price, _ := d.NewFromString(strings.Split(line, "'")[2])

		newProduct := Product{
			ProductId: id,
			Name:      name,
			Price:     price,
			Date:      dataLoader.dateStr,
			Type:      c.ProductType,
		}

		products = append(products, newProduct)
	}

	return products
}

func (dataLoader *DataLoader) persistProducts(jsonProducts []byte) {
	mutation := &api.Mutation{
		SetJson:   jsonProducts,
		CommitNow: true,
	}

	req := &api.Request{
		Mutations: []*api.Mutation{mutation},
		CommitNow: true,
	}

	_, err := newClient().NewTxn().Do(context.Background(), req)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Products loaded.")
}

func (dataLoader *DataLoader) loadBuyers() {
	fmt.Println("Loading buyers...")
	req, err := http.NewRequest("GET", c.BuyersURL, nil)
	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	var dateAsTimestamp string = fmt.Sprint(f.DateStringToTimestamp(dataLoader.dateStr))
	q.Add("date", dateAsTimestamp)

	req.URL.RawQuery = q.Encode()
	requestUrl := req.URL.String()

	resp, err := http.Get(requestUrl)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var buyers []Buyer
	uErr := json.Unmarshal(body, &buyers)
	if uErr != nil {
		log.Fatal(uErr)
	}

	jsonBuyers, _ := dataLoader.marshalJSON(&buyers)
	dataLoader.persistBuyers(jsonBuyers)
}

/*
	Custom marshaller to change the json tag of BuyerId field
	to "BuyerId".

	In the type declaration of the Buyer struct, the json tag of
	said field is "id" so that it matches the data returned from
	AWS. A change in the tag name is required so that the buyers json
	matches the structure of the Buyer node in the database.
*/
func (dataLoader *DataLoader) marshalJSON(buyers *[]Buyer) ([]byte, error) {
	type alias struct {
		BuyerId string
		Age     int
		Name    string
		Date    string `json:"Date,omitempty"`
		Type    string `json:"dgraph.type,omitempty"`
	}

	var a []alias = []alias{}
	for _, e := range *buyers {
		e.Type = c.BuyerType
		e.Date = dataLoader.dateStr
		a = append(a, alias(e))
	}

	return json.Marshal(&a)
}

func (dataLoader *DataLoader) persistBuyers(jsonBuyers []byte) {
	mutation := &api.Mutation{
		SetJson:   jsonBuyers,
		CommitNow: true,
	}

	req := &api.Request{
		Mutations: []*api.Mutation{mutation},
		CommitNow: true,
	}

	_, err := newClient().NewTxn().Do(context.Background(), req)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Buyers loaded.")
}

func (dataLoader *DataLoader) loadTransactions() {
	fmt.Println("Loading transactions...")
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
	dataLoader.persistTransactions([]byte(jsonTransactions))

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

func (dataLoader *DataLoader) persistTransactions(jsonTransactions []byte) {

	mutation := &api.Mutation{
		SetJson:   jsonTransactions,
		CommitNow: true,
	}

	_, err := dataLoader.txn.Mutate(context.Background(), mutation)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transactions loaded.")
}

/*
	Determines if the database has information about the
	requested node based on the date queried by the client.
	In that case, a request to AWS is not necessary.
*/
func (dataLoader *DataLoader) isDateRequestable() bool {
	query := fmt.Sprintf(`{
		q(func: eq(Date, "%s")){
				  uid
			  }
	  }`, dataLoader.dateStr)

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
