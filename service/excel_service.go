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

	// Agregar hojas adicionales
	es.addSummarySheet(f, data)
	es.addStocksSheet(f, data)
	es.addIndicesSheet(f, data)

	// Eliminar la hoja por defecto "Sheet1" si existe
	err = f.DeleteSheet("Sheet1")
	if err != nil {
		// Si no se puede eliminar, no es crítico, continuar
	}

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

	// Agregar información sobre límites de API si hay pocos índices
	if indices == 0 {
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row+1), "⚠️ ADVERTENCIA:")
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row+2), "No se obtuvieron datos de índices.")
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row+3), "Posible causa: Límite de API alcanzado (25 llamadas/día)")
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row+4), "Los límites se renuevan cada 24 horas.")
	}

	// Ajustar ancho de columna
	f.SetColWidth(summarySheet, "A", "A", 50)
}

// addStocksSheet agrega una hoja específica para stocks
func (es *ExcelService) addStocksSheet(f *excelize.File, data []models.StockData) {
	stocksSheet := "Stocks"
	_, err := f.NewSheet(stocksSheet)
	if err != nil {
		return
	}

	// Filtrar solo stocks
	var stocks []models.StockData
	for _, item := range data {
		if es.determineType(item.Symbol) == "Stock" {
			stocks = append(stocks, item)
		}
	}

	if len(stocks) == 0 {
		f.SetCellValue(stocksSheet, "A1", "No hay datos de stocks disponibles")
		return
	}

	// Configurar encabezados
	headers := []string{
		"Símbolo", "Fecha", "Apertura", "Máximo",
		"Mínimo", "Cierre", "Volumen", "Fuente", "Estado",
	}

	// Escribir encabezados
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(stocksSheet, cell, header)
	}

	// Escribir datos de stocks
	for i, stock := range stocks {
		row := i + 2
		f.SetCellValue(stocksSheet, fmt.Sprintf("A%d", row), stock.Symbol)
		f.SetCellValue(stocksSheet, fmt.Sprintf("B%d", row), stock.Date)
		f.SetCellValue(stocksSheet, fmt.Sprintf("C%d", row), stock.Open)
		f.SetCellValue(stocksSheet, fmt.Sprintf("D%d", row), stock.High)
		f.SetCellValue(stocksSheet, fmt.Sprintf("E%d", row), stock.Low)
		f.SetCellValue(stocksSheet, fmt.Sprintf("F%d", row), stock.Close)
		f.SetCellValue(stocksSheet, fmt.Sprintf("G%d", row), stock.Volume)
		f.SetCellValue(stocksSheet, fmt.Sprintf("H%d", row), stock.From)
		f.SetCellValue(stocksSheet, fmt.Sprintf("I%d", row), stock.Status)
	}

	// Ajustar ancho de columnas
	columns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	widths := []float64{12, 12, 12, 12, 12, 12, 15, 15, 10}

	for i, col := range columns {
		f.SetColWidth(stocksSheet, col, col, widths[i])
	}
}

// addIndicesSheet agrega una hoja específica para índices
func (es *ExcelService) addIndicesSheet(f *excelize.File, data []models.StockData) {
	indicesSheet := "Indices"
	_, err := f.NewSheet(indicesSheet)
	if err != nil {
		return
	}

	// Filtrar solo índices
	var indices []models.StockData
	for _, item := range data {
		if es.determineType(item.Symbol) == "Índice" {
			indices = append(indices, item)
		}
	}

	if len(indices) == 0 {
		f.SetCellValue(indicesSheet, "A1", "⚠️ DATOS DE ÍNDICES NO DISPONIBLES")
		f.SetCellValue(indicesSheet, "A3", "Posibles causas:")
		f.SetCellValue(indicesSheet, "A4", "• Límite de API alcanzado (Alpha Vantage: 25 llamadas/día)")
		f.SetCellValue(indicesSheet, "A5", "• Error de conectividad con el proveedor de datos")
		f.SetCellValue(indicesSheet, "A6", "• Símbolos de índices no soportados por la API actual")
		f.SetCellValue(indicesSheet, "A8", "Solución recomendada:")
		f.SetCellValue(indicesSheet, "A9", "• Esperar hasta mañana para que se renueven los límites de API")
		f.SetCellValue(indicesSheet, "A10", "• Considerar actualizar a un plan premium de la API")
		f.SetCellValue(indicesSheet, "A11", "• Usar solo datos de stocks disponibles")

		// Ajustar ancho de columna para el mensaje
		f.SetColWidth(indicesSheet, "A", "A", 60)
		return
	}

	// Configurar encabezados
	headers := []string{
		"Símbolo", "Fecha", "Apertura", "Máximo",
		"Mínimo", "Cierre", "Volumen", "Fuente", "Estado",
	}

	// Escribir encabezados
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(indicesSheet, cell, header)
	}

	// Escribir datos de índices
	for i, index := range indices {
		row := i + 2
		f.SetCellValue(indicesSheet, fmt.Sprintf("A%d", row), index.Symbol)
		f.SetCellValue(indicesSheet, fmt.Sprintf("B%d", row), index.Date)
		f.SetCellValue(indicesSheet, fmt.Sprintf("C%d", row), index.Open)
		f.SetCellValue(indicesSheet, fmt.Sprintf("D%d", row), index.High)
		f.SetCellValue(indicesSheet, fmt.Sprintf("E%d", row), index.Low)
		f.SetCellValue(indicesSheet, fmt.Sprintf("F%d", row), index.Close)
		f.SetCellValue(indicesSheet, fmt.Sprintf("G%d", row), index.Volume)
		f.SetCellValue(indicesSheet, fmt.Sprintf("H%d", row), index.From)
		f.SetCellValue(indicesSheet, fmt.Sprintf("I%d", row), index.Status)
	}

	// Ajustar ancho de columnas
	columns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	widths := []float64{12, 12, 12, 12, 12, 12, 15, 15, 10}

	for i, col := range columns {
		f.SetColWidth(indicesSheet, col, col, widths[i])
	}
}
