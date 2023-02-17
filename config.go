package main

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

func (ConfP *Config) getConfig(orgId string) {

	//Start of common section

	//Sunday is zero, Monday is 1, etc.
	var daysOfWeek = map[string]time.Weekday{
		"Sunday":    time.Sunday,
		"Monday":    time.Monday,
		"Tuesday":   time.Tuesday,
		"Wednesday": time.Wednesday,
		"Thursday":  time.Thursday,
		"Friday":    time.Friday,
		"Saturday":  time.Saturday,
	}

	//NOTE: Times 1970 and before give an error in mysql
	var zd time.Time
	ConfP.ZeroDate = zd

	ed, err := time.Parse(time.RFC3339, "2000-01-01T00:00:00Z")
	if err != nil {
		fmt.Println("empty time parse error: ", err)
	}
	ConfP.EmptyDate = ed

	configRows := queryConfigurationRows(orgId)

	ConfP.OrgId = orgId
	for i := 0; i < len(configRows); i++ {
		if configRows[i].Name == "Organization" {
			ConfP.OrgId = configRows[i].Value
		} else if configRows[i].Name == "OrganizationName" {
			ConfP.OrgName = configRows[i].Value
		} else if configRows[i].Name == "PlanningHorizon" {
			ConfP.PlanningHorizon, err = strconv.Atoi(configRows[i].Value)
			//TODO remove once financial planning rows in the database have 104 time periods
			ConfP.PlanningHorizon = 52 //temporary
			if err != nil {
				ConfP.PlanningHorizon = 104 //default
			}
		} else if configRows[i].Name == "DailyHorizon" {
			ConfP.DailyHorizon, err = strconv.Atoi(configRows[i].Value)
			if err != nil {
				ConfP.PlanningHorizon = 21 //default
			}
		} else if configRows[i].Name == "StartDayOfWeek" {
			ConfP.StartDayOfWeek = configRows[i].Value
		} else if configRows[i].Name == "OverrideDateTime" {
			//Override date and time uses RC3330 format in UTC: "2020-03-30T12:00:00+00:00"
			ConfP.OverrideDateTime = configRows[i].Value
			//TODO remove testing database start date of week = Monday, prior override date "2020-06-22T12:00:00+00:00"
		} else if configRows[i].Name == "PastDaysFcst" {
			ConfP.PastDaysFcst, err = strconv.Atoi(configRows[i].Value)
			if err != nil {
				ConfP.PastDaysSuply = 30 //default
			}
		} else if configRows[i].Name == "PastDaysSuply" {
			ConfP.PastDaysFcst, err = strconv.Atoi(configRows[i].Value)
			if err != nil {
				ConfP.PastDaysSuply = 7 //default
			}
		}
	}

	//Config info calculated

	//Start day of week as a number
	ConfP.WeekdayStartDayOfWeek = 1 //default
	if val, ok := daysOfWeek[ConfP.StartDayOfWeek]; ok {
		ConfP.WeekdayStartDayOfWeek = int(val)
	}

	//Check override date and time if one exists
	if ConfP.OverrideDateTime != "" {
		ConfP.CurrentDate, err = time.Parse(time.RFC3339, ConfP.OverrideDateTime)
		if err != nil {
			// No organization has been determined yet
			insertErrorLog("0", "", "", err, 1)
			//Default to today
			ConfP.CurrentDate = time.Now()
		}
	} else {
		//No override date and time, use the current date and time
		ConfP.CurrentDate = time.Now()
	}

	//Remove hours, minutes, seconds, nanoseconds
	y := ConfP.CurrentDate.Year()
	m := ConfP.CurrentDate.Month()
	d := ConfP.CurrentDate.Day()
	ConfP.CurrentDate = time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)

	//TODO remove
	fmt.Println("current date ", ConfP.CurrentDate)

	//Find the start of the week
	var w1 time.Weekday = ConfP.CurrentDate.Weekday()
	var WeekdayCurrentDate int
	WeekdayCurrentDate = int(w1)

	//Adjust for later start of week
	if WeekdayCurrentDate < ConfP.WeekdayStartDayOfWeek {
		//For example current date = Sunday, start day of week = Monday
		WeekdayCurrentDate = WeekdayCurrentDate + 7
	}

	//Move backwards one day at a time
	var i int = WeekdayCurrentDate
	ConfP.CurrentWeek = ConfP.CurrentDate
	for {
		if i <= ConfP.WeekdayStartDayOfWeek {
			break
		} else {
			i--
			ConfP.CurrentWeek = ConfP.CurrentWeek.AddDate(0, 0, -7)
		}
	}

	//Find the start date of the first week of the year
	//Adjust for later start of week
	var CurrentYear = ConfP.CurrentWeek.Year()
	var firstWeek time.Time = time.Date(CurrentYear, 1, 1, 0, 0, 0, 0, time.UTC)
	var fw time.Weekday = firstWeek.Weekday()
	var WeekdayFirstWeek int
	WeekdayFirstWeek = int(fw)
	if WeekdayFirstWeek < ConfP.WeekdayStartDayOfWeek {
		WeekdayFirstWeek = WeekdayCurrentDate + 7
	}

	//Move backwards one day at a time
	i = WeekdayFirstWeek
	for {
		if i <= ConfP.WeekdayStartDayOfWeek {
			break
		} else {
			i--
			firstWeek = firstWeek.AddDate(0, 0, -1)
		}
	}

	//Find the week number of the current date
	var weeks64 float64 = ConfP.CurrentWeek.Sub(firstWeek).Hours() / (24 * 7)
	var weeks int = int(math.Round(weeks64))
	//fmt.Println(weeks)

	//Assign
	ConfP.CurrentWeekNo = weeks + 1

	//TODO remove testing
	fmt.Println("ConfP.PlanningHorizon", ConfP.PlanningHorizon)

	//Create date slices
	var t time.Time = ConfP.CurrentWeek
	var WeeklyFutureDates = make([]time.Time, ConfP.PlanningHorizon)

	//TODO remove testing
	fmt.Println("len(WeeklyFutureDates)", len(WeeklyFutureDates))

	for w := 0; w < len(WeeklyFutureDates); w++ {
		//WeeklyFutureDates = append(WeeklyFutureDates, t)
		WeeklyFutureDates[w] = t
		//Increment the date by a week
		t = t.AddDate(0, 0, 7)
	}
	ConfP.EndPlanningHorizon = WeeklyFutureDates[len(WeeklyFutureDates)-1]
	fmt.Println("end of planning horizon: ", ConfP.EndPlanningHorizon)

	//Assign
	ConfP.WeeklyFutureDates = WeeklyFutureDates

	t = ConfP.CurrentWeek.AddDate(0, 0, -156*7)
	var WeeklyPastDates = make([]time.Time, 156)
	for w := 0; w < 156; w++ {
		//WeeklyPastDates = append(WeeklyPastDates, t)
		WeeklyPastDates[w] = t
		//Increment the date by a week
		t = t.AddDate(0, 0, 7)
	}

	//Assign
	ConfP.WeeklyPastDates = WeeklyPastDates

	//End of common section

	var f float64

	for i := 0; i < len(configRows); i++ {
		if configRows[i].Name == "DefaultResponsible" {
			ConfP.DefaultResponsible = configRows[i].Value
		} else if configRows[i].Name == "ForecastInMonth" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.ForecastInMonth = 52.0 //default
			} else {
				ConfP.ForecastInMonth = float64(f)
			}
		} else if configRows[i].Name == "ForecastInQuarter" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.ForecastInQuarter = 24 //default
			} else {
				ConfP.ForecastInQuarter = float64(f)
			}
		} else if configRows[i].Name == "ForecastInHalf" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.ForecastInHalf = 6.0 //default
			} else {
				ConfP.ForecastInHalf = float64(f)
			}
		} else if configRows[i].Name == "CalcWklyPcntBySKU" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.CalcWklyPcntBySKU = 70.0 //default
			} else {
				ConfP.CalcWklyPcntBySKU = float64(f)
			}
		} else if configRows[i].Name == "TrendYearThreshold" {
			ConfP.TrendYearThreshold, err = strconv.Atoi(configRows[i].Value)
			if err != nil {
				ConfP.TrendYearThreshold = 12 //default
			}
		} else if configRows[i].Name == "TrendHalfYearThreshold" {
			ConfP.TrendHalfYearThreshold, err = strconv.Atoi(configRows[i].Value)
			if err != nil {
				ConfP.TrendHalfYearThreshold = 24 //default
			}
		} else if configRows[i].Name == "TrendQuarterThreshold" {
			ConfP.TrendQuarterThreshold, err = strconv.Atoi(configRows[i].Value)
			if err != nil {
				ConfP.TrendQuarterThreshold = 52 //default
			}
		} else if configRows[i].Name == "SkuProfileWeightM1" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.SkuProfileWeightM1 = 1.0 //default
			} else {
				ConfP.SkuProfileWeightM1 = float64(f)
			}
		} else if configRows[i].Name == "SkuProfileWeightM2" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.SkuProfileWeightM2 = 0.5 //default
			} else {
				ConfP.SkuProfileWeightM2 = float64(f)
			}
		} else if configRows[i].Name == "SkuProfileWeightM3" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.SkuProfileWeightM3 = 0.25 //default
			} else {
				ConfP.SkuProfileWeightM3 = float64(f)
			}
		} else if configRows[i].Name == "SmoothWk0Pcnt" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.SmoothWk0Pcnt = 1.0 //default
			} else {
				ConfP.SmoothWk0Pcnt = float64(f)
			}
		} else if configRows[i].Name == "SmoothWk1Pcnt" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.SmoothWk1Pcnt = 0.5 //default
			} else {
				ConfP.SmoothWk1Pcnt = float64(f)
			}
		} else if configRows[i].Name == "SmoothWk2Pcnt" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.SmoothWk2Pcnt = 0.25 //default
			} else {
				ConfP.SmoothWk2Pcnt = float64(f)
			}
		} else if configRows[i].Name == "TrendLimitPcnt" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.TrendLimitPcnt = 0.25 //default
			} else {
				ConfP.TrendLimitPcnt = float64(f)
			}
		} else if configRows[i].Name == "TrendExceptionThresholdPcnt" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.TrendExceptionThresholdPcnt = 0.20 //default
			} else {
				ConfP.TrendExceptionThresholdPcnt = float64(f)
			}
		} else if configRows[i].Name == "AbnormalDemandFactor1" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.AbnormalDemandFactor1 = 12.0 //default
			} else {
				ConfP.AbnormalDemandFactor1 = float64(f)
			}
		} else if configRows[i].Name == "AbnormalDemandFactor2" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.AbnormalDemandFactor2 = 0.70 //default
			} else {
				ConfP.AbnormalDemandFactor2 = f
			}
		} else if configRows[i].Name == "AbnormalDemandMin" {
			f, err = strconv.ParseFloat(configRows[i].Value, 64)
			if err != nil {
				ConfP.AbnormalDemandMin = 0.50 //default
			} else {
				ConfP.AbnormalDemandMin = float64(f)
			}
		}
	}
}
