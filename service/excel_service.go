package service

import (
	"fmt"
	"time"

	"github.com/nescool101/fintrackgoland/models"
	"github.com/xuri/excelize/v2"
)

// ExcelService manejo de generación de archivos Excel
type ExcelService struct{}

// NewExcelService crea una nueva instancia del servicio Excel
func NewExcelService() *ExcelService {
	return &ExcelService{}
}

// GenerateStockReport genera un archivo Excel con datos de stocks/indices
func (es *ExcelService) GenerateStockReport(data []models.StockData, filename string) ([]byte, error) {
	// Crear nuevo archivo Excel
	f := excelize.NewFile()
	defer f.Close()

	// Crear hoja para el reporte
	sheetName := "Reporte_Financiero"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("error creando hoja: %v", err)
	}

	// Configurar encabezados
	headers := []string{
		"Tipo", "Símbolo", "Fecha", "Apertura", "Máximo",
		"Mínimo", "Cierre", "Volumen", "Fuente", "Estado",
	}

	// Escribir encabezados con formato
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Establecer estilo para encabezados
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6E6FA"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err == nil {
		f.SetRowStyle(sheetName, 1, 1, headerStyle)
	}

	// Escribir datos
	for i, stock := range data {
		row := i + 2

		// Determinar tipo (Stock o Indice)
		tipo := es.determineType(stock.Symbol)

		// Escribir datos en cada columna
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), tipo)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), stock.Symbol)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), stock.Date)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), stock.Open)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), stock.High)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), stock.Low)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), stock.Close)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), stock.Volume)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), stock.From)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), stock.Status)
	}

	// Ajustar ancho de columnas
	columns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	widths := []float64{10, 12, 12, 12, 12, 12, 12, 15, 15, 10}

	for i, col := range columns {
		f.SetColWidth(sheetName, col, col, widths[i])
	}

	// Agregar hoja de resumen
	es.addSummarySheet(f, data)

	// Establecer la hoja principal como activa
	f.SetActiveSheet(index)

	// Generar archivo en memoria
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("error generando archivo Excel: %v", err)
	}

	return buffer.Bytes(), nil
}

// determineType determina si el símbolo es un stock o índice
func (es *ExcelService) determineType(symbol string) string {
	indices := []string{"SPX", "NDX", "DJI", "NYA", "ES_F", "NQ_F", "^GSPC", "^IXIC", "^DJI", "^NYA"}
	for _, index := range indices {
		if symbol == index {
			return "Índice"
		}
	}
	return "Stock"
}

// addSummarySheet agrega una hoja de resumen al archivo Excel
func (es *ExcelService) addSummarySheet(f *excelize.File, data []models.StockData) {
	summarySheet := "Resumen"
	_, err := f.NewSheet(summarySheet)
	if err != nil {
		return
	}

	// Título del resumen
	f.SetCellValue(summarySheet, "A1", "RESUMEN DEL REPORTE FINANCIERO")
	f.SetCellValue(summarySheet, "A2", fmt.Sprintf("Generado el: %s", time.Now().Format("2006-01-02 15:04:05")))
	f.SetCellValue(summarySheet, "A3", fmt.Sprintf("Total de símbolos: %d", len(data)))

	// Contar tipos
	stocks, indices := 0, 0
	for _, item := range data {
		if es.determineType(item.Symbol) == "Índice" {
			indices++
		} else {
			stocks++
		}
	}

	f.SetCellValue(summarySheet, "A5", "DISTRIBUCIÓN POR TIPO:")
	f.SetCellValue(summarySheet, "A6", fmt.Sprintf("Stocks: %d", stocks))
	f.SetCellValue(summarySheet, "A7", fmt.Sprintf("Índices: %d", indices))

	// Fuentes de datos
	f.SetCellValue(summarySheet, "A9", "FUENTES DE DATOS:")
	sources := make(map[string]int)
	for _, item := range data {
		sources[item.From]++
	}

	row := 10
	for source, count := range sources {
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("%s: %d símbolos", source, count))
		row++
	}

	// Ajustar ancho de columna
	f.SetColWidth(summarySheet, "A", "A", 35)
}
