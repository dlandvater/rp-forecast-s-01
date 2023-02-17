package main

// Calculate weekly forecasts
func calcWeeklyForecasts(AnnualForecast float64, skuWeeklyPcnt []float64) []float64 {
	var ForecastBase = make([]float64, 52)

	//Calculate weekly forecasts
	//SKU start and end dates are not used here, instead later when the forecasts are broken down from months or quarters into weeks
	for i := 0; i < len(skuWeeklyPcnt); i++ {
		ForecastBase[i] = skuWeeklyPcnt[i] * AnnualForecast
	}

	return ForecastBase
}
