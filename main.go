package glassnode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	BaseURLV1    = "https://api.glassnode.com/v1/"
	MetricPrefix = "metrics/indicators/"
)

func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL: BaseURLV1,
		apiKey:  apiKey,
		http: &http.Client{
			Timeout: time.Minute,
		},
	}
}

type Client struct {
	BaseURL string
	apiKey  string
	http    *http.Client
}

func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}

		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return err
	}

	return nil
}

// APIOptionsList represents query parameters in a request
// more details under https://docs.glassnode.com/api/indicators
type APIOptionsList struct {
	// BTC
	Asset string
	// sopr
	Metric string
	// UNIX Timestamp
	Since int
	// UNIX Timestamp
	Until int
	// 1h, 24h
	Frequency string
	// JSON by default, CSV unsupported
	Format string
}

func GetMetricData(ctx context.Context, api Client, options *APIOptionsList) (MetricData, error) {
	baseURL, err := url.Parse(api.BaseURL)
	if err != nil {
		fmt.Println("GetMetricData malformed URL: ", err.Error())
		return nil, err
	}
	baseURL.Path += MetricPrefix
	baseURL.Path += options.Metric

	params := url.Values{}

	// required params
	params.Add("api_key", api.apiKey)
	params.Add("a", options.Asset)

	// optional params
	if options.Since != 0 {
		params.Add("s", fmt.Sprint(options.Since))
	}
	if options.Until != 0 {
		params.Add("u", fmt.Sprint(options.Until))
	}
	if options.Frequency != "" {
		params.Add("i", fmt.Sprint(options.Frequency))
	}
	if options.Format != "" {
		params.Add("f", fmt.Sprint(options.Frequency))
	}
	baseURL.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	res := MetricData{}
	if err := api.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return res, nil
}

type MetricData []DataPoint

type DataPoint struct {
	Time  int64   `json:"t"`
	Value float64 `json:"v"`
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Sample raw response:
//
// [
//     {
//         "t": 1615420800,
//         "v": 1.03415767306151
//     },
//     {
//         "t": 1615507200,
//         "v": 1.03174070656315
//     }
// ]
