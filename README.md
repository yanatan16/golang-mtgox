# mtgox

An implementation of a Mt. Gox client in Go. It uses the [websocket interface](https://en.bitcoin.it/wiki/MtGox/API/Streaming).

_Note_: This API mostly complete but untested in production.

## Documentation

Example below. [API Documentation on Godoc](http://godoc.org/github.com/yanatan16/golang-mtgox).

## Example

```go
func ExampleMtGoxApi() {
	gox, err := NewFromConfig(os.ExpandEnv("$MTGOX_CONFIG"))
	if err != nil {
		panic(err)
	}

	c, err := gox.Subscribe("ticker", "depth")
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range c {
			fmt.Println("Got ticker or depth:", msg)
		}
	}()

	orderchan, err := gox.SubmitOrder("bid", 100000000, 10000) // Both are in _int notation. Will add float later
	if err != nil {
		panic(err)
	}

	order := <- orderchan
	if order.Success {
		fmt.Println("Yay submitted an order!", order.Result)
	} else {
		fmt.Println("Boo. Failure!", order.Error)
	}
}
```

## License

MIT found in LICENSE file.
