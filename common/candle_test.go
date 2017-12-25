package common

import (
	"math"
	"testing"
)

func generateCandleChart() *CandleChart {
	candleChart := CreateNewCandleChart()
	candleChart.AddCandle(Candle{
		High:   24.60,
		Low:    24.20,
		Close:  24.28,
		Volume: 18000,
	})
	candleChart.AddCandle(Candle{
		High:   24.48,
		Low:    24.24,
		Close:  24.33,
		Volume: 7200,
	})
	candleChart.AddCandle(Candle{
		High:   24.56,
		Low:    23.43,
		Close:  24.44,
		Volume: 12000,
	})
	candleChart.AddCandle(Candle{
		High:   25.16,
		Low:    24.25,
		Close:  25.05,
		Volume: 20000,
	})
	return candleChart
}

func TestGetPastRelativeCandle(t *testing.T) {
	// GIVEN
	// CandleChart
	candleChart := generateCandleChart()

	// WHEN
	candle0 := candleChart.GetPastRelativeCandle(0)
	candle1 := candleChart.GetPastRelativeCandle(-1)
	candle2 := candleChart.GetPastRelativeCandle(-2)
	candleNilTooOld := candleChart.GetPastRelativeCandle(-100)
	candleNilPositive := candleChart.GetPastRelativeCandle(1)
	currentCandle := candleChart.CurrentCandle()

	// THEN
	if candle0 != currentCandle {
		t.Errorf("Candle0 was not the same as current candle")
	}
	if candle1.High != 24.56 && candle1.Close != 24.44 {
		t.Errorf("Candle1 was not properly found :%#v", candle1)
	}
	if candle2.High != 24.48 && candle1.Close != 24.33 {
		t.Errorf("Candle2 was not properly found :%#v", candle1)
	}
	if candleNilTooOld != nil {
		t.Errorf("candleNilTooOld was not nil")
	}
	if candleNilPositive != nil {
		t.Errorf("CandleNilPositive was not nil")
	}
}

func TestCompleteCurrentCandle(t *testing.T) {
	// GIVEN
	// CandleChart
	candleChart := generateCandleChart()

	// WHEN
	candleChart.CompleteCurrentCandle()
	candle := candleChart.CurrentCandle()

	// THEN
	// MFI test
	if math.Ceil(candle.Indicators["mfi"]) != 67 {
		t.Errorf("Candle MFI not correct %#v", candle)
	}
}
