package gdax

import (
	"fmt"
	"testing"
	"thierry/gocoin/common"
)

func generateSnapshotMessage() GdaxMessage {
	return GdaxMessage{
		Type: "snapshot",
		Bids: [][]string{
			{"1000", "2"},
			{"1020", "3"},
			{"1029", "5"},
		},
		Asks: [][]string{
			{"1033", "8"},
		},
	}
}

func generateChangeMessage(side string, price string, size string) GdaxMessage {
	return GdaxMessage{
		Type: "l2update",
		Changes: [][]string{
			{side, price, size},
		},
	}
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
	if len(orderBook) != 4 {
		t.Errorf("Order book should have length %d, has length %d", 4, len(orderBook))
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
	message := GdaxMessage{
		Type: "snapshot",
		Bids: [][]string{
			{"1020", "5"},
		},
		Asks: [][]string{
			{"1033", "8"},
		},
	}
	// Send changes
	orderBook := map[string]*common.Order{}
	updateOrderBook(message, orderBook)
	prices := map[string][]float64{"gdax": {0, 0, 0, 0}}

	// WHEN
	updateBestPrices(orderBook, prices)

	// THEN
	if prices["gdax"][0] != 1033.0 {
		t.Errorf("Wrong best buy price at %f, wanted %f", 1033.0, prices["gdax"][0])
	}
	if prices["gdax"][2] != 1020.0 {
		t.Errorf("Wrong best sell price at %f, wanted %f", 1020.0, prices["gdax"][2])
	}
}

func TestBestPricesWithChanges(t *testing.T) {
	// GIVEN
	message := GdaxMessage{
		Type: "snapshot",
		Bids: [][]string{
			{"1020", "5"},
		},
		Asks: [][]string{
			{"1033", "8"},
		},
	}
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
	prices := map[string][]float64{"gdax": {0, 0, 0, 0}}

	// WHEN
	updateBestPrices(orderBook, prices)
	fmt.Printf("%#v\n", prices)

	// THEN
	if prices["gdax"][0] != 1031.0 {
		t.Errorf("Wrong best buy price at %f, wanted %f", 1031.0, prices["gdax"][0])
	}
	if prices["gdax"][2] != 1025.0 {
		t.Errorf("Wrong best sell price at %f, wanted %f", 1025.0, prices["gdax"][2])
	}
}
