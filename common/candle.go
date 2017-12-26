package common

import (
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

var NUM_CANDLE = 60
var NUM_INDICATOR = 10

type Candle struct {
	Time       time.Time
	Open       decimal.Decimal
	High       decimal.Decimal
	Low        decimal.Decimal
	Close      decimal.Decimal
	Average    decimal.Decimal
	Volume     decimal.Decimal
	Indicators map[string]decimal.Decimal
}

type CandleChart struct {
	Chart       []Candle
	currElem    int
	totalCandle int
}

// Public

func (candle *Candle) String() string {
	return fmt.Sprintf("Candle{Time: %v, Open: %s, High: %s, Low: %s, Close: %s, Average: %s, Volume: %s}",
		candle.Time, candle.Open, candle.High, candle.Low, candle.Close, candle.Average, candle.Volume)
}

func (chart *CandleChart) AddCandle(candle Candle) {
	// TODO improve, this should not happen often, but check if we're missing candles
	// (case of no trades for some time)
	chart.currElem += 1
	if chart.currElem == NUM_CANDLE {
		chart.currElem = 0
	}
	chart.Chart[chart.currElem] = candle
	chart.totalCandle += 1
}

func (chart *CandleChart) CurrentCandle() *Candle {
	return &chart.Chart[chart.currElem]
}

func (chart *CandleChart) GetPastRelativeCandle(i int) *Candle {
	// Simple check, since it makes no sense to look for future candles
	if i > 0 || i < 0-len(chart.Chart) {
		return nil
	}
	// Abs
	i = -i
	pos := chart.currElem
	if i > chart.currElem {
		pos = len(chart.Chart) - i - chart.currElem
	} else {
		pos = chart.currElem - i
	}
	return &chart.Chart[pos]
}

func (chart *CandleChart) UpdateCurrentCandle(price, size decimal.Decimal) {
	chart.updateCandle(price, size, chart.currElem)
}

func (chart *CandleChart) UpdatePreviousCandle(price, size decimal.Decimal) {
	chart.updateCandle(price, size, chart.getPreviousElemId())
}

// These calculations could be optimized by saving current value into the candle itself
// e.g. EMA. Currently we're recalculating past values, although assuming this isn't
// a big performance hit as they only get called once a candle complete
func (chart *CandleChart) CompleteCurrentCandle() {
	mfiConfig := 14
	// shortEma, longEma, macdEmaSignal := 10, 26, 9
	// shortEma, longEma, macdEmaSignal := 4, 5, 3

	// Calculate average price, and create indicator array
	candle := chart.CurrentCandle()
	candle.Average = (candle.High.Add(candle.Low).Add(candle.Open).Add(candle.Close)).Div(decimal.NewFromFloat(4.0))
	candle.Indicators = make(map[string]decimal.Decimal, NUM_INDICATOR)

	// Calculate MFI
	// Money flow for all related candles
	var moneyFlowPositive, moneyFlowNegative, previousPrice decimal.Decimal
	three := decimal.NewFromFloat(3)
	// Start at candle current - mfiConfig
	for i := -mfiConfig; i <= 0; i++ {
		c := chart.GetPastRelativeCandle(i)
		price := (c.High.Add(c.Low).Add(c.Close)).Div(three)
		moneyFlow := price.Mul(c.Volume)
		if price.Cmp(previousPrice) > 0 {
			moneyFlowPositive = moneyFlowPositive.Add(moneyFlow)
		} else if price.Cmp(previousPrice) < 0 {
			moneyFlowNegative = moneyFlowNegative.Add(moneyFlow)
		}
		previousPrice = price
	}
	// Money ratio
	moneyFlowRatio := moneyFlowPositive.Div(moneyFlowNegative)
	hundred := decimal.NewFromFloat(100.0)
	candle.Indicators["mfi"] = hundred.Sub(hundred.Div((decimal.NewFromFloat(1.0).Add(moneyFlowRatio))))

	// // Calculate MACD
	// // Calculate previous period SMA
	// smaShort, smaLong, smaSignal := 0.0, 0.0, 0.0
	// for i := -(longEma*2); i < longEma; i++ {
	// 	c := chart.GetPastRelativeCandle(i)
	// 	smaLong += c.Close
	// }
	// smaLong /= longEma
	// for i := -(shortEma*2); i < shortEma; i++ {
	// 	c := chart.GetPastRelativeCandle(i)
	// 	smaShort += c.Close
	// }
	// smaShort /= shortEma
	// for i := -(signalEma*2); i < longEma; i++ {
	// 	c := chart.GetPastRelativeCandle(i)
	// 	smaLong += c.Close
	// }

	// emaShort, emaLong, emaSignal, yesterdayEmaShort, yesterdayEmaLong, yesterdayEmaSignal := 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
	// for i := -longEma; i <= 0; i++ {
	// 	// Long EMA
	// 	c := chart.GetPastRelativeCandle(i)
	// 	emaLong = calculateSingleEma(c.Close, longEma, yesterdayEmaLong)
	// 	yesterdayEmaLong = emaLong
	// 	// Short EMA
	// 	if i >= -shortEma {
	// 		emaShort = calculateSingleEma(c.Close, shortEma, yesterdayEmaShort)
	// 		yesterdayEmaShort = emaShort
	// 	}
	// 	// Signal EMA
	// 	if i >= -macdEmaSignal {
	// 		emaSignal = calculateSingleEma(emaShort-emaLong, macdEmaSignal, yesterdayEmaSignal)
	// 		yesterdayEmaSignal = emaSignal
	// 	}
	// }
	// macd := emaShort - emaLong
	// candle.Indicators["macd"] = macd
	// candle.Indicators["macdh"] = macd - emaSignal
	// fmt.Printf("emaShort %f, emaLong %f, macd %f, signal %f\n", emaShort, emaLong, macd, emaSignal)
}

func CreateNewCandleChart() *CandleChart {
	return &CandleChart{currElem: 0, Chart: make([]Candle, NUM_CANDLE)}
}

func CalculateEma(numbers []decimal.Decimal, startEma decimal.Decimal) []decimal.Decimal {
	period := len(numbers)
	emaList := make([]decimal.Decimal, period)
	lastEma := startEma
	for i := 0; i < period; i++ {
		emaList[i] = calculateSingleEma(numbers[i], period, lastEma)
		lastEma = emaList[i]
	}
	return emaList
}

// Private

func (chart *CandleChart) updateCandle(price, size decimal.Decimal, i int) {
	candle := &chart.Chart[i]

	if price.Cmp(candle.High) > 0 {
		candle.High = price
	} else if price.Cmp(candle.Low) < 0 {
		candle.Low = price
	}
	// Update close to last price
	candle.Close = price
	// And volume
	candle.Volume = candle.Volume.Add(size)
}

func (chart *CandleChart) getPreviousElemId() int {
	i := chart.currElem - 1
	if i == -1 {
		i = len(chart.Chart)
	}
	return i
}

func calculateSingleEma(price decimal.Decimal, numDays int, previousEma decimal.Decimal) decimal.Decimal {
	k := decimal.NewFromFloat(2 / (float64(numDays) + 1))
	one := decimal.NewFromFloat(1.0)
	return price.Mul(k).Add(previousEma.Mul(one.Sub(k)))
}
