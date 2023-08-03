package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"log"

	"github.com/tealeg/xlsx"
)

type Row struct {
	Date         string
	DayOfWeek    string
	Preferential bool
	Available    bool
	Impossible   bool
}

func main() {
	directoryPath := "files"
	csvFilesDirectoryPath := "csvFiles"

	err := convertXLSXToCSV(directoryPath, csvFilesDirectoryPath)
	if err != nil {
		fmt.Printf("Error converting XLSX to CSV: %s\n", err)
	}

	fileRowsMap, err := processDirectory(csvFilesDirectoryPath)
	if err != nil {
		panic(err)
	}

	mergedData := mergeData(fileRowsMap)

	// Write the merged data to a CSV file
	err = writeMergedDataToCSV(mergedData, "mergedData.csv")
	if err != nil {
		panic(err)
	}

	// Print the merged data
	for _, row := range mergedData {
		fmt.Println(row)
	}
}

func processDirectory(directoryPath string) (map[string][]Row, error) {
	fileRowsMap := make(map[string][]Row)

	// Get all file paths within the directory
	filePaths, err := getAllFilePaths(directoryPath)
	if err != nil {
		return nil, err
	}

	// Process each file
	for _, filePath := range filePaths {
		fileName := getFileNameWithoutExtension(filePath)

		// Process the CSV file
		rows, err := processCSV(filePath)
		if err != nil {
			return nil, err
		}

		// Store the rows in the map
		fileRowsMap[fileName] = rows
	}

	return fileRowsMap, nil
}

func getAllFilePaths(directoryPath string) ([]string, error) {
	var filePaths []string

	// Read the directory
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		return nil, err
	}

	// Get the file paths
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(directoryPath, file.Name())
			filePaths = append(filePaths, filePath)
		}
	}

	return filePaths, nil
}

func processCSV(filePath string) ([]Row, error) {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all the CSV records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var rows []Row

	// Process each record
	for _, record := range records {
		// Extract the relevant columns
		date := record[0]
		dayOfWeek := record[1]
		preferential := convertToBool(record[2])
		available := convertToBool(record[3])
		impossible := convertToBool(record[4])

		// Create a new Row object
		row := Row{
			Date:         date,
			DayOfWeek:    dayOfWeek,
			Preferential: preferential,
			Available:    available,
			Impossible:   impossible,
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func convertToBool(value string) bool {
	if strings.ToLower(value) == "x" {
		return true
	}
	return false
}

func processRow(row Row) {
	// Process the row data
	fmt.Printf("Date: %s, Day of Week: %s\n", row.Date, row.DayOfWeek)
	fmt.Printf("Preferential: %v, Available: %v, Impossible: %v\n", row.Preferential, row.Available, row.Impossible)
	// Add your logic here to work with the extracted values
}

func getFileNameWithoutExtension(filePath string) string {
	fileName := filepath.Base(filePath)
	extension := filepath.Ext(fileName)
	return strings.TrimSuffix(fileName, extension)
}

func mergeData(fileRowsMap map[string][]Row) [][]string {
	mergedData := [][]string{}

	// Create a map to track the unique dates
	uniqueDates := make(map[string]bool)

	firstRow := true

	// Iterate over the fileRowsMap
	for fileName, rows := range fileRowsMap {
		// Iterate over the rows of each file
		for _, row := range rows {
			// Check if the date is already in the uniqueDates map
			if _, ok := uniqueDates[row.Date]; !ok {
				// Add the date to the uniqueDates map
				uniqueDates[row.Date] = true

				if firstRow {
					// Add the header row to the mergedData
					mergedData = append(mergedData, []string{"Date", "Day of Week", "Preferential", "Available", "Impossible"})
					firstRow = false
					continue
				}

				// Create a new row for the merged data
				mergedRow := []string{
					row.Date,
					row.DayOfWeek,
					"",
					"",
					"",
				}

				// Update the merged row based on the row data
				if row.Preferential {
					mergedRow[2] = fileName
				}
				if row.Available {
					mergedRow[3] = fileName
				}
				if row.Impossible {
					mergedRow[4] = fileName
				}

				// Add the merged row to the mergedData
				mergedData = append(mergedData, mergedRow)
			} else {
				// Find the existing row in the mergedData for the date
				for i := range mergedData {
					if mergedData[i][0] == row.Date {
						// Update the existing row based on the row data
						if row.Preferential {
							if mergedData[i][2] != "" {
								mergedData[i][2] += " | " + fileName
							} else {
								mergedData[i][2] = fileName
							}
						}
						if row.Available {
							if mergedData[i][3] != "" {
								mergedData[i][3] += " | " + fileName
							} else {
								mergedData[i][3] = fileName
							}
						}
						if row.Impossible {
							if mergedData[i][4] != "" {
								mergedData[i][4] += " | " + fileName
							} else {
								mergedData[i][4] = fileName
							}
						}
						break
					}
				}
			}
		}
	}

	return mergedData
}

func writeMergedDataToCSV(mergedData [][]string, filePath string) error {
	// Create a new CSV file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write each row of the merged data to the CSV file
	for _, row := range mergedData {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func convertXLSXToCSV(inputDir, outputDir string) error {
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".xlsx") {
			xlsxFile, err := xlsx.OpenFile(filepath.Join(inputDir, file.Name()))
			if err != nil {
				log.Printf("Error opening XLSX file: %s\n", err)
				continue
			}

			for _, sheet := range xlsxFile.Sheets {
				csvFileName := strings.TrimSuffix(file.Name(), ".xlsx") + "_" + sheet.Name + ".csv"
				csvFile, err := os.Create(filepath.Join(outputDir, csvFileName))
				if err != nil {
					log.Printf("Error creating CSV file: %s\n", err)
					continue
				}

				for _, row := range sheet.Rows {
					var csvRow []string
					for _, cell := range row.Cells {
						csvRow = append(csvRow, cell.String())
					}
					_, err := csvFile.WriteString(strings.Join(csvRow, ",") + "\n")
					if err != nil {
						log.Printf("Error writing to CSV file: %s\n", err)
						continue
					}
				}

				csvFile.Close()
			}
		}
	}
