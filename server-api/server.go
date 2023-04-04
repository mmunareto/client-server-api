package main

import (
	"context"
	"encoding/json"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"net/http"
	"time"
)

var (
	dollarQuoteURL = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
)

type DollarQuoteResponse struct {
	Type Details `json:"USDBRL"`
}

type Details struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type DollarQuote struct {
	Name string `gorm:"primaryKey"`
	Bid  string `json:"bid"`
}

const (
	requestTimeout = 200 * time.Millisecond
	dbTimeout      = 10 * time.Millisecond
)

func main() {
	println("Server started!")
	http.HandleFunc("/cotacao", getDollarQuote)
	http.ListenAndServe(":8080", nil)
}

func getDollarQuote(w http.ResponseWriter, r *http.Request) {
	println("Getting a new dollar quote")
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	dollarQuoteResponse := makeRequestDollarQuote(ctx)
	dollarQuote := &DollarQuote{
		Bid: dollarQuoteResponse.Type.Bid,
	}
	saveQuote(dollarQuote)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(dollarQuote)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func makeRequestDollarQuote(ctx context.Context) DollarQuoteResponse {
	println("Fetching dollar quote from external API")
	req, err := http.NewRequestWithContext(ctx, "GET", dollarQuoteURL, nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	var dollarQuoteResponse DollarQuoteResponse
	err = json.Unmarshal(body, &dollarQuoteResponse)
	if err != nil {
		panic(err)
	}

	println("Dollar Quote found!")

	return dollarQuoteResponse
}

func saveQuote(quote *DollarQuote) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	db := initDatabase()
	db.WithContext(ctx).Save(&quote)
}

func initDatabase() (db *gorm.DB) {
	dsn := "server-api.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(DollarQuote{})

	return db
}
