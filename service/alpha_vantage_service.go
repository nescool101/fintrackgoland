package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/nescool101/fintrackgoland/models"
)

/* AlphaVantageService maneja las llamadas a la API de Alpha Vantage
 * Soporta hasta 500 llamadas gratis por día
 * Mapea índices a ETFs correspondientes para obtener datos equivalentes
 */
type AlphaVantageService struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	mu         sync.RWMutex
	results    []models.StockData
	failed     []string
}

/* AlphaVantageResponse estructura para respuesta de Alpha Vantage API */
type AlphaVantageResponse struct {
	MetaData struct {
		Information   string `json:"1. Information"`
		Symbol        string `json:"2. Symbol"`
		LastRefreshed string `json:"3. Last Refreshed"`
		OutputSize    string `json:"4. Output Size"`
		TimeZone      string `json:"5. Time Zone"`
	} `json:"Meta Data"`
	TimeSeriesDaily map[string]AlphaVantageData `json:"Time Series (Daily)"`
}

/* AlphaVantageData estructura para datos diarios de Alpha Vantage */
type AlphaVantageData struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
}

/* NewAlphaVantageService crea una nueva instancia del servicio Alpha Vantage */
func NewAlphaVantageService(apiKey string) *AlphaVantageService {
	return &AlphaVantageService{
		APIKey:  apiKey,
		BaseURL: "https://www.alphavantage.co/query",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		results: make([]models.StockData, 0),
		failed:  make([]string, 0),
	}
}

/* mapIndexToETF mapea símbolos de índices a ETFs correspondientes */
func (avs *AlphaVantageService) mapIndexToETF(symbol string) string {
	indexToETF := map[string]string{
		"SPX":  "SPY", // S&P 500 Index → SPDR S&P 500 ETF
		"NDX":  "QQQ", // Nasdaq 100 Index → Invesco QQQ Trust ETF
		"DJI":  "DIA", // Dow Jones Industrial Average → SPDR Dow Jones Industrial Average ETF
		"NYA":  "VTI", // NYSE Composite Index → Vanguard Total Stock Market ETF
		"ES_F": "SPY", // E-mini S&P 500 Futures → SPY ETF (proxy)
		"NQ_F": "QQQ", // E-mini Nasdaq 100 Futures → QQQ ETF (proxy)
	}

	if etf, exists := indexToETF[symbol]; exists {
		return etf
	}
	return symbol // Si no es un índice conocido, usar el símbolo original
}

/* GetStockData obtiene datos de un símbolo específico desde Alpha Vantage */
func (avs *AlphaVantageService) GetStockData(symbol string) (*models.StockData, error) {
	/* Mapear índice a ETF si es necesario */
	actualSymbol := avs.mapIndexToETF(symbol)

	/* Construir URL de la API */
	url := fmt.Sprintf("%s?function=TIME_SERIES_DAILY&symbol=%s&apikey=%s",
		avs.BaseURL, actualSymbol, avs.APIKey)

	/* Realizar petición HTTP */
	resp, err := avs.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error realizando petición a Alpha Vantage: %v", err)
	}
	defer resp.Body.Close()

	/* Leer respuesta */
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %v", err)
	}

	/* Parsear JSON */
	var avResponse AlphaVantageResponse
	if err := json.Unmarshal(body, &avResponse); err != nil {
		return nil, fmt.Errorf("error parseando JSON: %v", err)
	}

	/* Verificar si hay datos */
	if len(avResponse.TimeSeriesDaily) == 0 {
		return nil, fmt.Errorf("no hay datos disponibles para %s", symbol)
	}

	/* Obtener el dato más reciente */
	var latestDate string
	for date := range avResponse.TimeSeriesDaily {
		if latestDate == "" || date > latestDate {
			latestDate = date
		}
	}

	latestData := avResponse.TimeSeriesDaily[latestDate]

	/* Convertir strings a float64 */
	open, _ := strconv.ParseFloat(latestData.Open, 64)
	high, _ := strconv.ParseFloat(latestData.High, 64)
	low, _ := strconv.ParseFloat(latestData.Low, 64)
	close, _ := strconv.ParseFloat(latestData.Close, 64)
	volume, _ := strconv.ParseFloat(latestData.Volume, 64)

	/* Fecha ya está en formato string correcto para el modelo */

	/* Crear estructura StockData */
	stockData := &models.StockData{
		Status:     "OK",
		From:       "Alpha Vantage",
		Symbol:     symbol, // Usar símbolo original (SPX, no SPY)
		Date:       latestDate,
		Open:       open,
		High:       high,
		Low:        low,
		Close:      close,
		Volume:     volume,
		AfterHours: 0, // Alpha Vantage no proporciona datos after hours en esta consulta
		PreMarket:  0, // Alpha Vantage no proporciona datos pre-market en esta consulta
	}

	return stockData, nil
}

/* GetMultipleStockData obtiene datos de múltiples símbolos */
func (avs *AlphaVantageService) GetMultipleStockData(symbols []string) error {
	avs.mu.Lock()
	avs.results = make([]models.StockData, 0)
	avs.failed = make([]string, 0)
	avs.mu.Unlock()

	/* Canal para controlar concurrencia */
	semaphore := make(chan struct{}, 5) // Máximo 5 peticiones concurrentes
	var wg sync.WaitGroup

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Adquirir semáforo
			defer func() { <-semaphore }() // Liberar semáforo

			data, err := avs.GetStockData(sym)

			avs.mu.Lock()
			if err != nil {
				avs.failed = append(avs.failed, fmt.Sprintf("%s: %v", sym, err))
			} else {
				avs.results = append(avs.results, *data)
			}
			avs.mu.Unlock()

			/* Pequeña pausa para respetar límites de rate */
			time.Sleep(200 * time.Millisecond)
		}(symbol)
	}

	wg.Wait()
	return nil
}

/* GetResults retorna los resultados exitosos */
func (avs *AlphaVantageService) GetResults() []models.StockData {
	avs.mu.RLock()
	defer avs.mu.RUnlock()
	return avs.results
}

/* GetFailed retorna los símbolos que fallaron */
func (avs *AlphaVantageService) GetFailed() []string {
	avs.mu.RLock()
	defer avs.mu.RUnlock()
	return avs.failed
}

/* ClearResults limpia los resultados almacenados */
func (avs *AlphaVantageService) ClearResults() {
	avs.mu.Lock()
	defer avs.mu.Unlock()
	avs.results = make([]models.StockData, 0)
	avs.failed = make([]string, 0)
}

/* GetName retorna el nombre del proveedor */
func (avs *AlphaVantageService) GetName() string {
	return "Alpha Vantage"
}

/* GetDailyCallLimit retorna el límite diario de llamadas */
func (avs *AlphaVantageService) GetDailyCallLimit() int {
	return 500
}
