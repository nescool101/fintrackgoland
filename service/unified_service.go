package service

import (
	"fmt"
	"log"
	"sync"

	"github.com/nescool101/fintrackgoland/models"
)

/* HybridService combina FMP para stocks y Alpha Vantage para índices */
type HybridService struct {
	fmpService          *FMPService
	alphaVantageService *AlphaVantageService
	mu                  sync.RWMutex
	results             []models.StockData
	failed              []string
}

/* NewHybridService crea un servicio híbrido que usa ambas APIs */
func NewHybridService(fmpAPIKey, alphaVantageKey string) *HybridService {
	return &HybridService{
		fmpService:          NewFMPService(fmpAPIKey),
		alphaVantageService: NewAlphaVantageService(alphaVantageKey),
		results:             make([]models.StockData, 0),
		failed:              make([]string, 0),
	}
}

/* isIndex determina si un símbolo es un índice que debe usar Alpha Vantage */
func (hs *HybridService) isIndex(symbol string) bool {
	indices := map[string]bool{
		"SPX":  true, // S&P 500
		"NDX":  true, // Nasdaq 100
		"DJI":  true, // Dow Jones
		"NYA":  true, // NYSE Composite
		"ES_F": true, // E-mini S&P 500 Futures
		"NQ_F": true, // E-mini Nasdaq Futures
	}
	return indices[symbol]
}

/* GetStockData obtiene datos usando el proveedor apropiado */
func (hs *HybridService) GetStockData(symbol string) (*models.StockData, error) {
	if hs.isIndex(symbol) {
		return hs.alphaVantageService.GetStockData(symbol)
	}
	return hs.fmpService.GetStockData(symbol)
}

/* GetMultipleStockData obtiene datos de múltiples símbolos */
func (hs *HybridService) GetMultipleStockData(symbols []string) error {
	hs.mu.Lock()
	hs.results = make([]models.StockData, 0)
	hs.failed = make([]string, 0)
	hs.mu.Unlock()

	/* Separar símbolos por proveedor */
	var indexSymbols []string
	var stockSymbols []string

	for _, symbol := range symbols {
		if hs.isIndex(symbol) {
			indexSymbols = append(indexSymbols, symbol)
		} else {
			stockSymbols = append(stockSymbols, symbol)
		}
	}

	var wg sync.WaitGroup

	/* Procesar índices con Alpha Vantage */
	if len(indexSymbols) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("📊 Obteniendo %d índices usando Alpha Vantage: %v", len(indexSymbols), indexSymbols)

			err := hs.alphaVantageService.GetMultipleStockData(indexSymbols)
			if err != nil {
				log.Printf("Error obteniendo datos de Alpha Vantage: %v", err)
			}

			/* Combinar resultados */
			hs.mu.Lock()
			hs.results = append(hs.results, hs.alphaVantageService.GetResults()...)
			hs.failed = append(hs.failed, hs.alphaVantageService.GetFailed()...)
			hs.mu.Unlock()
		}()
	}

	/* Procesar stocks con FMP */
	if len(stockSymbols) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("📈 Obteniendo %d stocks usando FMP: %v", len(stockSymbols), stockSymbols)

			err := hs.fmpService.GetMultipleStockData(stockSymbols)
			if err != nil {
				log.Printf("Error obteniendo datos de FMP: %v", err)
			}

			/* Combinar resultados */
			hs.mu.Lock()
			hs.results = append(hs.results, hs.fmpService.GetResults()...)
			hs.failed = append(hs.failed, hs.fmpService.GetFailed()...)
			hs.mu.Unlock()
		}()
	}

	wg.Wait()
	return nil
}

/* FetchData obtiene datos de un símbolo específico para una fecha */
func (hs *HybridService) FetchData(symbol, date string) {
	data, err := hs.GetStockData(symbol)
	if err != nil {
		hs.mu.Lock()
		hs.failed = append(hs.failed, fmt.Sprintf("%s (%s): %v", symbol, date, err))
		hs.mu.Unlock()
		return
	}

	hs.mu.Lock()
	hs.results = append(hs.results, *data)
	hs.mu.Unlock()
}

/* FetchWeeklyData obtiene datos semanales usando ambos proveedores */
func (hs *HybridService) FetchWeeklyData(symbols, dates []string) error {
	log.Printf("🔄 Iniciando obtención híbrida de datos para %d símbolos", len(symbols))
	log.Printf("📅 Fechas: %v", dates)

	/* Para datos semanales, usamos GetMultipleStockData */
	return hs.GetMultipleStockData(symbols)
}

/* GetResults retorna todos los resultados combinados */
func (hs *HybridService) GetResults() []models.StockData {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.results
}

/* GetFailed retorna todos los fallos combinados */
func (hs *HybridService) GetFailed() []string {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.failed
}

/* ClearResults limpia todos los resultados */
func (hs *HybridService) ClearResults() {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.results = make([]models.StockData, 0)
	hs.failed = make([]string, 0)
	hs.fmpService.ClearResults()
	hs.alphaVantageService.ClearResults()
}

/* GetName retorna información del servicio híbrido */
func (hs *HybridService) GetName() string {
	return "Híbrido (FMP + Alpha Vantage)"
}

/* GetInfo retorna información detallada de ambos proveedores */
func (hs *HybridService) GetInfo() string {
	return fmt.Sprintf("FMP: %s (%d llamadas/día) + Alpha Vantage: %s (%d llamadas/día)",
		hs.fmpService.GetName(), hs.fmpService.GetDailyCallLimit(),
		hs.alphaVantageService.GetName(), hs.alphaVantageService.GetDailyCallLimit())
}

/* GetDailyCallLimit retorna el límite combinado */
func (hs *HybridService) GetDailyCallLimit() int {
	return hs.fmpService.GetDailyCallLimit() + hs.alphaVantageService.GetDailyCallLimit()
}

/* GetProviderForSymbol retorna qué proveedor se usa para un símbolo */
func (hs *HybridService) GetProviderForSymbol(symbol string) string {
	if hs.isIndex(symbol) {
		return "Alpha Vantage"
	}
	return "FMP"
}

/* GetSupportedSymbols retorna la lista de símbolos soportados */
func (hs *HybridService) GetSupportedSymbols() map[string]string {
	supported := make(map[string]string)

	/* Índices soportados por Alpha Vantage */
	indices := []string{"SPX", "NDX", "DJI", "NYA", "ES_F", "NQ_F"}
	for _, idx := range indices {
		supported[idx] = "Alpha Vantage (ETF proxy)"
	}

	/* Stocks soportados por FMP */
	supported["AAPL"] = "FMP"
	supported["MSFT"] = "FMP"
	supported["GOOGL"] = "FMP"
	supported["TSLA"] = "FMP"
	supported["NVDA"] = "FMP"

	return supported
}
