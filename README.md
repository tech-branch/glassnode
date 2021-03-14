### What's that?

API wrapper for Glassnode on-chain market intelligence indicators

#### How do I use it?

Let's ask Glassnode for the latest SOPR indicator value. A simple request could be built like:

```golang
apisample := NewClient("your_apikey")

opts := APIOptionsList{
    Asset:     "BTC",  // or "ETH"
    Metric:    "sopr", // or "nupl"
    Since:     0,      // UNIX timestamp, 0 means all
    Until:     0,      // UNIX timestamp, 0 means all
    Frequency: "",     // 1h, 24h.. default is 24h
    Format:    "",     // JSON by default, CSV unsupported
}

d, err := GetMetricData(context.Background(), *apisample, &opts)
if err != nil {
    ...
}

for i := 0; i < len(d); i++ {
    fmt.Printf("%f\n", d[i].Value)
    fmt.Printf(time.Unix(d[i].Time, 0))
}
```

Which would print a list of elements like the below:

```
1.026512
2021-03-13 00:00:00 +0000 GMT
```