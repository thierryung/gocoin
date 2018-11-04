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
	Chart    []float64
	CurrElem int
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
	// Start from next element (which should be the very beginning)
	currPos := fifo.CurrElem + 1
	if currPos == len(fifo.Chart) {
		currPos = 0
	}
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

// Send in starting money, returns number of shares, and fee
func buy(strat string, startingMoney, price, fee decimal.Decimal, t time.Time) (decimal.Decimal, decimal.Decimal) {
	res := startingMoney.Div(price)
	feeCalc := startingMoney.Mul(fee)
	fmt.Printf("%s: Buying at %s price for a total of %s. Fee %s. (%s)\n", t.Format("2006-01-02 15:04"), price, startingMoney, feeCalc, strat)
	return res, feeCalc
}

// Send in number of shares, returns money gained
func sell(strat string, numShare, price, fee decimal.Decimal, t time.Time, currentlyHoldingCandle int) (decimal.Decimal, decimal.Decimal) {
	res := numShare.Mul(price)
	feeCalc := fee.Mul(res)
	fmt.Printf("%s: Selling %s price for a total of %s, minus fee of %s (%s - %d)\n", t.Format("2006-01-02 15:04"), price, res, feeCalc, strat, currentlyHoldingCandle)
	return res, feeCalc
}

func output(candle *common.Candle) {
	if candle.Time.Unix() < 0 {
		return
	}
	file, err := os.OpenFile("res.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if _, err = file.WriteString(fmt.Sprintf("%s %s %s %s %s %s %s %f %f %f\n",
		candle.Time.Format(time.RFC3339),
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
}

func main() {

	startingMoney := decimal.NewFromFloat(1000.0)
	feeBuy := decimal.NewFromFloat(0.0)
	feeSell := decimal.NewFromFloat(0.0)

	candleChart := common.CreateNewCandleChart()

	// Read through a log file, grab data and add candles one at a time
	// file, err := os.Open("LTC-USD.txt")
	file, err := os.Open("data/december_gdax.txt")
	defer file.Close()

	if err != nil {
		fmt.Printf("Error opening file %v\n", err)
		return
	}

	reader := bufio.NewReader(file)

	var line string
	counter := 0
	lastSellCounter := 0
	lastMacdh := 0.0
	volumeData := &Fifo{CurrElem: 0, Chart: make([]float64, 5)}
	macdhData := &Fifo{CurrElem: 0, Chart: make([]float64, 3)}
	var numShare, lastPrice, feeCalc, gain decimal.Decimal
	var lastTime time.Time
	currentlyHoldingCandle := 0
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		if line == "" {
			continue
		}
		data := strings.Split(strings.Trim(line, "\n"), " ")
		t := stringToTime(data[0])
		open, _ := decimal.NewFromString(data[1])
		high, _ := decimal.NewFromString(data[2])
		low, _ := decimal.NewFromString(data[3])
		clos, _ := decimal.NewFromString(data[4])
		// average, _ := decimal.NewFromString(data[5])
		volume, _ := decimal.NewFromString(data[6])
		// mfi, _ := decimal.NewFromString(data[7])
		// macd, _ := decimal.NewFromString(data[8])
		// macdh, _ := decimal.NewFromString(data[9])

		if t.Before(lastTime) {
			panic(fmt.Sprintf("We got wrong time %s (%d) before %s (%d)\n", t, t.Unix(), lastTime, lastTime.Unix()))
		}
		lastTime = t

		candleChart.AddCandle(common.Candle{
			Time:   t,
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
			if currentlyHoldingCandle > 0 {
				// Sell when mfi > 90
				if candle.Indicators["mfi"] > 90 {
					gain, feeCalc = sell("mfi over 90", numShare, clos, feeSell, t, currentlyHoldingCandle)
					startingMoney = gain.Sub(feeCalc)
					currentlyHoldingCandle = 0
					lastSellCounter = counter
					// Sell when mfi > 0 and price decrease
				} else if candle.Indicators["mfi"] > 80 && clos.Cmp(lastPrice) < 0 {
					gain, feeCalc = sell("mfi over 80 and price decrease", numShare, clos, feeSell, t, currentlyHoldingCandle)
					startingMoney = gain.Sub(feeCalc)
					currentlyHoldingCandle = 0
					lastSellCounter = counter
					// Sell when macd < 0.2 and not growing
				} else if candle.Indicators["macdh"] < 0.2 && candle.Indicators["macdh"] < lastMacdh {
					gain, feeCalc = sell("mfi over 80 and price decrease", numShare, clos, feeSell, t, currentlyHoldingCandle)
					startingMoney = gain.Sub(feeCalc)
					currentlyHoldingCandle = 0
					lastSellCounter = counter
					// Sell when macd < 0.15
				} else if candle.Indicators["macdh"] < 0.15 {
					gain, feeCalc = sell("macdh < 0.15", numShare, clos, feeSell, t, currentlyHoldingCandle)
					startingMoney = gain.Sub(feeCalc)
					currentlyHoldingCandle = 0
					lastSellCounter = counter
				} else {
					currentlyHoldingCandle += 1
				}

				// If not, look at the buy strategy
			} else {
				// fmt.Printf("mfi: %f, macdh: %f\n", candle.Indicators["mfi"], candle.Indicators["macdh"])
				// Filter out when mfi > 60
				if candle.Indicators["mfi"] > 60 {
					// Filter out when we just sold last tick
				} else if lastSellCounter+1 == counter {
					// Filter out when volume is lower than average
				} else if volFloat < volumeData.GetAverage() {
					// Buy when macd > 20
				} else if candle.Indicators["macdh"] >= 0.20 {
					numShare, feeCalc = buy("macdh over 0.2", startingMoney, clos, feeBuy, t)
					startingMoney = decimal.NewFromFloat(0).Sub(feeCalc)
					currentlyHoldingCandle = 1
					// Buy when macd > 10 and increasing macdh
					// } else if candle.Indicators["macdh"] >= 0.10 && macdhData.IsIncreasingFromPositive() {
				} else if false && candle.Indicators["macdh"] >= 0.10 && candle.Indicators["macdh"] > lastMacdh {
					numShare, feeCalc = buy("increasing macdh over 0.1", startingMoney, clos, feeBuy, t)
					startingMoney = decimal.NewFromFloat(0).Sub(feeCalc)
					currentlyHoldingCandle = 1
				}
			}
			// output(candle)
		}

		lastPrice = clos
		lastMacdh = candle.Indicators["macdh"]

	}

	if currentlyHoldingCandle > 0 {
		gain, feeCalc = sell("mfi over 80 and price decrease", numShare, candleChart.CurrentCandle().Close, feeSell, candleChart.CurrentCandle().Time, currentlyHoldingCandle)
		startingMoney = gain.Sub(feeCalc)
	}

	// Strategy run complete, display gain loss
	fmt.Printf("\n\n==================\n\nWe ended with %s (%d candles)\n\n", startingMoney, counter)

	if err != io.EOF {
		fmt.Printf(" > Failed!: %v\n", err)
	}
}
