package mtgox

import (
	"code.google.com/p/go.net/websocket"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/yanatan16/mtgox/jsonws"
	"github.com/yanatan16/mtgox/msgwatch"
	"strings"
	"fmt"
	"io/ioutil"
	"time"
	exq "github.com/yanatan16/exchequer"
)

type StreamType string
type OrderType string

const (
	api_host      string = "wss://websocket.mtgox.com:443"
	api_path      string = "/mtgox"
	http_endpoint string = "http://mtgox.com/api/2"
	origin_url    string = "http://localhost"

	TICKER StreamType = "ticker"
	DEPTH  StreamType = "depth"
	TRADES StreamType = "trades"

	BID OrderType = "bid"
	ASK OrderType = "ask"
)

var (
	// TODO: use https://mtgox.com/api/2/stream/list_public
	channels map[string]string = make(map[string]string)
	randStr  chan string
)

type StreamingApi struct {
	*msgwatch.MsgWatcher
	ws     *jsonws.JsonWebsocket
	key    []byte
	secret []byte
}

type Config struct {
	Currencies []string
	Key        string
	Secret     string
}

type CallResult struct {
	Result map[string]interface{}
	Success bool
	Error string
}

func NewFromConfig(cfgfile string) (*StreamingApi, error) {
	file, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		return nil, err
	}

	m := new(Config)
	err = json.Unmarshal(file, m)
	if err != nil {
		return nil, err
	}

	return New(m.Key, m.Secret, m.Currencies...)
}

func New(key, secret string, currencies ...string) (*StreamingApi, error) {
	url := fmt.Sprintf("%s%s?Currency=%s", api_host, api_path, strings.Join(currencies, ","))
	config, _ := websocket.NewConfig(url, origin_url)
	ws, err := jsonws.New(config)

	if err != nil {
		return nil, err
	}

	api := &StreamingApi{
		MsgWatcher: msgwatch.New(ws.RecvForever()),
		ws:         ws,
	}

	api.key, err = hex.DecodeString(strings.Replace(key, "-", "", -1))
	if err != nil {
		return nil, err
	}

	api.secret, err = base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, err
	}

	remarks := api.Listen("remark")
	go func() {
		for remark := range remarks {
			fmt.Println("Remark:", remark)
		}
	}()

	return api, err
}

func (api *StreamingApi) Close() error {
	return api.ws.Close()
}

// Unsubscribe from a named channel
func (api *StreamingApi) Unsubscribe(name string) error {
	if key, ok := api.Channels[name]; ok {
		return api.ws.Send(map[string]interface{}{
			"op":      "unsubscribe",
			"channel": key,
		})
	}

	return errors.New("No channel with name " + name + ".")
}

// Subscribe to a type of channel. Returns the Listen(typ) listener.
func (api *StreamingApi) Subscribe(typ StreamType) (chan map[string]interface{}, error) {
	err := api.ws.Send(map[string]interface{}{
		"op":   "mtgox.subscribe",
		"type": string(typ),
	})

	if err != nil {
		return nil, err
	}

	return api.Listen("ticker"), nil
}

func (api *StreamingApi) sign(body []byte) ([]byte, error) {
	mac := hmac.New(sha512.New, api.secret)
	_, err := mac.Write(body)
	if err != nil {
		return nil, err
	}

	return mac.Sum(nil), nil
}

func (api *StreamingApi) authenticatedSend(msg map[string]interface{}) error {
	if api.key == nil || api.secret == nil {
		return errors.New("API Key or secret is invalid or missing.")
	}

	req, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	signedReq, err := api.sign(req)
	if err != nil {
		return err
	}

	reqid := msg["id"]

	fullReq := append(append(api.key, signedReq...), req...)
	encodedReq := base64.StdEncoding.EncodeToString(fullReq)

	fmt.Println("Request:", string(req))
	return api.ws.Send(map[string]interface{}{
		"op":      "call",
		"id":      reqid,
		"call":    encodedReq,
		"context": "mtgox.com",
	})
}

func (api *StreamingApi) call(endpoint string, params map[string]interface{}) (chan *CallResult, error) {
	if params == nil {
		params = make(map[string]interface{})
	}
	msg := map[string]interface{}{
		"call":   endpoint,
		"item":   "BTC",
		"params": params,
		"id":     <-ids,
		"nonce":  <-nonces,
	}

	err := api.authenticatedSend(msg)
	if err != nil {
		return nil, err
	}

	out := interpretResult(api.ListenResult(msg["id"].(string)))

	return out, nil
}

func interpretResult(in chan map[string]interface{}) (chan *CallResult) {
	out := make(chan *CallResult)
	go func () {
		res := <- in
		if op, err := exq.String(res, "op"); err == nil {
			if op == "remark" {
				out <- &CallResult{
					Success: false,
					Error: res["remark"].(string),
				}
			} else if op == "result" {
				out <- &CallResult{
					Success: true,
					Result: res["result"].(map[string]interface{}),
				}
			} else {
				out <- &CallResult{
					Success: false,
					Error: fmt.Sprintf("Don't know how to interpret: %s", res),
				}
			}
		} else {
			out <- &CallResult{
				Success: false,
				Error: fmt.Sprintf("Error accessing object: %s", err.Error()),
			}
		}
	}()
	return out
}

func (api *StreamingApi) Info() (chan *CallResult, error) {
	return api.call("private/info", nil)
}

func (api *StreamingApi) Address(description string) (chan *CallResult, error) {
	m := make(map[string]interface{})
	if description != "" {
		m["description"] = description
	}
	return api.call("bitcoin/address", m)
}

func (api *StreamingApi) AddPrivateKey(key string, desc string) (chan *CallResult, error) {
	return api.call("bitcoin/addpriv", map[string]interface{}{
		"key":         key,
		"keytype":     "auto",
		"description": desc,
	})
}

func (api *StreamingApi) AddWallet(walletdat, description string) (chan *CallResult, error) {
	m := map[string]interface{}{
		"wallet": walletdat,
	}
	if description != "" {
		m["description"] = description
	}
	return api.call("bitcoin/wallet_add", m)
}

func (api *StreamingApi) Send(addr string, amount_int uint64) (chan *CallResult, error) {
	return api.call("bitcoin/send_simple", map[string]interface{}{
		"address":    addr,
		"amount_int": amount_int,
	})
}

func (api *StreamingApi) IdKey() (chan *CallResult, error) {
	return api.call("private/idkey", nil)
}

func (api *StreamingApi) FullHistory(currency string, page int) (chan *CallResult, error) {
	m := map[string]interface{}{
		"currency": currency,
	}
	if page != 0 {
		m["page"] = page
	}
	return api.call("wallet/history", m)
}

func (api *StreamingApi) QueryHistory(currency string, typ string, begin, end *time.Time, page int) (chan *CallResult, error) {
	m := map[string]interface{}{
		"currency": currency,
	}
	if typ != "" {
		m["type"] = typ
	}
	if begin != nil {
		m["date_start"] = begin.String()
	}
	if end != nil {
		m["date_end"] = end.String()
	}
	if page != 0 {
		m["page"] = page
	}
	return api.call("wallet/history", m)
}

func (api *StreamingApi) SubmitOrder(typ string, amount, price uint64) (chan *CallResult, error) {
	return api.call("order/add", map[string]interface{}{
		"type":       typ,
		"amount_int": amount,
		"price_int":  price,
	})
}

func (api *StreamingApi) CancelOrder(oid string) (chan *CallResult, error) {
	return api.call("order/cancel", map[string]interface{}{
		"oid": oid,
	})
}

func (api *StreamingApi) Orders() (chan *CallResult, error) {
	return api.call("private/orders", nil)
}

func (api *StreamingApi) OrderResult(oid string) (chan *CallResult, error) {
	return api.call("order/result", map[string]interface{}{"oid": oid})
}

func (api *StreamingApi) AddressDetails(addr string) (chan *CallResult, error) {
	return api.call("bitcoin/addr_details", map[string]interface{}{"hash": addr})
}

func (api *StreamingApi) Lag() (chan *CallResult, error) {
	return api.call("order/lag", nil)
}
