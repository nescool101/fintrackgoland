package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nescool101/fintrackgoland/api"
	"github.com/nescool101/fintrackgoland/config"
	"github.com/nescool101/fintrackgoland/service"
	"github.com/nescool101/fintrackgoland/utils"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg := config.LoadConfig()

	/*
	 * Usar solo FMP para todo: stocks + índices
	 * FMP API: 250 llamadas gratis/día para stocks e índices
	 * Alpha Vantage tiene límite muy bajo (25 llamadas/día)
	 */
	log.Println("🚀 Usando solo FMP para stocks e índices")
	log.Printf("📊 FMP API Key: %s... (250 llamadas/día)", cfg.FMPAPIKey[:8])
	dataService := service.NewFMPService(cfg.FMPAPIKey)

	// Configurar Gin en modo release para producción
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Crear el router Gin
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Crear el manejador de API
	apiHandler := api.NewAPIHandler(cfg, dataService)

	// Rutas públicas
	router.GET("/health", apiHandler.HealthCheck)
	router.GET("/api/status", apiHandler.GetAPIStatus)
	router.GET("/api/indices", apiHandler.GetSupportedIndices)
	router.GET("/api/excel/full", apiHandler.SendFullReport) // GET /api/excel/full?date=2024-01-15&recipient=nescool101@gmail.com (todos los símbolos) - SIN AUTENTICACIÓN

	// Rutas protegidas con autenticación básica
	protected := router.Group("/api")
	protected.Use(apiHandler.BasicAuth())
	{
		/* Endpoints para obtener datos de mercado */
		protected.GET("/stock/:symbol", apiHandler.GetStockData)  // GET /api/stock/SPX?date=2024-01-15
		protected.GET("/stocks", apiHandler.GetMultipleStockData) // GET /api/stocks?symbols=SPX,NDX,DJI&date=2024-01-15
		protected.GET("/weekly", apiHandler.GetWeeklyData)        // GET /api/weekly?symbols=SPX,NDX (opcional)

		/* Endpoints para generar y enviar reportes */
		protected.POST("/report/send", apiHandler.SendWeeklyReport) // POST /api/report/send
		protected.POST("/excel/send", apiHandler.SendExcelReport)   // POST /api/excel/send?symbols=SPX,NDX&date=2024-01-15&recipient=nescool101@gmail.com
	}

	// Obtener símbolos y fechas para el cron
	symbols := service.GetExtendedSymbols()
	dates := getWeekDates()

	// Información de configuración
	targetIndices := service.GetTargetIndices()
	stockSymbols := service.GetStockSymbols()
	log.Printf("🎯 Índices objetivo configurados: %v", targetIndices)
	log.Printf("📈 Stocks configurados: %d símbolos", len(stockSymbols))
	log.Printf("📊 Total de símbolos a procesar: %d", len(symbols))
	log.Printf("🔑 Servicio FMP: %d stocks + %d índices = 250 llamadas/día total",
		len(stockSymbols), len(targetIndices))

	// Configurar servidor HTTP con Gin
	server := &gin.Engine{}
	*server = *router

	// Iniciar servidor HTTP en goroutine separada
	go func() {
		log.Println("🚀 Iniciando servidor REST API en puerto :8080")
		log.Println("📋 Endpoints disponibles:")
		log.Println("   GET  /health                    - Estado del servicio (público)")
		log.Println("   GET  /api/status                - Información de APIs (público)")
		log.Println("   GET  /api/indices               - Índices soportados (público)")
		log.Println("   GET  /api/excel/full            - Generar reporte completo (54 símbolos) - PÚBLICO")
		log.Println("   GET  /api/stock/:symbol         - Datos de un símbolo (protegido)")
		log.Println("   GET  /api/stocks                - Datos de múltiples símbolos (protegido)")
		log.Println("   GET  /api/weekly                - Datos semanales (protegido)")
		log.Println("   POST /api/report/send           - Enviar reporte semanal (protegido)")
		log.Println("   POST /api/excel/send            - Generar y enviar Excel (protegido)")

		// Configurar puerto desde variable de entorno o usar 8080 por defecto
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		// Escuchar en todas las interfaces (0.0.0.0) para Fly.io
		addr := "0.0.0.0:" + port
		log.Printf("🌐 Servidor escuchando en %s", addr)

		if err := router.Run(addr); err != nil {
			log.Fatalf("Error iniciando servidor: %v", err)
		}
	}()

	// Configurar programador cron si está habilitado
	if cfg.RunCron {
		c := cron.New(cron.WithSeconds())
		/* Ejecutar cada viernes a las 9:00 AM */
		_, err := c.AddFunc("0 0 9 ? * FRI", func() {
			log.Println("📅 Ejecutando obtención programada de datos financieros")
			runDataFetch(dataService, symbols, dates, cfg)
		})
		if err != nil {
			log.Fatalf("Error programando tarea cron: %v", err)
		}
		c.Start()
		defer c.Stop()
		log.Println("⏰ Cron configurado: cada viernes a las 9:00 AM")
	}

	// Escuchar señales del OS para apagado graceful
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("🛑 Apagando servidor gracefully...")

	log.Println("✅ Servidor finalizado correctamente")
}

// getWeekDates genera las fechas de la semana actual para obtener datos
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

func runDataFetch(ds service.DataProvider, symbols, dates []string, cfg *config.Config) {
	err := ds.FetchWeeklyData(symbols, dates)
	if err != nil {
		log.Printf("Error obteniendo datos semanales: %v", err)
	}

	results := ds.GetResults()
	excelData, err := utils.GenerateExcel(results)
	if err != nil {
		log.Printf("Failed to generate Excel: %v", err)
		return
	}

	failed := ds.GetFailed()
	failedMessage := ""
	if len(failed) > 0 {
		failedMessage = fmt.Sprintf("No fue posible recibir el stock de %s.", strings.Join(failed, ", "))
	}

	emailBody := "Adjunto encontrarás el reporte semanal de datos."
	if failedMessage != "" {
		emailBody += "\n\n" + failedMessage
	}

	err = utils.SendEmail(
		cfg.EmailHost,
		cfg.EmailPort,
		cfg.EmailUser,
		cfg.EmailPass,
		cfg.Recipient,
		"Weekly Data Report",
		emailBody,
		excelData,
		"WeeklyData.xlsx",
	)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return
	}

	log.Println("Process completed successfully")
}
