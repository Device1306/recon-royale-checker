package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

type DNSResponse struct {
	Status int `json:"Status"`
	Answer []struct {
		Data string `json:"data"`
	} `json:"Answer"`
}

var recordTypes = []string{"A", "AAAA", "TXT", "CNAME", "NS", "MX"}

func main() {
	// Define the flags for input file (-f) and output file (-o)
	filePath := flag.String("f", "", "Path to the file containing domain list")
	outputPath := flag.String("o", "", "Path to the output file (optional)")
	flag.Parse()

	if *filePath == "" {
		fmt.Println("Please provide a file with the -f flag.")
		os.Exit(1)
	}

	// Read domains from the file
	domains, err := readDomainsFromFile(*filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	results := make(chan string, len(domains))

	// Loop through each domain and query concurrently
	for _, domain := range domains {
		wg.Add(1)
		go func(domain string) {
			defer wg.Done()
			found := checkDomain(domain)
			if found {
				results <- domain
			}
		}(domain)
	}

	// Close the channel once all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var output []string
	for res := range results {
		output = append(output, res)
	}

	// Write results to the output file or print them to the console
	if *outputPath != "" {
		err := writeResultsToFile(*outputPath, output)
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
		} else {
			fmt.Printf("Results written to %s\n", *outputPath)
		}
	} else {
		// Print to console if no output file is specified
		for _, res := range output {
			fmt.Println(res)
		}
	}
}

// readDomainsFromFile reads domain names from the specified file
func readDomainsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var domains []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			domains = append(domains, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return domains, nil
}

// checkDomain checks a single domain for each DNS record type
func checkDomain(domain string) bool {
	for _, recordType := range recordTypes {
		record := queryDNS(domain, recordType)
		if record != "" {
			return true
		}
	}
	return false
}

// queryDNS queries Cloudflare's DNS over HTTPS API for a specific domain and record type
func queryDNS(domain, recordType string) string {
	url := fmt.Sprintf("https://cloudflare-dns.com/dns-query?name=%s&type=%s", domain, recordType)

	// Create a new HTTP request with the correct headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/dns-json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	// Parse the JSON response
	var dnsResp DNSResponse
	err = json.Unmarshal(body, &dnsResp)
	if err != nil || dnsResp.Status != 0 || len(dnsResp.Answer) == 0 {
		return ""
	}

	// Return the first record found
	return dnsResp.Answer[0].Data
}

// writeResultsToFile writes the results to the specified output file
func writeResultsToFile(filePath string, results []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, result := range results {
		_, err := writer.WriteString(result + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
