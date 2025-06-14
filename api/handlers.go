package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nescool101/fintrackgoland/config"
	"github.com/nescool101/fintrackgoland/models"
	"github.com/nescool101/fintrackgoland/service"
	"github.com/nescool101/fintrackgoland/utils"
)

type APIHandler struct {
	Config       *config.Config
	DataService  service.DataProvider
	ExcelService *service.ExcelService
	EmailService *service.EmailService
}

// NewAPIHandler crea una nueva instancia del manejador de API
func NewAPIHandler(cfg *config.Config, ds service.DataProvider) *APIHandler {
	return &APIHandler{
		Config:       cfg,
		DataService:  ds,
		ExcelService: service.NewExcelService(),
		EmailService: service.NewEmailService(cfg.EmailHost, cfg.EmailPort, cfg.EmailUser, cfg.EmailPass),
	}
}

// HealthCheck endpoint para verificar el estado del servicio
func (h *APIHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}

// GetStockData endpoint para obtener datos de un símbolo específico
func (h *APIHandler) GetStockData(c *gin.Context) {
	symbol := c.Param("symbol")
	dateStr := c.Query("date")

	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "El símbolo es requerido",
		})
		return
	}

	// Si no se proporciona fecha, usar la fecha actual
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Validar formato de fecha
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Formato de fecha inválido. Use YYYY-MM-DD",
		})
		return
	}

	// Limpiar resultados previos
	h.DataService.ClearResults()

	// Obtener datos
	h.DataService.FetchData(symbol, dateStr)

	results := h.DataService.GetResults()
	failed := h.DataService.GetFailed()

	response := gin.H{
		"symbol":  symbol,
		"date":    dateStr,
		"success": len(results) > 0,
	}

	if len(results) > 0 {
		response["data"] = results[0]
	}

	if len(failed) > 0 {
		response["errors"] = failed
	}

	status := http.StatusOK
	if len(results) == 0 {
		status = http.StatusNotFound
		response["message"] = "No se encontraron datos para el símbolo y fecha especificados"
	}

	c.JSON(status, response)
}

// GetMultipleStockData endpoint para obtener datos de múltiples símbolos
func (h *APIHandler) GetMultipleStockData(c *gin.Context) {
	symbolsParam := c.Query("symbols")
	dateStr := c.Query("date")

	if symbolsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Al menos un símbolo es requerido",
		})
		return
	}

	symbols := strings.Split(symbolsParam, ",")
	for i, symbol := range symbols {
		symbols[i] = strings.TrimSpace(symbol)
	}

	// Si no se proporciona fecha, usar la fecha actual
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Validar formato de fecha
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Formato de fecha inválido. Use YYYY-MM-DD",
		})
		return
	}

	// Limpiar resultados previos
	h.DataService.ClearResults()

	// Obtener datos para todos los símbolos
	dates := []string{dateStr}
	h.DataService.FetchWeeklyData(symbols, dates)

	results := h.DataService.GetResults()
	failed := h.DataService.GetFailed()

	c.JSON(http.StatusOK, gin.H{
		"symbols":    symbols,
		"date":       dateStr,
		"total":      len(symbols),
		"successful": len(results),
		"failed":     len(failed),
		"data":       results,
		"errors":     failed,
	})
}

// GetWeeklyData endpoint para obtener datos de la semana actual
func (h *APIHandler) GetWeeklyData(c *gin.Context) {
	symbolsParam := c.Query("symbols")

	var symbols []string
	if symbolsParam != "" {
		symbols = strings.Split(symbolsParam, ",")
		for i, symbol := range symbols {
			symbols[i] = strings.TrimSpace(symbol)
		}
	} else {
		// Usar símbolos extendidos por defecto
		symbols = service.GetExtendedSymbols()
	}

	// Obtener fechas de la semana actual
	dates := getWeekDates()

	// Limpiar resultados previos
	h.DataService.ClearResults()

	// Obtener datos
	h.DataService.FetchWeeklyData(symbols, dates)

	results := h.DataService.GetResults()
	failed := h.DataService.GetFailed()

	c.JSON(http.StatusOK, gin.H{
		"symbols":    symbols,
		"dates":      dates,
		"total":      len(symbols) * len(dates),
		"successful": len(results),
		"failed":     len(failed),
		"data":       results,
		"errors":     failed,
	})
}

// GetSupportedIndices endpoint para listar los índices objetivo soportados
func (h *APIHandler) GetSupportedIndices(c *gin.Context) {
	indices := service.GetTargetIndices()
	allSymbols := service.GetExtendedSymbols()

	c.JSON(http.StatusOK, gin.H{
		"target_indices":   indices,
		"all_symbols":      allSymbols,
		"total_indices":    len(indices),
		"total_symbols":    len(allSymbols),
		"api_provider":     "FMP (Financial Modeling Prep)",
		"daily_free_calls": 250,
	})
}

// SendWeeklyReport endpoint para generar y enviar el reporte semanal
func (h *APIHandler) SendWeeklyReport(c *gin.Context) {
	// Usar símbolos extendidos
	symbols := service.GetExtendedSymbols()
	dates := getWeekDates()

	// Limpiar resultados previos
	h.DataService.ClearResults()

	// Obtener datos
	h.DataService.FetchWeeklyData(symbols, dates)

	results := h.DataService.GetResults()
	failed := h.DataService.GetFailed()

	// Generar Excel
	excelData, err := utils.GenerateExcel(results)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generando archivo Excel: " + err.Error(),
		})
		return
	}

	// Preparar mensaje de email
	emailBody := "Adjunto encontrarás el reporte semanal de datos financieros."
	if len(failed) > 0 {
		emailBody += "\n\nFallos: No fue posible obtener datos para: " + strings.Join(failed, ", ")
	}

	// Enviar email
	err = utils.SendEmail(
		h.Config.EmailHost,
		h.Config.EmailPort,
		h.Config.EmailUser,
		h.Config.EmailPass,
		h.Config.Recipient,
		"Reporte Semanal de Datos Financieros",
		emailBody,
		excelData,
		"ReporteSemanal.xlsx",
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error enviando email: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Reporte semanal enviado exitosamente",
		"symbols":    len(symbols),
		"successful": len(results),
		"failed":     len(failed),
		"dates":      dates,
	})
}

// GetAPIStatus endpoint para obtener información sobre el estado de las APIs
func (h *APIHandler) GetAPIStatus(c *gin.Context) {
	response := gin.H{
		"timestamp": time.Now().Format(time.RFC3339),
		"api": gin.H{
			"name":               "Financial Modeling Prep",
			"free_calls_per_day": 250,
			"supports_indices":   true,
			"status":             "active",
			"website":            "https://financialmodelingprep.com",
		},
		"supported_indices": service.GetTargetIndices(),
		"total_symbols":     len(service.GetExtendedSymbols()),
	}

	c.JSON(http.StatusOK, response)
}

// BasicAuth middleware para autenticación básica
func (h *APIHandler) BasicAuth() gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		h.Config.AuthUsername: h.Config.AuthPassword,
	})
}

// isIndex determina si un símbolo es un índice
func (h *APIHandler) isIndex(symbol string) bool {
	indices := []string{"SPX", "NDX", "DJI", "NYA", "ES_F", "NQ_F", "^GSPC", "^IXIC", "^DJI", "^NYA"}
	for _, index := range indices {
		if symbol == index {
			return true
		}
	}
	return false
}

// getWeekDates genera las fechas de la semana actual
func getWeekDates() []string {
	now := time.Now()

	// Encontrar el lunes de esta semana
	weekday := int(now.Weekday())
	if weekday == 0 { // Domingo
		weekday = 7
	}
	monday := now.AddDate(0, 0, -weekday+1)

	var dates []string
	for i := 0; i < 5; i++ { // Lunes a viernes
		date := monday.AddDate(0, 0, i)
		dates = append(dates, date.Format("2006-01-02"))
	}

	return dates
}

// SendExcelReport endpoint para generar y enviar reporte Excel específico
func (h *APIHandler) SendExcelReport(c *gin.Context) {
	// Obtener parámetros opcionales
	symbolsParam := c.Query("symbols")
	dateStr := c.Query("date")
	recipientEmail := c.Query("recipient")

	// Usar emails por defecto si no se proporciona
	if recipientEmail == "" {
		recipientEmail = "nescool101@gmail.com,paulocesarcelis@gmail.com"
	}

	// Determinar símbolos
	var symbols []string
	if symbolsParam != "" {
		symbols = strings.Split(symbolsParam, ",")
		for i, symbol := range symbols {
			symbols[i] = strings.TrimSpace(symbol)
		}
	} else {
		// Usar todos los símbolos por defecto (índices + stocks)
		// Para evitar problemas de rate limiting, usar solo los índices objetivo por defecto
		// Los usuarios pueden especificar símbolos específicos si necesitan stocks
		symbols = service.GetTargetIndices()

		// TODO: En el futuro, implementar procesamiento por lotes para incluir todos los stocks
		// symbols = service.GetExtendedSymbols()
	}

	// Determinar fecha
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Validar formato de fecha
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Formato de fecha inválido. Use YYYY-MM-DD",
		})
		return
	}

	// Limpiar resultados previos
	h.DataService.ClearResults()

	// Obtener datos
	dates := []string{dateStr}
	err := h.DataService.FetchWeeklyData(symbols, dates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error obteniendo datos: " + err.Error(),
		})
		return
	}

	results := h.DataService.GetResults()
	failed := h.DataService.GetFailed()

	// Verificar si hay datos
	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No se encontraron datos para los símbolos especificados",
		})
		return
	}

	// Generar archivo Excel
	filename := fmt.Sprintf("Reporte_Financiero_%s.xlsx", dateStr)
	excelData, err := h.ExcelService.GenerateStockReport(results, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generando archivo Excel: " + err.Error(),
		})
		return
	}

	// Preparar mensaje de email
	subject := fmt.Sprintf("📊 Reporte Financiero - %s", dateStr)
	emailBody := fmt.Sprintf("Se adjunta el reporte financiero para la fecha %s.", dateStr)

	// Contar tipos de datos obtenidos
	stocks, indices := 0, 0
	for _, result := range results {
		if h.isIndex(result.Symbol) {
			indices++
		} else {
			stocks++
		}
	}

	if len(failed) > 0 {
		emailBody += fmt.Sprintf("\n\n⚠️ Advertencia: No fue posible obtener datos para: %s", strings.Join(failed, ", "))

		// Verificar si faltan índices específicamente
		failedIndices := 0
		for _, failedSymbol := range failed {
			if h.isIndex(failedSymbol) {
				failedIndices++
			}
		}

		if failedIndices > 0 && indices == 0 {
			emailBody += `

<div class="warning">
<strong>📊 Información sobre índices:</strong><br>
No se pudieron obtener datos de índices. Posibles causas:<br>
• Límite de API alcanzado (Alpha Vantage: 25 llamadas/día)<br>
• Los límites de API se renuevan cada 24 horas<br>
• Considere usar solo datos de stocks o actualizar a un plan premium
</div>`
		}
	}

	// Enviar email
	err = h.EmailService.SendExcelReport(recipientEmail, subject, emailBody, excelData, filename, len(results))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error enviando email: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Reporte enviado exitosamente",
		"recipient":        recipientEmail,
		"date":             dateStr,
		"symbols_total":    len(symbols),
		"symbols_success":  len(results),
		"symbols_failed":   len(failed),
		"excel_filename":   filename,
		"excel_size_bytes": len(excelData),
		"data_summary":     results,
	})
}

// SendFullReport endpoint para generar reporte completo con todos los símbolos (stocks + índices)
func (h *APIHandler) SendFullReport(c *gin.Context) {
	dateStr := c.Query("date")
	recipientEmail := c.Query("recipient")

	// Usar emails por defecto si no se proporciona
	if recipientEmail == "" {
		recipientEmail = "nescool101@gmail.com,paulocesarcelis@gmail.com"
	}

	// Determinar fecha basada en la hora del servidor
	if dateStr == "" {
		now := time.Now()

		// Si es antes de las 3 PM (15:00), usar el día anterior
		// Si es después de las 3 PM, usar el día actual
		if now.Hour() < 15 {
			// Usar día anterior
			yesterday := now.AddDate(0, 0, -1)
			dateStr = yesterday.Format("2006-01-02")
			log.Printf("⏰ Hora actual: %s (antes de 3 PM) - Usando fecha del día anterior: %s",
				now.Format("15:04:05"), dateStr)
		} else {
			// Usar día actual
			dateStr = now.Format("2006-01-02")
			log.Printf("⏰ Hora actual: %s (después de 3 PM) - Usando fecha actual: %s",
				now.Format("15:04:05"), dateStr)
		}
	}

	// Validar formato de fecha
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Formato de fecha inválido. Use YYYY-MM-DD",
		})
		return
	}

	// Limpiar resultados previos
	h.DataService.ClearResults()

	// Obtener todos los símbolos
	allSymbols := service.GetExtendedSymbols()
	dates := []string{dateStr}

	// Procesar en lotes para evitar rate limiting
	batchSize := 10
	var allResults []models.StockData
	var allFailed []string

	for i := 0; i < len(allSymbols); i += batchSize {
		end := i + batchSize
		if end > len(allSymbols) {
			end = len(allSymbols)
		}

		batch := allSymbols[i:end]

		// Limpiar resultados del lote anterior
		h.DataService.ClearResults()

		// Procesar lote
		err := h.DataService.FetchWeeklyData(batch, dates)
		if err != nil {
			// Continuar con el siguiente lote si hay error
			continue
		}

		// Recopilar resultados del lote
		batchResults := h.DataService.GetResults()
		batchFailed := h.DataService.GetFailed()

		allResults = append(allResults, batchResults...)
		allFailed = append(allFailed, batchFailed...)

		// Pausa pequeña entre lotes para evitar rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	// Verificar si hay datos
	if len(allResults) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No se encontraron datos para ningún símbolo",
		})
		return
	}

	// Generar archivo Excel
	filename := fmt.Sprintf("Reporte_Completo_%s.xlsx", dateStr)
	excelData, err := h.ExcelService.GenerateStockReport(allResults, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generando archivo Excel: " + err.Error(),
		})
		return
	}

	// Preparar mensaje de email
	subject := fmt.Sprintf("📊 Reporte Completo - %s", dateStr)
	emailBody := fmt.Sprintf("Se adjunta el reporte completo con %d símbolos para la fecha %s.", len(allResults), dateStr)

	// Contar tipos de datos obtenidos
	stocks, indices := 0, 0
	for _, result := range allResults {
		if h.isIndex(result.Symbol) {
			indices++
		} else {
			stocks++
		}
	}

	emailBody += fmt.Sprintf("\n\n📈 Resumen del reporte:")
	emailBody += fmt.Sprintf("\n• Stocks obtenidos: %d", stocks)
	emailBody += fmt.Sprintf("\n• Índices obtenidos: %d", indices)

	if len(allFailed) > 0 {
		emailBody += fmt.Sprintf("\n\n⚠️ Advertencia: No fue posible obtener datos para %d símbolos: %s", len(allFailed), strings.Join(allFailed, ", "))

		// Verificar si faltan índices específicamente
		failedIndices := 0
		for _, failedSymbol := range allFailed {
			if h.isIndex(failedSymbol) {
				failedIndices++
			}
		}

		if failedIndices > 0 && indices == 0 {
			emailBody += `

<div class="warning">
<strong>📊 Información sobre índices:</strong><br>
No se pudieron obtener datos de índices. Posibles causas:<br>
• Límite de API alcanzado (Alpha Vantage: 25 llamadas/día)<br>
• Los límites de API se renuevan cada 24 horas<br>
• Sistema ahora usa solo FMP (250 llamadas/día para stocks e índices)
</div>`
		}
	}

	// Enviar email
	err = h.EmailService.SendExcelReport(recipientEmail, subject, emailBody, excelData, filename, len(allResults))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error enviando email: " + err.Error(),
		})
		return
	}

	// Agregar header personalizado con la fecha procesada
	c.Header("X-Processed-Date", dateStr)
	c.Header("X-Server-Time", time.Now().Format("2006-01-02 15:04:05"))
	c.Header("X-Date-Logic", func() string {
		if c.Query("date") != "" {
			return "manual"
		}
		if time.Now().Hour() < 15 {
			return "auto-previous-day"
		}
		return "auto-current-day"
	}())

	c.JSON(http.StatusOK, gin.H{
		"message":           "Reporte completo enviado exitosamente",
		"recipient":         recipientEmail,
		"date":              dateStr,
		"symbols_total":     len(allSymbols),
		"symbols_success":   len(allResults),
		"symbols_failed":    len(allFailed),
		"excel_filename":    filename,
		"excel_size_bytes":  len(excelData),
		"batches_processed": (len(allSymbols) + batchSize - 1) / batchSize,
		"data_summary":      allResults,
		"date_logic": func() string {
			if c.Query("date") != "" {
				return "Fecha especificada manualmente"
			}
			if time.Now().Hour() < 15 {
				return "Fecha automática: día anterior (antes de 3 PM)"
			}
			return "Fecha automática: día actual (después de 3 PM)"
		}(),
		"server_time": time.Now().Format("2006-01-02 15:04:05"),
	})
}
