package mm

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
)

type Portfolio struct {
	Investment float64
	Capital    float64
	Assets     map[string]Asset
	AssetValue float64
}

type Asset struct {
	Symbol       string
	Positions    []TradeEvent
	CurrentPrice float64
}

func NewPortfolio(investment float64, capital float64) *Portfolio {
	return &Portfolio{
		Investment: investment,
		Capital:    capital,
		Assets:     make(map[string]Asset),
	}
}
func (p *Portfolio) Clone() *Portfolio {
	p2 := *p
	for k, _ := range p2.Assets {
		p2.Assets[k] = p.Assets[k].Clone()
	}
	return &p2
}

func (a Asset) Clone() Asset {
	a2 := a
	a2.Positions = make([]TradeEvent, len(a.Positions))
	copy(a2.Positions, a.Positions)
	return a2
}

func (a Asset) Size() int {
	total := 0
	for _, p := range a.Positions {
		total += p.Size
	}
	return total
}
func (a Asset) AssetValue() float64 {
	return a.CurrentPrice * float64(a.Size())
}
func (a Asset) AveragePrice() float64 {
	total := 0.0
	for _, p := range a.Positions {
		total += float64(p.Size) * p.Price
	}
	return total / float64(a.Size())
}

func (p *Portfolio) Buy(symbol string, buy TradeEvent) {
	var cost float64
	cost = buy.Price * float64(buy.Size)
	if e, exists := p.Assets[symbol]; exists {
		e.Positions = append(e.Positions, buy)
		e.CurrentPrice = buy.Price
	} else {
		a := Asset{
			Symbol:       symbol,
			Positions:    []TradeEvent{buy},
			CurrentPrice: buy.Price,
		}
		p.Assets[symbol] = a
	}
	p.AssetValue += cost
	p.Capital -= cost
}

func (p *Portfolio) Sell(symbol string, sell TradeEvent) ([]TradeEvent, float64, error) {
	spew.Dump("Portfolio.Sell", symbol, sell, p)
	if a, ok := p.Assets[symbol]; ok {
		// add proceedings to capital
		proceedings := sell.Price * float64(sell.Size)
		p.Capital += proceedings
		// calculate P&L
		pnl := proceedings - (a.AveragePrice() * float64(sell.Size))

		if a.Size() == sell.Size {
			// sold completely
			delete(p.Assets, symbol)
			p.RecalculateAssetValue()
			spew.Dump("Portfolio.Sell-exit(full)", p, a.Positions, pnl)
			return a.Positions, pnl, nil
		} else {
			unsoldPortions := []TradeEvent{}
			soldPortions := []TradeEvent{}
			// remove sold portion
			soldSize := sell.Size
			for _, p := range a.Positions {
				if soldSize == 0 {
					break
				}
				var portion TradeEvent
				apportionedSize := 0
				switch {
				case p.Size > soldSize:
					apportionedSize = soldSize
					soldSize = 0
					portion, _ = p.SoldPartially(sell.Size, sell.Price)
					unsoldPortions = append(unsoldPortions, p)
				case p.Size <= soldSize:
					apportionedSize = p.Size
					soldSize -= apportionedSize
					portion = p
				}
				soldPortions = append(soldPortions, portion)
				a.Positions = unsoldPortions
			}
			p.RecalculateAssetValue()
			spew.Dump("Portfolio.Sell-exit(partial)", p, soldPortions, pnl)
			return soldPortions, pnl, nil
		}
	} else {
		return nil, 0, errors.New(fmt.Sprintf("%s - asset not found", symbol))
	}
}

func (p *Portfolio) RecalculateAssetValue() {
	p.AssetValue = 0
	for _, a := range p.Assets {
		p.AssetValue += a.AssetValue()
	}
}

func (p *Portfolio) UpdateAssetCurrentValue(symbol string, price float64) error {
	if a, ok := p.Assets[symbol]; ok {
		// update the current price of the asset
		a.CurrentPrice = price
		p.RecalculateAssetValue()
		return nil
	} else {
		return errors.New(fmt.Sprintf("%s - asset not found", symbol))
	}
}
