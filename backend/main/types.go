package main

type TransactionHolder struct {
	Transactions []Transaction
}

type ProductHolder struct {
	Products []Product
}

type BuyersById struct {
	Buyers []Buyer `json:"buyersById"`
}

type BuyerIdEndpoint struct {
	TransactionHistory  TransactionCollection
	BuyersWithSameIp    BuyerCollection
	RecommendedProducts []Product
}

type BuyerCollection struct {
	Buyers []Buyer
	Count  int
}

type TransactionCollection struct {
	Transactions []Transaction
	Count        int
}

type CollectionCount struct {
	CountArray []Count
}

type Count struct {
	Total int
}

type BuyerRequestParams struct {
	PageBParam     int
	PageSizeBParam int
	PageTParam     int
	PageSizeTParam int
}

type key string
