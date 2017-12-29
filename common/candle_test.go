package common

import (
	"github.com/shopspring/decimal"
	"math"
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

func generateCandleChartWithInit(num int, price decimal.Decimal, volume decimal.Decimal) *CandleChart {
	candleChart := CreateNewCandleChart()
	for i := 0; i < num; i++ {
		candleChart.AddCandle(Candle{
			Open:   price,
			High:   price,
			Low:    price,
			Close:  price,
			Volume: volume,
		})
	}
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

func TestCalculateMfi(t *testing.T) {
	// GIVEN
	candleChart := generateCandleChart()

	// WHEN
	res := candleChart.CalculateMfi(14)

	// THEN
	if math.Ceil(res) != 67 {
		t.Errorf("Candle MFI not correct %f", res)
	}
}

func TestCalculateEma(t *testing.T) {
	// GIVEN
	num := []decimal.Decimal{decimal.NewFromFloat(15837),
		decimal.NewFromFloat(15808.8),
		decimal.NewFromFloat(15810),
		decimal.NewFromFloat(15826),
		decimal.NewFromFloat(15815.01),
		decimal.NewFromFloat(15801),
		decimal.NewFromFloat(15780)}

	// WHEN
	res := CalculateEma(num, 4, decimal.NewFromFloat(15841.6625))

	// THEN
	// Approximation by 1000 should be exact
	if res[len(res)-1].Floor().Cmp(decimal.NewFromFloat(15799.0)) != 0 {
		t.Errorf("%v", res)
	}
}

func TestCalculateEmaAgain(t *testing.T) {
	// GIVEN
	num := []decimal.Decimal{decimal.NewFromFloat(277.8),
		decimal.NewFromFloat(278.78),
		decimal.NewFromFloat(278.94),
		decimal.NewFromFloat(280),
		decimal.NewFromFloat(281.88),
		decimal.NewFromFloat(281.99),
		decimal.NewFromFloat(282.49),
		decimal.NewFromFloat(282.92),
		decimal.NewFromFloat(281),
		decimal.NewFromFloat(281.96)}

	// WHEN
	res := CalculateEma(num, 3, decimal.NewFromFloat(277.0225))

	// THEN
	// Approximation by 100 should be exact
	if res[len(res)-1].Mul(decimal.NewFromFloat(100.0)).Floor().Cmp(decimal.NewFromFloat(28183.0)) != 0 {
		t.Errorf("%v", res)
	}
}

func TestCalculateMacd(t *testing.T) {
	// GIVEN
	candleChart := CreateNewCandleChart()
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(60.63)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(60.56)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(60.1)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(58.94)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(58.64)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(59.56)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(58.81)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(59.67)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(58.92)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(58.18)})

	// WHEN
	macd, macdh := candleChart.CalculateMacd(4, 5, 2)

	// THEN
	// Verify approximate to a few decimals
	// We use ceil for negative numbers, and floor for positive
	if math.Ceil(macd*1000000) != -122741 {
		t.Errorf("Candle Macd not correct %f", macd)
	}
	if math.Ceil(macdh*1000000) != -12693 {
		t.Errorf("Candle Macdh not correct %f", macdh)
	}
}

func TestCalculateMacdAgain(t *testing.T) {
	// GIVEN
	candleChart := CreateNewCandleChart()
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(60.7)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(61.77)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(62.35)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(62.59)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(62.58)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(62.36)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(62.29)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(62.22)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(61.69)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(62.43)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(61.83)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(60.64)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(60.43)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(59.91)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(59.82)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(59.59)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(59.57)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(60.24)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(60.63)})
	candleChart.AddCandle(Candle{Close: decimal.NewFromFloat(60.56)})

	// WHEN
	macd, macdh := candleChart.CalculateMacd(5, 8, 4)

	// THEN
	// Verify approximate to a few decimals.
	// We use ceil for negative numbers, and floor for positive
	if math.Ceil(macd*1000000) != -97082 {
		t.Errorf("Candle Macd not correct %f", macd)
	}
	if math.Floor(macdh*1000000) != 113533 {
		t.Errorf("Candle Macdh not correct %f", macdh)
	}
}
