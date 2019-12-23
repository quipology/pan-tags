/*
Purpose: This tool will create tags against PANOS.
*/
package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const envFile = ".env" // Environment file

var (
	wg     sync.WaitGroup
	apiKey string // For storing API key
)

func readEnvFile() []byte {
	// Get absoulute path of program
	abs, err := filepath.Abs(os.Args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Get base directory of program
	base := filepath.Dir(abs)
	// Read env file
	envBytes, err := ioutil.ReadFile(filepath.Join(base, envFile))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return envBytes
}

func getAPIKey(e []byte) string {
	// Parse env data for API key
	re := regexp.MustCompile(`API_KEY=([^\s]+)`)
	reResults := re.FindStringSubmatch(string(e))
	// Check size of re slice to ensure a match for group 1 was found
	if len(reResults) != 2 {
		fmt.Println("API key not found, exiting..")
		os.Exit(1)
	}
	return re.FindStringSubmatch(string(e))[1]
}

func createClient() *http.Client {
	// Create HTTP transport
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	// Create HTTP client
	return &http.Client{Transport: tr}
}

func getTags(fn string) []string {
	f, err := os.Open(fn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	// Create scanner
	s := bufio.NewScanner(f)
	// Store slice of tags in t
	var t []string
	// Append each tag found to t
	for s.Scan() {
		t = append(t, s.Text())
	}
	return t
}

func getPANs(e []byte) []string {
	re := regexp.MustCompile(`PAN=([^\s]+)`)
	reResults := re.FindStringSubmatch(string(e))
	if len(reResults) != 2 {
		fmt.Println("No PAN devices found in .env file, exiting..")
		os.Exit(1)
	}
	return strings.Split(reResults[1], ",")
}

func createTag(tag, pan string, c *http.Client) bool {
	tagSetXPath := "type=config&action=set&xpath=/config/shared/tag"
	// Generate encoded query string
	encodedQuery := url.QueryEscape(fmt.Sprintf("<entry name='%v'/>", tag))
	encodedQuery = strings.Replace(encodedQuery, "%26", "%26amp;", -1)
	// URL
	url := fmt.Sprintf("https://%v/api/?key=%v&%v&element=%v", pan, apiKey, tagSetXPath, encodedQuery)
	// Generate GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Execute request
	fmt.Printf("Attempting to create tag '%v' on %v..\n", tag, pan)
	counter := 1
	for counter <= 3 { // # of retries before confirming no connectivity
		resp, err := c.Do(req)
		if err != nil {
			counter++
			if counter > 3 {
				counter = 3
				break
			}
			fmt.Printf("Unable to connect to %v, retrying..(attempt #%v)\n", pan, counter)
			continue
		}
		// Read reponse
		if resp.StatusCode == 200 {
			fmt.Println(fmt.Sprintf("Tag:'%v' was successfully created.", tag))
			return true
		}
		fmt.Println(fmt.Sprintf("Something went wrong when attempting to add tag:'%v'", tag))
		return true
	}
	fmt.Printf("Unable to connect to: %v after %v attempts.\n", pan, counter)
	return false
}

func commitChanges(pan string, c *http.Client) {
	commitXPath := "type=commit"
	// Generate encoded query string
	encodedQuery := url.QueryEscape("<commit><description>New tags added - saving changes.</description></commit>")
	encodedQuery = strings.Replace(encodedQuery, "%26", "%26amp;", -1)
	// URL
	url := fmt.Sprintf("https://%v/api/?key=%v&%v&cmd=%v", pan, apiKey, commitXPath, encodedQuery)
	// Generate GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Execute request
	fmt.Printf("Committing changes on %v..\n", pan)
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("Failed to commit changes on %v: %v\n", pan, err)
	}
	if resp.StatusCode == 200 {
		fmt.Printf("Changes committed successfully on %v\n", pan)
		return
	}
	fmt.Printf("Something went wrong when attempting to commit changes on:'%v'\n", pan)
}

func main() {
	// Check if tag file was provided
	if len(os.Args) != 2 {
		fmt.Printf("usage: %v <filename>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	// Get env file data
	envBytes := readEnvFile()

	// Get API Key
	apiKey = getAPIKey(envBytes)

	// Get list of PANs
	pFWs := getPANs(envBytes)

	// Get tags from provided file
	tags := getTags(os.Args[1])

	// Create HTTPS client
	client := createClient()

	for _, pan := range pFWs {
		wg.Add(1)
		go func(pan string) {
			defer wg.Done()
			count := 0 // For keeping count on how many tags were created
			for _, tag := range tags {
				ok := createTag(tag, pan, client)
				if !ok { // Return if creation of tag failed on PAN device
					break
				} else {
					count++ // Add to counter since tag was created successfully
				}
			}
			// Commit changes on current PAN device if at least 1 tag was successfully created
			if count > 0 {
				commitChanges(pan, client)
			}
		}(pan)

	}
	wg.Wait() // Wait for all PAN devices to finish
	fmt.Println("Tag processing completed, exiting..")
}
