package quotes

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

type QuoteData struct {
	Symbol  string
	Dates   []time.Time
	Opens   []float64
	Highs   []float64
	Lows    []float64
	Closes  []float64
	Volumes []float64
}

type Quote struct {
	Date     time.Time
	Open     float64
	High     float64
	Low      float64
	Close    float64
	AdjClose float64
	Volume   float64
}

func LoadFromFile(csvfile string, symbol string, skip int) *QuoteData {
	csvreader, err := os.Open(csvfile)
	if err != nil {
		log.Printf("Unable to open %s: %v", csvfile, err)
	}
	reader := csv.NewReader(bufio.NewReader(csvreader))
	reader.ReuseRecord = true
	reader.FieldsPerRecord = -1
	return Load(reader, symbol, skip)
}

func (q *QuoteData) String() string {
	return fmt.Sprintf("%s Open:%d Close:%d High:%d, Low:%d, Volume:%d",
		q.Symbol, len(q.Opens), len(q.Closes), len(q.Highs), len(q.Lows), len(q.Volumes))

}

func Load(reader *csv.Reader, symbol string, skip int) *QuoteData {
	var quotes []Quote
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
			if line[i] == "null" {
				line[i] = "0"
			}
		}
		var q = Quote{}
		if q.Date, err = time.Parse("02-01-2006", line[0]); err != nil {
			log.Printf("%d: (Date) %v", lineno, err)
			continue
		}
		if q.Open, err = strconv.ParseFloat(line[1], 64); err != nil {
			log.Printf("%d: (Open) %v", lineno, err)
			continue
		}
		if q.High, err = strconv.ParseFloat(line[2], 64); err != nil {
			log.Printf("%d: (High) %v", lineno, err)
			continue
		}
		if q.Low, err = strconv.ParseFloat(line[3], 64); err != nil {
			log.Printf("%d: (Low) %v", lineno, err)
			continue
		}
		if q.Close, err = strconv.ParseFloat(line[4], 64); err != nil {
			log.Printf("%d: (Close) %v", lineno, err)
			continue
		}
		if q.AdjClose, err = strconv.ParseFloat(line[5], 64); err != nil {
			log.Printf("%d: (AdjClose) %v", lineno, err)
			continue
		}
		if q.Volume, err = strconv.ParseFloat(line[6], 64); err != nil {
			log.Printf("%d: (Volume) %v", lineno, err)
			continue
		}
		quotes = append(quotes, q)
	}
	return convert(quotes, symbol)
}

func convert(quotes []Quote, symbol string) *QuoteData {
	if len(quotes) == 0 {
		return &QuoteData{
			Symbol: symbol,
		}
	}
	qd := &QuoteData{
		Symbol:  symbol,
		Dates:   make([]time.Time, len(quotes)),
		Opens:   make([]float64, len(quotes)),
		Highs:   make([]float64, len(quotes)),
		Lows:    make([]float64, len(quotes)),
		Closes:  make([]float64, len(quotes)),
		Volumes: make([]float64, len(quotes)),
	}
	for i, q := range quotes {
		qd.Dates[i] = q.Date
		qd.Opens[i] = q.Open
		qd.Highs[i] = q.High
		qd.Lows[i] = q.Low
		qd.Closes[i] = q.Close
		qd.Volumes[i] = q.Volume
	}
	return qd
}
