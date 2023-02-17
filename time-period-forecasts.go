package main

import (
	"database/sql"
	"time"
)

// TODO replace with row id generator
var nextRowId uint64 = 1000

// Calculate weekly forecasts, accumulate into monthly, quarterly, or half-year forecasts
// Assumptions and adjustments are applied later. These are only baseline statistical forecasts
func forecastTimePeriod(ForecastBase []float64, ForecastRows []ForecastBaseline, SkuP *SKU) {
	var WeeklyFutureDates = ConfigP.WeeklyFutureDates
	var baseFcstType string = "B"
	var days int64
	var weeks int64
	var sold float64 = 0
	var nextWklyFcst int64 = 0
	var fcstRowQty float64 = 0
	var update bool
	var span int64
	var startDate time.Time
	var endDate time.Time

	//Determine the forecast time period
	if SkuP.AnnualForecast > ConfigP.ForecastInMonth {
		days = 7
	} else if SkuP.AnnualForecast > ConfigP.ForecastInQuarter {
		days = 28
	} else if SkuP.AnnualForecast > ConfigP.ForecastInHalf {
		days = 91
	} else {
		days = 182
	}
	weeks = days / 7

	//Check for a change in forecast rows time period
	//TODO here and drp, what if the row set is nil?
	lengthFcstRows := len(ForecastRows)
	if len(ForecastRows) > 0 {
		if ForecastRows[0].Days != days {
			//Different time span, delete all existing forecasts
			for i := 0; i < len(ForecastRows); i++ {
				deleteFcstRow(ForecastRows[i].RowId, SkuP)
			}
			lengthFcstRows = 0
		}
	}

	//check that forecast rows exist
	if lengthFcstRows > 0 {
		//Check starting point (first weekly forecast, first forecast database row)
		if (WeeklyFutureDates[0].After(ForecastRows[0].StartDate)) &&
			(WeeklyFutureDates[0].Before(ForecastRows[0].EndDate)) &&
			(len(ForecastRows) > 0) {
			//Case 1: First date in weekly forecast array is in the middle of the first forecast row
			//For example, today is 3/30/xxxx, forecast start date = 3/16/xxxx, forecast end date = 4/12/xxxx
			//Leave the forecast row as is, advance to what would be the next forecast row start date

			//TODO remove
			//		fmt.Println(WeeklyFutureDates[0])
			//		fmt.Println(ForecastRows[0])

			//skip, advance to next weekly forecast
			nextWklyFcst = skip(WeeklyFutureDates, ForecastRows[0].EndDate)

		} else if (WeeklyFutureDates[0].Equal(ForecastRows[0].StartDate)) &&
			(len(ForecastRows) > 0) {
			//Case 2: First date in weekly forecast array equals first forecast row start date
			//Accumulate the weekly forecasts, check dates, days & quantity, update as needed

			//accumulate
			nextWklyFcst, fcstRowQty = acculmulate(nextWklyFcst, weeks, ForecastBase)

			//End date of the forecast should be the start date + days - one day
			span = days - 1
			startDate = WeeklyFutureDates[0]
			endDate = startDate.AddDate(0, 0, int(span))

			//TODO remove
			//		fmt.Println(startDate, endDate, span)
			//		fmt.Println(ForecastRows[0])

			//Check for differences, returns whether or not to update forecast row in the database
			update = checkDifferences(startDate, endDate, days, fcstRowQty, ForecastRows[0])

			//If appropriate, update forecast database row
			if update == true {
				//Selling price is optional, default to null for baseline
				var sp sql.NullFloat64
				sp.Valid = false
				//sp.Float64 = float64(SkuP.SellingPrice)

				//Selling price for calculated baseline forecasts is always empty.
				updateFcstRow(ForecastRows[0].RowId, SkuP, baseFcstType, days, startDate, endDate, fcstRowQty, sold, sp)
			}
		} else if (WeeklyFutureDates[0].Before(ForecastRows[0].StartDate)) &&
			(len(ForecastRows) > 0) {

			//Case 3: First date in weekly forecast array is earlier than the first forecast row start date
			//Accumulate the weekly forecasts, update the forecast row

			//accumulate
			nextWklyFcst, fcstRowQty = acculmulate(nextWklyFcst, weeks, ForecastBase)

			//End date of the forecast should be the start date + days - one day
			span = days - 1
			startDate = WeeklyFutureDates[0]
			endDate = startDate.AddDate(0, 0, int(span))

			//TODO remove
			//		fmt.Println(startDate, endDate, span)
			//		fmt.Println(ForecastRows[0])

			//Selling price is optional, default to null for baseline
			var sp sql.NullFloat64
			sp.Valid = false
			//sp.Float64 = float64(SkuP.SellingPrice)

			//Update forecast database row
			updateFcstRow(ForecastRows[0].RowId, SkuP, baseFcstType, days, startDate, endDate, fcstRowQty, sold, sp)

		} else if (WeeklyFutureDates[0].After(ForecastRows[0].EndDate)) &&
			(len(ForecastRows) > 0) {

			//Case 4: Forecast is past its end date but has not been deleted (should not happen)
			//Delete the forecast row

			//Update forecast database row
			deleteFcstRow(ForecastRows[0].RowId, SkuP)
		}

		//Second and later forecast rows, first was done above
		//End date of the forecast should be the start date + days - one day, except at the end of the array
		for i := 1; i < len(ForecastRows); i++ {

			//Check if all weekly forecasts have been used
			if nextWklyFcst > 51 {
				//Delete remaining forecast rows
				deleteFcstRow(ForecastRows[i].RowId, SkuP)
			} else {
				//accumulate
				startDate = WeeklyFutureDates[nextWklyFcst]
				nextWklyFcst, fcstRowQty = acculmulate(nextWklyFcst, weeks, ForecastBase)

				//check for end of weekly forecasts, days are less than other forecast rows
				if nextWklyFcst > 51 {

					//End date of the forecast is the end of the weekly forecast array, days will vary
					endDate = WeeklyFutureDates[51].AddDate(0, 0, 6)
					span = int64(endDate.Sub(startDate).Hours() / 24)

					//TODO remove
					//fmt.Println(startDate, endDate, span)
					//fmt.Println(ForecastRows[i])

					//Selling price is optional, default to null for baseline
					var sp sql.NullFloat64
					sp.Valid = false
					//sp.Float64 = float64(SkuP.SellingPrice)

					//Do not check for differences, always update forecast rows at the end of the weekly forecast array
					updateFcstRow(ForecastRows[i].RowId, SkuP, baseFcstType, span+1, startDate, endDate, fcstRowQty, sold, sp)

				} else {

					//End date of the forecast is the start date + days - one day
					span = int64(days - 1)
					endDate = startDate.AddDate(0, 0, int(span))

					//TODO remove
					//fmt.Println(startDate, endDate, span)
					//fmt.Println(ForecastRows[i])

					//Check for differences, returns whether or not to update forecast row in the database
					update = checkDifferences(startDate, endDate, days, fcstRowQty, ForecastRows[i])

					//Selling price is optional, default to null for baseline
					var sp sql.NullFloat64
					sp.Valid = false
					//sp.Float64 = float64(SkuP.SellingPrice)

					//If appropriate, update forecast database row
					if update == true {
						updateFcstRow(ForecastRows[i].RowId, SkuP, baseFcstType, days, startDate, endDate, fcstRowQty, sold, sp)
					}
				}
			}
		}
	}

	//Remaining weekly forecasts, forecast rows are all used
	for nextWklyFcst < 52 {
		//accumulate
		startDate = WeeklyFutureDates[nextWklyFcst]
		nextWklyFcst, fcstRowQty = acculmulate(nextWklyFcst, weeks, ForecastBase)

		//check for end of weekly forecasts, days are less than other forecast rows
		if nextWklyFcst > 51 {

			//End date of the forecast is the end of the weekly forecast array, days will vary
			endDate = WeeklyFutureDates[51].AddDate(0, 0, 6)
			//Conform to other forecasts (multiples of 7 days)
			span = int64(endDate.Sub(startDate).Hours()/24) + 1

			//TODO remove
			//fmt.Println(startDate, endDate, span)

			//Selling price is optional, default to null for baseline
			var sp sql.NullFloat64
			sp.Valid = false
			//sp.Float64 = float64(SkuP.SellingPrice)

			_ = insertFcstRow(SkuP, baseFcstType, span, startDate, endDate, fcstRowQty, sold, sp)

		} else {

			//End date of the forecast is the start date + days - one day
			span = int64(days - 1)
			endDate = startDate.AddDate(0, 0, int(span))

			//TODO remove
			//fmt.Println(startDate, endDate, span)

			//Selling price is optional, default to null for baseline
			var sp sql.NullFloat64
			sp.Valid = false
			//sp.Float64 = float64(SkuP.SellingPrice)

			_ = insertFcstRow(SkuP, baseFcstType, days, startDate, endDate, fcstRowQty, sold, sp)
		}
	}

	//TODO remove
	//	fmt.Println("end of time-period-forecasts")
}

// Skip to next forecast start date
func skip(WeeklyFutureDates []time.Time, forecastEndDate time.Time) int64 {

	var nextElement int64

	//Find the first weekly forecast element after the forecast end date
	for i := 0; i < len(WeeklyFutureDates); i++ {
		if WeeklyFutureDates[i].After(forecastEndDate) {
			nextElement = int64(i)
			break
		}
	}
	return nextElement
}

// Accumulate forecasts
func acculmulate(nextWklyFcst int64, weeks int64, ForecastBase []float64) (int64, float64) {

	var accumulator float64

	var end = nextWklyFcst + weeks
	if end > 52 {
		end = 52
	}
	for i := nextWklyFcst; i < end; i++ {
		accumulator = accumulator + ForecastBase[i]
	}

	return end, accumulator
}

// Check differences
func checkDifferences(startDate time.Time, endDate time.Time, days int64, fcstRowQty float64, ForecastRow ForecastBaseline) bool {

	var update bool = true //default to update forecast row
	var daysOk bool = false
	var startDateOk bool = false
	var endDateOk bool = false
	var quantityOk bool = false

	if days == ForecastRow.Days {
		daysOk = true
	}

	//Check start date
	if startDate.Equal(ForecastRow.StartDate) {
		startDateOk = true
	}

	//Check days
	//Check end date
	if endDate.Equal(ForecastRow.EndDate) {
		endDateOk = true
	}

	//Check quantity
	var difference float64 = fcstRowQty - ForecastRow.Quantity
	if difference < 0 {
		difference = 0 - difference
	}
	if ForecastRow.Quantity > 0 {
		difference = difference / ForecastRow.Quantity
		if difference < 0.02 {
			//Using 2% as a tolerance
			quantityOk = true
		}
	} else {
		//Cannot evaluate a difference
		quantityOk = false
	}

	if startDateOk && endDateOk && quantityOk && daysOk {
		update = false
	}

	return update
}
