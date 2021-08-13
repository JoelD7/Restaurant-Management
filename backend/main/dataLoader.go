package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	c "module/constants"
	f "module/utils"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	d "github.com/shopspring/decimal"
)

type Buyer struct {
	BuyerId string
	Age     int
	Name    string
	Date    string `json:"Date,omitempty"`
	Type    string `json:"dgraph.type,omitempty"`
}

/*
	Struct used to match the fields returned
	from AWS.
*/
type BuyerUnmarshall struct {
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

func (dataLoader *DataLoader) loadRestaurantData() (string, error) {
	ok, dateErr := dataLoader.isDateRequestable()
	if dateErr != nil {
		fmt.Println(dateErr)
		return "", dateErr
	}

	if !ok {
		fmt.Printf("The restaurant data for date %s has already loaded.\n", dataLoader.dateStr)
		return fmt.Sprintf("The restaurant data for date %s has already loaded.\n", dataLoader.dateStr), nil
	}

	functions := make([]func() error, 0)
	functions = append(functions, dataLoader.loadBuyers)
	functions = append(functions, dataLoader.loadTransactions)
	functions = append(functions, dataLoader.loadProducts)

	waitGroup := sync.WaitGroup{}

	for i := range functions {
		waitGroup.Add(1)

		go func(function func() error) {
			function()

			waitGroup.Done()
		}(functions[i])
	}

	waitGroup.Wait()

	return "All data succesfully loaded", nil
}

func (dataLoader *DataLoader) loadProducts() error {
	fmt.Println("Loading products...")

	req, reqErr := http.NewRequest("GET", c.ProductURL, nil)
	if reqErr != nil {
		fmt.Println(reqErr)
		return reqErr
	}

	q := req.URL.Query()
	var dateAsTimestamp string = fmt.Sprint(f.DateStringToTimestamp(dataLoader.dateStr))
	q.Add("date", dateAsTimestamp)

	req.URL.RawQuery = q.Encode()
	requestUrl := req.URL.String()

	resp, respErr := http.Get(requestUrl)
	if respErr != nil {
		fmt.Printf("Error in response for GET request: '%s' | %v\n", requestUrl, respErr)
		return respErr
	}

	defer resp.Body.Close()
	body, resBodyError := io.ReadAll(resp.Body)
	if resBodyError != nil {
		fmt.Printf("Error while reading response body for request '%s' | %v\n", requestUrl, resBodyError)
		return resBodyError
	}

	rawProductsLines := strings.Split(string(body), "\n")
	products, parseErr := dataLoader.parseProducts(rawProductsLines)
	if parseErr != nil {
		return fmt.Errorf("error while parsing products | %w", parseErr)
	}

	jsonProducts, jsonErr := json.Marshal(products)

	if jsonErr != nil {
		fmt.Printf("Error while marshalling products for database upload | %v\n", jsonErr)
		return jsonErr
	}

	persistErr := dataLoader.persistProducts(jsonProducts)
	if persistErr != nil {
		fmt.Println(persistErr)
		return persistErr
	}

	return nil
}

func (dataLoader *DataLoader) parseProducts(rawProductsLines []string) ([]Product, error) {
	var products []Product

	for _, line := range rawProductsLines {
		lineSections := strings.Split(line, "'")

		if len(lineSections) < 3 {
			continue
		}

		var id, name string
		var price d.Decimal
		var priceErr error

		/*
			In this case the splitting have to rules change because if
			they don't, a string such as:
			<6621fd74'"Campbell's soup chicken & sausage"'3625> would
			have <s soup chicken & sausage> as its third element after the
			split, leading to an error, because the code expects the third
			element to be the price of the product.
		*/
		if strings.Contains(line, `"`) {
			id = lineSections[0]
			firstQuotePos := strings.Index(line, `"`)
			lastQuotePos := strings.LastIndex(line, `"`)
			name = line[firstQuotePos+1 : lastQuotePos]
			price, priceErr = d.NewFromString(lineSections[len(lineSections)-1])
		} else {
			id = lineSections[0]
			name = lineSections[1]
			price, priceErr = d.NewFromString(lineSections[2])
		}

		if priceErr != nil {
			fmt.Printf("parseProducts: Error while casting products prices from string to decimal.Decimal | %v\n", priceErr)
			return nil, priceErr
		}

		newProduct := Product{
			ProductId: id,
			Name:      name,
			Price:     price,
			Date:      dataLoader.dateStr,
			Type:      c.ProductType,
		}

		products = append(products, newProduct)
	}

	return products, nil
}

func (dataLoader *DataLoader) persistProducts(jsonProducts []byte) error {
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
		fmt.Printf("Error while persisting new products | %v\n", err)
		return err
	}

	fmt.Println("Products loaded.")
	return nil
}

func (dataLoader *DataLoader) loadBuyers() error {
	fmt.Println("Loading buyers...")

	req, reqErr := http.NewRequest("GET", c.BuyersURL, nil)
	if reqErr != nil {
		fmt.Printf("Error in GET request '%s' | %v\n", c.BuyersURL, reqErr)
		return reqErr
	}

	q := req.URL.Query()
	var dateAsTimestamp string = fmt.Sprint(f.DateStringToTimestamp(dataLoader.dateStr))
	q.Add("date", dateAsTimestamp)

	req.URL.RawQuery = q.Encode()
	requestUrl := req.URL.String()

	resp, resErr := http.Get(requestUrl)
	if resErr != nil {
		fmt.Printf("Error in response for GET request '%s' | %v\n", requestUrl, resErr)
		return resErr
	}

	defer resp.Body.Close()
	body, bodyReadErr := io.ReadAll(resp.Body)
	if bodyReadErr != nil {
		fmt.Printf("Error reading body of response for GET request '%s' | %v\n", requestUrl, bodyReadErr)
		return bodyReadErr
	}

	var buyers []BuyerUnmarshall
	uErr := json.Unmarshal(body, &buyers)
	if uErr != nil {
		fmt.Printf("Error while unmarshalling buyers object obtained from response for GET request '%s'\n| %v", requestUrl, uErr)
		return uErr
	}

	jsonBuyers, mErr := dataLoader.marshalJSON(&buyers)
	if mErr != nil {
		fmt.Printf("Error while marshalling buyers object for database persistence |%v\n", mErr)
		return mErr
	}

	persistErr := dataLoader.persistBuyers(jsonBuyers)
	if persistErr != nil {
		fmt.Println(persistErr)
	}

	return nil
}

/*
	Convert BuyerUnmarshall to Buyer
*/
func (dataLoader *DataLoader) marshalJSON(buyers *[]BuyerUnmarshall) ([]byte, error) {
	var a []Buyer = []Buyer{}
	for _, e := range *buyers {
		e.Type = c.BuyerType
		e.Date = dataLoader.dateStr
		a = append(a, Buyer(e))
	}

	return json.Marshal(&a)
}

func (dataLoader *DataLoader) persistBuyers(jsonBuyers []byte) error {
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
		fmt.Printf("Error while persisting buyers to database: %v", err)
		return err
	}

	fmt.Println("Buyers loaded.")
	return nil
}

func (dataLoader *DataLoader) loadTransactions() error {
	fmt.Println("Loading transactions...")

	req, reqErr := http.NewRequest("GET", c.TransactionsURL, nil)

	if reqErr != nil {
		fmt.Printf("Error in GET request '%s' | %v \n", c.TransactionsURL, reqErr)
		return reqErr
	}

	q := req.URL.Query()
	var dateAsTimestamp string = fmt.Sprint(f.DateStringToTimestamp(dataLoader.dateStr))
	q.Add("date", dateAsTimestamp)

	req.URL.RawQuery = q.Encode()
	query := req.URL.String()

	resp, resErr := http.Get(query)
	if resErr != nil {
		fmt.Printf("Error in response for GET request '%s' | %v\n", query, resErr)
		return resErr
	}

	defer resp.Body.Close()
	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		fmt.Printf("Error while reading response body of GET request '%s' | %v\n", query, bodyErr)
		return bodyErr
	}

	rawTransactions := strings.Split(string(body), "#")
	var transactions []Transaction = dataLoader.parseTransactions(rawTransactions)
	rawJsonTransactions, mErr := json.Marshal(transactions)

	if mErr != nil {
		fmt.Printf("Error while marshalling transactions for database persistence | %v\n", mErr)
		return mErr
	}

	/*
		Replace all appearances of the unicode null character: \u0000 with an
		empty string.
	*/
	jsonTransactions := strings.Replace(string(rawJsonTransactions), "\\u0000", "", -1)
	persistErr := dataLoader.persistTransactions([]byte(jsonTransactions))

	if persistErr != nil {
		fmt.Println(persistErr)
	}

	return nil

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

func (dataLoader *DataLoader) persistTransactions(jsonTransactions []byte) error {
	mutation := &api.Mutation{
		SetJson:   jsonTransactions,
		CommitNow: true,
	}

	_, err := dataLoader.txn.Mutate(context.Background(), mutation)

	if err != nil {
		fmt.Printf("Error while persisting transactions: %v\n", err)
		return err
	}

	fmt.Println("Transactions loaded.")
	return nil
}

/*
	Determines if the database has information about the
	requested node based on the date queried by the client.
	In that case, a request to AWS is not necessary.
*/
func (dataLoader *DataLoader) isDateRequestable() (bool, error) {
	//Parse the date to the format the database uses for dates: RFC3339
	t, parseErr := time.Parse(c.DateLayout, dataLoader.dateStr)
	if parseErr != nil {
		fmt.Printf("Error while parsing string '%s' to date | %v\n", dataLoader.dateStr, parseErr)
		return false, parseErr
	}

	date := t.Format(c.DateLayoutRFC3339)

	query := fmt.Sprintf(`{
		q(func: eq(Date, "%s")){
				  uid
			  }
	  }`, date)

	res, resErr := dataLoader.txn.Query(ctx, query)
	if resErr != nil {
		fmt.Printf("Error while making query: '%s' to database | %v\n", query, resErr)
		return false, resErr
	}

	resultSize := res.Metrics.NumUids["uid"]

	if resultSize > 0 {
		return false, nil
	} else {
		return true, nil
	}
}
