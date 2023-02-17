package main

// Calculate trend based on comparable weeks from prior years
func calculateTrend(skuP *SKU, priorYearsDemand []float64, Demand []float64) float64 {

	var Trend float64
	var endingIndex int
	var span int

	//Spans based on annual forecast and thresholds from configuration
	if priorYearsDemand[2] < float64(ConfigP.TrendYearThreshold) {
		span = 52
	} else if priorYearsDemand[2] < float64(ConfigP.TrendHalfYearThreshold) {
		span = 26
	} else if priorYearsDemand[2] < float64(ConfigP.TrendQuarterThreshold) {
		span = 12
	} else {
		span = 12
	}

	//two years ago
	endingIndex = 103
	twoYearsAgoComparableDemand := comparableQty(endingIndex, span, Demand)

	//last year
	endingIndex = 155
	priorYearComparableDemand := comparableQty(endingIndex, span, Demand)

	//trend is the percentage difference in the comparable quantities
	//two years ago: 100, last year: 120, trend = (120-100) / 100 = +20%
	//two years ago: 100, last year: 80, trend = (80-100) / 100 = -20%
	//priorYearsDemand elements: 0 = minus 3 year, 1 = minus 2 year, 2 = last year
	if twoYearsAgoComparableDemand > 0 {
		Trend = ((priorYearComparableDemand - twoYearsAgoComparableDemand) / twoYearsAgoComparableDemand)
	} else {
		//Cannot calculate a trend if basis for comparison is zero
		Trend = 0
	}
	//math.Abs requires float64
	//	var Trend64 = float64(Trend)
	//	var absTrend64 = math.Abs(Trend64)
	//	var absTrend32 = float64(absTrend64)

	var absTrend = Trend
	if absTrend < 0 {
		absTrend = 0 - absTrend
	}

	if absTrend > ConfigP.TrendLimitPcnt {
		//Write trend limited exception
		var ex Exception
		ex.RowId = createRowId(skuP.OrgId, "exceptions")
		ex.OrgId = skuP.OrgId
		ex.ExceptionNo = 23
		ex.ItemId = skuP.ItemId
		ex.LocationId = skuP.LocationId
		ex.Responsible = skuP.ResponsibleDemand
		ex.CreateDate = ConfigP.CurrentDate
		ex.ExceptionQty1 = Trend
		ex.ExceptionQty2 = ConfigP.TrendLimitPcnt
		insertException(ex)

		if Trend < 0 {
			Trend = 0 - ConfigP.TrendLimitPcnt
		} else {
			Trend = ConfigP.TrendLimitPcnt
		}
	} else if absTrend > ConfigP.TrendExceptionThresholdPcnt {
		//Write trend tolerance exception
		var ex Exception
		ex.RowId = createRowId(ConfigP.OrgId, "exception")
		ex.OrgId = ConfigP.OrgId
		ex.ExceptionNo = 24
		ex.ItemId = skuP.ItemId
		ex.LocationId = skuP.LocationId
		ex.Responsible = SkuP.ResponsibleDemand
		ex.CreateDate = ConfigP.CurrentDate
		ex.ExceptionQty1 = Trend
		ex.ExceptionQty2 = ConfigP.TrendExceptionThresholdPcnt
		insertException(ex)
	}
	return Trend
}

// General purpose replacement for any year
func comparableQty(endingIndex int, span int, Demand []float64) float64 {
	//Get a total for the year
	var comparisonQty float64

	//Accumulate selected elements
	for i := endingIndex - span; i <= endingIndex; i++ {
		comparisonQty = comparisonQty + Demand[i]
	}
	return comparisonQty
}
