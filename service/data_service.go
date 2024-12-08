package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/nescool101/fintrackgoland/models"
)

type DataService struct {
	APIKey  string
	Results []models.StockData
	Failed  []string
	Mutex   sync.Mutex
	Client  *http.Client
}

func NewDataService(apiKey string) *DataService {
	return &DataService{
		APIKey: apiKey,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (ds *DataService) FetchData(symbol, date string) {
	url := fmt.Sprintf("https://api.polygon.io/v1/open-close/%s/%s?adjusted=false&apiKey=%s", symbol, date, ds.APIKey)
	resp, err := ds.Client.Get(url)
	if err != nil {
		log.Printf("Error fetching data for %s on %s: %v", symbol, date, err)
		ds.addFailed(symbol, date)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-OK HTTP status: %d for %s on %s", resp.StatusCode, symbol, date)
		ds.addFailed(symbol, date)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body for %s on %s: %v", symbol, date, err)
		ds.addFailed(symbol, date)
		return
	}

	var data models.StockData
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Error unmarshaling JSON for %s on %s: %v", symbol, date, err)
		ds.addFailed(symbol, date)
		return
	}

	if data.Status != "OK" {
		log.Printf("API returned status '%s' for %s on %s", data.Status, symbol, date)
		ds.addFailed(symbol, date)
		return
	}

	data.Date = date // Ensure date is set
	ds.addResult(data)
}

func (ds *DataService) addResult(data models.StockData) {
	ds.Mutex.Lock()
	defer ds.Mutex.Unlock()
	ds.Results = append(ds.Results, data)
}

func (ds *DataService) addFailed(symbol, date string) {
	ds.Mutex.Lock()
	defer ds.Mutex.Unlock()
	ds.Failed = append(ds.Failed, fmt.Sprintf("%s (%s)", symbol, date))
}

func (ds *DataService) FetchWeeklyData(symbols []string, dates []string) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent requests to 5

	for _, symbol := range symbols {
		for _, date := range dates {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(sym, dt string) {
				defer wg.Done()
				ds.FetchData(sym, dt)
				<-semaphore
			}(symbol, date)
		}
	}

	wg.Wait()
}
