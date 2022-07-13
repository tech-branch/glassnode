package glassnode

import (
	"context"
	"os"
	"testing"
)

func Test_tvconv(t *testing.T) {
	a := `
	[
		{
			"t":1604361600,
			"o":{
				"ma128":7.671518382631376e+22,
				"ma14":8.49032655528302e+22,
				"ma200":7.285316967299924e+22,
				"ma25":8.449559630140737e+22,
				"ma40":8.390750464697435e+22,
				"ma60":8.150679497286024e+22,
				"ma9":8.435624413591933e+22,
				"ma90":7.893158075608794e+22
			}
		}
	]
	`

	b, err := UnmarshalJSON([]byte(a))
	if err != nil {
		t.Logf("Failed Unmarshalling mock json object: %s", err.Error())
		t.FailNow()
	}

	c := b.([]TimeOptions)

	if c[0].Time != 1604361600 {
		t.Logf("Options element parsing failed: Time %d doesnt match the mock value", c[0].Time)
		t.Fail()

	}

	if c[0].Options["ma9"] < 100 {
		t.Logf("Options element parsing failed: MA9 %f should be a large number", c[0].Options["ma9"])
		t.Fail()
	}

}

func Test_toconv(t *testing.T) {
	a := `
	[
	{
		"t": 1586217600,
		"v": 0.04205656140305256
	},
	{
		"t": 1586304000,
		"v": 0.04161501005133026
	},
	{
		"t": 1586908800,
		"v": 0.02936962662277417
	},
	{
		"t": 1586995200,
		"v": 0.028548190409882022
	}
	]
	`

	b, err := UnmarshalJSON([]byte(a))
	if err != nil {
		t.Logf("Failed Unmarshalling mock json object: %s", err.Error())
		t.FailNow()
	}

	c := b.([]TimeValue)

	if c[0].Time != 1586217600 {
		t.Logf("TimeValue element parsing failed: Time %d doesnt match the mock value", c[0].Time)
		t.Fail()
	}

	if c[0].Value < 0.042 {
		t.Logf("TimeValue element parsing failed: Value %f should be larger than .042", c[0].Value)
		t.Fail()
	}
}

func Test_liveintegration(t *testing.T) {
	apik := os.Getenv("GLASSNODE_API_KEY")
	if apik == "" {
		t.Log("live API Key is required for this test")
		t.SkipNow()
	}
	apisample := NewClient(apik)

	yesterday := YesterdayTimestamp()

	opts := APIOptionsList{
		Asset:     "BTC",        // or "ETH"
		Metric:    "sopr",       // or "nupl"
		Since:     yesterday,    // UNIX timestamp, 0 means all
		Until:     0,            // UNIX timestamp, 0 means all
		Frequency: "",           // 1h, 24h.. default is 24h
		Category:  "indicators", // either indicators, market, mining etc
	}

	d, err := GetMetricData(context.Background(), *apisample, &opts)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	e := d.([]TimeValue)

	if e[0].Value > 2 || e[0].Value < 0 {
		t.Log("SOPR should be between 0 and 2")
		t.Fail()
	}
}

func Test_liveintegration_directmappings(t *testing.T) {
	apik := os.Getenv("GLASSNODE_API_KEY")
	if apik == "" {
		t.Log("live API Key is required for this test")
		t.SkipNow()
	}
	apisample := NewClient(apik)

	opts := APIOptionsList{
		Asset:    "BTC",               // or "ETH"
		Category: "indicators",        // either indicators, market, mining etc
		Metric:   "difficulty_ribbon", // or "nupl"
		DirectMapping: map[string]string{
			"s": "1586280246",
			"u": "1587144246",
		},
	}

	d, err := GetMetricData(context.Background(), *apisample, &opts)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	e := d.([]TimeOptions)

	if e[0].Time != 1586217600 {
		t.Logf("First record time should be %d but is %d", 1586217600, e[0].Time)
		t.Fail()
	}
	if e[0].Options["ma9"] < 100 {
		t.Logf("ma9 val should be greater than %f but is %f", 100.0, e[0].Options["ma9"])
		t.Fail()
	}
}

func Test_constructURL(t *testing.T) {

	apisample := NewClient("x")

	opts := APIOptionsList{
		Asset:    "BTC",
		Category: "indicators",
		Metric:   "difficulty_ribbon",
		DirectMapping: map[string]string{
			"s": "123",
			"u": "456",
		},
	}
	expectation := "https://api.glassnode.com/v1/metrics/indicators/difficulty_ribbon?a=BTC&api_key=x&s=123&u=456"
	result, err := constructURL(*apisample, &opts)

	if err != nil {
		t.Fail()
	}

	if expectation != result {
		t.Logf("Result: %s didn't match expectation: %s", result, expectation)
		t.Fail()
	}

}
