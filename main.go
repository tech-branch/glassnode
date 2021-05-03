package glassnode

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	BaseURLV1     = "https://api.glassnode.com/v1/"
	MetricsPrefix = "metrics/"
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

func (c *Client) sendRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("[sendRequest] HTTP request unsuccessful: %d", res.StatusCode)
	}

	bResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("[sendRequest] couldnt read response body: %s", err.Error())
	}

	return bResponse, nil
}

// APIOptionsList represents query parameters in a request
// more details under https://docs.glassnode.com/api/indicators
type APIOptionsList struct {
	// Asset (required) indicates which asset you'd like to request, eg. BTC.
	// stands for the a parameter, translates to: &a=BTC in the final URL
	Asset string

	// Metric (required) indicates which metric you'd like to request, eg. sopr.
	// Represents part of the URL path, translates to: /sopr in the final URL
	Metric string

	// Category (required) is the URL modificator, can be one of: market, derivatives etc, check glassnode.
	// Represents part of the URL path, translates to: /market/<metric> in the final URL
	Category string

	// DirectMapping forms a sql-like union with the other parameters, must follow Glassnode documentation.
	// May be something like {"s": "123"} which adds &s=123 to the request
	DirectMapping map[string]string

	// Since is a UNIX Timestamp indicating the starting point for the fetched dataset.
	// Stands for the s parameter, translates to: &s=<value> in the final URL
	Since int

	// Until is a UNIX Timestamp indicating the ending point for the fetched dataset.
	// Stands for the u parameter, translates to: &u=<value> in the final URL
	Until int

	// Frequency specifies the data interval, usually 1h, 24h, check glassnode.
	// Stands for the i parameter, translates to: &i=<value> in the final URL
	Frequency string

	// Format - defaults to JSON and that's the only one supported in this lib so far.
	// Only here to indicate it's not supported.
	// If you speficy "f" in DirectMappings, it'll be erased.
	// Format string
}

func GetMetricData(ctx context.Context, api Client, options *APIOptionsList) (interface{}, error) {

	// -----------------
	// Parse API options
	// -----------------

	fullURL, err := constructURL(api, options)
	if err != nil {
		return nil, err
	}

	// -------------------
	// Construct a request
	// -------------------

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("[GetMetricData] error wrapping request: %s", err.Error())
	}

	req = req.WithContext(ctx)
	res, err := api.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("[GetMetricData] request errored: %s", err.Error())
	}

	// ----------------------
	// Unmarshal the response
	// ----------------------

	parsedResponse, err := UnmarshalJSON(res)
	if err != nil {
		return nil, fmt.Errorf("[GetMetricData] error parsing response: %s", err.Error())
	}

	return parsedResponse, nil
}

func constructURL(api Client, options *APIOptionsList) (string, error) {

	// --------
	// Base URL
	// --------

	baseURL, err := url.Parse(api.BaseURL)
	if err != nil {
		return "", fmt.Errorf("[GetMetricData] couldn't parse url: %s", err.Error())
	}
	baseURL.Path += MetricsPrefix

	// -------------
	// Data Category
	// -------------

	if options.Category == "" {
		return "", fmt.Errorf("APIOptionsList.Category appears to be empty but is required")
	}
	baseURL.Path += options.Category + "/"

	// ---------------
	// Specific metric
	// ---------------

	if options.Metric == "" {
		return "", fmt.Errorf("APIOptionsList.Metric appears to be empty but is required")
	}
	baseURL.Path += options.Metric

	// ---------------
	// Parsing helpers
	// ---------------

	unrefinedParams := make(map[string]string)
	finalParams := url.Values{}

	//
	// apply raw params, if any
	//

	for key, value := range options.DirectMapping {
		unrefinedParams[key] = value
	}

	//
	// required params
	//

	if api.apiKey == "" {
		return "", fmt.Errorf("api key appears to be empty but is required")
	}
	unrefinedParams["api_key"] = api.apiKey

	if options.Asset != "" {
		unrefinedParams["a"] = options.Asset
	}
	if unrefinedParams["a"] == "" {
		return "", fmt.Errorf("parameter a (Asset) appears to be empty but is required")
	}

	//
	// optional params
	//

	if options.Since != 0 {
		unrefinedParams["s"] = fmt.Sprint(options.Since)
	}
	if options.Until != 0 {
		unrefinedParams["u"] = fmt.Sprint(options.Until)
	}
	if options.Frequency != "" {
		unrefinedParams["i"] = fmt.Sprint(options.Frequency)
	}

	//
	// Unsupported options:
	if unrefinedParams["f"] != "" {
		return "", fmt.Errorf("parameter f (Format) shouldn't be specified")
	}

	// -----------------------
	// Construct the final URL
	// -----------------------

	for key, value := range unrefinedParams {
		finalParams.Add(key, value)
	}
	baseURL.RawQuery = finalParams.Encode()

	return baseURL.String(), nil
}

// -----------
// Data models
// -----------

type TimeValue struct {
	Time  int64   `json:"t"`
	Value float64 `json:"v"`
}

type TimeOptions struct {
	Time    int64              `json:"t"`
	Options map[string]float64 `json:"o"`
}

// -----------------------
// Dual-type unmarshalling
// -----------------------

func UnmarshalJSON(b []byte) (interface{}, error) {
	//
	// Attempt the first type conversion
	//
	tv := []TimeValue{}
	err := json.Unmarshal(b, &tv)

	//
	// no error, but we need to make sure we unmarshalled into the correct type
	//
	if err == nil && tv[0].Value != 0.0 {
		// it appears to be the TimeValue type, return.
		return tv, nil
	}

	// So it appears it's not the TimeValue type, now check for errors
	// and attempt unmarshalling into TimeOptions
	//
	// abort if we have an error other than the wrong type
	//
	if _, ok := err.(*json.UnmarshalTypeError); err != nil && !ok {
		return nil, err
	}

	//
	// Unmarshalling into TimeOptions
	//
	to := []TimeOptions{}
	err = json.Unmarshal(b, &to)
	if err != nil {
		return nil, err
	}

	return to, nil
}

// -----
// Utils
// -----

// YesterdayTimestamp returns an int timestamp of the current time minus 24h
func YesterdayTimestamp() int {
	dt := time.Now().Add(-24 * time.Hour)
	return int(dt.Unix())
}
