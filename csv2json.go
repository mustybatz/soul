package main

import (
	"errors"
	"flag"
	"os"
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

func main() {
	getFileData()
}
