package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type USDBRL struct {
	Price     string `json:"bid"`
	Timestamp string `json:"timestamp"`
}

func NewUSDBRL(price string, timestamp string) *USDBRL {
	return &USDBRL{
		Price:     price,
		Timestamp: timestamp,
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Request timed out, exceeded 300ms")
		}
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data map[string]USDBRL
	if err := json.Unmarshal([]byte(string(body)), &data); err != nil {
		fmt.Println("Error parsing JSON:", err)
		panic(err)
	}
	quote := NewUSDBRL(data["USDBRL"].Price, data["USDBRL"].Timestamp)
	writeFile(quote)
}

func writeFile(quote *USDBRL) error {
	f, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte("Dólar: " + quote.Price + "\n"))
	if err != nil {
		return err
	}
	return nil
}
