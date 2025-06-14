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

// DataProvider interfaz para proveedores de datos financieros
type DataProvider interface {
	FetchData(symbol, date string)
	FetchWeeklyData(symbols []string, dates []string) error
	GetResults() []models.StockData
	GetFailed() []string
	ClearResults()
}

// FMPService utiliza la API de Financial Modeling Prep para obtener datos financieros
type FMPService struct {
	APIKey  string
	Results []models.StockData
	Failed  []string
	Mutex   sync.Mutex
	Client  *http.Client
}

// FMPResponse estructura para la respuesta de FMP API
type FMPResponse struct {
	Date             string  `json:"date"`
	Open             float64 `json:"open"`
	High             float64 `json:"high"`
	Low              float64 `json:"low"`
	Close            float64 `json:"close"`
	AdjustedClose    float64 `json:"adjClose"`
	Volume           float64 `json:"volume"`
	UnadjustedVolume float64 `json:"unadjustedVolume"`
	Change           float64 `json:"change"`
	ChangePercent    float64 `json:"changePercent"`
	Vwap             float64 `json:"vwap"`
	Label            string  `json:"label"`
	ChangeOverTime   float64 `json:"changeOverTime"`
}

// NewFMPService crea una nueva instancia del servicio FMP
func NewFMPService(apiKey string) *FMPService {
	return &FMPService{
		APIKey: apiKey,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchData obtiene datos de un símbolo específico para una fecha dada
func (fs *FMPService) FetchData(symbol, date string) {
	// Convertir los símbolos a formato FMP
	fmpSymbol := convertToFMPSymbol(symbol)

	// URL para datos históricos diarios
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/historical-price-full/%s?from=%s&to=%s&apikey=%s", fmpSymbol, date, date, fs.APIKey)

	resp, err := fs.Client.Get(url)
	if err != nil {
		log.Printf("Error obteniendo datos para %s en %s: %v", symbol, date, err)
		fs.addFailed(symbol, date)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Estado HTTP no válido: %d para %s en %s", resp.StatusCode, symbol, date)
		fs.addFailed(symbol, date)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error leyendo respuesta para %s en %s: %v", symbol, date, err)
		fs.addFailed(symbol, date)
		return
	}

	var response struct {
		Symbol     string        `json:"symbol"`
		Historical []FMPResponse `json:"historical"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error deserializando JSON para %s en %s: %v", symbol, date, err)
		fs.addFailed(symbol, date)
		return
	}

	if len(response.Historical) == 0 {
		log.Printf("Sin datos históricos para %s en %s", symbol, date)
		fs.addFailed(symbol, date)
		return
	}

	// Tomar el primer registro (debería ser el único para la fecha específica)
	data := response.Historical[0]
	stockData := models.StockData{
		Status:     "OK",
		Symbol:     symbol,
		Date:       date,
		Open:       data.Open,
		High:       data.High,
		Low:        data.Low,
		Close:      data.Close,
		Volume:     data.Volume,
		AfterHours: 0, // FMP no proporciona datos afterhours en este endpoint
		PreMarket:  0, // FMP no proporciona datos premarket en este endpoint
	}

	fs.addResult(stockData)

	// Pausa de 1 segundo para evitar saturar el servidor en llamadas individuales
	time.Sleep(1 * time.Second)
}

// FetchWeeklyData obtiene datos para múltiples símbolos y fechas
func (fs *FMPService) FetchWeeklyData(symbols []string, dates []string) error {
	var wg sync.WaitGroup
	// Límite de 5 solicitudes concurrentes para el plan gratuito
	semaphore := make(chan struct{}, 5)

	for _, symbol := range symbols {
		for _, date := range dates {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(sym, dt string) {
				defer wg.Done()
				fs.FetchData(sym, dt)
				// Pausa de 1 segundo para evitar saturar el servidor
				time.Sleep(1 * time.Second)
				<-semaphore
			}(symbol, date)
		}
	}

	wg.Wait()
	return nil
}

// GetResults devuelve los resultados obtenidos
func (fs *FMPService) GetResults() []models.StockData {
	fs.Mutex.Lock()
	defer fs.Mutex.Unlock()
	return fs.Results
}

// GetFailed devuelve la lista de fallos
func (fs *FMPService) GetFailed() []string {
	fs.Mutex.Lock()
	defer fs.Mutex.Unlock()
	return fs.Failed
}

// ClearResults limpia los resultados y fallos
func (fs *FMPService) ClearResults() {
	fs.Mutex.Lock()
	defer fs.Mutex.Unlock()
	fs.Results = make([]models.StockData, 0)
	fs.Failed = make([]string, 0)
}

// addResult añade un resultado exitoso de forma thread-safe
func (fs *FMPService) addResult(data models.StockData) {
	fs.Mutex.Lock()
	defer fs.Mutex.Unlock()
	fs.Results = append(fs.Results, data)
}

// addFailed añade un símbolo fallido de forma thread-safe
func (fs *FMPService) addFailed(symbol, date string) {
	fs.Mutex.Lock()
	defer fs.Mutex.Unlock()
	fs.Failed = append(fs.Failed, fmt.Sprintf("%s (%s)", symbol, date))
}

/* GetStockData obtiene datos de un símbolo específico */
func (fs *FMPService) GetStockData(symbol string) (*models.StockData, error) {
	/* Limpiar resultados previos */
	fs.ClearResults()

	/* Obtener fecha de hoy */
	today := time.Now().Format("2006-01-02")

	/* Obtener datos */
	fs.FetchData(symbol, today)

	/* Verificar resultados */
	results := fs.GetResults()
	if len(results) == 0 {
		return nil, fmt.Errorf("no se encontraron datos para %s", symbol)
	}

	return &results[0], nil
}

/* GetMultipleStockData obtiene datos de múltiples símbolos */
func (fs *FMPService) GetMultipleStockData(symbols []string) error {
	/* Limpiar resultados previos */
	fs.ClearResults()

	/* Obtener fecha de hoy */
	today := time.Now().Format("2006-01-02")
	dates := []string{today}

	/* Obtener datos semanales */
	fs.FetchWeeklyData(symbols, dates)

	return nil
}

/* GetName retorna el nombre del proveedor */
func (fs *FMPService) GetName() string {
	return "Financial Modeling Prep"
}

/* GetDailyCallLimit retorna el límite diario de llamadas */
func (fs *FMPService) GetDailyCallLimit() int {
	return 250
}

// convertToFMPSymbol convierte símbolos a formato FMP
func convertToFMPSymbol(symbol string) string {
	symbolMap := map[string]string{
		"SPX":  "^GSPC", // S&P 500 Index
		"NDX":  "^IXIC", // NASDAQ Composite (o usar ^NDX)
		"DJI":  "^DJI",  // Dow Jones Industrial Average
		"NYA":  "^NYA",  // NYSE Composite Index
		"ES_F": "ES=F",  // E-mini S&P 500 Futures
		"NQ_F": "NQ=F",  // E-mini NASDAQ 100 Futures
		"ES=F": "ES=F",
		"NQ=F": "NQ=F",
	}

	if fmpSymbol, exists := symbolMap[symbol]; exists {
		return fmpSymbol
	}

	// Si no hay mapeo específico, devolver el símbolo original
	return symbol
}

/* GetTargetIndices devuelve la lista de índices objetivo solicitados */
func GetTargetIndices() []string {
	return []string{
		"SPX",  // S&P 500 Index
		"NDX",  // NASDAQ Composite
		"DJI",  // Dow Jones Industrial Average
		"NYA",  // NYSE Composite Index
		"ES_F", // E-mini S&P 500 Futures
		"NQ_F", // E-mini NASDAQ 100 Futures
	}
}

/* GetStockSymbols devuelve todos los símbolos de stocks para obtención diaria */
func GetStockSymbols() []string {
	return []string{
		/* ETFs and major indexes */
		"SPY", "QQQ", "IWM", "DIA", "SMH",

		/* Bond and inverse ETFs */
		"TLT", "PSQ", "SH",

		/* Individual stocks */
		"NFLX", "COST", "NVDA", "META", "MSFT", "AMZN", "GOOG", "AAPL", "TSLA", "PLTR",
		"AMD", "MSTR", "LLY", "AVGO", "UNH", "PFE",

		/* Special format stocks */
		"BRK.B", // Berkshire Hathaway Class B

		/* Commodities */
		"GLD", "SLV",

		/* Leveraged ETFs */
		"TQQQ", "SQQQ", "UPRO", "SPXS", "UDOW", "SDOW", "URTY", "SRTY",

		/* Sector ETFs (SPDR Select Sector) */
		"XLC",  // Communication Services
		"XLF",  // Financial
		"XLE",  // Energy
		"XLK",  // Technology
		"XLY",  // Consumer Discretionary
		"XLI",  // Industrial
		"XLB",  // Materials
		"XLRE", // Real Estate
		"XLP",  // Consumer Staples
		"XLV",  // Health Care
		"XLU",  // Utilities
		"ETH",  // Ethereum (if available as stock symbol)

		/* Crypto ETFs */
		"IBIT", // iShares Bitcoin
	}
}

/* GetExtendedSymbols devuelve todos los símbolos: índices + stocks */
func GetExtendedSymbols() []string {
	indices := GetTargetIndices()
	stocks := GetStockSymbols()

	/* Combinar ambas listas */
	allSymbols := make([]string, 0, len(indices)+len(stocks))
	allSymbols = append(allSymbols, indices...)
	allSymbols = append(allSymbols, stocks...)

	return allSymbols
}
