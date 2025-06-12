package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nescool101/fintrackgoland/config"
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

	// Usar email por defecto si no se proporciona
	if recipientEmail == "" {
		recipientEmail = "nescool101@gmail.com"
	}

	// Determinar símbolos
	var symbols []string
	if symbolsParam != "" {
		symbols = strings.Split(symbolsParam, ",")
		for i, symbol := range symbols {
			symbols[i] = strings.TrimSpace(symbol)
		}
	} else {
		// Usar índices objetivo por defecto
		symbols = service.GetTargetIndices()
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

	if len(failed) > 0 {
		emailBody += fmt.Sprintf("\n\nAdvertencia: No fue posible obtener datos para: %s", strings.Join(failed, ", "))
	}

	// Enviar email
	err = h.EmailService.SendExcelReport(recipientEmail, subject, emailBody, excelData, filename)
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
