package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func main() {
	req, err := http.NewRequest("GET", "https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/transactions", nil)

	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	q.Add("date", "1624680000")
	req.URL.RawQuery = q.Encode()
	query := req.URL.String()

	resp, err := http.Get(query)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	transactions := strings.Split(string(body), "#")

	for i := 0; i < len(transactions); i++ {
		line := transactions[i]
		size := len(line)

		if size == 0 {
			continue
		}

		transactionId := line[0:12]
		buyerId := line[12:21]
		deviceRgx := regexp.MustCompile(`[a-z]`)
		productRgx := regexp.MustCompile(`\(`)

		deviceIndex := deviceRgx.FindStringIndex(line[21 : size-1])[0]
		fmt.Println(deviceIndex)
		if deviceIndex == 0 {
			continue
		}

		productIndex := productRgx.FindStringIndex(line[21 : size-1])[0]

		ip := line[21 : deviceIndex+21]
		device := line[deviceIndex+21 : productIndex+21]
		products := line[productIndex+21:]

		fmt.Printf("transactionId: %s, buyerId: %s, ip: %s, device: %s, products: %s\n", transactionId, buyerId, ip, device, products)
	}

}
