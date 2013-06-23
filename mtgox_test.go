package mtgox

import (
	"fmt"
	"testing"
	"time"
	"os"
)

var gox *StreamingApi

func init () {
	var err error
	gox, err = NewFromConfig(os.ExpandEnv("$HOME/.gogox/config.json"))
	if err != nil {
		panic(err)
	}
}

func TestTickerReceive(t *testing.T) {
	ticker, err := gox.Subscribe("ticker")
	if err != nil {
		t.Error(err)
	}

	select {
	case tick := <-ticker:
		channel := tick["channel"]

		if channel != "d5f06780-30a8-4a48-a2f8-7ed181b4a13f" {
			t.Error("Unexpected mesasged received as ticker.BTCUSD")
		}

	case <-time.After(10 * time.Second):
		t.Error("No tickers received after five seconds.")
	}
}

func TestInfoCall(t *testing.T) {
	infoc, err := gox.Info()
	if err != nil {
		t.Error(err)
	}

	select {
	case info := <-infoc:
		fmt.Println("Info:", info)
	case <-time.After(10 * time.Second):
		t.Error("No message after a second.")
	}
}
