package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	Type DetailsResponse `json:"USDBRL"`
}

type DetailsResponse struct {
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
	ID         int    `gorm:"primaryKey"`
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

const (
	requestTimeout = 200 * time.Millisecond
	dbTimeout      = 10 * time.Millisecond
)

func main() {
	println("Server started!")
	http.HandleFunc("/cotacao", getDollarQuote)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func getDollarQuote(w http.ResponseWriter, r *http.Request) {
	println("Getting a new dollar quote")
	ctx := r.Context()
	dollarQuoteResponse := makeRequestDollarQuote(ctx)
	details := dollarQuoteResponse.Type
	dollarQuote := &DollarQuote{
		Code:       details.Code,
		Codein:     details.Codein,
		Name:       details.Name,
		High:       details.High,
		Low:        details.Low,
		VarBid:     details.VarBid,
		PctChange:  details.PctChange,
		Bid:        details.Bid,
		Ask:        details.Ask,
		Timestamp:  details.Timestamp,
		CreateDate: details.CreateDate,
	}
	saveQuote(ctx, dollarQuote)

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
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

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

func saveQuote(ctx context.Context, quote *DollarQuote) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	db := initDatabase()
	db.WithContext(ctx).Save(&quote)

	var dollarQuoteSaved DollarQuote
	db.Find(&dollarQuoteSaved)
	fmt.Printf("DollarQuote saved: %v\n", dollarQuoteSaved)
}

func initDatabase() (db *gorm.DB) {
	dsn := "server-api.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(DollarQuote{})
	if err != nil {
		panic(err)
	}

	return db
}
