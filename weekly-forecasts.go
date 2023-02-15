package main

// Calculate weekly forecasts
func calcWeeklyForecasts(AnnualForecast float32, skuWeeklyPcnt []float32) []float32 {
	var ForecastBase = make([]float32, 52)

	//Calculate weekly forecasts
	//SKU start and end dates are not used here, instead later when the forecasts are broken down from months or quarters into weeks
	for i := 0; i < len(skuWeeklyPcnt); i++ {
		ForecastBase[i] = skuWeeklyPcnt[i] * AnnualForecast
	}

	return ForecastBase
}
