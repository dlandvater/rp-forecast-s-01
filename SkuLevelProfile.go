package main

import "fmt"

//Calculate weekly percentages for fast-moving items

// Replace abnormal demands and promotion weeks with substitute values based on the year's demand and the profile weekly percentages
func calculateSkuProfile(Demand []float32) []float32 {

	var SkuWeeklyPcnt = make([]float32, 52)
	var SkuWeeklyTotals = make([]float32, 52)
	var weightedWeeklyTotals float32
	var pcntTotals float32

	//three years ago
	SkuWeeklyTotals = accumulate(0, ConfigP.SkuProfileWeightM3, Demand, SkuWeeklyTotals)

	//two years ago
	SkuWeeklyTotals = accumulate(52, ConfigP.SkuProfileWeightM2, Demand, SkuWeeklyTotals)

	//last year
	SkuWeeklyTotals = accumulate(104, ConfigP.SkuProfileWeightM1, Demand, SkuWeeklyTotals)

	//Smoothing
	var SkuWeeklyTotalsSmooth = smooth(ConfigP, SkuWeeklyTotals)

	//Normalize
	for i := 0; i < len(SkuWeeklyTotals); i++ {
		weightedWeeklyTotals = weightedWeeklyTotals + SkuWeeklyTotalsSmooth[i]
	}

	//Convert to weekly percentages
	for i := 0; i < len(SkuWeeklyTotalsSmooth); i++ {
		SkuWeeklyPcnt[i] = SkuWeeklyTotalsSmooth[i] / weightedWeeklyTotals
	}

	//Check percentage totals, should be 100%
	for i := 0; i < len(SkuWeeklyPcnt); i++ {
		pcntTotals = pcntTotals + SkuWeeklyPcnt[i]
	}
	if pcntTotals < 0.98 || pcntTotals > 1.02 {
		//TODO exception if not 100% within tolerance
		fmt.Println("SKU wkly pcnt outside tolerance")
	}

	return SkuWeeklyPcnt
}

// Multiply the demands by the weight for each year
func accumulate(incr int, weight float32, Demand []float32, SkuWeeklyTotals []float32) []float32 {

	for i := 0; i < 52; i++ {
		SkuWeeklyTotals[i] = Demand[i+incr] * weight
	}

	return SkuWeeklyTotals
}

// Smooth the profile using smoothing weights for the week, the week before & after, and two weeks before & after
func smooth(ConfigInfo *Config, SkuWeeklyTotals []float32) []float32 {

	var SkuWeeklyTotalsSmooth = make([]float32, 52)

	//Normalize weights
	var totalSmoothWeights = ConfigInfo.SmoothWk0Pcnt + 2*ConfigInfo.SmoothWk1Pcnt + 2*ConfigInfo.SmoothWk2Pcnt
	var ratio = 1 / totalSmoothWeights
	var SmoothWk0Pcnt = ratio * ConfigInfo.SmoothWk0Pcnt
	var SmoothWk1Pcnt = ratio * ConfigInfo.SmoothWk1Pcnt
	var SmoothWk2Pcnt = ratio * ConfigInfo.SmoothWk2Pcnt

	var w0, w1, wm1, w2, wm2 int

	for i := 0; i < 52; i++ {
		w0 = i
		w1 = wrap(i + 1)
		wm1 = wrap(i - 1)
		w2 = wrap(i + 2)
		wm2 = wrap(i - 2)

		SkuWeeklyTotalsSmooth[i] = SkuWeeklyTotals[w0]*SmoothWk0Pcnt +
			(SkuWeeklyTotals[w1]+SkuWeeklyTotals[wm1])*SmoothWk1Pcnt +
			(SkuWeeklyTotals[w2]+SkuWeeklyTotals[wm2])*SmoothWk2Pcnt
	}

	return SkuWeeklyTotalsSmooth
}

// Wrap around the ends
func wrap(index int) int {

	if index > 51 {
		index = index - 52
	}
	if index < 0 {
		index = index + 52
	}
	return index
}
