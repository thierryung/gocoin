package main

import (
	"fmt"
	"runtime"
	// "github.com/Jeffail/gabs"
	// "os"
	// "thierry/gocoin/bitfinex"
	// "thierry/gocoin/bitmex"
	// "thierry/gocoin/common"
	"thierry/gocoin/gdax"
	"time"
)

func timer(prices map[string][]float64) {
	for true {
		// if prices["gdax"][0] <= 1 ||
		// 	prices["gdax"][2] <= 1 ||
		// 	prices["bitfinex"][0] <= 1 ||
		// 	prices["bitfinex"][2] <= 1 ||
		// 	prices["bitmex"][0] <= 1 ||
		// 	prices["bitmex"][2] <= 1 {
		// 	continue
		// }
		fmt.Printf("Gdax - %f (%f) - %f (%f)\n",
			prices["gdax"][0],
			prices["gdax"][1],
			prices["gdax"][2],
			prices["gdax"][3],
		)
		// fmt.Printf("Bitfinex - %f (%f) - %f (%f)\n",
		// 	prices["bitfinex"][0],
		// 	prices["bitfinex"][1],
		// 	prices["bitfinex"][2],
		// 	prices["bitfinex"][3],
		// )
		// fmt.Printf("Bitmex - %f (%f) - %f (%f)\n",
		// 	prices["bitmex"][0],
		// 	prices["bitmex"][1],
		// 	prices["bitmex"][2],
		// 	prices["bitmex"][3],
		// )
		// file, err := os.OpenFile("output.txt", os.O_APPEND|os.O_WRONLY, 0600)
		// if err != nil {
		// 	panic(err)
		// }
		// defer file.Close()
		// if _, err = file.WriteString(fmt.Sprintf("%s###%f###%f###%f###%f###%f###%f\n",
		// 	time.Now().Format(time.RFC3339),
		// 	prices["gdax"][0], prices["gdax"][2],
		// 	prices["bitfinex"][0], prices["bitfinex"][2],
		// 	prices["bitmex"][0], prices["bitmex"][2])); err != nil {
		// 	panic(err)
		// }
		time.Sleep(1 * time.Second)
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
	// bitmexOrderBook := map[string]*common.Order{}
	// bitfinexOrderBook := map[string]*common.Order{}
	// gdaxOrderBook := map[string]*common.Order{}

	// Buy Price, Size Avail, Sell Price, Size Avail, Last Ticker Price, Volume
	// prices := map[string][]float64{
	// 	"gdax":     {0, 0, 0, 0, 0, 0},
	// 	// "bitfinex": {0, 0, 0, 0},
	// 	// "bitmex":   {0, 0, 0, 0},
	// }
	go gdax.Update()//gdaxOrderBook, prices)
	// go bitfinex.Update(bitfinexOrderBook, prices)
	// go bitmex.Update(bitmexOrderBook, prices)
	// go timer(prices)
	// go mem()

	// orderBook := map[float64]Order{}
	// jsonParsed, _ := gabs.ParseJSON([]byte(`[33919,[[8861.5,1,0.011275],[8857.9,1,0.1],[8856.6,2,0.27655864]]]`))
	// row := jsonParsed.Index(1)
	// data, _ := row.Index(0).ArrayCount()
	// if data > 0 {
	// 	listSnapshot, _ := row.Children()
	// 	for _, order := range listSnapshot {
	// 		fmt.Printf("%#v\n", order.Index(0).Data())
	// 	}
	// }
	// row, err := jsonParsed.Search("data").Children()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// for _, order := range row {
	// 	orderParsed, _ := order.ChildrenMap()
	// 	orderBook[orderParsed["id"].Data().(float64)] = Order{Id: orderParsed["id"].Data().(float64), Side: orderParsed["side"].Data().(string), Size: orderParsed["size"].Data().(float64), Price: orderParsed["price"].Data().(float64)}
	// }
	// fmt.Printf("%#v", orderBook)

	waiter := <-queue
	println(waiter)
}
