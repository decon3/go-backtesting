package cfg

import (
	"github.com/spf13/viper"
)

type RiskConfiguration interface {
	InitialCash() float64
	PortfolioRisk() float64
	TradeRisk() float64
}

type riskCfg struct {
	initialCash   float64
	portfolioRisk float64
	tradeRisk     float64
	period        int
}

var rcfg riskCfg = riskCfg{
	initialCash: -1,
}
var dataFolder = ""

func (r *riskCfg) InitialCash() float64 {
	return r.initialCash
}
func (r *riskCfg) PortfolioRisk() float64 {
	return r.portfolioRisk
}
func (r *riskCfg) TradeRisk() float64 {
	return r.tradeRisk
}
func (r *riskCfg) Period() int {
	return r.period
}

func GetConfiguration() (*riskCfg, string) {
	if rcfg.initialCash != -1 {
		return &rcfg, dataFolder
	}

	viper.SetDefault("portfolioRisk", 0.01)
	viper.SetDefault("tradeRisk", 0.03)
	viper.SetDefault("initialCash", 100000)
	viper.SetDefault("dataFolder", "data")

	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	rcfg = riskCfg{
		initialCash:   viper.GetFloat64("initialCash"),
		portfolioRisk: viper.GetFloat64("portfolioRisk"),
		tradeRisk:     viper.GetFloat64("tradeRisk"),
		period:    viper.GetInt("period"),
	}
	dataFolder = viper.GetString("dataFolder")
	return &rcfg, dataFolder
}
