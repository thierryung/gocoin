package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Jeffail/gabs"
	ws "github.com/gorilla/websocket"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type Order struct {
	Id    float64
	Side  string
	Size  float64
	Price float64
	Time  time.Time
}

type GdaxMessage struct {
	Type          string     `json:"type"`
	ProductId     string     `json:"product_id"`
	TradeId       int        `json:"trade_id,number"`
	OrderId       string     `json:"order_id"`
	Sequence      int64      `json:"sequence,number"`
	MakerOrderId  string     `json:"maker_order_id"`
	TakerOrderId  string     `json:"taker_order_id"`
	Time          time.Time  `json:"time,string"`
	RemainingSize float64    `json:"remaining_size,string"`
	NewSize       float64    `json:"new_size,string"`
	OldSize       float64    `json:"old_size,string"`
	Size          float64    `json:"size,string"`
	Price         float64    `json:"price,string"`
	Side          string     `json:"side"`
	Reason        string     `json:"reason"`
	OrderType     string     `json:"order_type"`
	Funds         float64    `json:"funds,string"`
	NewFunds      float64    `json:"new_funds,string"`
	OldFunds      float64    `json:"old_funds,string"`
	Message       string     `json:"message"`
	Bids          [][]string `json:"bids,omitempty"`
	Asks          [][]string `json:"asks,omitempty"`
	Changes       [][]string `json:"changes,omitempty"`
}

type BitfinexTrade struct {
	ID        int64
	Timestamp int64
	Amount    float64
	Price     float64
}

type ExchangePrice struct {
	Name  string
	Price float64
}

// JSONEncode encodes structure data into JSON
func JSONEncode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// JSONDecode decodes JSON data into a structure
func JSONDecode(data []byte, to interface{}) error {
	if !strings.Contains(reflect.ValueOf(to).Type().String(), "*") {
		return errors.New("json decode error - memory address not supplied")
	}
	return json.Unmarshal(data, to)
}

func goGdax(prices map[string]float64) {
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.gdax.com", nil)
	if err != nil {
		println(err.Error())
	}

	subscribe := map[string]string{
		"type":       "subscribe",
		"product_id": "BTC-USD",
	}
	if err := wsConn.WriteJSON(subscribe); err != nil {
		println(err.Error())
	}
	message := GdaxMessage{}
	for true {
		if err := wsConn.ReadJSON(&message); err != nil {
			println(err.Error())
			break
		}
		heyChannelAllowed := map[string]bool{"done": true, "received": true, "open": true}
		_, filter := heyChannelAllowed[message.Type]

		if !filter && message.Reason != "canceled" {
			//fmt.Printf("%#v\n", message)
			prices["gdax"] = message.Price
		}
	}
}

func goBitfinex(prices map[string]float64) {
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://api.bitfinex.com/ws/2", nil)
	if err != nil {
		println(err.Error())
	}

	subscribe := map[string]string{
		"event":   "subscribe",
		"channel": "trades",
		"symbol":  "tBTCUSD",
	}
	if err := wsConn.WriteJSON(subscribe); err != nil {
		println(err.Error())
	}
	for true {
		msgType, resp, err := wsConn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}

		switch msgType {
		case ws.TextMessage:
			var result interface{}
			err := JSONDecode(resp, &result)
			if err != nil {
				fmt.Println(err)
				continue
			}

			switch reflect.TypeOf(result).String() {
			case "map[string]interface {}":
				eventData := result.(map[string]interface{})
				event := eventData["event"]
				fmt.Printf("event is %#v\n", event)

				switch event {
				case "subscribed":
					println("subscribed")
				case "auth":
					status := eventData["status"].(string)

					if status == "OK" {
						println("auth ok")
					} else if status == "fail" {
						println("auth NOT ok")
					}
				}
			case "[]interface {}":
				chanData := result.([]interface{})
				//chanID := int(chanData[0].(float64))
				if chanData[1] != "tu" {
					continue
				}
				switch len(chanData) {
				case 2:
					fmt.Printf("lots of trades: %#v", chanData)
				case 3:
					trade := chanData[2].([]interface{})
					message := BitfinexTrade{ID: int64(trade[0].(float64)), Timestamp: int64(trade[1].(float64)), Amount: trade[2].(float64), Price: trade[3].(float64)}
					//fmt.Printf("%#v\n", message)
					prices["bitfinex"] = message.Price
				}
			}
		}

	}
}

func goBitmex(orderBook map[float64]*Order, prices map[string]float64) {
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://www.bitmex.com/realtime?subscribe=orderBookL2:XBTUSD", nil)
	if err != nil {
		println(err.Error())
	}

	counter := 0

	for true {
		msgType, resp, err := wsConn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
		if msgType != ws.TextMessage {
			continue
		}
		jsonParsed, err := gabs.ParseJSON(resp)
		if err != nil {
			fmt.Println(err)
			break
		}
		table, ok := jsonParsed.Search("table").Data().(string)
		if !ok || table != "orderBookL2" {
			continue
		}
		action, _ := jsonParsed.Search("action").Data().(string)
		row, err := jsonParsed.Search("data").Children()
		if err != nil {
			fmt.Println(err)
			continue
		}
		if action == "delete" {
			for _, order := range row {
				orderParsed, _ := order.ChildrenMap()
				delete(orderBook, orderParsed["id"].Data().(float64))
			}

		} else {
			for _, order := range row {
				orderParsed, _ := order.ChildrenMap()
				id := orderParsed["id"].Data().(float64)
				if _, ok := orderBook[id]; !ok {
					orderBook[id] = &Order{Id: id}
				}
				if side, ok := orderParsed["side"].Data().(string); ok {
					orderBook[id].Side = side
				}
				if size, ok := orderParsed["size"].Data().(float64); ok {
					orderBook[id].Size = size
				}
				if price, ok := orderParsed["price"].Data().(float64); ok {
					orderBook[id].Price = price
				}
			}
		}

		// Get highest buy price, so we can short sell it
		counter += 1
		highest := 0.0
		if counter%1 == 0 {
			for _, order := range orderBook {
				if order.Side == "Buy" && highest < order.Price {
					highest = order.Price
				}
			}
			prices["bitmex"] = highest
		}
	}
}

func timer(prices map[string]float64) {
	for true {
		file, err := os.OpenFile("output.txt", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		if _, err = file.WriteString(fmt.Sprintf("%s###gdax:%f###bitfinex:%f###bitmex:%f\n", time.Now().Format(time.RFC3339), prices["gdax"], prices["bitfinex"], prices["bitmex"])); err != nil {
			panic(err)
		}
		time.Sleep(3 * time.Second)
	}
}

func mem() {
	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("\nAlloc = %v\nTotalAlloc = %v\nSys = %v\nNumGC = %v\n\n", m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC)
		time.Sleep(5 * time.Second)
	}
}

func main() {
	queue := make(chan int)
	bitmexOrderBook := map[float64]*Order{}

	prices := map[string]float64{"gdax": 0, "bitfinex": 0, "bitmex": 0}
	go goGdax(prices)
	go goBitfinex(prices)
	go goBitmex(bitmexOrderBook, prices)
	go timer(prices)
	go mem()

	//orderBook := map[float64]Order{}
	//jsonParsed, err := gabs.ParseJSON([]byte(`{"table":"orderBookL2","keys":["symbol","id","side"],"types":{"symbol":"symbol","id":"long","side":"symbol","size":"long","price":"float"},"foreignKeys":{"symbol":"instrument","side":"side"},"attributes":{"symbol":"grouped","id":"sorted"},"action":"partial","data":[{"symbol":"XBTUSD","id":8700000050,"side":"Sell","size":1000,"price":999999.5},{"symbol":"XBTUSD","id":8728338000,"side":"Sell","size":100000,"price":716620},{"symbol":"XBTUSD","id":8785065000,"side":"Sell","size":10000,"price":149350},{"symbol":"XBTUSD","id":8792053400,"side":"Sell","size":1000,"price":79466},{"symbol":"XBTUSD","id":8792065000,"side":"Sell","size":10000,"price":79350},{"symbol":"XBTUSD","id":8792065050,"side":"Sell","size":150000,"price":79349.5},{"symbol":"XBTUSD","id":8792898450,"side":"Sell","size":1,"price":71015.5},{"symbol":"XBTUSD","id":8796065000,"side":"Sell","size":10000,"price":39350},{"symbol":"XBTUSD","id":8797398100,"side":"Sell","size":10,"price":26019},{"symbol":"XBTUSD","id":8798011200,"side":"Sell","size":3999,"price":19888},{"symbol":"XBTUSD","id":8798065000,"side":"Sell","size":10000,"price":19350},{"symbol":"XBTUSD","id":8798065050,"side":"Sell","size":150000,"price":19349.5},{"symbol":"XBTUSD","id":8798468200,"side":"Sell","size":46,"price":15318},{"symbol":"XBTUSD","id":8798468350,"side":"Sell","size":44,"price":15316.5},{"symbol":"XBTUSD","id":8798468400,"side":"Sell","size":40,"price":15316},{"symbol":"XBTUSD","id":8798468450,"side":"Sell","size":43,"price":15315.5},{"symbol":"XBTUSD","id":8798468600,"side":"Sell","size":49,"price":15314},{"symbol":"XBTUSD","id":8798500000,"side":"Sell","size":35000,"price":15000},{"symbol":"XBTUSD","id":8798670000,"side":"Sell","size":1,"price":13300},{"symbol":"XBTUSD","id":8798716800,"side":"Sell","size":5000,"price":12832},{"symbol":"XBTUSD","id":8798750000,"side":"Sell","size":85000,"price":12500},{"symbol":"XBTUSD","id":8798770000,"side":"Sell","size":1,"price":12300},{"symbol":"XBTUSD","id":8798822050,"side":"Sell","size":3000,"price":11779.5},{"symbol":"XBTUSD","id":8798859050,"side":"Sell","size":100,"price":11409.5},{"symbol":"XBTUSD","id":8798880500,"side":"Sell","size":10000,"price":11195},{"symbol":"XBTUSD","id":8798885850,"side":"Sell","size":100,"price":11141.5},{"symbol":"XBTUSD","id":8798900200,"side":"Sell","size":3000,"price":10998}]}`))
	//row, err := jsonParsed.Search("data").Children()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//for _, order := range row {
	//	orderParsed, _ := order.ChildrenMap()
	//	orderBook[orderParsed["id"].Data().(float64)] = Order{Id: orderParsed["id"].Data().(float64), Side: orderParsed["side"].Data().(string), Size: orderParsed["size"].Data().(float64), Price: orderParsed["price"].Data().(float64)}
	//}
	//fmt.Printf("%#v", orderBook)

	waiter := <-queue
	println(waiter)
}
