/*
Purpose: This tool will create tags against PANOS.
*/
package main

import (
	"bufio"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const envFile = ".env"
const tagXPath = "type=config&action=get&xpath=/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/tag"

type Tags struct {
	XMLName xml.Name `xml:"response"`
	Results Result   `xml:"result"`
	TagList []Tag    `xml:"tag"`
}

type Tag struct {
	Key string `xml:"entry name,attr"`
	Value string `xml:"`
}

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
	envBytes, err := ioutil.ReadFile(base + "/" + envFile)
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

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %v <filename>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	// Get env file data
	envBytes := readEnvFile()

	// Get API Key
	apiKey := getAPIKey(envBytes)
	fmt.Println(apiKey)

	// Get list of PANs
	pFWs := getPANs(envBytes)
	// Get tags from provided file
	tags := getTags(os.Args[1])
	for _, i := range tags {
		fmt.Println(i)
	}

	//------------------
	url := fmt.Sprintf("https://%v/api/?key=%v&%v", pFWs[0], apiKey, tagXPath)
	fmt.Println(url)
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(b))
	fmt.Println(resp.Status, "-", resp.StatusCode)

	var t Tags
	err = xml.Unmarshal(b, &t)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(t)

}
