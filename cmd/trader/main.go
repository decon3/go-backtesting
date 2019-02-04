package main

import (
	"log"
	"os"
	"pkg/mm"
	"pkg/quotes"
	"strconv"
	"time"
)

func main() {
	portfolioHistory := []*mm.Portfolio{}
	tradeHistory := []*mm.TradeHistoryEntry{}
	investment, _ := strconv.ParseFloat(os.Args[1], 64)
	portfolio := mm.NewPortfolio(investment, investment)
	trades := mm.LoadTrades("trades.csv")
	for _, t := range trades {
		high, low := GetChannel(t.Symbol, t.Date)
		t.Execute(t.Price, t.Size, high, low) // grades the execution
		if t.Action == mm.Buy {
			portfolio.Buy(t.Symbol, *t)
		} else {
			buys, pnl, err := portfolio.Sell(t.Symbol, *t)
			if err != nil {
				for _, buy := range buys {
					the := &mm.TradeHistoryEntry{
						Capital:     buy.Decision.Capital,
						BoughtOn:    buy.Date,
						BoughtPrice: buy.Price,
						SoldOn:      t.Date,
						SoldPrice:   t.Price,
						Pnl:         pnl,
						Size:        t.Size,
						Cost:        buy.Cost,
						StopLoss:    buy.Decision.StopLoss,
						Target:      buy.Decision.Target,
						BuyGrade:    buy.Grade,
						SellGrade:   t.Grade,
					}
					tradeHistory = append(tradeHistory, the)
				}
			}
		}
		portfolioHistory = append(portfolioHistory, portfolio.Clone())
		log.Printf("Investment:%.2f Capital:%.2f AssetValue:%.2f ", portfolio.Investment, portfolio.Capital, portfolio.AssetValue)
	}
}

var quotecache map[string]*quotes.QuoteData = make(map[string]*quotes.QuoteData, 1)

func GetChannel(symbol string, date time.Time) (high float64, low float64) {
	a, ok := quotecache[symbol]
	if ok == false {
		a = quotes.LoadFromFile(symbol+".NS.aqh", symbol, 3)
		if a != nil {
			quotecache[symbol] = a
		}
	}
	if a == nil {
		return 0, 0
	}
	found := -1
	if len(a.Dates) == 0 {
		return 0, 0
	}
	for i, v := range a.Dates {
		if v == date {
			found = i
			break
		}
	}
	if found == -1 {
		return 0, 0
	}
	high = a.Highs[found]
	low = a.Lows[found]
	return high, low
}
