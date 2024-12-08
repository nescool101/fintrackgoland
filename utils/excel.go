package utils

import (
	"bytes"
	"github.com/nescool101/fintrackgoland/models"
	"github.com/xuri/excelize/v2"
	"log"
	"strconv"
)

func GenerateExcel(data []models.StockData) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Weekly Data"
	index, _ := f.NewSheet(sheet)

	headers := []string{"Status", "From", "Symbol", "Date", "Open", "High", "Low", "Close", "Volume", "AfterHours", "PreMarket"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}

	for row, record := range data {
		r := row + 2 // Starting from row 2
		f.SetCellValue(sheet, "A"+strconv.Itoa(r), record.Status)
		f.SetCellValue(sheet, "B"+strconv.Itoa(r), record.From)
		f.SetCellValue(sheet, "C"+strconv.Itoa(r), record.Symbol)
		f.SetCellValue(sheet, "D"+strconv.Itoa(r), record.Date)
		f.SetCellValue(sheet, "E"+strconv.Itoa(r), record.Open)
		f.SetCellValue(sheet, "F"+strconv.Itoa(r), record.High)
		f.SetCellValue(sheet, "G"+strconv.Itoa(r), record.Low)
		f.SetCellValue(sheet, "H"+strconv.Itoa(r), record.Close)
		f.SetCellValue(sheet, "I"+strconv.Itoa(r), record.Volume)
		f.SetCellValue(sheet, "J"+strconv.Itoa(r), record.AfterHours)
		f.SetCellValue(sheet, "K"+strconv.Itoa(r), record.PreMarket)
	}

	f.SetActiveSheet(index)

	var fileBuffer bytes.Buffer
	if err := f.Write(&fileBuffer); err != nil {
		log.Printf("Error generating Excel file: %v", err)
		return nil, err
	}

	return fileBuffer.Bytes(), nil
}
