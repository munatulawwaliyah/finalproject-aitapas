package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type AIModelConnector struct {
	Client *http.Client
}

type Inputs struct {
	Table map[string][]string `json:"table"`
	Query string              `json:"query"`
}

type Response struct {
	Answer      string   `json:"answer"`
	Coordinates [][]int  `json:"coordinates"`
	Cells       []string `json:"cells"`
	Aggregator  string   `json:"aggregator"`
}

func CsvToSlice(data string) (map[string][]string, error) {
	r := csv.NewReader(strings.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 1 {
		return nil, fmt.Errorf("no data found in CSV")
	}

	table := make(map[string][]string)
	headers := records[0]

	for _, header := range headers {
		table[header] = []string{}
	}

	for _, record := range records[1:] {
		for i, value := range record {
			table[headers[i]] = append(table[headers[i]], value)
		}
	}
	return table, nil// TODO: replace this
}

func (c *AIModelConnector) ConnectAIModel(payload interface{}, token string) (Response, error) {
	url := "https://api-inference.huggingface.co/models/google/tapas-base-finetuned-wtq"
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Response{}, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return Response{}, err
	}

	return response, nil // TODO: replace this
}

func GetRecommendation(query string, answer string) string {
	recommendations := map[string]string{
		"energy consumption": "Turn off the TV when not in use to save energy.",
		"electricity cost":   "Use LED bulbs to reduce electricity costs.",
	}
	for key, recommendation := range recommendations {
		if strings.Contains(strings.ToLower(query), key) {
			return recommendation
		}
	}
	return "Consider general energy-saving tips to reduce your consumption."
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	token := os.Getenv("HUGGINGFACE_TOKEN")
	if token == "" {
		log.Fatalf("HUGGINGFACE_TOKEN is required")
	}

	filePath := "data-series.csv"
	csvFile, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening CSV file: %v", err)
	}
	defer csvFile.Close()

	csvData, err := io.ReadAll(csvFile)
	if err != nil {
		log.Fatalf("Error reading csv file: %v", err)
	}
	
	table, err := CsvToSlice(string(csvData))
	if err != nil {
		log.Fatalf("Error converting CSV to slice: %v", err)
	}

	fmt.Print("Enter your query: ")
	var query string
	fmt.Scanln(&query)

	client := &http.Client{}
	connector := AIModelConnector{Client: client}
	
	input := Inputs{
		Table: table,
		Query: query,
	}

	response, err := connector.ConnectAIModel(input, token)
	if err != nil {
		log.Fatalf("Error connecting to AI model: %v", err)
	}

	fmt.Printf("Answer: %s\n", response.Answer)
	fmt.Printf("Recommendation: %s\n", GetRecommendation(query, response.Answer))
}