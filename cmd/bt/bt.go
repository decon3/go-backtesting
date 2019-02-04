package main

import (
	"fmt"
	gbt "github.com/dirkolbrich/gobacktest"
	"github.com/dirkolbrich/gobacktest/algo"
	"github.com/dirkolbrich/gobacktest/data"
	"io/ioutil"
	"log"
	"pkg/cfg"
	"strings"
)

func main() {
	_, dataFolder := cfg.GetConfiguration()
	symbols := getSymbols(dataFolder)
	pl1 := 0.00
	pl2 := 0.00
	for _, symbol := range symbols {
		fmt.Printf("***************** %s *******************\n", symbol)
		pl1 += testEma(symbol)
		pl2 += testEmaConfirmedByMfi(symbol)
	}
	fmt.Printf("Total PL: %.2f (EMA) %.2f (EMA+MFI)\n", pl1, pl2)
}

func testEma(symbol string) float64 {
	cfg, dataFolder := cfg.GetConfiguration()
	// initiate a new backtester
	test := gbt.New()

	// define and load symbols
	symbols := []string{symbol}
	test.SetSymbols(symbols)

	// create a data provider and load the data into the backtest
	mydata := &data.BarEventFromCSVFile{FileDir: "./" + dataFolder + "/"}
	mydata.Load(symbols)
	test.SetData(mydata)

	// create a new strategy with an algo stack and load into the backtest
	strategy := gbt.NewStrategy("basic")
	period := algo.RunMonthly()
	switch {
	case cfg.Period() == 1:
		period = algo.RunDaily()
	case cfg.Period() < 22:
		period = algo.RunWeekly()
	case cfg.Period() < 64:
		period = algo.RunMonthly()
	case cfg.Period() < 200:
		period = algo.RunQuarterly()
	default:
		period = algo.RunYearly()
	}
	strategy.SetAlgo(
		period,
		algo.If(
			// condition
			algo.And(
				algo.BiggerThan(algo.EMA(50), algo.EMA(200)),
				algo.NotInvested(),
			),
			// action
			algo.CreateSignal("buy"), // create a buy signal
		),
		algo.If(
			// condition
			algo.And(
				algo.Or(
					algo.HasHitStopLoss(),
					algo.SmallerThan(algo.EMA(50), algo.EMA(200)),
				),
				algo.IsInvested(),
			),
			// action
			algo.CreateSignal("exit"), // create a sell signal
		),
	)

	p := gbt.NewPortfolio()
	log.Printf("prisk:%.2f, trisk:%.2f", cfg.PortfolioRisk(), cfg.TradeRisk())
	r := gbt.NewRiskManager(cfg.PortfolioRisk(), cfg.TradeRisk(), .20)
	p.SetRiskManager(r)
	p.SetInitialCash(cfg.InitialCash())

	// create an asset and append to strategy
	strategy.SetChildren(gbt.NewAsset("BAJFINANCE.NS"))

	// load the strategy into the backtest
	test.SetStrategy(strategy)
	test.SetPortfolio(p)

	c := gbt.FixedCommission{Commission: 6.75}
	t := gbt.PercentageTransactionTax{
		Tax: 0.0000325,
	}
	e := gbt.PercentageExchangeFee{
		ExchangeFee: 0.0000015,
	}
	test.SetExchange(&gbt.Exchange{
		Symbol:         "NSE",
		Commission:     &c,
		ExchangeFee:    &e,
		TransactionTax: &t,
	})

	// run the backtest
	err := test.Run()
	if err != nil {
		fmt.Printf("err: %v", err)
	}

	// print the result of the test
	test.Stats().PrintResult()
	pc := gbt.Casher(p)
	pl := pc.Cash() - pc.InitialCash()
	fmt.Printf("Initial Cash: %.2f. Current cash: %.2f. P/L:%.2f\n",
		pc.InitialCash(), pc.Cash(), pl)
	test.Reset()
	return pl
}

func testEmaConfirmedByMfi(symbol string) float64 {
	cfg, dataFolder := cfg.GetConfiguration()
	// initiate a new backtester
	test := gbt.New()

	// define and load symbols
	symbols := []string{symbol}
	test.SetSymbols(symbols)

	// create a data provider and load the data into the backtest
	mydata := &data.BarEventFromCSVFile{FileDir: "./" + dataFolder + "/"}
	mydata.Load(symbols)
	test.SetData(mydata)

	// create a new strategy with an algo stack and load into the backtest
	strategy := gbt.NewStrategy("basic")
	period := algo.RunMonthly()
	switch {
	case cfg.Period() == 1:
		period = algo.RunDaily()
	case cfg.Period() < 22:
		period = algo.RunWeekly()
	case cfg.Period() < 64:
		period = algo.RunMonthly()
	case cfg.Period() < 200:
		period = algo.RunQuarterly()
	default:
		period = algo.RunYearly()
	}
	strategy.SetAlgo(
		period,
		algo.If(
			// condition
			algo.And(
				algo.And(
					algo.BiggerThan(algo.EMA(50), algo.EMA(200)),
					algo.BiggerThan(algo.Mfi(10), algo.Number(50)),
				),
				algo.NotInvested(),
			),
			// action
			algo.CreateSignal("buy"), // create a buy signal
		),
		algo.If(
			// condition
			algo.And(
				algo.Or(
					algo.HasHitStopLoss(),
					algo.Or(
						algo.HasHitTarget(),
						algo.And(
							algo.SmallerThan(algo.EMA(50), algo.EMA(200)),
							algo.SmallerThan(algo.Mfi(10), algo.Number(50)),
						),
					),
				),
				algo.IsInvested(),
			),
			// action
			algo.CreateSignal("exit"), // create a sell signal
		),
	)

	p := gbt.NewPortfolio()
	log.Printf("prisk:%.2f, trisk:%.2f", cfg.PortfolioRisk(), cfg.TradeRisk())
	r := gbt.NewRiskManager(cfg.PortfolioRisk(), cfg.TradeRisk(), .20)
	p.SetRiskManager(r)
	p.SetInitialCash(cfg.InitialCash())

	// create an asset and append to strategy
	strategy.SetChildren(gbt.NewAsset("BAJFINANCE.NS"))

	// load the strategy into the backtest
	test.SetStrategy(strategy)
	test.SetPortfolio(p)

	c := gbt.FixedCommission{Commission: 6.75}
	t := gbt.PercentageTransactionTax{
		Tax: 0.0000325,
	}
	e := gbt.PercentageExchangeFee{
		ExchangeFee: 0.0000015,
	}
	test.SetExchange(&gbt.Exchange{
		Symbol:         "NSE",
		Commission:     &c,
		ExchangeFee:    &e,
		TransactionTax: &t,
	})

	// run the backtest
	err := test.Run()
	if err != nil {
		fmt.Printf("err: %v", err)
	}

	// print the result of the test
	test.Stats().PrintResult()
	pc := gbt.Casher(p)
	pl := pc.Cash() - pc.InitialCash()
	fmt.Printf("Initial Cash: %.2f. Current cash: %.2f. P/L:%.2f\n",
		pc.InitialCash(), pc.Cash(), pl)
	test.Reset()
	return pl
}

func getSymbols(dataFolder string) []string {
	files, err := ioutil.ReadDir(dataFolder)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d files in '%s'\n", len(files), dataFolder)
	symbols := []string{}
	for _, file := range files {
		tokens := strings.Split(file.Name(), ".")
		if len(tokens) != 2 || tokens[1] != "csv" {
			continue
		}
		symbols = append(symbols, tokens[0])
	}
	return symbols
}
