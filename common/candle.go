package common

import (
	"time"
)

var NUM_CANDLE = 60

type Candle struct {
	Time       time.Time
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Average    float64
	Volume     float64
	Indicators map[string]float64
}

type CandleChart struct {
	Chart    []Candle
	currElem int
}

// Public

func (chart *CandleChart) AddCandle(candle Candle) {
	// TODO improve, this should not happen often, but check if we're missing candles
	// (case of no trades for some time)
	chart.currElem += 1
	if chart.currElem == NUM_CANDLE {
		chart.currElem = 0
	}
	chart.Chart[chart.currElem] = candle
}

func (chart *CandleChart) CurrentCandle() Candle {
	return chart.Chart[chart.currElem]
}

func (chart *CandleChart) UpdateCurrentCandle(price, size float64) {
	chart.updateCandle(price, size, chart.currElem)
}

func (chart *CandleChart) UpdatePreviousCandle(price, size float64) {
	chart.updateCandle(price, size, chart.getPreviousElemId())
}

func (chart *CandleChart) CompleteCurrentCandle() {
}

func CreateNewCandleChart() *CandleChart {
	return &CandleChart{currElem: 0, Chart: make([]Candle, NUM_CANDLE)}
}

// Private

func (chart *CandleChart) updateCandle(price, size float64, i int) {
	candle := &chart.Chart[i]

	if price > candle.High {
		candle.High = price
	} else if price < candle.Low {
		candle.Low = price
	}
	// Update close to last price
	candle.Close = price
	// Also update average
	candle.Average = (candle.High + candle.Low + candle.Open + candle.Close) / 4
	// And volume
	candle.Volume += size
}

func (chart *CandleChart) getPreviousElemId() int {
	i := chart.currElem - 1
	if i == -1 {
		i = len(chart.Chart)
	}
	return i
}
