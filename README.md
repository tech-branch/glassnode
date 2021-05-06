### What's that?

API wrapper for Glassnode on-chain market intelligence data

#### How do I use it?

Let's ask Glassnode for the latest SOPR indicator value. 
A simple request could be built like this:

```golang
import (
	"context"
	"fmt"
	"time"

	"github.com/tech-branch/glassnode"
)

func main() {
	apisample := glassnode.NewClient("your_api_key")

	yesterday := glassnode.YesterdayTimestamp()

	opts := glassnode.APIOptionsList{
		Asset:    "BTC",        // or "ETH"
		Category: "indicators", // market, derivatives etc
		Metric:   "sopr",       // or "nupl"
		Since:    yesterday,    // UNIX timestamp, 0 means all
		Until:    0,            // UNIX timestamp, 0 means all
	}

	d, err := glassnode.GetMetricData(context.Background(), *apisample, &opts)
	if err != nil {
		fmt.Errorf("couldn't fetch metric data: %s", err.Error())
	}

	// See "Why multiple return types" below for explanation
	e := d.([]TimeValue)

	for i := 0; i < len(e); i++ {
		fmt.Printf("%f\n", e[i].Value)
		fmt.Print(time.Unix(e[i].Time, 0))
	}
}

```

Which would print something like the below:

```
1.026512
2021-03-13 00:00:00 +0000 GMT
```

In case you didn't update the API key, you would receive a 401 error:

```
[GetMetricData] request errored: [sendRequest] HTTP request unsuccessful: 401
```

### Why multiple return types?

Unfortunately, there's no simple way of telling which type of response we'll get from Glassnode. Therefore the unmarshalling process returns either one of:
- `[]TimeValue` 
- `[]TimeOptions`

Glassnode has two types of responses: 
- list of single value maps 
- list of multiple values maps 

Check Glassnode API documentation to see specific examples, here are some simplified ones:

- `TimeOptions` type:

```json
	[
		{
			"t":1604361600,
			"o":{
				"ma128":7.671518382631376e+22,
				...
				"ma90":7.893158075608794e+22
			}
		}
	]
```

- `TimeValue` type:

```json
[
	{
		"t": 1586217600,
		"v": 0.04205656140305256
	},
	...
	{
		"t": 1586995200,
		"v": 0.028548190409882022
	}
]
```

### There's more!

There's another way of specifying parameters in case you need it. Not every parameter option is implemented by the APIOptionsList but additional parameters can be specified using a map like presented below.

```golang
opts := APIOptionsList{
	Asset:    "BTC",               // or "ETH"
	Category: "indicators",        // either indicators, market, mining etc
	Metric:   "difficulty_ribbon", // or "nupl"
	DirectMapping: map[string]string{
		"s": "1586280246",  // equivalent of Since
		"u": "1587144246",	// equivalent of Until
	},
}
```

Please see the tests file for complete examples.
