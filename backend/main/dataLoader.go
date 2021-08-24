package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	c "module/constants"
	f "module/utils"
	"net/http"
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
	Type    string `json:"dgraph.type,omitempty"`
}

type BuyerHolder struct {
	Buyers []Buyer
}

type Product struct {
	ProductId string
	Name      string
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

type LoadResponse struct {
	Buyers       []Buyer
	Products     []Product
	Transactions []Transaction
}

func (dataLoader *DataLoader) loadRestaurantData() ([]byte, error) {

	waitGroup := sync.WaitGroup{}
	errorChan := make(chan error)
	wgDone := make(chan bool)
	productsChan := make(chan []Product, 1)
	transactionsChan := make(chan []Transaction, 1)
	buyersChan := make(chan []Buyer, 1)

	waitGroup.Add(3)
	go dataLoader.loadProducts(errorChan, productsChan, &waitGroup)
	go dataLoader.loadBuyers(errorChan, buyersChan, &waitGroup)
	go dataLoader.loadTransactions(errorChan, transactionsChan, &waitGroup)

	go func() {
		waitGroup.Wait()
		close(wgDone)
	}()

	select {
	case err := <-errorChan:
		return nil, err

	case <-wgDone:
		dataLoader.txn.Commit(context.Background())

		dataLoaded := &LoadResponse{
			Buyers:       <-buyersChan,
			Products:     <-productsChan,
			Transactions: <-transactionsChan,
		}

		jsonData, err := json.Marshal(dataLoaded)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal loaded restaurant data: %w", err)
		}

		return jsonData, nil
	}
}

func (dataLoader *DataLoader) loadProducts(errChan chan<- error, productsChan chan<- []Product, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	fmt.Println("Loading products...")

	rawProductsLines, err := dataLoader.fetchProductsFromAWS()
	if err != nil {
		errChan <- err
		return
	}

	products, err := dataLoader.parseProducts(rawProductsLines)
	if err != nil {
		errChan <- fmt.Errorf("error while parsing products | %w", err)
		return
	}

	jsonProducts, err := json.Marshal(products)
	if err != nil {
		fmt.Printf("Error while marshalling products for database upload | %v\n", err)
		errChan <- err
		return
	}

	err = dataLoader.persistProducts(jsonProducts)
	if err != nil {
		errChan <- err
		return
	}

	productsChan <- products
	close(productsChan)
}

func (dataLoader *DataLoader) fetchProductsFromAWS() ([]string, error) {
	req, err := http.NewRequest("GET", c.ProductURL, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	q := req.URL.Query()
	timestamp, err := f.DateStringToTimestamp(dataLoader.dateStr)
	if err != nil {
		return nil, err
	}

	var dateAsTimestamp string = fmt.Sprint(timestamp)
	q.Add("date", dateAsTimestamp)

	req.URL.RawQuery = q.Encode()
	requestUrl := req.URL.String()

	resp, err := http.Get(requestUrl)
	if err != nil {
		fmt.Printf("Error in response for GET request: '%s' | %v\n", requestUrl, err)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error while reading response body for request '%s' | %v\n", requestUrl, err)
		return nil, err
	}

	defer func() {
		err = resp.Body.Close()
	}()
	if err != nil {
		return nil, err
	}

	rawProductsLines := strings.Split(string(body), "\n")
	return rawProductsLines, nil
}

func (dataLoader *DataLoader) parseProducts(rawProductsLines []string) ([]Product, error) {
	addedProductIds, err := dataLoader.getPersistedProductsIds()
	if err != nil {
		return nil, err
	}

	var products []Product

	for _, line := range rawProductsLines {
		// c89db54f'Campbell's minestrone italian style slow simmered soup'8841
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
			name = strings.ReplaceAll(line[firstQuotePos+1:lastQuotePos], "&quot;", "'")
			price, err = d.NewFromString(lineSections[len(lineSections)-1])
		} else {
			id = lineSections[0]
			name = strings.ReplaceAll(lineSections[1], "&quot;", "'")
			price, err = d.NewFromString(lineSections[2])
		}

		if err != nil {
			fmt.Printf("parseProducts: Error while casting products prices from string to decimal.Decimal | %v\n", priceErr)
			return nil, err
		}

		newProduct := Product{
			ProductId: id,
			Name:      name,
			Price:     price,
			Type:      c.ProductType,
		}

		if !f.ArrayContains(addedProductIds, id) && name != "" && name != "null" {
			products = append(products, newProduct)
			addedProductIds = append(addedProductIds, id)
		}
	}

	return products, nil
}

func (dataLoader *DataLoader) getPersistedProductsIds() ([]string, error) {
	var addedProductIds []string

	query := `{
		products(func: type(Product)){
			  expand(_all_){}
		}
	  }`

	res, err := dataLoader.txn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving products from database | %w", err)
	}

	var productHolder ProductHolder
	err = json.Unmarshal(res.Json, &productHolder)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshalling products retrieved from database | %w", err)
	}

	for _, product := range productHolder.Products {
		addedProductIds = append(addedProductIds, product.ProductId)
	}

	return addedProductIds, nil
}

func (dataLoader *DataLoader) persistProducts(jsonProducts []byte) error {
	mutation := &api.Mutation{
		SetJson: jsonProducts,
	}

	req := &api.Request{
		Mutations: []*api.Mutation{mutation},
	}

	_, err := dataLoader.txn.Do(context.Background(), req)

	if err != nil {
		fmt.Printf("Error while persisting new products | %v\n", err)
		return err
	}

	fmt.Println("Products loaded.")
	return nil
}

func (dataLoader *DataLoader) loadBuyers(errChan chan<- error, buyersChan chan<- []Buyer, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	fmt.Println("Loading buyers...")

	unfilteredBuyers, err := dataLoader.fetchBuyersFromAWS()
	if err != nil {
		errChan <- err
		return
	}

	var buyers []BuyerUnmarshall

	addedBuyerIds, err := dataLoader.getPersistedBuyersIds()
	if err != nil {
		errChan <- err
		return
	}

	for _, b := range unfilteredBuyers {
		if !f.ArrayContains(addedBuyerIds, b.BuyerId) {
			buyers = append(buyers, b)
			addedBuyerIds = append(addedBuyerIds, b.BuyerId)
		}
	}

	jsonBuyers, err := dataLoader.marshalBuyers(&buyers)
	if err != nil {
		fmt.Printf("Error while marshalling buyers object for database persistence |%v\n", err)
		errChan <- err
		return
	}

	err = dataLoader.persistBuyers(jsonBuyers)
	if err != nil {
		errChan <- fmt.Errorf("error while persisting buyers | %w", err)
		return
	}

	var buyersRes []Buyer
	err = json.Unmarshal(jsonBuyers, &buyersRes)
	if err != nil {
		errChan <- err
		return
	}

	buyersChan <- buyersRes
	close(buyersChan)

}

func (dataLoader *DataLoader) fetchBuyersFromAWS() ([]BuyerUnmarshall, error) {
	//Form request URL
	req, err := http.NewRequest("GET", c.BuyersURL, nil)
	if err != nil {
		fmt.Printf("Error while forming GET request '%s' | %v\n", c.BuyersURL, err)
		return nil, err
	}

	q := req.URL.Query()
	timestamp, err := f.DateStringToTimestamp(dataLoader.dateStr)
	if err != nil {
		return nil, err
	}

	var dateAsTimestamp string = fmt.Sprint(timestamp)
	q.Add("date", dateAsTimestamp)

	req.URL.RawQuery = q.Encode()
	requestUrl := req.URL.String()

	// Make GET request
	resp, err := http.Get(requestUrl)
	if err != nil {
		fmt.Printf("Error in response for GET request '%s' | %v\n", requestUrl, err)
		return nil, err
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body of response for GET request '%s' | %v\n", requestUrl, err)
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
	}()
	if err != nil {
		return nil, err
	}

	var unfilteredBuyers []BuyerUnmarshall
	err = json.Unmarshal(body, &unfilteredBuyers)
	if err != nil {
		fmt.Printf("Error while unmarshalling buyers object obtained from response for GET request '%s'\n| %v", requestUrl, err)
		return nil, err
	}

	return unfilteredBuyers, nil
}

func (dataLoader *DataLoader) getPersistedBuyersIds() ([]string, error) {
	var addedIds []string
	query := `{
		buyers(func: type(Buyer)){
			  expand(_all_){}
		}
	  }`

	res, err := dataLoader.txn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error while fetching buyers from database %w", err)
	}

	var buyerHolder BuyerHolder
	err = json.Unmarshal(res.Json, &buyerHolder)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshalling buyers retrieved from database | %w", err)
	}

	for _, buyer := range buyerHolder.Buyers {
		addedIds = append(addedIds, buyer.BuyerId)
	}

	return addedIds, nil
}

/*
	Convert BuyerUnmarshall to Buyer
*/
func (dataLoader *DataLoader) marshalBuyers(buyers *[]BuyerUnmarshall) ([]byte, error) {
	var a []Buyer = []Buyer{}
	for _, e := range *buyers {
		e.Type = c.BuyerType
		a = append(a, Buyer(e))
	}

	return json.Marshal(&a)
}

func (dataLoader *DataLoader) persistBuyers(jsonBuyers []byte) error {
	mutation := &api.Mutation{
		SetJson: jsonBuyers,
	}

	req := &api.Request{
		Mutations: []*api.Mutation{mutation},
	}

	_, err := dataLoader.txn.Do(context.Background(), req)

	if err != nil {
		fmt.Printf("Error while persisting buyers to database: %v", err)
		return err
	}

	fmt.Println("Buyers loaded.")
	return nil
}

func (dataLoader *DataLoader) loadTransactions(errChan chan<- error, transactionsChan chan<- []Transaction, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	fmt.Println("Loading transactions...")

	rawTransactions, err := dataLoader.fetchTransactionsFromAWS()
	if err != nil {
		errChan <- err
		return
	}

	var transactions []Transaction = dataLoader.parseTransactions(rawTransactions)
	jsonTransactions, err := json.Marshal(transactions)

	if err != nil {
		fmt.Printf("Error while marshalling transactions for database persistence | %v\n", err)
		errChan <- err
		return
	}

	err = dataLoader.persistTransactions(jsonTransactions)

	if err != nil {
		errChan <- fmt.Errorf("failed to persist transactions | %w", err)
		return
	}

	transactionsChan <- transactions
	close(transactionsChan)
}

func (dataLoader *DataLoader) fetchTransactionsFromAWS() ([]string, error) {
	req, err := http.NewRequest("GET", c.TransactionsURL, nil)

	if err != nil {
		fmt.Printf("Error in GET request '%s' | %v \n", c.TransactionsURL, err)
		return nil, err
	}

	q := req.URL.Query()
	timestamp, err := f.DateStringToTimestamp(dataLoader.dateStr)
	if err != nil {
		return nil, err
	}

	var dateAsTimestamp string = fmt.Sprint(timestamp)
	q.Add("date", dateAsTimestamp)

	req.URL.RawQuery = q.Encode()
	query := req.URL.String()

	resp, err := http.Get(query)
	if err != nil {
		fmt.Printf("Error in response for GET request '%s' | %v\n", query, err)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error while reading response body of GET request '%s' | %v\n", query, err)
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
	}()
	if err != nil {
		return nil, err
	}

	//Replace null characters with '||'
	bodyWithBars := strings.ReplaceAll(string(body), "\x00", "|")
	rawTransactions := strings.Split(bodyWithBars, "||")
	return rawTransactions, nil
}

func (dataLoader *DataLoader) parseTransactions(rawTransactions []string) []Transaction {
	var transactions []Transaction
	transactionsQty := len(rawTransactions)

	for i := 0; i < transactionsQty; i++ {
		transactionString := rawTransactions[i]
		size := len(transactionString)
		transactionStringArr := strings.Split(transactionString, "|")

		if size == 0 {
			continue
		}

		transactionId := string(transactionStringArr[0])[1:]
		buyerId := transactionStringArr[1]
		ip := transactionStringArr[2]
		device := transactionStringArr[3]
		products := string(transactionStringArr[4])[1 : len(transactionStringArr[4])-1]

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
	defer f.TimeTrack(time.Now(), "persistTransactions")
	mutation := &api.Mutation{
		SetJson: jsonTransactions,
	}

	_, err := dataLoader.txn.Mutate(context.Background(), mutation)

	if err != nil {
		fmt.Printf("Error while persisting transactions: %v\n", err)
		return err
	}

	return nil
}

/*
	Determines if the database has information about the
	requested node based on the date queried by the client.
	In that case, a request to AWS is not necessary.
*/
func (dataLoader *DataLoader) isDateRequestable() (bool, error) {
	//Parse the date to the format the database uses for dates: RFC3339
	t, err := time.Parse(c.DateLayout, dataLoader.dateStr)
	if err != nil {
		fmt.Printf("Error while parsing string '%s' to date | %v\n", dataLoader.dateStr, err)
		return false, err
	}

	date := t.Format(c.DateLayoutRFC3339)

	query := fmt.Sprintf(`{
		q(func: eq(Date, "%s")){
				  uid
			  }
	  }`, date)

	res, err := dataLoader.txn.Query(ctx, query)
	if err != nil {
		fmt.Printf("Error while making query: '%s' to database | %v\n", query, err)
		return false, err
	}

	resultSize := res.Metrics.NumUids["uid"]

	if resultSize > 0 {
		return false, nil
	} else {
		return true, nil
	}
}
