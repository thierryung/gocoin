package main

import (
	"bufio"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
	"os"
	"strconv"
	"strings"
	"thierry/gocoin/common"
	"time"
)

type Fifo struct {
	Chart       []float64
	CurrElem    int
}

func (fifo *Fifo) AddNew(value float64) {
	fifo.CurrElem += 1
	if fifo.CurrElem == len(fifo.Chart) {
		fifo.CurrElem = 0
	}
	fifo.Chart[fifo.CurrElem] = value
}

func (fifo *Fifo) GetAverage() float64 {
	total := 0.0
	for i := 0; i < len(fifo.Chart); i++ {
		total += fifo.Chart[i]
	}
	return total / float64(len(fifo.Chart))
}

func (fifo *Fifo) IsIncreasingFromPositive() bool {
	// Starting from 0 ensure we're starting from positive value
	lastValue := 0.0
	currPos := fifo.CurrElem
	for i := 0; i < len(fifo.Chart); i++ {
		if lastValue < fifo.Chart[currPos] {
			return false
		}
		currPos += 1
		if currPos == len(fifo.Chart) {
			currPos = 0
		}
	}
	return true
}

func stringToTime(s string) time.Time {
	i, err := strconv.ParseInt(s, 10, 64)
    if err != nil {
        panic(err)
    }
    return time.Unix(i, 0)
}

// Send in starting money, returns number of shares
func buy(startingMoney decimal.Decimal, price decimal.Decimal) decimal.Decimal {
	res := startingMoney.Div(price)
	fmt.Printf("Buying %s shares at %s price for a total of %s\n", res, price, startingMoney)
	return res
}

// Send in number of shares, returns money gained
func sell(numShare decimal.Decimal, price decimal.Decimal) decimal.Decimal {
	res := numShare.Mul(price)
	fmt.Printf("Selling %s shares at %s price for a total of %s\n", numShare, price, res)
	return res
}

func main() {

	startingMoney := decimal.NewFromFloat(1000.0)

	candleChart := common.CreateNewCandleChart()

	// Read through a log file, grab data and add candles one at a time
	file, err := os.Open("LTC-USD.txt")
	defer file.Close()

	if err != nil {
		fmt.Printf("Error opening file %v\n", err)
		return
	}

	reader := bufio.NewReader(file)

	var line string
	counter := 0
	lastSellCounter := 0
	volumeData := &Fifo{CurrElem: 0, Chart: make([]float64, 5)}
	macdhData := &Fifo{CurrElem: 0, Chart: make([]float64, 3)}
	var numShare, lastPrice decimal.Decimal
	currentlyHolding := false
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		if line == "" {
			continue
		}
		data := strings.Split(strings.Trim(line, "\n"), " ")
		time := stringToTime(data[0])
		open, _ := decimal.NewFromString(data[1])
		high, _ := decimal.NewFromString(data[2])
		low, _ := decimal.NewFromString(data[3])
		clos, _ := decimal.NewFromString(data[4])
		// average, _ := decimal.NewFromString(data[5])
		volume, _ := decimal.NewFromString(data[6])
		// mfi, _ := decimal.NewFromString(data[7])
		// macd, _ := decimal.NewFromString(data[8])
		// macdh, _ := decimal.NewFromString(data[9])

		candleChart.AddCandle(common.Candle{
			Time:   time,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  clos,
			Volume: volume,
		})
		candleChart.CompleteCurrentCandle()
		candle := candleChart.CurrentCandle()
		volFloat, _ := volume.Float64()
		volumeData.AddNew(volFloat)
		macdhData.AddNew(candle.Indicators["macdh"])
		counter += 1

		// After 60 candles have been added, start running strategy
		if counter > 60 {
			// Check if we're currently holding, we're looking at the sell strategy
			if currentlyHolding {
				// Sell when mfi > 90
				if candle.Indicators["mfi"] > 90 {
					startingMoney = sell(numShare, clos)
					currentlyHolding = false
					lastSellCounter = counter
				// Sell when mfi > 0 and price decrease
				} else if candle.Indicators["mfi"] > 80 && clos.Cmp(lastPrice) < 0 {
					startingMoney = sell(numShare, clos)
					currentlyHolding = false
					lastSellCounter = counter
				}

			// If not, look at the buy strategy
			} else {
				// fmt.Printf("mfi: %f, macdh: %f\n", candle.Indicators["mfi"], candle.Indicators["macdh"])
				// Filter out when mfi > 60
				if candle.Indicators["mfi"] > 60 {
				// Filter out when we just sold last tick
				} else if lastSellCounter + 1 == counter {
				// Filter out when volume is lower than average
				} else if volFloat < volumeData.GetAverage() {
				// Buy when macd > 30
				} else if candle.Indicators["macdh"] >= 0.10 {
					numShare = buy(startingMoney, clos)
					currentlyHolding = true
				// Buy when macd > 10 and increasing macdh
				} else if candle.Indicators["macdh"] >= 0.10 && macdhData.IsIncreasingFromPositive() {
					numShare = buy(startingMoney, clos)
					currentlyHolding = true
				}
			}
		}

		lastPrice = clos

	}
	// Strategy run complete, display gain loss
	fmt.Printf("\n\n==================\n\nWe ended with %s\n\n", startingMoney)

	if err != io.EOF {
		fmt.Printf(" > Failed!: %v\n", err)
	}
}
