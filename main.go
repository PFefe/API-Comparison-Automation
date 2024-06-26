package main

import (
	"encoding/json"
	"fmt"
	"github.com/nsf/jsondiff"
	"io"
	"net/http"
	"os"
	"time"
)

// Request represents the structure of a request
type Request struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

// Response represents the structure of a response
type Response struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

// RequestResponse represents the combined structure
type RequestResponse struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

// readAuthToken reads the authorization token from a file
func readAuthToken(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	var token struct {
		Authorization string `json:"authorization"`
	}
	err = json.Unmarshal(
		byteValue,
		&token,
	)
	if err != nil {
		return "", err
	}

	return token.Authorization, nil
}

// readRequestResponse reads the request and response from a JSON file
func readRequestResponse(filePath string) (*RequestResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var reqRes RequestResponse
	err = json.Unmarshal(
		byteValue,
		&reqRes,
	)
	if err != nil {
		return nil, err
	}

	return &reqRes, nil
}

// sendAPIRequest sends the API request and returns the response
func sendAPIRequest(req Request, authToken string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 30 * time.Second, // Increase the timeout duration
	}

	httpReq, err := http.NewRequest(
		req.Method,
		req.URL,
		nil,
	)
	if err != nil {
		return nil, err
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(
			key,
			value,
		)
	}

	httpReq.Header.Add(
		"authorization",
		authToken,
	)

	/*	// Log the headers before sending the request
		fmt.Printf(
			"Final Request Headers for URL %s:\n",
			req.URL,
		)*/
	/*	for key, values := range httpReq.Header {
			for _, value := range values {
				fmt.Printf(
					"%s: %s\n",
					key,
					value,
				)
			}
		}
	*/
	return client.Do(httpReq)
}

// compareResponses compares the original and new responses
func compareResponses(original, new json.RawMessage) (jsondiff.Difference, string) {
	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(
		original,
		new,
		&opts,
	)
	return diff, explanation
}

func main() {
	authTokenFilePath := "data/token.json" // Replace with the actual path to your auth token file
	authToken, err := readAuthToken(authTokenFilePath)
	if err != nil {
		fmt.Printf(
			"Error reading auth token: %s\n",
			err,
		)
		return
	}

	// Example file paths
	filePaths := []string{
		"requests/request_agents-breakdown-statistics.json",
		"requests/request_charts-leads.json",
		"requests/request_nsights-v1-credits.json",
		"requests/request_agents-insights-top-statistics?type=leads.json",
		"requests/request_charts-listings.json",
		"requests/request_overview.json",
		"requests/request_agents-insights-top-statistics?type=listings.json",
		"requests/request_top-communities.json",
		"requests/request_agents-insights-top-statistics?type=quality_score.json",
		"requests/request_whatsapp-insights-daily.json",
		"requests/request_agents.json",
		"requests/request_insights-v1-filters.json",
		"requests/request_whatsapp-insights-hourly.json",
		"requests/request_calls-insights-daily.json",
		"requests/request_insights-v1-inventory.json",
		"requests/request_whatsapp-insights-last7days.json",
		"requests/request_calls-insights-hourly.json",
		"requests/request_leads-insights-last7days.json",
		"requests/request_whatsapp-insights.json",
		"requests/request_calls-insights-last7days.json",
		"requests/request_leads-insights.json",
		"requests/request_calls-insights.json",
		"requests/request_listings-lpl.json",
	}

	for _, filePath := range filePaths {
		reqRes, err := readRequestResponse(filePath)
		if err != nil {
			fmt.Printf(
				"Error reading request/response file: %s\n",
				err,
			)
			continue
		}

		resp, err := sendAPIRequest(
			reqRes.Request,
			authToken,
		)
		if err != nil {
			fmt.Printf(
				"Error sending API request: %s\n",
				err,
			)
			continue
		}
		//check if the response status is 200
		if resp.StatusCode != 200 {
			fmt.Printf(
				Red+
					"Error: Response status is not 200 for the endpoint: %s\n"+Reset,
				reqRes.Request.URL,
			)
			continue
		}

		defer resp.Body.Close()

		/*		// Log the request URL
				fmt.Printf(
					"Request URL: %s\n",
					reqRes.Request.URL,
				)
				// Log the request method
				fmt.Printf(
					"Request method: %s\n",
					reqRes.Request.Method,
				)
				// Log the request headers
				fmt.Printf(
					"Request headers: %v\n",
					resp.Request.Header,
				)
				// Log the response status
				fmt.Printf(
					"Response status: %d\n",
					resp.StatusCode,
				)*/

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf(
				"Error reading response body: %s\n",
				err,
			)
			continue
		}

		/*		// Log the response body
				fmt.Printf(
					"Response body: %s\n",
					string(body),
				)*/

		originalResponse := reqRes.Response.Body
		newResponse := json.RawMessage(body)

		diff, explanation := compareResponses(
			originalResponse,
			newResponse,
		)

		// log the difference and skip the explanation if the difference value is "FullMatch"

		diffrence := diff.String()

		if diffrence == "FullMatch" {
			fmt.Printf(
				Green+" ✓ No Difference: %s for the endpoint: %s \n"+Reset,
				diff,
				reqRes.Request.URL,
			)
			continue
		} else {
			fmt.Printf(
				Red+" ✗ Difference Found: %s for the endpoint: %s \n"+Reset,
				diff,
				reqRes.Request.URL,
			)
			fmt.Printf(
				"Explanation: %s\n",
				explanation,
			)
		}
	}
}
