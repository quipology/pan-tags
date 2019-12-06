/*
Purpose: This tool will create tags against PANOS.
*/
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

const envFile = ".env"

func getAPIKey() string {
	// Get absoulute path of program
	abs, err := filepath.Abs(os.Args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Get base directory of program
	base := filepath.Dir(abs)
	// Read env file
	envBytes, err := ioutil.ReadFile(base + "/" + envFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Parse env data for API key
	re := regexp.MustCompile(`API_KEY=([^\n\r\s]+)`)
	reResults := re.FindStringSubmatch(string(envBytes))
	if len(reResults) != 2 {
		fmt.Println("API key not found, exiting..")
		os.Exit(1)
	}
	return re.FindStringSubmatch(string(envBytes))[1]
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %v <filename>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	// Source file
	// tagFile := os.Args[1]

	// Get API Key
	apiKey := getAPIKey()
	fmt.Println(apiKey)

}
