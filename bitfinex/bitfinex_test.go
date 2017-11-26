package bitfinex

import (
	"fmt"
	"github.com/Jeffail/gabs"
	"testing"
	"thierry/gocoin/common"
)

func generateSnapshotMessage() *gabs.Container {
	jsonParsed, _ := gabs.ParseJSON([]byte(`[33919,[[8861.5,1,0.011275],[8857.9,1,0.1],[8856.6,2,0.27655864]]]`))
	return jsonParsed
}

func generateChangeMessage(side string, price string, size string) *gabs.Container {
	if side == "sell" {
		size = "-" + size
	}
	jsonParsed, _ := gabs.ParseJSON([]byte(`[33919,[` + price + `,12,` + size + `]]`))
	return jsonParsed
}

func TestUpdateOrderBookSnapshot(t *testing.T) {
	// GIVEN
	// Orderbook snapshot
	message := generateSnapshotMessage()
	// Order book
	orderBook := map[string]*common.Order{}

	// WHEN
	updateOrderBook(message, orderBook)

	// THEN
	if len(orderBook) != 3 {
		t.Errorf("Order book should have length %d, has length %d", 3, len(orderBook))
	}
}

func TestUpdateOrderBookUpdateChange(t *testing.T) {
	// GIVEN
	// Orderbook changes
	messageBuy1 := generateChangeMessage("buy", "1025", "3")
	messageBuy2 := generateChangeMessage("buy", "1026", "3")
	messageSell1 := generateChangeMessage("sell", "1030", "3")
	messageSell2 := generateChangeMessage("sell", "1031", "3")
	// Change one buy price
	messageBuy3 := generateChangeMessage("buy", "1026", "3")
	orderBook := map[string]*common.Order{}

	// WHEN
	updateOrderBook(messageBuy1, orderBook)
	updateOrderBook(messageBuy2, orderBook)
	updateOrderBook(messageSell1, orderBook)
	updateOrderBook(messageSell2, orderBook)
	updateOrderBook(messageBuy3, orderBook)

	// THEN
	if len(orderBook) != 4 {
		t.Errorf("Order book should have length %d, has length %d", 4, len(orderBook))
	}
}

func TestBestPrices(t *testing.T) {
	// GIVEN
	message, _ := gabs.ParseJSON([]byte(`[33919,[[1000.5,1,1.5],[1005.3,1,1.1],[1100.32,1,-0.3]]]`))
	// Send changes
	orderBook := map[string]*common.Order{}
	updateOrderBook(message, orderBook)
	prices := map[string][]float64{"bitfinex": {0, 0, 0, 0}}

	// WHEN
	updateBestPrices(orderBook, prices)

	// THEN
	if prices["bitfinex"][0] != 1100.32 {
		t.Errorf("Wrong best buy price at %f, wanted %f", 1100.32, prices["bitfinex"][0])
	}
	if prices["bitfinex"][2] != 1005.3 {
		t.Errorf("Wrong best sell price at %f, wanted %f", 1005.3, prices["bitfinex"][2])
	}
}

func TestBestPricesWithChanges(t *testing.T) {
	// GIVEN
	message, _ := gabs.ParseJSON([]byte(`[33919,[[1020,1,5],[1033,1,-8]]]`))
	// Orderbook changes
	messageBuy1 := generateChangeMessage("buy", "1020", "0")
	messageBuy2 := generateChangeMessage("buy", "1025", "2")
	messageSell1 := generateChangeMessage("sell", "1033", "0")
	messageSell2 := generateChangeMessage("sell", "1032", "3")
	messageSell3 := generateChangeMessage("sell", "1031", "3")
	// Send changes
	orderBook := map[string]*common.Order{}
	updateOrderBook(message, orderBook)
	updateOrderBook(messageBuy1, orderBook)
	updateOrderBook(messageBuy2, orderBook)
	updateOrderBook(messageSell1, orderBook)
	updateOrderBook(messageSell2, orderBook)
	updateOrderBook(messageSell3, orderBook)
	prices := map[string][]float64{"bitfinex": {0, 0, 0, 0}}

	// WHEN
	updateBestPrices(orderBook, prices)
	fmt.Printf("%#v\n", prices)

	// THEN
	if prices["bitfinex"][0] != 1031.0 {
		t.Errorf("Wrong best buy price at %f, wanted %f", 1031.0, prices["bitfinex"][0])
	}
	if prices["bitfinex"][2] != 1025.0 {
		t.Errorf("Wrong best sell price at %f, wanted %f", 1025.0, prices["bitfinex"][2])
	}
}
