package mm

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

func LoadTrades(csvfile string) []*TradeEvent {
	csvreader, err := os.Open(csvfile)
	if err != nil {
		log.Printf("Unable to open %s: %v", csvfile, err)
	}
	reader := csv.NewReader(bufio.NewReader(csvreader))
	reader.ReuseRecord = true
	reader.FieldsPerRecord = -1
	return Load(reader, 1)
}

func Load(reader *csv.Reader, skip int) []*TradeEvent {
	var trades []*TradeEvent
	var lineno = 0
	for {
		lineno++
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("error reading file: %v", err)
			break
		}
		if skip > 0 && lineno <= skip {
			continue
		}
		for i := 0; i < len(line); i++ {
			if line[i] == "null" || line[i] == "" {
				line[i] = "0"
			}
		}
		var t = &TradeEvent{}
		var d = &TradeDecision{}
		t.Decision = d
		f := 0
		t.Symbol = line[f]

		f++
		if t.Date, err = time.Parse("2006/01/02", line[f]); err != nil {
			log.Printf("%d: (%d:Date) %v", lineno, f, err)
			continue
		}
		f++
		if line[f] == "BUY" {
			t.Action = Buy
		} else {
			t.Action = Sell
		}
		f++
		if t.Size, err = strconv.Atoi(line[f]); err != nil {
			log.Printf("%d: (%d:Size) %v", lineno, f, err)
			continue
		}
		f++
		if t.Price, err = strconv.ParseFloat(line[f], 64); err != nil {
			log.Printf("%d: (%d:Price) %v", lineno, f, err)
			continue
		}
		f++
		if d.Capital, err = strconv.ParseFloat(line[f], 64); err != nil {
			log.Printf("%d: (%d:Capital) %v", lineno, f, err)
			continue
		}
		f++
		if d.CapitalRisk, err = strconv.ParseFloat(line[f], 64); err != nil {
			log.Printf("%d: (%d:CapitalRisk) %v", lineno, f, err)
			continue
		}
		f++
		if d.TradeRisk, err = strconv.ParseFloat(line[f], 64); err != nil {
			log.Printf("%d: (%d:TradeRisk) %v", lineno, f, err)
			continue
		}
		f++
		if d.Close, err = strconv.ParseFloat(line[f], 64); err != nil {
			log.Printf("%d: (%d:Close) %v", lineno, f, err)
			continue
		}
		f++
		if d.StopLoss, err = strconv.ParseFloat(line[f], 64); err != nil {
			log.Printf("%d: (%d:StopLoss) %v", lineno, f, err)
			continue
		}
		f++
		if d.Target, err = strconv.ParseFloat(line[f], 64); err != nil {
			log.Printf("%d: (%d:Target) %v", lineno, f, err)
			continue
		}
		f++
		if line[f] == "Z" {
			t.CalculateCost(ZerodhaRates)
		} else {
			t.CalculateCost(FivePaisaRates)
		}
		trades = append(trades, t)
	}
	return trades
}
