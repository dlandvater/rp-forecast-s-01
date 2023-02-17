package main

import (
	"cloud.google.com/go/spanner"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/iterator"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//Start of common section

type UserRow struct {
	RowId        uint64
	OrgId        string
	UserId       string
	Email        string
	ReadRights   string
	UpdateRights string
	ReadS        []string
	UpdateS      []string
}

type ConfigRow struct {
	RowId     string
	OrgId     string
	Name      string
	LabelText string
	Value     string
	DataTyp   string
	MinValue  string
	MaxValue  string
}

type Config struct {
	RowId                               uint64
	OrgId                               string
	OrgName                             string
	StartDayOfWeek                      string
	WeekdayStartDayOfWeek               int
	OverrideDateTime                    string
	CurrentDate                         time.Time
	CurrentWeek                         time.Time
	CurrentWeekNo                       int
	EmptyDate                           time.Time
	ZeroDate                            time.Time
	PlanningHorizon                     int
	DailyHorizon                        int
	WeeklyFutureDates                   []time.Time
	WeeklyPastDates                     []time.Time
	EndPlanningHorizon                  time.Time
	DefaultResponsible                  string
	TrendYearThreshold                  int
	TrendHalfYearThreshold              int
	TrendQuarterThreshold               int
	TrendExceptionThresholdPcnt         float64
	TrendLimitPcnt                      float64
	AbnormalDemandFactor1               float64
	AbnormalDemandFactor2               float64
	AbnormalDemandMin                   float64
	SkuProfileWeightM3                  float64
	SkuProfileWeightM2                  float64
	SkuProfileWeightM1                  float64
	ForecastInMonth                     float64
	ForecastInQuarter                   float64
	ForecastInHalf                      float64
	CalcWklyPcntBySKU                   float64
	SmoothWk0Pcnt                       float64
	SmoothWk1Pcnt                       float64
	SmoothWk2Pcnt                       float64
	NotSelectableLevelLocationHierarchy int
	NotSelectableLevelItemHierarchy     int
	PastDaysFcst                        int
	PastDaysSuply                       int
}

type Exception struct {
	RowId          string
	OrgId          string
	ExceptionNo    int
	ItemId         string
	LocationId     string
	Responsible    string
	CreateDate     time.Time
	DeleteDate     time.Time
	ExceptionDate1 time.Time
	ExceptionDate2 time.Time
	ExceptionDate3 time.Time
	ExceptionQty1  float64
	ExceptionQty2  float64
	ExceptionQty3  float64
	DeleteCode     string
}

func createRowId(orgId string, tableName string) string {
	var n int64
	var typ string

	//TODO can increase the number of digits in the time component if needed.

	// Type of table
	switch tableName {
	case "error_log":
		typ = "01"
	case "queue_forecast":
		typ = "02"
	case "queue_supply":
		typ = "03"
	case "queue_loadbuilding":
		typ = "04"
	case "queue_on_hand_on_order":
		typ = "05"
	case "exceptions":
		typ = "06"
	case "forecast":
		typ = "07"
	case "planned_orders":
		typ = "08"
	case "financial_planning":
		typ = "09"
	default:
		typ = "0"
	}
	//	rowId := createRowId(orgId, "error_log")
	// Create unique row ID using first 14 digits of nano time reversed + orgId + type of table
	time.Sleep(time.Nanosecond)
	n = time.Now().UnixNano()
	s := strconv.FormatInt(n, 10) // Nano time to string
	r := StringReverse(s)         // Reverse time to prevent database hot spots
	t := r[0:15]                  //Truncate high-end digits
	rowId := t + orgId + typ
	return rowId
}

func StringReverse(InputString string) (ResultString string) {
	// iterating and prepending
	for _, c := range InputString {
		ResultString = string(c) + ResultString
	}
	return
}

func insertErrorLog(orgId string, itemId string, locationId string, err error, skip int) {

	_, s, n, _ := runtime.Caller(skip)
	ss := strings.Split(s, "/")
	s = ss[len(ss)-1]
	errSource := fmt.Sprintf("%s%s%d", s, ":", n)
	if len(errSource) > 50 {
		errSource = errSource[0:50]
	}
	errMsg := err.Error()
	if len(errMsg) > 1024 {
		errMsg = errMsg[0:1023]
	}
	sCol := []string{"row_id", "org_id", "create_date", "item_id", "location_id", "source", "message"}
	rowId := createRowId(orgId, "error_log")

	m := []*spanner.Mutation{
		spanner.InsertOrUpdate("error_log", sCol, []interface{}{rowId, orgId, time.Now(), itemId, locationId,
			errSource, errMsg}),
	}
	_, err = dataClient.Apply(ctx, m)
	if err != nil {
		log.Printf("write error log rp-forecast", err)
	}
}

func insertBatchLog(orgId string, bMsg string) {

	if len(bMsg) > 1024 {
		bMsg = bMsg[0:1023]
	}
	rowId := createRowId(orgId, "error_log")
	sCol := []string{"row_id", "org_id", "create_date", "message"}

	m := []*spanner.Mutation{
		spanner.InsertOrUpdate("error_log", sCol, []interface{}{rowId, orgId, time.Now(), bMsg}),
	}
	_, err := dataClient.Apply(ctx, m)
	log.Printf("write batch log rp-forecast", err)
}

func queryConfigurationRows(orgId string) []ConfigRow {

	var configRows []ConfigRow
	var counter int

	sSQL := fmt.Sprintf(`SELECT row_id, org_id, name, label_text, value, data_typ, min_value, max_value 
		FROM config WHERE org_id = '%s' ORDER BY name ;`, orgId)

	stmt := spanner.Statement{SQL: sSQL}
	iter := dataClient.Single().Query(ctx, stmt)
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			return configRows
		}
		if err != nil {
			// No organization has been determined yet
			insertErrorLog("0", "", "", err, 1)
			return configRows
		}
		nextConfig := ConfigRow{}
		if err2 := row.Columns(&nextConfig.RowId, &nextConfig.OrgId, &nextConfig.Name, &nextConfig.LabelText,
			&nextConfig.Value, &nextConfig.DataTyp, &nextConfig.MinValue, &nextConfig.MaxValue); err2 != nil {
			// No organization has been determined yet
			insertErrorLog("0", "", "", err, 1)
			return configRows
		}
		configRows = append(configRows, nextConfig)
		counter++
	}
	if counter == 0 {
		err0 := errors.New("Config rows not found")
		// No organization has been determined yet
		insertErrorLog("0", "", "", err0, 1)
	}
	return configRows
}

func insertException(exception Exception) {

	sCol := []string{"row_id", "org_id", "exception_no", "create_date", "delete_date", "item_id", "location_id",
		"responsible", "date1", "date2", "date3", "quantity1", "quantity2", "quantity3", "delete_code"}
	rowId := createRowId(exception.OrgId, "error_log")

	m := []*spanner.Mutation{
		spanner.InsertOrUpdate("error_log", sCol, []interface{}{rowId, exception.OrgId, exception.ExceptionNo,
			exception.CreateDate, exception.DeleteDate, exception.ItemId, exception.LocationId, exception.Responsible,
			exception.ExceptionDate1, exception.ExceptionDate2, exception.ExceptionDate3,
			exception.ExceptionQty1, exception.ExceptionQty2, exception.ExceptionQty3, exception.DeleteCode}),
	}
	_, err := dataClient.Apply(ctx, m)
	if err != nil {
		insertErrorLog(exception.OrgId, exception.ItemId, exception.LocationId, err, 1)
	}
}

// Update or insert a financial planning row
func updateInsertFinancialPlanningRow(fplanType string, weeklyQty []float64, skuP *SKU) {

	var rowId string

	/* NOTE: not using update or insert here because the formula for calculating the row ID uses time and it will be
	different for an update when compared to the initial insert. So, the result would be duplicate rows. */

	// Section 1: get the rowId if a row already exists.
	sSQL := fmt.Sprintf(`SELECT row_id FROM financial_planning WHERE org_id = '%s' AND item_id = '%s' 
        AND location_id = '%s' AND type = '%s' `,
		skuP.OrgId, skuP.ItemId, skuP.LocationId, fplanType)

	stmt := spanner.Statement{SQL: sSQL}
	iter := dataClient.Single().Query(ctx, stmt)
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			// Need to insert a new row, generate row ID
			rowId = createRowId(skuP.OrgId, "financial_planning")
			break
		}
		if err != nil {
			insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
			return
		}
		err = row.Columns(&rowId)
		if err != nil {
			insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
			return
		} else {
			// Use rowId to update financial planning row
			break
		}
	}
	// Section 2: update or insert the row.
	sCol := []string{"row_id", "org_id", "item_id", "location_id", "type", "WEEK_1", "WEEK_2", "WEEK_3", "WEEK_4", "WEEK_5",
		"WEEK_6", "WEEK_7", "WEEK_8", "WEEK_9", "WEEK_10", "WEEK_11", "WEEK_12", "WEEK_13", "WEEK_14", "WEEK_15", "WEEK_16",
		"WEEK_17", "WEEK_18", "WEEK_19", "WEEK_20", "WEEK_21", "WEEK_22", "WEEK_23", "WEEK_24", "WEEK_25", "WEEK_26", "WEEK_27",
		"WEEK_28", "WEEK_29", "WEEK_30", "WEEK_31", "WEEK_32", "WEEK_33", "WEEK_34", "WEEK_35", "WEEK_36", "WEEK_37", "WEEK_38",
		"WEEK_39", "WEEK_40", "WEEK_41", "WEEK_42", "WEEK_43", "WEEK_44", "WEEK_45", "WEEK_46", "WEEK_47", "WEEK_48", "WEEK_49",
		"WEEK_50", "WEEK_51", "WEEK_52"}

	m := []*spanner.Mutation{
		spanner.InsertOrUpdate("financial_planning", sCol, []interface{}{rowId, skuP.OrgId, skuP.ItemId, skuP.LocationId,
			fplanType, weeklyQty[0], weeklyQty[1], weeklyQty[2], weeklyQty[3], weeklyQty[4], weeklyQty[5], weeklyQty[6],
			weeklyQty[7], weeklyQty[8], weeklyQty[9], weeklyQty[10], weeklyQty[11], weeklyQty[12], weeklyQty[13],
			weeklyQty[14], weeklyQty[15], weeklyQty[16], weeklyQty[17], weeklyQty[18], weeklyQty[19], weeklyQty[20],
			weeklyQty[21], weeklyQty[22], weeklyQty[23], weeklyQty[24], weeklyQty[25], weeklyQty[26], weeklyQty[27],
			weeklyQty[28], weeklyQty[29], weeklyQty[30], weeklyQty[31], weeklyQty[32], weeklyQty[33], weeklyQty[34],
			weeklyQty[35], weeklyQty[36], weeklyQty[37], weeklyQty[38], weeklyQty[39], weeklyQty[40], weeklyQty[41],
			weeklyQty[42], weeklyQty[43], weeklyQty[44], weeklyQty[45], weeklyQty[46], weeklyQty[47], weeklyQty[48],
			weeklyQty[49], weeklyQty[50], weeklyQty[51]}),
	}
	_, err := dataClient.Apply(ctx, m)
	if err != nil {
		insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
	}
}

// Delete an existing financial planning row
func deleteFinancialPlanningRow(rowId string, orgId string, itemId string, locationId string) {

	// uses slice of mutations
	var m *spanner.Mutation
	var mm []*spanner.Mutation

	m = spanner.Delete("financial_planning", spanner.Key{rowId})
	mm = append(mm, m)

	_, err := dataClient.Apply(ctx, mm)
	if err != nil {
		insertErrorLog(orgId, itemId, locationId, err, 1)
	}
}

func queryUserRows() []UserRow {

	var userRows []UserRow

	sSQL := "SELECT row_id,org_id,user_id,read_rights,update_rights " +
		"FROM users " +
		"ORDER BY user_id ;"

	stmt := spanner.Statement{SQL: sSQL}
	iter := dataClient.Single().Query(ctx, stmt)
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			return userRows
		}
		if err != nil {
			// No organization has been determined yet
			insertErrorLog("0", "", "", err, 1)
			return userRows
		}
		nextUser := UserRow{}
		if err2 := row.Columns(&nextUser.RowId, &nextUser.OrgId, &nextUser.UserId, &nextUser.ReadRights, &nextUser.UpdateRights); err2 != nil {
			// No organization has been determined yet
			insertErrorLog("0", "", "", err, 1)
			return userRows
		}
		//convert rights to slice
		s := nextUser.ReadRights
		S := strings.Split(s, ",")
		nextUser.ReadS = S

		s = nextUser.UpdateRights
		S = strings.Split(s, ",")
		nextUser.UpdateS = S

		//TODO remove testing
		fmt.Println("nextUser", nextUser.UserId, nextUser.OrgId)

		userRows = append(userRows, nextUser)
	}
	return userRows
}

// Delete exceptions for a SKU
// func deleteSkuExceptions(exception_list string, skuP *SKU) {
//
//		//Not using mutation delete - don't have row IDs, may want to revise this at some point
//		_, err := dataClient.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
//			stmt := spanner.Statement{
//				SQL: fmt.Sprintf(`DELETE FROM exceptions WHERE "+
//					"org_id = '%s' AND item_id = '%s' AND location_id = '%s' AND exception_no IN %s `,
//					skuP.OrgId, skuP.ItemId, skuP.LocationId, exception_list),
//			}
//			_, err := txn.Update(ctx, stmt)
//			if err != nil {
//				insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
//			}
//			return nil
//		})
//		if err != nil {
//			insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
//		}
//	}
func deleteSkuExceptions(exception_list string, skuP *SKU) {

	_, err := dataClient.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL: fmt.Sprintf(`DELETE FROM exceptions WHERE 
				org_id = '%s' AND item_id = '%s' AND location_id = '%s' AND exception_no IN %s `,
				skuP.OrgId, skuP.ItemId, skuP.LocationId, exception_list),
		}
		iter := txn.Query(ctx, stmt)
		defer iter.Stop()
		for {
			row, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			var (
				orgId      string
				itemId     string
				locationId string
			)
			if err := row.Columns(&orgId, &itemId, &locationId); err != nil {
				insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
				return err
			}
			fmt.Println("%s %s %s\n", orgId, itemId, locationId)
		}
		fmt.Println("%d record(s) deleted.\n", iter.RowCount)
		return nil
	})
	if err != nil {
		insertErrorLog(skuP.OrgId, skuP.ItemId, skuP.LocationId, err, 1)
	}
	return
}

//End of common section
