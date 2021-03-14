### What's that?

API wrapper for Glassnode on-chain market intelligence indicators

#### How do I use it?

Let's ask Glassnode for the latest SOPR indicator value. 
A simple request could be built like the below

```golang
import (
	"context"
	"fmt"
	"time"

	"github.com/tech-branch/glassnode"
)

func main() {
	apisample := glassnode.NewClient("your_api_key")

    yesterdayAndSome := int(time.Now().Unix() - 87000)

	opts := glassnode.APIOptionsList{
		Asset:     "BTC",            // or "ETH"
		Metric:    "sopr",           // or "nupl"
		Since:     yesterdayAndSome, // UNIX timestamp, 0 means all
		Until:     0,                // UNIX timestamp, 0 means all
		Frequency: "",               // 1h, 24h.. default is 24h
		Format:    "",               // JSON by default, CSV unsupported
	}

	d, err := glassnode.GetMetricData(context.Background(), *apisample, &opts)
	if err != nil {
		println(err.Error())
	}

	for i := 0; i < len(d); i++ {
		fmt.Printf("%f\n", d[i].Value)
		fmt.Print(time.Unix(d[i].Time, 0))
	}
}

```

Which would print something like the below:

```
1.026512
2021-03-13 00:00:00 +0000 GMT
```

In case you didn't update the API key, error would be printed like:

```
unknown error, status code: 401
```
