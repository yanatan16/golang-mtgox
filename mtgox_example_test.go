package mtgox

import (
	"fmt"
	"os"
)

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