package gdax

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"strconv"
	"thierry/gocoin/common"
	"time"
)

type GdaxSubscribe struct {
	Type       string              `json:"type"`
	Channels   []map[string]string `json:"channels"`
	ProductIds []string            `json:"product_ids"`
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

func Update(orderBook map[string]*common.Order, prices map[string][]float64) {
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.gdax.com", nil)
	if err != nil {
		println(err.Error())
	}

	subscribe := GdaxSubscribe{
		Type:       "subscribe",
		Channels:   []map[string]string{{"name": "level2"}},
		ProductIds: []string{"BTC-USD"},
	}

	if err := wsConn.WriteJSON(subscribe); err != nil {
		println(err.Error())
	}

	message := GdaxMessage{}
	counter := 0

	for true {
		if err := wsConn.ReadJSON(&message); err != nil {
			println(err.Error())
			break
		}
		updateOrderBook(message, orderBook)

		// Get highest buy (that we can sell to) / lowest sell (that we can buy from) price
		counter += 1
		// TODO: Reduce the counter cooldown
		if counter%10 == 0 {
			updateBestPrices(orderBook, prices)
		}
	}
}

func updateBestPrices(orderBook map[string]*common.Order, prices map[string][]float64) {
	buy, buySize, sell, sellSize := 0.0, 0.0, 0.0, 0.0
	for _, order := range orderBook {
		if order.Side == "sell" && (buy > order.Price || buy == 0.0) && order.Size > 0 {
			buy = order.Price
			buySize = order.Size
		} else if order.Side == "buy" && sell < order.Price && order.Size > 0 {
			sell = order.Price
			sellSize = order.Size
		}
	}
	prices["gdax"][0] = buy
	prices["gdax"][1] = buySize
	prices["gdax"][2] = sell
	prices["gdax"][3] = sellSize
}

func updateOrderBook(message GdaxMessage, orderBook map[string]*common.Order) {
	var err error
	if message.Type == "snapshot" {
		for _, order := range message.Bids {
			// Gdax level2 is easier, but only provides price level data, which we're using as id
			size, price := 0.0, 0.0
			if price, err = strconv.ParseFloat(order[0], 64); err != nil {
				println(err.Error())
				continue
			}
			if size, err = strconv.ParseFloat(order[1], 64); err != nil {
				println(err.Error())
				continue
			}
			id := "buy-" + strconv.FormatFloat(price, 'f', common.PRECISION_DECIMAL, 64)
			if _, ok := orderBook[id]; !ok {
				orderBook[id] = &common.Order{Id: id, Side: "buy"}
			}
			orderBook[id].Size = size
			orderBook[id].Price = price
		}
		fmt.Printf("Processed %d bids in Gdax snapshots\n", len(message.Bids))
		for _, order := range message.Asks {
			// Gdax level2 is easier, but only provides price level data, which we're using as id
			size, price := 0.0, 0.0
			if price, err = strconv.ParseFloat(order[0], 64); err != nil {
				println(err.Error())
				continue
			}
			if size, err = strconv.ParseFloat(order[1], 64); err != nil {
				println(err.Error())
				continue
			}
			id := "sell-" + strconv.FormatFloat(price, 'f', common.PRECISION_DECIMAL, 64)
			if _, ok := orderBook[id]; !ok {
				orderBook[id] = &common.Order{Id: id, Side: "sell"}
			}
			orderBook[id].Size = size
			orderBook[id].Price = price
		}
		fmt.Printf("Processed %d asks in Gdax snapshots\n", len(message.Asks))

	} else if message.Type == "l2update" {
		for _, order := range message.Changes {
			side := order[0]
			size, price := 0.0, 0.0
			if price, err = strconv.ParseFloat(order[1], 64); err != nil {
				println(err.Error())
				continue
			}
			if size, err = strconv.ParseFloat(order[2], 64); err != nil {
				println(err.Error())
				continue
			}
			id := side + "-" + strconv.FormatFloat(price, 'f', common.PRECISION_DECIMAL, 64)
			if _, ok := orderBook[id]; !ok {
				orderBook[id] = &common.Order{Id: id}
			}
			orderBook[id].Size = size
			orderBook[id].Side = side
			orderBook[id].Price = price
		}
	} else {
		fmt.Println("Message type is " + message.Type)
	}
}
