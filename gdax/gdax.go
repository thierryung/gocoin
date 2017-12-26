package gdax

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"os"
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
	Size          string     `json:"size"`
	Price         string     `json:"price"`
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

func Update() {
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.gdax.com", nil)
	if err != nil {
		println(err.Error())
	}

	var candleCharts map[string]*common.CandleChart = make(map[string]*common.CandleChart)

	subscribe := GdaxSubscribe{
		Type:       "subscribe",
		Channels:   []map[string]string{ /*{"name": "level2"}, */ {"name": "matches"}},
		ProductIds: []string{"BTC-USD", "LTC-USD", "ETH-USD"},
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
		if message.Type == "match" {
			updateMatch(message, candleCharts)

		} else {
			// updateOrderBook(message, orderBook)
		}

		// Get highest buy (that we can sell to) / lowest sell (that we can buy from) price
		counter += 1
		// TODO: Reduce the counter cooldown
		if counter%10 == 0 {
			// updateBestPrices(orderBook, prices)
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

func updateMatch(message GdaxMessage, candleCharts map[string]*common.CandleChart) {
	// Check if we're still in the current minute, or we need a new one
	// Note: There is a small possibility of misattributing the match to the wrong candle,
	// we're still doing some logic processing to attribute match to current or past candle,
	// but not more than that (e.g. issues could appear if match received is older than a minute)
	t := message.Time
	productId := message.ProductId
	if _, ok := candleCharts[productId]; !ok {
		candleCharts[productId] = common.CreateNewCandleChart()
	}
	candleChart := candleCharts[productId]

	// Decimal package
	price, _ := decimal.NewFromString(message.Price)
	size, _ := decimal.NewFromString(message.Size)

	// Check if message belongs to this current or previous candle
	currentCandle := candleChart.CurrentCandle()
	if currentCandle.Time.UnixNano()+60000000000 <= t.UnixNano() {

		// Update current candle with indicators
		candleChart.CompleteCurrentCandle()

		// Following output could be improved. Right now we are waiting for the next message
		// to indicate a new candle, and possibly loosing a few seconds of headstart.
		go output(message.ProductId, *currentCandle)

		// Add new candle
		candleChart.AddCandle(common.Candle{
			Time:    t.Truncate(time.Minute),
			Open:    price,
			High:    price,
			Low:     price,
			Close:   price,
			Average: price,
			Volume:  size,
		})
	} else {
		// We're handling the edge case here, if we already created a new current candle, but
		// for some reason we just got a match from past minute, we need to handle past minute candle
		if currentCandle.Time.UnixNano() > t.UnixNano() {
			candleChart.UpdatePreviousCandle(price, size)
		} else {
			candleChart.UpdateCurrentCandle(price, size)
		}
	}
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

func output(productId string, candle common.Candle) {
	if candle.Time.Unix() < 0 {
		return
	}
	file, err := os.OpenFile(productId+".txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if _, err = file.WriteString(fmt.Sprintf("%d %f %f %f %f %f %f %f %f %f\n",
		candle.Time.Unix(),
		candle.Open,
		candle.High,
		candle.Low,
		candle.Close,
		candle.Average,
		candle.Volume,
		candle.Indicators["mfi"],
		candle.Indicators["macd"],
		candle.Indicators["macdh"])); err != nil {
		fmt.Printf("ERROR WHILE WRITING")
	}
	fmt.Printf("%s %s %f %f %f %f %f %f %f %f %f\n",
		productId,
		candle.Time.Format(time.RFC3339),
		candle.Open,
		candle.High,
		candle.Low,
		candle.Close,
		candle.Average,
		candle.Volume,
		candle.Indicators["mfi"],
		candle.Indicators["macd"],
		candle.Indicators["macdh"],
	)
}
