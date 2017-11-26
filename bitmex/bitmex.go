package bitmex

import (
	"fmt"
	"github.com/Jeffail/gabs"
	ws "github.com/gorilla/websocket"
	// "os"
	"strconv"
	"thierry/gocoin/common"
)

func Update(orderBook map[string]*common.Order, prices map[string][]float64) {
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
		if order.Side == "Sell" && (buy > order.Price || buy == 0.0) && order.Size > 0 {
			buy = order.Price
			buySize = order.Size
		} else if order.Side == "Buy" && sell < order.Price && order.Size > 0 {
			sell = order.Price
			sellSize = order.Size
		}
	}
	prices["bitmex"][0] = buy
	prices["bitmex"][1] = buySize
	prices["bitmex"][2] = sell
	prices["bitmex"][3] = sellSize
}

func updateOrderBook(jsonParsed *gabs.Container, orderBook map[string]*common.Order) {
	table, ok := jsonParsed.Search("table").Data().(string)
	if !ok || table != "orderBookL2" {
		return
	}
	action, _ := jsonParsed.Search("action").Data().(string)
	row, err := jsonParsed.Search("data").Children()
	if err != nil {
		fmt.Println(err)
		return
	}
	if action == "delete" {
		for _, order := range row {
			orderParsed, _ := order.ChildrenMap()
			delete(orderBook, strconv.FormatFloat(orderParsed["id"].Data().(float64), 'f', common.PRECISION_DECIMAL, 64))
		}

	} else {
		for _, order := range row {
			orderParsed, _ := order.ChildrenMap()
			id := strconv.FormatFloat(orderParsed["id"].Data().(float64), 'f', common.PRECISION_DECIMAL,  64)
			if _, ok := orderBook[id]; !ok {
				orderBook[id] = &common.Order{Id: id}
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
}
