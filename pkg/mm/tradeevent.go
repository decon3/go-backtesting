package mm

import (
	"errors"
	"time"
)

type TradeAction int

const (
	Buy TradeAction = iota + 1
	Sell
)

type TradeEvent struct {
	Symbol   string
	Date     time.Time
	Size     int
	Price    float64
	Action   TradeAction
	Cost     float64
	Decision *TradeDecision
	High     float64
	Low      float64
	Grade    string
}

type TradeHistoryEntry struct {
	Capital     float64
	BoughtOn    time.Time
	BoughtPrice float64
	SoldOn      time.Time
	SoldPrice   float64
	Pnl         float64
	Size        int
	Cost        float64
	StopLoss    float64
	Target      float64
	BuyGrade    string
	SellGrade   string
}

type TradeDecision struct {
	Capital     float64
	CapitalRisk float64
	TradeRisk   float64
	Close       float64
	StopLoss    float64
	Target      float64
	Weekly      string
	Daily       string
	Decision    string
}

func (t *TradeEvent) Execute(price float64, size int, high float64, low float64) {
	t.Price = price
	t.Size = size
	t.High = high
	t.Low = low
	t.GradeExecution(high, low)
}

func (d *TradeDecision) CalculateSize(price float64) (int, error) {
	if d.Capital <= 0 {
		return 0, errors.New("Capital is missing")
	}
	if d.CapitalRisk <= 0 {
		return 0, errors.New("CapitalRisk is missing")
	}
	if d.CapitalRisk > 1 || d.TradeRisk > 1 {
		return 0, errors.New("CapitalRisk and TradeRisk should be less than 1")
	}
	if d.TradeRisk <= 0 {
		return 0, errors.New("TradeRisk is missing")
	}
	return int((d.Capital * d.CapitalRisk) / (price * d.TradeRisk)), nil
}

func (t *TradeEvent) ValidateTradeCompletion() error {
	if t.Date.After(time.Now()) {
		return errors.New("Date cannot be in future")
	}
	if t.Size <= 0 {
		return errors.New("Size is missing")
	}
	if t.Price <= 0 {
		return errors.New("Price has to be a positive number")
	}
	if t.Action != Buy && t.Action != Sell {
		return errors.New("Action can only be 'Buy' or 'Sell'")
	}
	if t.Grade != "A" && t.Grade != "B" && t.Grade != "C" {
		return errors.New("Trade has not been graded")
	}
	return t.Decision.Validate(t.Action)
}

func (d *TradeDecision) Validate(a TradeAction) error {
	if a == Buy {
		if d.Capital <= 0 {
			return errors.New("Capital is missing")
		}
		if d.CapitalRisk <= 0 {
			return errors.New("CapitalRisk is missing")
		}
		if d.TradeRisk <= 0 {
			return errors.New("TradeRisk is missing")
		}
		if d.StopLoss <= 0 {
			return errors.New("TradeRisk is missing")
		}
		if d.Target <= 0 {
			return errors.New("TradeRisk is missing")
		}
	}
	if d.Close <= 0 {
		return errors.New("Close is missing")
	}
	return nil
}

func (t *TradeEvent) GradeExecution(high float64, low float64) string {
	if high <= 0 || low <= 0 || high < low {
		t.Grade = "C"
		return "C"
	}
	t.High = high
	t.Low = low
	percentOfChannel := (t.High - t.Low)
	if t.Action == Buy {
		switch {
		case percentOfChannel <= 25:
			t.Grade = "A"
		case percentOfChannel <= 50:
			t.Grade = "B"
		default:
			t.Grade = "C"
		}
	} else {
		switch {
		case percentOfChannel >= 75:
			t.Grade = "A"
		case percentOfChannel >= 50:
			t.Grade = "B"
		default:
			t.Grade = "C"
		}
	}
	return t.Grade
}

func (t *TradeEvent) SoldPartially(size int, price float64) (TradeEvent, error) {
	if t.Action == Buy {
		t2 := t

		// reduce cost proportionately
		saleTurnover := float64(size) * price
		buyTurnover := float64(t.Size) * t.Price
		acquisitionCostOfSoldPortion := t.Cost * (saleTurnover / buyTurnover)
		t.Cost -= acquisitionCostOfSoldPortion

		// update remaining portion
		t.Size -= size

		// update the sold portion and return it
		t2.Size = size
		t2.Cost = acquisitionCostOfSoldPortion
		return *t2, nil
	} else {
		return TradeEvent{}, errors.New("Cannot sell: This is not a buy event")
	}
}

func (t *TradeEvent) CalculateCost(rates TradingRates) {
	turnover := float64(t.Size) * t.Price
	ttax := turnover * rates.TurnoverTax
	sebi := turnover * rates.Sebi
	stamp := turnover * rates.StampDuty
	exchange := turnover * rates.Exchange
	var brokerage, demat float64
	if t.Action == Buy {
		brokerage = rates.BrokerageBuy
	} else {
		brokerage = rates.BrokerageSell
		demat = rates.Demat
	}
	gst := (exchange + brokerage) * rates.Gst
	t.Cost = ttax + sebi + stamp + exchange + brokerage + demat + gst
}

type TradingRates struct {
	TurnoverTax   float64
	Sebi          float64
	Exchange      float64
	StampDuty     float64
	StampDutyMax  float64
	BrokerageBuy  float64
	BrokerageSell float64
	Demat         float64
	Gst           float64
}

var FivePaisaRates TradingRates = TradingRates{
	TurnoverTax:   0.00025,
	Sebi:          0.0000015,
	Exchange:      0.0000325,
	StampDuty:     0.001,
	StampDutyMax:  100,
	Gst:           0.18,
	BrokerageBuy:  10,
	BrokerageSell: 10,
	Demat:         18.5,
}
var ZerodhaRates TradingRates = TradingRates{
	TurnoverTax:   0.00025,
	Sebi:          0.0000015,
	Exchange:      0.0000325,
	StampDuty:     0.001,
	StampDutyMax:  100,
	Gst:           0.18,
	BrokerageBuy:  0,
	BrokerageSell: 0,
	Demat:         13.5,
}
