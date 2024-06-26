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

	httpReq.Header.Set(
		"Authorization",
		authToken,
	)

	/*	fmt.Printf(
			"Sending request to %s with method %s\n",
			req.URL,
			req.Method,
		)
		fmt.Printf(
			"Request headers: %v\n",
			httpReq.Header,
		)*/

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
		"requests/request_leads-insights.json",
		"requests/request_nsights-v1-credits.json",
		// Add other file paths
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

		defer resp.Body.Close()

		// Log the request URL
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
			reqRes.Request.Headers,
		)
		// Log the response status
		fmt.Printf(
			"Response status: %d\n",
			resp.StatusCode,
		)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf(
				"Error reading response body: %s\n",
				err,
			)
			continue
		}

		// Log the response body
		fmt.Printf(
			"Response body: %s\n",
			string(body),
		)

		originalResponse := reqRes.Response.Body
		newResponse := json.RawMessage(body)

		diff, explanation := compareResponses(
			originalResponse,
			newResponse,
		)
		fmt.Printf(
			"Difference: %s\n",
			diff,
		)
		fmt.Printf(
			"Explanation: %s for the file %s\n",
			explanation,
			filePath,
		)
	}
}
