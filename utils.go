package glassnode

import (
	"fmt"
	"net/url"
	"path"
)

func constructURL(api Client, options *APIOptionsList) (string, error) {

	// --------
	// Base URL
	// --------

	baseURL, err := url.Parse(api.BaseURL)
	if err != nil {
		return "", fmt.Errorf("[GetMetricData] couldn't parse url: %s", err.Error())
	}
	baseURL.Path += path.Join(baseURL.Path, MetricsPrefix)

	// -------------
	// Data Category
	// -------------

	if options.Category == "" {
		return "", fmt.Errorf("APIOptionsList.Category appears to be empty but is required")
	}
	baseURL.Path += path.Join(baseURL.Path, options.Category)

	// ---------------
	// Specific metric
	// ---------------

	if options.Metric == "" {
		return "", fmt.Errorf("APIOptionsList.Metric appears to be empty but is required")
	}
	baseURL.Path += path.Join(baseURL.Path, options.Metric)

	finalParams, err := makeParams(api.apiKey, options)
	if err != nil {
		return "", fmt.Errorf("[GetMetricData] couldn't prepare parameters: %s", err.Error())
	}
	baseURL.RawQuery = finalParams.Encode()

	return baseURL.String(), nil
}

func makeParams(apiKey string, options *APIOptionsList) (*url.Values, error) {
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

	if apiKey == "" {
		return nil, fmt.Errorf("api key appears to be empty but is required")
	}
	unrefinedParams["api_key"] = apiKey

	if options.Asset != "" {
		unrefinedParams["a"] = options.Asset
	}
	if unrefinedParams["a"] == "" {
		return nil, fmt.Errorf("parameter a (Asset) appears to be empty but is required")
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
		return nil, fmt.Errorf("parameter f (Format) shouldn't be specified")
	}

	// -----------------------
	// Construct the final URL
	// -----------------------

	for key, value := range unrefinedParams {
		finalParams.Add(key, value)
	}
	return &finalParams, nil
}
