package main

import (
	"math"
)

// Replace abnormal demands and promotion weeks with substitute values based on the year's demand and the profile weekly percentages
func replaceAbnormalDemand(skuP *SKU, ProfileInfo *Profile, Sales []float32, Demand []float32, Symbol []string) ([]float32, []float32) {

	//elements: 0 = minus 3 year, 1 = minus 2 year, 2 = last year
	var priorYearsDemand = make([]float32, 3, 3)

	//three years ago
	priorYearsDemand[0] = replace(0, skuP, ProfileInfo, Demand, Symbol)

	//two years ago
	priorYearsDemand[1] = replace(52, skuP, ProfileInfo, Demand, Symbol)

	//last year
	priorYearsDemand[2] = replace(104, skuP, ProfileInfo, Demand, Symbol)

	return priorYearsDemand, Demand
}

// General purpose replacement of abnormal or promotion demands useful for any year
func replace(incr int, skuP *SKU, ProfileInfo *Profile, Demand []float32, Symbol []string) float32 {

	//Get a total for the year
	var totalDemand float32
	var differenceRatio float32
	var inverseFcst float64
	var inverseFcstPwr float64
	var abnormalDemandThreshold float32
	var totalPcnt float32
	var adjTotalDemand float32
	var substituteDemand float32
	var difference float32
	//Use shifted profile percentages since the demand slice starts with the current date
	var profilePcnt []float32
	profilePcnt = ProfileInfo.ShiftedWeeklyPcnt

	//Total demand for the year
	for i := 0; i < 52; i++ {
		//use if not abnormal or promotion
		if Symbol[incr+i] == "" {
			totalDemand = totalDemand + Demand[incr+i]
			profilePcnt[i] = ProfileInfo.ShiftedWeeklyPcnt[i]
			totalPcnt = totalPcnt + profilePcnt[i]
		}
	}

	//Adjust for missing normal percentages,
	//for example total demand = 100 total percentages for normal weeks = 50%, adjusted total = 200
	if totalPcnt > 0 {
		adjTotalDemand = totalDemand / totalPcnt
	} else {
		//With zero total normal percentages, use total demand
		adjTotalDemand = totalDemand
	}

	for i := 0; i < 52; i++ {
		//Calculate substitute value for each week
		substituteDemand = adjTotalDemand * profilePcnt[i]

		//Calculate the difference as a percentage
		difference = Demand[incr+i] - substituteDemand
		if substituteDemand > 0 {
			differenceRatio = difference / substituteDemand
		} else {
			//Default to no difference.
			differenceRatio = 0
		}

		//Calculate the threshold for abnormal demand - positive variance only
		if substituteDemand > 0 && difference > 0 {
			inverseFcst = (1 / float64(substituteDemand))
			inverseFcstPwr = math.Pow(inverseFcst, ConfigP.AbnormalDemandFactor2)
			abnormalDemandThreshold = (float32(inverseFcstPwr) * ConfigP.AbnormalDemandFactor1) + ConfigP.AbnormalDemandMin
		} else {
			//Has the effect of resetting the difference ratio where there is no difference.
			abnormalDemandThreshold = 0
		}

		//Check for abnormal demand
		if differenceRatio > abnormalDemandThreshold &&
			Symbol[incr+i] == "" &&
			SkuP.LastReplenishmentDate.Valid == false &&
			SkuP.EndDate.Valid == false &&
			(SkuP.AnnualForecastOverride.Valid && SkuP.AnnualForecastOverride.Float64 != 0 ||
				SkuP.AnnualForecastOverride.Valid == false) &&
			Demand[incr+i] > 1 {
			//Write abnormal demand exception
			var ex Exception
			ex.OrgId = ConfigP.OrgId
			ex.ExceptionNo = 11
			ex.ItemId = SkuP.ItemId
			ex.LocationId = SkuP.LocationId
			ex.Responsible = SkuP.ResponsibleDemand
			ex.CreateDate = ConfigP.CurrentDate
			ex.ExceptionDate1 = ConfigP.WeeklyPastDates[incr+i]
			ex.ExceptionQty1 = Demand[incr+i]
			insertException(ex)
		}

		//Replace abnormal and promotion
		if Symbol[incr+i] != "" {
			//Replace demand with substitute Demand
			Demand[incr+i] = substituteDemand
		}
	}
	return adjTotalDemand
}
