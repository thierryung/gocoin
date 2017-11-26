package common

import "time"

type Order struct {
	Id    string
	Side  string
	Size  float64
	Price float64
	Time  time.Time
}
