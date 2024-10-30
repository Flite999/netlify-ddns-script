package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AccessToken string
	ZoneId      string
	Hostname    string
}

type DNSRecord struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

func loadConfig() (Config, error) {
	var config Config
	err := godotenv.Load(".env")
	if err != nil {
		return config, err
	}
	config.AccessToken = os.Getenv("ACCESS_TOKEN")
	config.ZoneId = os.Getenv("ZONE_ID")
	config.Hostname = os.Getenv("HOSTNAME")
	return config, nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println("Access Token:", config.AccessToken)
	fmt.Println("DNS Records:", config.ZoneId)

	var prevIp string

	for {
		ipAdd := checkIP()
		if ipAdd != prevIp {
			// Update DNS record - delete previous record and create new record since there is no update option
			records := getDNSRecords(config, prevIp)
			deleteDNSRecord(config, records)
			createDNSRecord(config, ipAdd)

			fmt.Println("DNS record updated successfully")
			prevIp = ipAdd
		}

		// Wait for 12 hours
		time.Sleep(12 * time.Hour)
	}
}

func checkIP() string {
	fmt.Println("Function executed at:", time.Now())

	cmd := exec.Command("curl", "ipinfo.io/ip")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing curl command:", err)
		return ""
	}
	fmt.Println("Current IP address:", string(output))

	return string(output)
}

func getDNSRecords(config Config, prevIP string) []string {
	// get all DNS records for zone id
	url := fmt.Sprintf("https://api.netlify.com/api/v1/dns_zones/%s/dns_records", config.ZoneId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to execute HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read HTTP response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to get DNS record: %s", body)
	}

	var records []DNSRecord
	err = json.Unmarshal(body, &records)
	if err != nil {
		log.Fatalf("Failed to parse DNS records: %v", err)
	}

	var matchingRecordIds []string
	for _, record := range records {
		if record.Value == prevIP {
			matchingRecordIds = append(matchingRecordIds, record.Id)
		}
	}

	return matchingRecordIds
}

func deleteDNSRecord(config Config, records []string) {
	if len(records) == 0 {
		fmt.Println("No DNS records to delete")
		return
	}
	for _, recordId := range records {
		// delete a DNS record for zone id
		url := fmt.Sprintf("https://api.netlify.com/api/v1/dns_zones/%s/dns_records/%s", config.ZoneId, recordId)
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			log.Fatalf("Failed to create HTTP request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+config.AccessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Failed to execute HTTP request: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read HTTP response: %v", err)
		}

		if resp.StatusCode != http.StatusNoContent {
			log.Fatalf("Failed to delete DNS record: %s", body)
		}

		fmt.Println("DNS record deleted successfully")
	}
}

func createDNSRecord(config Config, ipAdd string) {
	// create a new DNS record for zone id
	url := fmt.Sprintf("https://api.netlify.com/api/v1/dns_zones/%s/dns_records", config.ZoneId)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.AccessToken)
	//@ will automatically set hostname to zone id hostname
	reqBody := fmt.Sprintf(`{
		"type": "A",
		"hostname": "%s",
		"value": "%s",
		"ttl": 3600
	}`, config.Hostname, ipAdd)

	req.Body = io.NopCloser(strings.NewReader(reqBody))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to execute HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read HTTP response: %v", err)
	}
	fmt.Printf("Response Status Code: %v\n", resp.StatusCode)
	if resp.StatusCode != 201 {
		log.Fatalf("Failed to create DNS record: %s", body)
	}

	fmt.Println("DNS record created successfully")
}
