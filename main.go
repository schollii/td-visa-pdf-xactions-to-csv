// "07cafa57-162c-4ce9-b8af-92ad47f01152.pdf") // Read local pdf file

/*
 * PDF to text: Extract all text for each page of a pdf file.
 *
 * Run as: go run pdf_extract_text.go input.pdf
 */

package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

func init() {
	// Make sure to load your metered License API key prior to using the library.
	// If you need a key, you can sign up and create a free one at https://cloud.unidoc.io
	err := license.SetMeteredKey(os.Getenv(`UNIDOC_LICENSE_API_KEY`))
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: go run pdf_extract_text.go input.pdf year\n")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	datesYearStr := os.Args[2]
	datesYear, err := strconv.Atoi(datesYearStr)
	if err != nil {
		fmt.Printf("Could not convert year to number: %v\n", datesYearStr)
	}

	table, err := getTransactions(inputPath, datesYear)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Number of rows found: %v\n", len(table))
	if err := saveTableToCSV(table, inputPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(2)
	}
}

// outputPdfText prints out contents of PDF file to stdout.
func getTransactions(inputPath string, datesYear int) ([][]string, error) {
	fmt.Printf("Processing %v\n", inputPath)

	f, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return nil, err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, err
	}
	// TD encrypts its PDFs but without a password
	if isEncrypted {
		password := []byte{}
		success, err := pdfReader.Decrypt(password)
		if err != nil {
			return nil, err
		}
		if !success {
			return nil, fmt.Errorf("could not decrypt PDF file")
		}
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, err
	}

	var startMonth string
	firstRow := true

	table := [][]string{}
	table = append(table, []string{"Date", "Description", "Amount"})
	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return nil, err
		}

		ex, err := extractor.New(page)
		if err != nil {
			return nil, err
		}

		text, err := ex.ExtractText()
		if err != nil {
			return nil, err
		}

		// Separate the rows and columns of the table.
		rows := strings.Split(text, "\n")
		const RE_DATE = `[A-Z]{3} +[0-9]+`
		const RE_ROW_START = RE_DATE + ` +` + RE_DATE + ` `
		const RE_AMOUNT_CAD = `-?\$[0-9,.]+`
		//const RE_AMOUNT_USD = `-?[0-9,.]+ USD`
		//const RE_AMOUNT = RE_AMOUNT_CAD + `|` + RE_AMOUNT_USD
		RE_COLS := fmt.Sprintf(`(%v) +%v +(.*) +(%v)`, RE_DATE, RE_DATE, RE_AMOUNT_CAD)

		for i, row := range rows {
			if !regexp.MustCompile(RE_ROW_START).MatchString(row) {
				continue
			}

			// we have a row

			// if a row does not end with a dollar amount then there are more rows to concatenate
			for j := 1; !regexp.MustCompile(RE_AMOUNT_CAD + ` *$`).MatchString(row); j++ {
				row += " " + rows[i+j]
			}

			// extract columns from the row
			cols := regexp.MustCompile(RE_COLS).FindStringSubmatch(row)
			if len(cols) < 4 {
				fmt.Printf("ERROR: could not parse the row: '%v'\n", row)
				continue
			}

			// add the year to the date (note col[0] is the full row)
			yearOffset := 0
			currentMonth := strings.SplitN(cols[1], " ", 2)[0]
			if firstRow {
				startMonth = currentMonth
				firstRow = false
			} else {
				// it's possible that row is for next year, if date is JAN
				if startMonth != currentMonth && startMonth == "DEC" {
					yearOffset = 1
				}
			}
			currentYear := datesYear + yearOffset
			cols[1] = fmt.Sprintf("%v,%v", cols[1], currentYear)

			// save it
			table = append(table, cols[1:])
		}
	}

	return table, nil
}

func saveTableToCSV(table [][]string, inputFilePath string) error {
	//outputDir, inputFilename := filepath.Split(inputFilePath)
	outputFilePath := strings.TrimSuffix(inputFilePath, filepath.Ext(inputFilePath)) + ".csv"
	file, err := os.Create(outputFilePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.WriteAll(table); err != nil {
		return err
	}

	fmt.Printf("Saved to file %v\n", outputFilePath)
	return nil
}
