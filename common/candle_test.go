package common

import (
	"github.com/shopspring/decimal"
	"testing"
)

func generateCandleChart() *CandleChart {
	candleChart := CreateNewCandleChart()
	candleChart.AddCandle(Candle{
		Open:   decimal.NewFromFloat(24.00),
		High:   decimal.NewFromFloat(24.60),
		Low:    decimal.NewFromFloat(24.20),
		Close:  decimal.NewFromFloat(24.28),
		Volume: decimal.NewFromFloat(18000),
	})
	candleChart.AddCandle(Candle{
		Open:   decimal.NewFromFloat(24.00),
		High:   decimal.NewFromFloat(24.48),
		Low:    decimal.NewFromFloat(24.24),
		Close:  decimal.NewFromFloat(24.33),
		Volume: decimal.NewFromFloat(7200),
	})
	candleChart.AddCandle(Candle{
		Open:   decimal.NewFromFloat(24.00),
		High:   decimal.NewFromFloat(24.56),
		Low:    decimal.NewFromFloat(23.43),
		Close:  decimal.NewFromFloat(24.44),
		Volume: decimal.NewFromFloat(12000),
	})
	candleChart.AddCandle(Candle{
		Open:   decimal.NewFromFloat(24.00),
		High:   decimal.NewFromFloat(25.16),
		Low:    decimal.NewFromFloat(24.25),
		Close:  decimal.NewFromFloat(25.05),
		Volume: decimal.NewFromFloat(20000),
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
	if candle1.High.Cmp(decimal.NewFromFloat(24.56)) != 0 || candle1.Close.Cmp(decimal.NewFromFloat(24.44)) != 0 {
		t.Errorf("Candle1 was not properly found: %v", candle1)
	}
	if candle2.High.Cmp(decimal.NewFromFloat(24.48)) != 0 || candle2.Close.Cmp(decimal.NewFromFloat(24.33)) != 0 {
		t.Errorf("Candle2 was not properly found: %v", candle2)
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
	candleChart := generateCandleChart()

	// WHEN
	candleChart.CompleteCurrentCandle()
	candle := candleChart.CurrentCandle()

	// THEN
	// MFI test
	if candle.Indicators["mfi"].Ceil().Cmp(decimal.NewFromFloat(67.0)) != 0 {
		t.Errorf("Candle MFI not correct %#v", candle)
	}
}

func TestCalculateEma(t *testing.T) {
	// GIVEN
	// num := []float64{277.8, 278.78, 278.94, 280}

	// // WHEN
	// res := CalculateEma(num, 277.0225)

	// THEN
	// Approximation by 100 should be exact
	// if Math.Floor(res[0] * 100) != 27733
}
