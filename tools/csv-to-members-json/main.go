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
	inputFile := "./tools/csv-to-members-json/team_usernames.csv"
	outputFile := flag.String("output", "src/data/members.json", "Path to output JSON file")
	flag.Parse()

	if inputFile == "" {
		fmt.Println("Input CSV path cannot be empty")
		os.Exit(1)
	}

	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
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

	// Accept both the spreadsheet's current "Status" heading and the older
	// "Member Status" heading.
	if _, ok := headerMap["Member Status"]; !ok {
		if idx, found := headerMap["Status"]; found {
			headerMap["Member Status"] = idx
		}
	}
	requiredHeaders := []string{"Name", "LeetCode Username", "CodeChef Username", "Codeforces Username", "AtCoder Username"}
	for _, req := range requiredHeaders {
		if _, ok := headerMap[req]; !ok {
			fmt.Printf("Missing required CSV header: %s\n", req)
			os.Exit(1)
		}
	}

	var members []Member
	if existing, err := os.ReadFile(*outputFile); err == nil {
		// The existing file may have a trailing comma from a previous manual
		// edit. Clean that common JSON formatting error before decoding it.
		if err := json.Unmarshal(stripTrailingCommas(existing), &members); err != nil {
			fmt.Printf("Error reading existing output JSON: %v\n", err)
			os.Exit(1)
		}
	} else if !os.IsNotExist(err) {
		fmt.Printf("Error opening output file: %v\n", err)
		os.Exit(1)
	}
	if members == nil {
		members = make([]Member, 0)
	}
	allowedStatuses := map[string]bool{
		"Junior Team":       true,
		"Senior Team":       true,
		"Retired":           true,
		"Board Team Member": true,
	}

	for i, record := range records[1:] {
		if len(record) != len(headers) {
			fmt.Printf("Row %d: expected %d columns, got %d\n", i+2, len(headers), len(record))
			os.Exit(1)
		}
		name := strings.TrimSpace(record[headerMap["Name"]])
		if name == "" {
			fmt.Printf("Row %d: Name cannot be empty\n", i+2)
			os.Exit(1)
		}
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

	fmt.Printf("Successfully appended %d records to %s (%d total)\n", len(records)-1, *outputFile, len(members))
}

func stripTrailingCommas(data []byte) []byte {
	cleaned := make([]byte, 0, len(data))
	inString := false
	escaped := false
	for i, b := range data {
		if inString {
			cleaned = append(cleaned, b)
			if escaped {
				escaped = false
			} else if b == '\\' {
				escaped = true
			} else if b == '"' {
				inString = false
			}
			continue
		}
		if b == '"' {
			inString = true
			cleaned = append(cleaned, b)
			continue
		}
		if b == ',' {
			j := i + 1
			for j < len(data) && (data[j] == ' ' || data[j] == '\n' || data[j] == '\r' || data[j] == '\t') {
				j++
			}
			if j < len(data) && (data[j] == '}' || data[j] == ']') {
				continue
			}
		}
		cleaned = append(cleaned, b)
	}
	return cleaned
}

func parseUsername(record []string, headerMap map[string]int, key string) *string {
	if idx, ok := headerMap[key]; ok && idx < len(record) {
		val := strings.TrimSpace(record[idx])
		if val != "" && val != "-" {
			return &val
		}
	}
	return nil
}
