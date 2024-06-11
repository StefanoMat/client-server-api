package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
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

var db *sql.DB

func main() {
	// Connect to SQLite database
	// If the database does not exist, it will be created
	var err error
	db, err = sql.Open("sqlite3", "./database.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()
	// Create a table
	sqlStmt := `
 CREATE TABLE IF NOT EXISTS quote_history (
	 id INTEGER PRIMARY KEY AUTOINCREMENT,
	 price REAL NOT NULL,
	 timestamp DATETIME NOT NULL
 );
 `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		fmt.Printf("Error creating table: %s\n", err)
		return
	}

	fmt.Println("Database and table created successfully.")

	mux := http.NewServeMux()
	mux.HandleFunc("/dolar", DolarPrice)
	http.ListenAndServe(":8080", mux)
}

func DolarPrice(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	//pegar o json e dar unmarshel
	var data map[string]USDBRL
	if err := json.Unmarshal([]byte(string(body)), &data); err != nil {
		fmt.Println("Error parsing JSON:", err)
		panic(err)
	}
	quote := NewUSDBRL(data["USDBRL"].Price, data["USDBRL"].Timestamp)
	err = insertQuote(db, quote)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")

	w.Write([]byte(body))
}

func insertQuote(db *sql.DB, quote *USDBRL) error {
	stmt, err := db.Prepare("INSERT INTO quote_history (price, timestamp) VALUES (?, ?);")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(quote.Price, quote.Timestamp)
	if err != nil {
		panic(err)
	}
	return nil
}
