package main

import (
	"cloud.google.com/go/spanner"
	"database/sql"
	"fmt"
	"google.golang.org/api/iterator"
	"time"
)

//NOTE: UserP.Org_id is used with a user token, ConfigP.Org_id is used with a pub sub message.

func queryOneProfile(skuP *SKU) *Profile {

	var profileId string
	var profileRow = new(Profile)
	var weeklyPcnt = make([]float64, 52)

	if skuP.ProfileId.Valid {
		profileId = skuP.ProfileId.StringVal
	}

	// Start with a flat profile to avoid out of bounds errors
	for i := 0; i < len(weeklyPcnt); i++ {
		weeklyPcnt[i] = 0.01923
	}

	sSQL := fmt.Sprintf(`SELECT row_id,org_id,profile_id,responsible, 
		weekly_percent1,weekly_percent2,weekly_percent3,weekly_percent4,weekly_percent5,weekly_percent6,weekly_percent7, 
		weekly_percent8,weekly_percent9,weekly_percent10,weekly_percent11,weekly_percent12,weekly_percent13,weekly_percent14, 
		weekly_percent15,weekly_percent16,weekly_percent17,weekly_percent18,weekly_percent19,weekly_percent20,weekly_percent21, 
		weekly_percent22,weekly_percent23,weekly_percent24,weekly_percent25,weekly_percent26,weekly_percent27,weekly_percent28, 
		weekly_percent29,weekly_percent30,weekly_percent31,weekly_percent32,weekly_percent33,weekly_percent34,weekly_percent35, 
		weekly_percent36,weekly_percent37,weekly_percent38,weekly_percent39,weekly_percent40,weekly_percent41,weekly_percent42, 
		weekly_percent43,weekly_percent44,weekly_percent45,weekly_percent46,weekly_percent47,weekly_percent48,weekly_percent49, 
		weekly_percent50,weekly_percent51,weekly_percent52 
		FROM profiles WHERE org_id = '%s' AND profile_id = '%s' ;`, skuP.OrgId, profileId)

	stmt := spanner.Statement{SQL: sSQL}
	iter := dataClient.Single().Query(ctx, stmt)
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			var ex Exception
			ex.RowId = createRowId(ConfigP.OrgId, "exceptions")
			ex.OrgId = ConfigP.OrgId
			ex.ExceptionNo = 113
			ex.ItemId = SkuP.ItemId
			ex.LocationId = SkuP.LocationId
			ex.Responsible = SkuP.ResponsibleDemand
			ex.CreateDate = ConfigP.CurrentDate
			insertException(ex)

			profileRow.WeeklyPcnt = weeklyPcnt
			return profileRow
		}
		if err != nil {
			insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
			return profileRow
		}
		profileRowDB := ProfileDatabase{}
		err = row.Columns(&profileRowDB.RowId, &profileRowDB.OrgId, &profileRowDB.ProfileId, &profileRowDB.Responsible,
			&profileRowDB.WeeklyPcnt1, &profileRowDB.WeeklyPcnt2, &profileRowDB.WeeklyPcnt3, &profileRowDB.WeeklyPcnt4,
			&profileRowDB.WeeklyPcnt5, &profileRowDB.WeeklyPcnt6, &profileRowDB.WeeklyPcnt7, &profileRowDB.WeeklyPcnt8,
			&profileRowDB.WeeklyPcnt9, &profileRowDB.WeeklyPcnt10, &profileRowDB.WeeklyPcnt11, &profileRowDB.WeeklyPcnt12,
			&profileRowDB.WeeklyPcnt13, &profileRowDB.WeeklyPcnt14, &profileRowDB.WeeklyPcnt15, &profileRowDB.WeeklyPcnt16,
			&profileRowDB.WeeklyPcnt17, &profileRowDB.WeeklyPcnt18, &profileRowDB.WeeklyPcnt19, &profileRowDB.WeeklyPcnt20,
			&profileRowDB.WeeklyPcnt21, &profileRowDB.WeeklyPcnt22, &profileRowDB.WeeklyPcnt23, &profileRowDB.WeeklyPcnt24,
			&profileRowDB.WeeklyPcnt25, &profileRowDB.WeeklyPcnt26, &profileRowDB.WeeklyPcnt27, &profileRowDB.WeeklyPcnt28,
			&profileRowDB.WeeklyPcnt29, &profileRowDB.WeeklyPcnt30, &profileRowDB.WeeklyPcnt31, &profileRowDB.WeeklyPcnt32,
			&profileRowDB.WeeklyPcnt33, &profileRowDB.WeeklyPcnt34, &profileRowDB.WeeklyPcnt35, &profileRowDB.WeeklyPcnt36,
			&profileRowDB.WeeklyPcnt37, &profileRowDB.WeeklyPcnt38, &profileRowDB.WeeklyPcnt39, &profileRowDB.WeeklyPcnt40,
			&profileRowDB.WeeklyPcnt41, &profileRowDB.WeeklyPcnt42, &profileRowDB.WeeklyPcnt43, &profileRowDB.WeeklyPcnt44,
			&profileRowDB.WeeklyPcnt45, &profileRowDB.WeeklyPcnt46, &profileRowDB.WeeklyPcnt47, &profileRowDB.WeeklyPcnt48,
			&profileRowDB.WeeklyPcnt49, &profileRowDB.WeeklyPcnt50, &profileRowDB.WeeklyPcnt51, &profileRowDB.WeeklyPcnt52)
		if err != nil {
			insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
			return profileRow
		}
		//columns to array
		weeklyPcnt[0] = profileRowDB.WeeklyPcnt1
		weeklyPcnt[1] = profileRowDB.WeeklyPcnt2
		weeklyPcnt[2] = profileRowDB.WeeklyPcnt3
		weeklyPcnt[3] = profileRowDB.WeeklyPcnt4
		weeklyPcnt[4] = profileRowDB.WeeklyPcnt5
		weeklyPcnt[5] = profileRowDB.WeeklyPcnt6
		weeklyPcnt[6] = profileRowDB.WeeklyPcnt7
		weeklyPcnt[7] = profileRowDB.WeeklyPcnt8
		weeklyPcnt[8] = profileRowDB.WeeklyPcnt9
		weeklyPcnt[9] = profileRowDB.WeeklyPcnt10
		weeklyPcnt[10] = profileRowDB.WeeklyPcnt11
		weeklyPcnt[11] = profileRowDB.WeeklyPcnt12
		weeklyPcnt[12] = profileRowDB.WeeklyPcnt13
		weeklyPcnt[13] = profileRowDB.WeeklyPcnt14
		weeklyPcnt[14] = profileRowDB.WeeklyPcnt15
		weeklyPcnt[15] = profileRowDB.WeeklyPcnt16
		weeklyPcnt[16] = profileRowDB.WeeklyPcnt17
		weeklyPcnt[17] = profileRowDB.WeeklyPcnt18
		weeklyPcnt[18] = profileRowDB.WeeklyPcnt19
		weeklyPcnt[19] = profileRowDB.WeeklyPcnt20
		weeklyPcnt[20] = profileRowDB.WeeklyPcnt21
		weeklyPcnt[21] = profileRowDB.WeeklyPcnt22
		weeklyPcnt[22] = profileRowDB.WeeklyPcnt23
		weeklyPcnt[23] = profileRowDB.WeeklyPcnt24
		weeklyPcnt[24] = profileRowDB.WeeklyPcnt25
		weeklyPcnt[25] = profileRowDB.WeeklyPcnt26
		weeklyPcnt[26] = profileRowDB.WeeklyPcnt27
		weeklyPcnt[27] = profileRowDB.WeeklyPcnt28
		weeklyPcnt[28] = profileRowDB.WeeklyPcnt29
		weeklyPcnt[29] = profileRowDB.WeeklyPcnt30
		weeklyPcnt[30] = profileRowDB.WeeklyPcnt31
		weeklyPcnt[31] = profileRowDB.WeeklyPcnt32
		weeklyPcnt[32] = profileRowDB.WeeklyPcnt33
		weeklyPcnt[33] = profileRowDB.WeeklyPcnt34
		weeklyPcnt[34] = profileRowDB.WeeklyPcnt35
		weeklyPcnt[35] = profileRowDB.WeeklyPcnt36
		weeklyPcnt[36] = profileRowDB.WeeklyPcnt37
		weeklyPcnt[37] = profileRowDB.WeeklyPcnt38
		weeklyPcnt[38] = profileRowDB.WeeklyPcnt39
		weeklyPcnt[39] = profileRowDB.WeeklyPcnt40
		weeklyPcnt[40] = profileRowDB.WeeklyPcnt41
		weeklyPcnt[41] = profileRowDB.WeeklyPcnt42
		weeklyPcnt[42] = profileRowDB.WeeklyPcnt43
		weeklyPcnt[43] = profileRowDB.WeeklyPcnt44
		weeklyPcnt[44] = profileRowDB.WeeklyPcnt45
		weeklyPcnt[45] = profileRowDB.WeeklyPcnt46
		weeklyPcnt[46] = profileRowDB.WeeklyPcnt47
		weeklyPcnt[47] = profileRowDB.WeeklyPcnt48
		weeklyPcnt[48] = profileRowDB.WeeklyPcnt49
		weeklyPcnt[49] = profileRowDB.WeeklyPcnt50
		weeklyPcnt[50] = profileRowDB.WeeklyPcnt51
		weeklyPcnt[51] = profileRowDB.WeeklyPcnt52

		//load struct
		profileRow.RowId = profileRowDB.RowId
		profileRow.OrgId = profileRowDB.OrgId
		profileRow.ProfileId = profileRowDB.ProfileId
		profileRow.Responsible = profileRowDB.Responsible
		profileRow.WeeklyPcnt = weeklyPcnt

		//Shift profile percentages to the current week
		//For example, profile percentages start at week 1 and extend to week 52
		//If the current date is week 10, then create a new slice of profile percentages starting at 10
		profileRow.ShiftedWeeklyPcnt = shift(ConfigP.CurrentWeekNo-1, weeklyPcnt)

		// Should only be one row
		break
	}
	return profileRow
}

func querySkuOneItemOneLocation(itemId string, locationId string) (*SKU, error) {

	var err error
	var row *spanner.Row
	var skuRow = new(SKU)

	sSQL := fmt.Sprintf(`SELECT row_id,org_id,item_id,location_id,category_id,profile_id,
	  responsible_demand,annual_forecast_override,trend_override,start_date,end_date,last_replenishment_date,
	  is_highly_seasonal,season_start_date,season_end_date,cost
		FROM sku WHERE org_id = '%s' AND item_id = '%s' AND location_id = '%s' `, ConfigP.OrgId, itemId, locationId)

	stmt := spanner.Statement{SQL: sSQL}
	iter := dataClient.Single().Query(ctx, stmt)
	defer iter.Stop()
	for {
		row, err = iter.Next()
		if err == iterator.Done {
			var ex Exception
			ex.RowId = createRowId(ConfigP.OrgId, "exception")
			ex.OrgId = ConfigP.OrgId
			ex.ExceptionNo = 111
			ex.ItemId = SkuP.ItemId
			ex.LocationId = SkuP.LocationId
			ex.Responsible = SkuP.ResponsibleDemand
			ex.CreateDate = ConfigP.CurrentDate
			insertException(ex)
			return skuRow, err
		}
		if err != nil {
			insertErrorLog(ConfigP.OrgId, itemId, locationId, err, 1)
			return skuRow, err
		}
		err := row.Columns(&skuRow.RowId, &skuRow.OrgId, &skuRow.ItemId, &skuRow.LocationId, &skuRow.CategoryId, &skuRow.ProfileId,
			&skuRow.ResponsibleDemand, &skuRow.AnnualForecastOverride, &skuRow.TrendOverride, &skuRow.StartDate, &skuRow.EndDate,
			&skuRow.LastReplenishmentDate, &skuRow.IsHighlySeasonal, &skuRow.SeasonStartDate, &skuRow.SeasonEndDate,
			&skuRow.Cost)
		if err != nil {
			insertErrorLog(ConfigP.OrgId, itemId, locationId, err, 1)
			return skuRow, err
		}
		// Should only be one row
		break
	}
	return skuRow, err
}

func querySalesHistoryRows(skuP *SKU) []SalesHistory {

	var salesHistoryRows []SalesHistory

	sSQL := fmt.Sprintf(`SELECT row_id, org_id,item_id,location_id,postal_code,start_date,sale_qty,
        promotion,abnormal_demand,adjusted_sale_qty 
		FROM history_sales_weekly 
		WHERE org_id = '%s' AND item_id = '%s' AND location_id = '%s' 
		ORDER BY start_date ;`, skuP.OrgId, skuP.ItemId, skuP.LocationId)

	stmt := spanner.Statement{SQL: sSQL}
	iter := dataClient.Single().Query(ctx, stmt)
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			return salesHistoryRows
		}
		if err != nil {
			insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
			return salesHistoryRows
		}
		nextSalesHistory := SalesHistory{}
		if err := row.Columns(&nextSalesHistory.RowId, &nextSalesHistory.OrgId, &nextSalesHistory.ItemId,
			&nextSalesHistory.LocationId, &nextSalesHistory.PostalCode, &nextSalesHistory.StartDate,
			&nextSalesHistory.SaleQty, &nextSalesHistory.Promotion, &nextSalesHistory.AbnormalDemand,
			&nextSalesHistory.AdjustedQty); err != nil {
			insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
			return salesHistoryRows
		}
		salesHistoryRows = append(salesHistoryRows, nextSalesHistory)
	}
	return salesHistoryRows
}

func queryForecastBaselineRows(skuP *SKU) []ForecastBaseline {

	var forecastRows []ForecastBaseline

	sSQL := fmt.Sprintf(`SELECT row_id, org_id, item_id, location_id, type, days, start_date, end_date, quantity, sold, selling_price 
		FROM forecast 
		WHERE org_id = '%s' AND item_id = '%s' AND location_id = '%s' 
		ORDER BY start_date ;`, skuP.OrgId, skuP.ItemId, skuP.LocationId)

	stmt := spanner.Statement{SQL: sSQL}
	iter := dataClient.Single().Query(ctx, stmt)
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			return forecastRows
		}
		if err != nil {
			insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
			return forecastRows
		}
		nextForecast := ForecastBaseline{}
		if err := row.Columns(&nextForecast.RowId, &nextForecast.OrgId, &nextForecast.ItemId, &nextForecast.LocationId,
			&nextForecast.Type, &nextForecast.Days, &nextForecast.StartDate, &nextForecast.EndDate, &nextForecast.Quantity,
			&nextForecast.Sold, &nextForecast.SellingPrice); err != nil {
			insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
			return forecastRows
		}
		forecastRows = append(forecastRows, nextForecast)
	}
	return forecastRows
}

func queryOneCategory(skuP *SKU) *Category {

	var categoryRow = new(Category)

	sSQL := fmt.Sprintf(`SELECT row_id,org_id,location_id,responsible,profile_id
		FROM category WHERE org_id = '%s' AND category_id = '%s' AND location_id = '%s' `, skuP.OrgId, skuP.CategoryId,
		skuP.LocationId)

	stmt := spanner.Statement{SQL: sSQL}
	iter := dataClient.Single().Query(ctx, stmt)
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			var ex Exception
			ex.RowId = createRowId(ConfigP.OrgId, "exceptions")
			ex.OrgId = ConfigP.OrgId
			ex.ExceptionNo = 112
			ex.ItemId = SkuP.ItemId
			ex.LocationId = SkuP.LocationId
			ex.Responsible = SkuP.ResponsibleDemand
			ex.CreateDate = ConfigP.CurrentDate
			insertException(ex)
			return categoryRow
		}
		if err != nil {
			insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
			return categoryRow
		}
		err = row.Columns(&categoryRow.RowId, &categoryRow.OrgId, &categoryRow.LocationId, &categoryRow.Responsible,
			&categoryRow.ProfileId)
		if err != nil {
			insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
			return categoryRow
		}
		// Should only be one row
		break
	}
	return categoryRow
}

// Update an existing forecast row
func updateFcstRow(rowId string, skuP *SKU, baseFcstType string, days int,
	startDate time.Time, endDate time.Time, quantity float64, sold float64, sellingPrice sql.NullFloat64) {

	sCol := []string{"row_id", "days", "start_date", "end_date", "quantity"}

	m := []*spanner.Mutation{
		spanner.Update("forecast", sCol, []interface{}{rowId, days, startDate, endDate, quantity}),
	}
	_, err := dataClient.Apply(ctx, m)
	if err != nil {
		insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
	}
}

// Insert a new forecast row
func insertFcstRow(skuP *SKU, baseFcstType string, days int,
	startDate time.Time, endDate time.Time, quantity float64, sold float64, sellingPrice sql.NullFloat64) string {

	sCol := []string{"row_id", "org_id", "item_id", "location_id", "type", "days", "start_date", "end_date", "quantity", "sold",
		"selling_price"}
	rowId := createRowId(skuP.OrgId, "forecast")

	//Selling price is optional, default to null for baseline
	var sp sql.NullFloat64
	sp.Valid = false
	//sp.Float64 = float64(SkuP.SellingPrice)

	m := []*spanner.Mutation{
		spanner.InsertOrUpdate("forecast", sCol, []interface{}{rowId, skuP.OrgId, skuP.ItemId, skuP.LocationId,
			baseFcstType, days, startDate, endDate, quantity, sold, sp}),
	}
	_, err := dataClient.Apply(ctx, m)
	if err != nil {
		insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
	}
	return rowId
}

// Delete an existing forecast row
func deleteFcstRow(rowId string, skuP *SKU) {

	m := []*spanner.Mutation{
		spanner.Delete("forecast", spanner.Key{rowId}),
	}
	_, err := dataClient.Apply(ctx, m)
	if err != nil {
		insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
	}
}

// Update SKU row
func updateSkuRow(skuP *SKU) {

	sCol := []string{"row_id", "annual_forecast", "trend", "last_reforecast_date"}

	m := []*spanner.Mutation{
		spanner.Update("sku", sCol, []interface{}{skuP.RowId, skuP.AnnualForecast, skuP.Trend, ConfigP.CurrentDate}),
	}
	_, err := dataClient.Apply(ctx, m)
	if err != nil {
		insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
	}
}
