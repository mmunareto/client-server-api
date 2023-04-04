package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	cotacaoURL = "http://localhost:8080/cotacao"
)

type DollarQuote struct {
	Name string `json:"name"`
	Bid  string `json:"bid"`
}

const requestTimeout = 300 * time.Millisecond

func main() {
	dollarQuote := getDollarQuote()

	createFile("cotacao.txt", dollarQuote.Bid)
}

func getDollarQuote() DollarQuote {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", cotacaoURL, nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var dollarQuote DollarQuote
	err = json.Unmarshal(body, &dollarQuote)
	if err != nil {
		panic(err)
	}

	return dollarQuote
}

func createFile(fileName string, content string) {
	f, err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		panic(err)
	}

	size, err := f.WriteString("Dólar:" + content)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Arquivo criado com sucesso! Tamanho %v bytes\n", size)
}
