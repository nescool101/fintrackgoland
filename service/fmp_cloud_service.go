package service

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nescool101/fintrackgoland/models"
)

// FMPCloudService utiliza la librería fmpcloud-go para obtener datos financieros
type FMPCloudService struct {
	APIKey  string
	Results []models.StockData
	Failed  []string
	Mutex   sync.Mutex
	// Nota: Para usar la librería fmpcloud-go, necesitarías instalarla:
	// go get -u github.com/spacecodewor/fmpcloud-go
	// Client  *fmpcloud.APIClient // Comentado hasta instalar la librería
}

// NewFMPCloudService crea una nueva instancia del servicio usando fmpcloud-go
func NewFMPCloudService(apiKey string) *FMPCloudService {
	return &FMPCloudService{
		APIKey: apiKey,
		// Para usar fmpcloud-go, descomenta las siguientes líneas después de instalarlo:
		// Client: fmpcloud.NewAPIClient(fmpcloud.Config{APIKey: apiKey}),
	}
}

// FetchData obtiene datos usando la librería fmpcloud-go
func (fcs *FMPCloudService) FetchData(symbol, date string) {
	// Convertir símbolos al formato apropiado
	// fmpSymbol := convertToFMPSymbol(symbol) // Comentado temporalmente

	// Ejemplo de cómo usarías fmpcloud-go (descomenta después de instalar):
	/*
		fmpSymbol := convertToFMPSymbol(symbol)
		quote, err := fcs.Client.Stock.Quote(fmpSymbol)
		if err != nil {
			log.Printf("Error obteniendo cotización para %s: %v", symbol, err)
			fcs.addFailed(symbol, date)
			return
		}

		stockData := models.StockData{
			Status:     "OK",
			Symbol:     symbol,
			Date:       date,
			Open:       quote.Open,
			High:       quote.DayHigh,
			Low:        quote.DayLow,
			Close:      quote.Price,
			Volume:     float64(quote.Volume),
			AfterHours: quote.AfterMarketPrice,
			PreMarket:  quote.PreMarketPrice,
		}

		fcs.addResult(stockData)
	*/

	// Implementación temporal usando la API directa
	log.Printf("FMPCloudService: Procesando %s para %s (implementación temporal)", symbol, date)

	// Simulación temporal - reemplazar con implementación real
	stockData := models.StockData{
		Status:     "OK",
		Symbol:     symbol,
		Date:       date,
		Open:       100.0,
		High:       105.0,
		Low:        95.0,
		Close:      102.0,
		Volume:     1000000,
		AfterHours: 0,
		PreMarket:  0,
	}

	fcs.addResult(stockData)

	// Pausa de 1 segundo para evitar saturar el servidor
	time.Sleep(1 * time.Second)
}

// FetchWeeklyData obtiene datos para múltiples símbolos y fechas
func (fcs *FMPCloudService) FetchWeeklyData(symbols []string, dates []string) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Límite para API gratuita

	for _, symbol := range symbols {
		for _, date := range dates {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(sym, dt string) {
				defer wg.Done()
				fcs.FetchData(sym, dt)
				// Pausa adicional de 1 segundo para evitar saturar el servidor
				time.Sleep(1 * time.Second)
				<-semaphore
			}(symbol, date)
		}
	}

	wg.Wait()
}

// addResult añade un resultado exitoso de forma thread-safe
func (fcs *FMPCloudService) addResult(data models.StockData) {
	fcs.Mutex.Lock()
	defer fcs.Mutex.Unlock()
	fcs.Results = append(fcs.Results, data)
}

// addFailed añade un símbolo fallido de forma thread-safe
func (fcs *FMPCloudService) addFailed(symbol, date string) {
	fcs.Mutex.Lock()
	defer fcs.Mutex.Unlock()
	fcs.Failed = append(fcs.Failed, fmt.Sprintf("%s (%s)", symbol, date))
}
