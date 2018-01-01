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
	Indicators map[string]float64
}

type CandleChart struct {
	Chart       []Candle
	currElem    int
	totalCandle int
}

// Public

func (candle *Candle) String() string {
	return fmt.Sprintf("Candle{Time: %v, Open: %s, High: %s, Low: %s, Close: %s, Average: %s, Volume: %s, Indicators: %v}",
		candle.Time, candle.Open, candle.High, candle.Low, candle.Close, candle.Average, candle.Volume, candle.Indicators)
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
	if i <= chart.currElem {
		pos = chart.currElem - i
	} else {
		pos = len(chart.Chart) - i + chart.currElem
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
// a big performance hit as they only get called once a candle is complete
func (chart *CandleChart) CompleteCurrentCandle() {
	mfiConfig := 14
	emaShortConfig, emaLongConfig, macdEmaSignalConfig := 10, 26, 9

	// Calculate average price, and create indicator array
	candle := chart.CurrentCandle()
	candle.Average = (candle.High.Add(candle.Low).Add(candle.Open).Add(candle.Close)).Div(decimal.NewFromFloat(4.0))
	candle.Indicators = make(map[string]float64, NUM_INDICATOR)

	// Calculate MFI
	candle.Indicators["mfi"] = chart.CalculateMfi(mfiConfig)

	// Calculate MACD. We need at least double the longest period for this.
	macd, macdh := chart.CalculateMacd(emaShortConfig, emaLongConfig, macdEmaSignalConfig)
	candle.Indicators["macd"], candle.Indicators["macdh"] = macd, macdh
}

func (chart *CandleChart) CalculateMfi(days int) float64 {
	if chart.totalCandle < days {
		return 0.0
	}
	// Money flow for all related candles
	var moneyFlowPositive, moneyFlowNegative, previousPrice decimal.Decimal
	three := decimal.NewFromFloat(3)
	// Get starter price
	candle := chart.GetPastRelativeCandle(-days)
	previousPrice = (candle.High.Add(candle.Low).Add(candle.Close)).Div(three)
	// Start at candle current - mfiConfig
	for i := -days + 1; i <= 0; i++ {
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
	moneyFlowRatio := moneyFlowPositive
	if !moneyFlowNegative.Equal(decimal.NewFromFloat(0.0)) {
		moneyFlowRatio = moneyFlowPositive.Div(moneyFlowNegative)
	}
	hundred := decimal.NewFromFloat(100.0)
	res, _ := hundred.Sub(hundred.Div((decimal.NewFromFloat(1.0).Add(moneyFlowRatio)))).Float64()
	return res
}

func (chart *CandleChart) CalculateMacd(emaShortConfig, emaLongConfig, macdEmaSignalConfig int) (float64, float64) {
	// We can only calculate MACD if we have enough candles
	if chart.totalCandle < emaLongConfig*2 || chart.totalCandle < macdEmaSignalConfig*2 {
		return 0.0, 0.0
	}
	// Calculate starter SMA
	smaList := make([]decimal.Decimal, emaLongConfig)
	emaList := make([]decimal.Decimal, emaLongConfig)
	smaNum, emaNum := 0, 0
	for i := -(emaLongConfig * 2) + 1; i <= 0; i++ {
		c := chart.GetPastRelativeCandle(i)
		// Use the first few for SMA
		if smaNum < emaLongConfig {
			smaList[smaNum] = c.Close
			smaNum += 1
			// Then for EMA
		} else {
			emaList[emaNum] = c.Close
			emaNum += 1
		}
	}
	// For short SMA, we only need a slice of the data
	smaShort := calculateSma(smaList[len(smaList)-emaShortConfig:])
	smaLong := calculateSma(smaList)

	// Calculate EMA from SMA starting point
	emaLong := CalculateEma(emaList, emaLongConfig, smaLong)
	emaShort := CalculateEma(emaList, emaShortConfig, smaShort)

	// Calculate MACD line
	macd := make([]decimal.Decimal, emaLongConfig)
	for i := 0; i < emaLongConfig; i++ {
		macd[i] = emaShort[i].Sub(emaLong[i])
	}

	// Calculate Signal line
	macdSignalSma := calculateSma(macd[:macdEmaSignalConfig])
	macdSignalList := CalculateEma(macd[macdEmaSignalConfig:], macdEmaSignalConfig, macdSignalSma)

	macdValue := macd[len(macd)-1]
	macdSignalRes := macdSignalList[len(macdSignalList)-1]
	macdhRes, _ := macdValue.Sub(macdSignalRes).Float64()
	macdRes, _ := macdValue.Float64()

	return macdRes, macdhRes
}

func CreateNewCandleChart() *CandleChart {
	return &CandleChart{currElem: 0, Chart: make([]Candle, NUM_CANDLE)}
}

func CalculateEma(numbers []decimal.Decimal, period int, startEma decimal.Decimal) []decimal.Decimal {
	l := len(numbers)
	emaList := make([]decimal.Decimal, l)
	lastEma := startEma
	for i := 0; i < l; i++ {
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

func calculateSma(numbers []decimal.Decimal) decimal.Decimal {
	var sma decimal.Decimal
	for i := 0; i < len(numbers); i++ {
		sma = sma.Add(numbers[i])
	}
	return sma.Div(decimal.NewFromFloat(float64(len(numbers))))
}
