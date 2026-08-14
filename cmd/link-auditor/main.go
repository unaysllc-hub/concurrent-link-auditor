package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/unaysllc-hub/concurrent-link-auditor/auditor"
)

func main() {
	workers := flag.Int("workers", 8, "number of concurrent requests")
	timeout := flag.Duration("timeout", 10*time.Second, "timeout for each request")
	jsonOutput := flag.Bool("json", false, "print a JSON report")
	input := flag.String("input", "", "optional file containing one URL per line")
	flag.Parse()
	urls := append([]string{}, flag.Args()...)
	if *input != "" {
		file, err := os.Open(*input)
		if err != nil {
			fatal(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if value := strings.TrimSpace(scanner.Text()); value != "" && !strings.HasPrefix(value, "#") {
				urls = append(urls, value)
			}
		}
		if err := scanner.Err(); err != nil {
			fatal(err)
		}
	}
	if len(urls) == 0 {
		fatal(fmt.Errorf("provide URLs as arguments or with -input"))
	}
	client := &http.Client{Timeout: *timeout}
	results := auditor.Check(context.Background(), client, urls, *workers)
	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(results); err != nil {
			fatal(err)
		}
	} else {
		for _, result := range results {
			fmt.Println(result.String())
		}
	}
	for _, result := range results {
		if !result.Healthy() {
			os.Exit(1)
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "link-auditor:", err)
	os.Exit(2)
}
