package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nescool101/fintrackgoland/config"
	"github.com/nescool101/fintrackgoland/service"
	"github.com/nescool101/fintrackgoland/utils"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg := config.LoadConfig()

	dataService := service.NewDataService(cfg.APIKey)

	symbols := []string{
		"SPX", "NDX", "DJI", "SPY", "QQQ", "IWM", "DIA", "SMH", "TLT", "ES=F",
		"NQ=F", "NVDA", "META", "MSFT", "AMZN", "GOOG", "AAPL", "TSLA", "PLTR",
		"AMD", "MSTR", "GLD", "SLV", "NG=F", "TQQQ", "SQQQ", "UPRO", "SPXS",
		"UDOW", "SDOW", "URTY", "SRTY",
	}

	dates := getWeekDates()

	// Setup HTTP server for health checks and test endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/test", basicAuth(testHandler(cfg, dataService, symbols, dates), "admin", "password123")) // Replace with secure credentials

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start HTTP server in a separate goroutine
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Setup cron scheduler
	if cfg.RunCron {
		c := cron.New(cron.WithSeconds())
		_, err := c.AddFunc("0 0 9 ? * FRI", func() {
			log.Println("Starting scheduled data fetch")
			runDataFetch(dataService, symbols, dates, cfg)
		})
		if err != nil {
			log.Fatalf("Error scheduling cron job: %v", err)
		}
		c.Start()
		defer c.Stop()
	}

	// Listen for OS interrupt signals to gracefully shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down gracefully...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Server exited properly")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func testHandler(cfg *config.Config, ds *service.DataService, symbols, dates []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Manual data fetch triggered via /test endpoint")
		runDataFetch(ds, symbols, dates, cfg)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Data fetch and email sending initiated successfully. Check logs for details."))
	}
}

func runDataFetch(ds *service.DataService, symbols, dates []string, cfg *config.Config) {
	ds.FetchWeeklyData(symbols, dates)

	excelData, err := utils.GenerateExcel(ds.Results)
	if err != nil {
		log.Printf("Failed to generate Excel: %v", err)
		return
	}

	failedMessage := ""
	if len(ds.Failed) > 0 {
		failedMessage = fmt.Sprintf("No fue posible recibir el stock de %s.", strings.Join(ds.Failed, ", "))
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

// basicAuth is a middleware function that provides basic HTTP authentication.
func basicAuth(handler http.HandlerFunc, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}
