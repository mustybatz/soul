package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type inputFile struct {
	filepath  string
	separator string
	pretty    bool
}

func getFileData() (inputFile, error) {
	// We need to validate that we are getting the correct number of arguments
	if len(os.Args) < 2 {
		return inputFile{}, errors.New("A filepath argument is required.")
	}

	// Defining option flags- For this, we are using the flag package from the standard lib
	// We need to define three arguments: the flag's name, the default value and a short description (displayed with the option --help)
	separator := flag.String("separator", "comma", "Column separator")
	pretty := flag.Bool("pretty", false, "Generate pretty JSON")

	flag.Parse() // This will parse all the arguments from the terminal

	fileLocation := flag.Arg(0) // The only argument (that is not a flag option) is the file location (CSV file)

	// Validating Whether or not we received "commas" or "semicolon" from the parsed arguments.
	// If we didn't receive any of those, we should return an error.

	if !(*separator == "comma" || *separator == "semicolon") {
		return inputFile{}, errors.New("Only comma or semicolon separators are allowed")
	}

	// If we got to this endpoint, our program arguments are validated.
	// We return the corresponding struct instance with all the required data

	return inputFile{fileLocation, *separator, *pretty}, nil
}

func checkIfValidFile(filename string) (bool, error) {
	// Checking if entered file is CSV by using the filepath package from standard lib
	if fileExtension := filepath.Ext(filename); fileExtension != ".csv" {
		return false, fmt.Errorf("File %s is not CSV", filename)
	}

	// Checking if filepath entered belongs to an existing file, We use the stat method from the
	// standard lib

	if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
		return false, fmt.Errorf("File %s does not exist", filename)
	}

	// If we got into this point it means that the file is valid
	return true, nil
}

func exitGracefully(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func check(e error) {
	if e != nil {
		exitGracefully(e)
	}
}

func processLine(headers []string, dataList []string) (map[string]string, error) {
	// Validating if we're getting the same number of headers and columns, Otherwise
	// we return an error.

	if len(headers) != len(dataList) {
		return nil, errors.New("Line doesn't match headers format. Skipping")
	}

	// Creating the map we're going to populate
	recordMap := make(map[string]string)
	// For each header we're going to set a new map key with the corresponding column value
	for i, name := range headers {
		recordMap[name] = dataList[i]
	}

	// Returning the generated map
	return recordMap, nil
}

func processCsvFile(fileData inputFile, writerChannel chan<- map[string]string) {
	// Opening our file for reading
	file, err := os.Open(fileData.filepath)
	// Checking for errors, we shouldn't get any.
	check(err)
	// Don't forget to close the file once everything is done.
	defer file.Close()

	// Defining a "headers" and "line" slice
	var headers, line []string
	// Initializing our CSV reader
	reader := csv.NewReader(file)
	// The default character separator is coma (,) so we need to change to
	// semicolon (;) if we get that option from the terminal
	if fileData.separator == "semicolon" {
		reader.Comma = ';'
	}

	// Reding the first line, where we will find our headers.
	headers, err = reader.Read()
	check(err) // again, error checking

	// Now we're going to iterate over each line from the CSV file
	for {

		// We read one row (line) from the CSV.
		// This line is a string slice, whith each element representing a column.
		line, err = reader.Read()

		// If we get to EOF, we close the channel and break the for-loop
		if err == io.EOF {
			close(writerChannel)
			break
		} else if err != nil {
			exitGracefully(err) // If this happens, we got an unexpected error
		}

		// Processing a CSV line
		record, err := processLine(headers, line)

		if err != nil { // If we get an error here, it means we got a wrong number of columns, so we skip this line
			fmt.Printf("Line: %sError: %s\n", line, err)
			continue
		}

		// Otherwise, we send the processed record to the writer channel
		writerChannel <- record
	}

}

func getJSONFunc(pretty bool) (func(map[string]string) string, string) {
	// Declaring the variables we're going to return at the end
	var jsonFunc func(map[string]string) string
	var breakLine string
	if pretty { //Pretty is enabled, so we should return a well-formatted JSON file (multi-line)
		breakLine = "\n"
		jsonFunc = func(record map[string]string) string {
			jsonData, _ := json.MarshalIndent(record, "   ", "   ") // By doing this we're ensuring the JSON generated is indented and multi-line
			return "   " + string(jsonData)                         // Transforming from binary data to string and adding the indent characets to the front
		}
	} else { // Now pretty is disabled so we should return a compact JSON file (one single line)
		breakLine = "" // It's an empty string because we never break lines when adding a new JSON object
		jsonFunc = func(record map[string]string) string {
			jsonData, _ := json.Marshal(record) // Now we're using the standard Marshal function, which generates JSON without formating
			return string(jsonData)             // Transforming from binary data to string
		}
	}

	return jsonFunc, breakLine // Returning everythinbg
}

func writeJSONFile(csvPath string, writerChannel <-chan map[string]string, done chan<- bool, pretty bool) {
	writeString := createStringWriter(csvPath)
	jsonFunc, breakLine := getJSONFunc(pretty)

	fmt.Println("Writing JSON file...")

	writeString("["+breakLine, false)
	first := true
	for {
		record, more := <-writerChannel
		if more {
			if !first {
				writeString(","+breakLine, false)
			} else {
				first = false
			}

			jsonData := jsonFunc(record)
			writeString(jsonData, false)
		} else {
			writeString(breakLine+"]", true)
			fmt.Println("Completed!")
			done <- true
			break
		}
	}
}

func createStringWriter(csvPath string) func(string, bool) {
	jsonDir := filepath.Dir(csvPath)                                                       // Getting the directory where the CSV file is
	jsonName := fmt.Sprintf("%s.json", strings.TrimSuffix(filepath.Base(csvPath), ".csv")) // Declaring the JSON filename, using the CSV file name as base
	finalLocation := filepath.Join(jsonDir, jsonName)                                      // Declaring the JSON file location, using the previous variables as base.

	// Opening the JSON file what we want to start writing
	f, err := os.Create(finalLocation)
	check(err)

	// This is the function we want to return, we're going to use it to write the JSON file
	return func(data string, close bool) { // 2 arguments: The piece of text we want to write, and whether or not we should close the file.
		_, err := f.WriteString(data)
		check(err)

		// If close is "true", it means there are isn't more data to be written.
		if close {
			f.Close()
		}
	}

}

func main() {

	// Showing usefull info when the user enters the --help option
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <csvFile>\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	// Getting the file data that was entered by the user
	fileData, err := getFileData()

	if err != nil {
		exitGracefully(err)
	}

	// Validating the file entered
	if _, err := checkIfValidFile(fileData.filepath); err != nil {
		exitGracefully(err)
	}

	// Declaring the channels that our go-routines are going to use
	writerChannel := make(chan map[string]string)
	done := make(chan bool)

	// Running both of our go-routines, the first one is responsible for reading and the other for writing
	go processCsvFile(fileData, writerChannel)
	go writeJSONFile(fileData.filepath, writerChannel, done, fileData.pretty)

	// Waiting for the done channel to receive a value, so we can terminate program execution
	<-done

}
