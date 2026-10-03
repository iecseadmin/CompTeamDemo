package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

type Member struct {
	Name               string  `json:"name"`
	LeetcodeUsername   *string `json:"leetcodeUsername"`
	CodeforcesUsername *string `json:"codeforcesUsername"`
	CodechefUsername   *string `json:"codechefUsername"`
	AtcoderUsername    *string `json:"atcoderUsername"`
	MemberStatus       string  `json:"memberStatus"`
}

func main() {
	inputFile := flag.String("input", "", "Path to input CSV file")
	outputFile := flag.String("output", "", "Path to output JSON file")
	flag.Parse()

	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Usage: go run main.go -input <input.csv> -output <output.json>")
		os.Exit(1)
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV: %v\n", err)
		os.Exit(1)
	}

	if len(records) == 0 {
		fmt.Println("CSV file is empty")
		os.Exit(1)
	}

	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.TrimSpace(h)] = i
	}

	requiredHeaders := []string{"Name", "LeetCode Username", "CodeChef Username", "Codeforces Username", "AtCoder Username"}
	for _, req := range requiredHeaders {
		if _, ok := headerMap[req]; !ok {
			fmt.Printf("Missing required CSV header: %s\n", req)
			os.Exit(1)
		}
	}

	var members []Member
	seenNames := make(map[string]bool)
	allowedStatuses := map[string]bool{
		"Junior Team":       true,
		"Senior Team":       true,
		"Retired":           true,
		"Board Team Member": true,
	}

	for i, record := range records[1:] {
		name := strings.TrimSpace(record[headerMap["Name"]])
		if name == "" {
			fmt.Printf("Row %d: Name cannot be empty\n", i+2)
			os.Exit(1)
		}
		if seenNames[name] {
			fmt.Printf("Row %d: Duplicate member name found: %s\n", i+2, name)
			os.Exit(1)
		}
		seenNames[name] = true

		status := "Junior Team"
		if idx, ok := headerMap["Member Status"]; ok && idx < len(record) {
			s := strings.TrimSpace(record[idx])
			if s != "" {
				if !allowedStatuses[s] {
					fmt.Printf("Row %d: Invalid Member Status: '%s'. Allowed values are Junior Team, Senior Team, Retired, Board Team Member.\n", i+2, s)
					os.Exit(1)
				}
				status = s
			}
		}

		members = append(members, Member{
			Name:               name,
			LeetcodeUsername:   parseUsername(record, headerMap, "LeetCode Username"),
			CodeforcesUsername: parseUsername(record, headerMap, "Codeforces Username"),
			CodechefUsername:   parseUsername(record, headerMap, "CodeChef Username"),
			AtcoderUsername:    parseUsername(record, headerMap, "AtCoder Username"),
			MemberStatus:       status,
		})
	}

	outData, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		fmt.Printf("Error generating JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputFile, outData, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted %d records to %s\n", len(members), *outputFile)
}

func parseUsername(record []string, headerMap map[string]int, key string) *string {
	if idx, ok := headerMap[key]; ok && idx < len(record) {
		val := strings.TrimSpace(record[idx])
		if val != "" {
			return &val
		}
	}
	return nil
}
