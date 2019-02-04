package main

import (
	//	"github.com/vdobler/chart"
	"fmt"
	"github.com/jinzhu/now"
	"io/ioutil"
	"log"
	"os"
	"pkg/quotes"
	"pkg/talib"
	"strconv"
	"strings"
	"time"
)

type MoneyManagement struct {
	Capital       float64
	RiskOnCapital float64
	RiskOnTrade   float64
	ProfitOnTrade float64
}

func main() {
	if len(os.Args) < 5 {
		panic("Usage: quoter capital risk target capitalRisk")
	}
	mm := MoneyManagement{}
	mm.Capital, _ = strconv.ParseFloat(os.Args[1], 64)
	mm.RiskOnTrade, _ = strconv.ParseFloat(os.Args[2], 64)
	mm.ProfitOnTrade, _ = strconv.ParseFloat(os.Args[3], 64)
	mm.RiskOnCapital, _ = strconv.ParseFloat(os.Args[4], 64)
	symbols := getSymbols()
	var totalProfit float64 = 0

	for _, symbol := range symbols {
		//	check(symbol)
		log.Printf("**********  %s   ***********", symbol)
		qtd := quotes.LoadFromFile(symbol+".NS.aqh", symbol, 3)
		totalProfit += BackTestMovingAverages(qtd, &mm)
	}
	log.Printf("Total profit: %.0f", totalProfit)
}

func getSymbols() []string {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	symbols := []string{}
	for _, file := range files {
		tokens := strings.Split(file.Name(), ".")
		if len(tokens) != 3 || tokens[2] != "aqh" {
			continue
		}
		symbols = append(symbols, tokens[0])
	}
	return symbols
}

/*
func check(symbol string) {
	qtd := quotes.LoadFromFile(symbol+".NS.aqh", symbol, 3)
	ema20 := talib.Ema(qtd.Closes, 20)
	ema50 := talib.Ema(qtd.Closes, 50)
	dumper := talib.NewDumper(symbol+"_ema", 2, 2, 400, 300)
	defer dumper.Close()
	c := chart.StripChart{}
	c.AddData("20 ema", ema20, chart.Style{})
	c.AddData("50 ema", ema50, chart.Style{})
	c.Title = symbol + " - ema"
	c.XRange.Label = "Closing Price"
	c.Key.Pos = "icr"
	e20 := ema20[len(ema20)-1]
	e50 := ema50[len(ema50)-1]
	close := qtd.Closes[len(qtd.Closes)-1]
		log.Printf("%s: %.2f. EMA20: %.2f EMA50: %.2f. Buillish? %v. Buy? %v Sell? %v",
			symbol, close, e20, e50, e20 > e50, BullishCrossover(qtd, ema20, ema50), BearishCrossover(qtd, ema20, ema50))
	dumper.Plot(&c)
}
*/

func BullishCrossover(qtd *quotes.QuoteData, ema20 []float64, ema50 []float64, index int) bool {
	e20 := ema20[index]
	e20a := ema20[index-1]
	e50 := ema50[index]
	e50a := ema50[index-1]
	return e20 > e50 && e20a < e50a && qtd.Closes[index] > e20
}

func BearishCrossover(qtd *quotes.QuoteData, ema20 []float64, ema50 []float64, index int) bool {
	e20 := ema20[index]
	e20a := ema20[index-1]
	e50 := ema50[index]
	e50a := ema50[index-1]
	return e20 < e50 && e20a > e50a && qtd.Closes[index] < e20
}

func BackTestMovingAverages(qtd *quotes.QuoteData, mm *MoneyManagement) float64 {
	var tradebook []*Trade = []*Trade{}
	var position *Trade = nil
	totalCloses := len(qtd.Closes)
	ema20 := talib.Sma(qtd.Closes, 20)
	ema50 := talib.Sma(qtd.Closes, 50)
	aroonUp, aroonDn := talib.Aroon(qtd.Highs, qtd.Lows, 20)
	tradingCap := mm.Capital

	for i, todayclose := range qtd.Closes {
		if todayclose < 20 {
			continue
		}
		if i <= 200 { // not the first 200 days
			continue
		}
		if i+1 == totalCloses {
			continue
		}
		if position == nil && BullishCrossover(qtd, ema20, ema50, i) && IsTrendingUp(aroonUp[i], aroonDn[i]) {
			price := qtd.Highs[i+1]
			size := CalculateTradeSize(tradingCap, price, price*mm.RiskOnTrade, mm.RiskOnCapital)
			if size < 1 { // insufficient capital
				continue
			}
			position = &Trade{
				// next day
				Date:     qtd.Dates[i+1],
				Size:     CalculateTradeSize(tradingCap, price, price*mm.RiskOnTrade, mm.RiskOnCapital),
				Ema20:    ema20[i],
				Ema50:    ema50[i],
				StopLoss: price - price*mm.RiskOnTrade,
				Target:   price + price*mm.ProfitOnTrade,
				AroonUp:  aroonUp[i],
				AroonDn:  aroonDn[i],
				Price:    price,
				Buy:      true,
			}
			position.Cost()
			tradebook = append(tradebook, position)
		}
		if position != nil &&
			HasTrendReversed(aroonUp[i], aroonDn[i]) &&
			(BearishCrossover(qtd, ema20, ema50, i) ||
				position.HasHitStopLoss(todayclose) ||
				position.HasHitTarget(todayclose)) {
			sell := &Trade{
				// next day
				Date:    qtd.Dates[i+1],
				Size:    position.Size,
				Ema20:   ema20[i],
				Ema50:   ema50[i],
				AroonUp: aroonUp[i],
				AroonDn: aroonDn[i],
				Price:   qtd.Lows[i+1],
				Buy:     false,
			}
			sell.Cost()
			sell.CalculatePL(position)
			tradebook = append(tradebook, sell)
			tradingCap += sell.PL

			log.Printf("(e20:%.2f e50:%.2f tu:%.2f td:%.2f) %s - (e20:%.2f e50:%.2f tu:%.2f td:%.2f) %s %.0f [%s] -- %.2f (%.2f)",
				position.Ema20, position.Ema50, position.AroonUp, position.AroonDn, position.String(),
				sell.Ema20, sell.Ema50, sell.AroonUp, sell.AroonDn, sell.String(),
				sell.Date.Sub(position.Date).Hours()/24, position.SellReason(todayclose),
				sell.PL, tradingCap)
			position = nil
		}
	}
	log.Printf("CAPITAL: %.2f, P/L: %.2f Total Trades:%d", tradingCap, tradingCap-mm.Capital, len(tradebook)/2)
	return tradingCap - mm.Capital
}

func IsTrendingUp(aroonUp float64, aroonDn float64) bool {
	return aroonUp > 70 && aroonDn < 40
}

func HasTrendReversed(aroonUp float64, aroonDn float64) bool {
	return aroonUp < 50 && aroonDn > 25
}

func (position *Trade) SellReason(currentPrice float64) string {
	if position.HasHitStopLoss(currentPrice) {
		return "SL"
	}
	if position.HasHitTarget(currentPrice) {
		return "Target"
	}
	return "Bearish"
}

func (position *Trade) HasHitStopLoss(currentPrice float64) bool {
	return currentPrice < position.StopLoss
}
func (position *Trade) HasHitTarget(currentPrice float64) bool {
	return currentPrice > position.Target
}
func CalculateTradeSize(capital float64, price float64, risk float64, capRisk float64) float64 {
	return float64(int(capital * capRisk / risk))
}

type Portfolio struct {
	Symbol string
	Bought *Trade
}

type Trade struct {
	Date     time.Time
	Size     float64
	Price    float64
	Buy      bool
	Ema20    float64
	Ema50    float64
	Demat    float64
	StopLoss float64
	AroonUp  float64
	AroonDn  float64
	Target   float64
	Tax      float64
	PL       float64
}

func (t *Trade) Cost() {
	turnover := t.Size * t.Price
	if t.Buy {
		ttax := turnover * .0001
		sebi := turnover * .0000325
		stamp := turnover * .0000015
		exchange := turnover * .0000034
		gst := exchange * .18
		t.Tax = ttax + sebi + stamp + exchange + gst
	} else {
		ttax := turnover * .0001
		sebi := turnover * .0000325
		stamp := turnover * .0000015
		exchange := turnover * .0000034
		t.Demat = 13.2
		gst := exchange + t.Demat*.18
		t.Tax = ttax + sebi + stamp + exchange + gst
	}
}
func (t *Trade) CalculatePL(t2 *Trade) {
	if t2.Buy == false {
		panic("No short positions allowed")
	}

	if t.Buy {
		panic("No short positions allowed")
	}

	if t.Size > t2.Size {
		panic("No short positions allowed")
	}

	t.PL = (t.Size * t.Price) - (t2.Size * t2.Price)
	t.PL -= t.Tax + t.Demat
	t.PL -= t2.Tax + t2.Demat

}

func (t *Trade) String() string {
	if t.Buy {
		return fmt.Sprintf("%s Sz:%.0f, Rs.%.2f, %s",
			t.Date.Format("2006.01.02"), t.Size, t.Price, t.ReportTrend())
	} else {
		return fmt.Sprintf("%s Rs.%.2f, %s",
			t.Date.Format("2006.01.02"), t.Price, t.ReportTrend())
	}
}

func (t *Trade) ReportTrend() string {
	if IsTrendingUp(t.AroonUp, t.AroonDn) {
		return "UP"
	}
	if HasTrendReversed(t.AroonUp, t.AroonDn) {
		return "DN"
	}
	return "NT"
}

type MovingAverageStrategy struct {
	UseEmaNotSma bool    `json:UseEmaNotSma`
	Period1      int     `json:Period1`
	Period2      int     `json:Period2`
	AroonUpBuy   float64 `json:AroonUpBuy`
	AroonUpSell  float64 `json:AroonUpSell`
	AroonDnBuy   float64 `json:AroonDnBuy`
	AroonDnSell  float64 `json:AroonDnSell`
	UseMfi       bool    `json:UseMfi`
}
