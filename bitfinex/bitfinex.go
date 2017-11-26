package bitfinex

import (
	"fmt"
	"github.com/Jeffail/gabs"
	ws "github.com/gorilla/websocket"
	"strconv"
	"thierry/gocoin/common"
)

func Update(orderBook map[string]*common.Order, prices map[string][]float64) {
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://api.bitfinex.com/ws/2", nil)
	if err != nil {
		println(err.Error())
	}

	subscribe := map[string]string{
		"event":   "subscribe",
		"channel": "book",
		"symbol":  "tBTCUSD",
		"prec":    "P0",
		"freq":    "F0",
	}
	if err := wsConn.WriteJSON(subscribe); err != nil {
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
			continue
		}

		updateOrderBook(jsonParsed, orderBook)

		// Get highest buy price, so we can short sell it
		counter += 1
		if counter%1 == 0 {
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
	prices["bitfinex"][0] = buy
	prices["bitfinex"][1] = buySize
	prices["bitfinex"][2] = sell
	prices["bitfinex"][3] = sellSize
}

func updateOrderBook(jsonParsed *gabs.Container, orderBook map[string]*common.Order) {
	// Ignore events
	exists := jsonParsed.Exists("event")
	if exists {
		return
	}

	row := jsonParsed.Index(1)
	isUpdateCheck, _ := row.Index(0).ArrayCount()
	// We have array of array, it's a snapshot
	if isUpdateCheck > 0 {
		listSnapshot, _ := row.Children()
		countBuy, countSell := 0, 0
		for _, order := range listSnapshot {
			if c, _ := order.ArrayCount(); c < 3 {
				continue
			}
			price := order.Index(0).Data().(float64)
			numOrder := order.Index(1).Data().(float64)
			size := order.Index(2).Data().(float64)
			side := "unset"
			if size > 0 {
				side = "buy"
				countBuy += 1
			} else if size < 0 {
				side = "sell"
				countSell += 1
				size = 0 - size
			}
			if numOrder == 0 {
				size = 0
			}

			id := side + "-" + strconv.FormatFloat(price, 'f', common.PRECISION_DECIMAL, 64)
			if _, ok := orderBook[id]; !ok {
				orderBook[id] = &common.Order{Id: id}
			}
			orderBook[id].Side = side
			orderBook[id].Size = size
			orderBook[id].Price = price
		}
		// Updates
	} else {
		if c, _ := row.ArrayCount(); c < 3 {
			return
		}
		price := row.Index(0).Data().(float64)
		numOrder := row.Index(1).Data()
		size := row.Index(2).Data().(float64)
		side := "unset"
		if size > 0 {
			side = "buy"
		} else if size < 0 {
			side = "sell"
			size = 0 - size
		}
		if numOrder == 0.0 {
			size = 0
		}

		id := side + "-" + strconv.FormatFloat(price, 'f', common.PRECISION_DECIMAL, 64)
		if _, ok := orderBook[id]; !ok {
			orderBook[id] = &common.Order{Id: id}
		}
		orderBook[id].Side = side
		orderBook[id].Size = size
		orderBook[id].Price = price
	}
}
