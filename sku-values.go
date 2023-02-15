package main

import (
	"time"
)

// Inherits or calculates values for a SKU based on settings at a higher level, or dates that are in the past
func (SkuP *SKU) getSkuValues(skuP *SKU) {

	//Check for valid season start and end dates for highly seasonal SKUs
	SkuP.inSeason = false
	if SkuP.IsHighlySeasonal == "S" && SkuP.SeasonStartDate.Valid && SkuP.SeasonEndDate.Valid {
		var seasonStartDate time.Time = SkuP.SeasonStartDate.Time
		var seasonEndDate time.Time = SkuP.SeasonEndDate.Time

		//Cases
		//Check if dates need adjustment to the current year
		// current date: 2021-02-15
		// season start: 2019-01-01, season end: 2019-03-01
		// add year - start: 2020-01-01, season end; 2020-03-01
		// add year - start: 2021-01-01, season end; 2021-03-01
		//OK
		// current date: 2021-02-15
		// season start: 2019-04-01, season end: 2019-04-01
		// add year - start: 2020-04-01, season end; 2020-06-01
		// add year - start: 2021-04-01, season end; 2021-06-01
		//OK
		// current date: 2021-02-15
		// season start: 2018-10-01, season end: 2019-01-15
		// add year - 2019-10-01, season end: 2020-01-15
		// add year - 2020-10-01, season end: 2021-01-15
		// add year - 2021-10-01, season end: 2022-01-15
		//OK
		//Cases 1: Season is more than a year in the past - adjust dates
		if seasonStartDate.Before(ConfigP.CurrentDate.AddDate(-1, 0, 0)) {
			for seasonEndDate.Before(ConfigP.CurrentDate) {
				seasonStartDate = seasonStartDate.AddDate(1, 0, 0)
				seasonEndDate = seasonEndDate.AddDate(1, 0, 0)
			}
			SkuP.SeasonStartDate.Time = seasonStartDate
			SkuP.SeasonEndDate.Time = seasonEndDate

		} else if seasonStartDate == ConfigP.CurrentDate.AddDate(-1, 0, 0) ||
			(seasonStartDate.After(ConfigP.CurrentDate.AddDate(-1, 0, 0)) &&
				seasonStartDate.Before(ConfigP.CurrentDate) && seasonEndDate.Before(ConfigP.CurrentDate)) {
			//Case 2 - Season is a year or less in the past - adjust dates
			// current date: 2021-02-15
			// season start: 2020-10-01, season end: 2021-01-15
			// add year - 2020-10-01, season end: 2021-01-15
			//OK
			SkuP.inSeason = false
			SkuP.SeasonStartDate.Time = seasonStartDate.AddDate(1, 0, 0)
			SkuP.SeasonEndDate.Time = seasonEndDate.AddDate(1, 0, 0)
		} else if seasonStartDate.After(ConfigP.CurrentDate) {
			//Case 3 - today is before the season start date - season in the future
			SkuP.inSeason = false

		} else if seasonStartDate.Before(ConfigP.CurrentDate) && seasonEndDate.After(ConfigP.CurrentDate) {
			//Case 4 - today is after season start date and before season end date - in season
			SkuP.inSeason = true
		} else {
			//Highly seasonal invalid exception
			var ex Exception
			ex.RowId = createRowId(ConfigP.OrgId, "exception_messages")
			ex.OrgId = ConfigP.OrgId
			ex.ExceptionNo = 126
			ex.ItemId = SkuP.ItemId
			ex.LocationId = SkuP.LocationId
			ex.Responsible = SkuP.ResponsibleDemand
			ex.CreateDate = ConfigP.CurrentDate
			insertException(ex)
		}
	}

	//TODO remove testing
	//fmt.Println("SkuP.SeasonStartDate: ", SkuP.SeasonStartDate.Time)
	//fmt.Println("SkuP.SeasonEndDate: ", SkuP.SeasonEndDate.Time)
}
