package main

import (
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/spanner"
	"context"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
)

var (
	ctx               context.Context
	dataClient        *spanner.Client
	cert              string
	token             string
	config            Config
	ConfigP           *Config
	SkuP              *SKU
	CategoryP         *Category
	ProfileP          *Profile
	SalesHistoryRowsP []SalesHistory
	ForecastRowsP     []ForecastBaseline
)

func main() {

	fmt.Println("start of main")

	token = mustGetenv("PUBSUB_VERIFICATION_TOKEN") // token is used to verify push requests.

	ctx = context.Background()

	client, err := pubsub.NewClient(ctx, mustGetenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		fmt.Println("client err: ", err)
		log.Fatal(err)
	}
	defer client.Close()

	topicName := mustGetenv("PUBSUB_TOPIC_FORECAST")
	topic = client.Topic(topicName)

	http.HandleFunc("/pubsub/push", pushHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	//TODO remove testing
	//fmt.Println("port:", port)

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func pushHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	var counter int
	//TODO outgoing triggers
	var triggersSupply []NetChangeTrigger
	var NetChgSupply string

	/* for local testing:
	1. create json such as: [{"OrgId":"1","ItemId":"2920667","LocationId":"12"},{"OrgId":"1","ItemId":"2920667","LocationId":"10"}]
	2. encode base 64 using https://www.base64encode.net/ or similar to give: W3siT3JnSWQiOiIxIiwiSXRlbUlkIjoiMjkyMDY2NyIsIkxvY2F0aW9uSWQiOiIxMiJ9LHsiT3JnSWQiOiIxIiwiSXRlbUlkIjoiMjkyMDY2NyIsIkxvY2F0aW9uSWQiOiIxMCJ9XQ==
	3. put this in the body using Postman or similar:
	{
	    "message": {
	        "data": "W3siT3JnSWQiOiIxIiwiSXRlbUlkIjoiMjkyMDY2NyIsIkxvY2F0aW9uSWQiOiIxMiJ9LHsiT3JnSWQiOiIxIiwiSXRlbUlkIjoiMjkyMDY2NyIsIkxvY2F0aW9uSWQiOiIxMCJ9XQ=="
	    }
	}
	4. add parameters: token <token value>>
	5. use localhost:<debugging port> for url
	*/

	ctx = context.Background()

	//Extract the trigger list from the pub sub push message
	triggersForecast := retrieveTriggers(w, r)

	// Client
	dataClient, err = spanner.NewClient(ctx, "projects/rp-database-s-01/instances/rp-combined/databases/retail")
	// Emulator client
	//dataClient, err = spanner.NewClient(ctx, "projects/rp-forecast-s-01/instances/rp-combined/databases/retail")
	if err != nil {
		log.Println("new spanner client error", err)
		insertErrorLog("0", "", "", err, 1)
	}
	defer dataClient.Close()

	ConfigP = &config
	if len(triggersForecast) > 0 {
		ConfigP.getConfig(triggersForecast[0].OrgId)
	} else {
		//TODO convert to error log?
		fmt.Println("no triggers")
		return
	}

	//start log
	insertBatchLog(ConfigP.OrgId, "START")

	//Loop through the triggers
	for _, trigger := range triggersForecast {

		itemId := trigger.ItemId
		locationId := trigger.LocationId

		//TODO remove
		fmt.Println("org / SKU: ", ConfigP.OrgId, itemId, locationId)

		//Load slices from database
		//SKU
		SkuP, err = querySkuOneItemOneLocation(itemId, locationId)

		//Without a SKU row, cannot calculate a forecast.
		if err == nil {

			//Delete forecast exceptions for this SKU
			fcstExceptions := "(11,23,24,111,112,113,121,122,126)"
			deleteSkuExceptions(fcstExceptions, SkuP)

			//Category
			CategoryP = queryOneCategory(SkuP)

			//Profile - SKU level overrides the category profile
			if SkuP.ProfileId.Valid == false {
				SkuP.ProfileId.Valid = true
				SkuP.ProfileId.StringVal = CategoryP.ProfileId
			}
			ProfileP, _ = getProfile(SkuP)

			//TODO As appropriate revise/assign SKU values based on settings at a higher level
			SkuP.getSkuValues(SkuP)

			//Sales history for a SKU, return count not needed
			SalesHistoryRowsP = querySalesHistoryRows(SkuP)

			//Forecast-baseline for a SKU, return count not needed
			ForecastRowsP = queryForecastBaselineRows(SkuP)

			var Sales, Demand, Symbol = bucketSalesHistory(SkuP, SalesHistoryRowsP)

			var priorYearsDemand []float64 //elements: 0 = minus 3 year, 1 = minus 2 year, 2 = last year
			priorYearsDemand, Demand = replaceAbnormalDemand(SkuP, ProfileP, Sales, Demand, Symbol)

			//Calculate trend, check for override
			if SkuP.TrendOverride.Valid {
				if SkuP.TrendOverride.Float64 < -100 || SkuP.TrendOverride.Float64 > 100 {
					//Ignore trend override if invalid
					SkuP.Trend = calculateTrend(SkuP, priorYearsDemand, Demand)
					//Write exception for invalid trend override
					var ex Exception
					ex.OrgId = ConfigP.OrgId
					ex.ExceptionNo = 122
					ex.ItemId = SkuP.ItemId
					ex.LocationId = SkuP.LocationId
					ex.Responsible = SkuP.ResponsibleDemand
					ex.CreateDate = ConfigP.CurrentDate
					insertException(ex)
				} else {
					SkuP.Trend = float64(SkuP.TrendOverride.Float64)
				}
			} else {
				SkuP.Trend = calculateTrend(SkuP, priorYearsDemand, Demand)
			}

			//Calculate annual forecast, check for override
			if SkuP.AnnualForecastOverride.Valid {
				if SkuP.AnnualForecastOverride.Float64 < 0 || SkuP.AnnualForecastOverride.Float64 > 999999999 {
					//Ignore override annual forecast.
					SkuP.AnnualForecast = (float64(1.0) + SkuP.Trend) * priorYearsDemand[2]
					//Write exception for invalid annual forecast override
					var ex Exception
					ex.OrgId = ConfigP.OrgId
					ex.ExceptionNo = 121
					ex.ItemId = SkuP.ItemId
					ex.LocationId = SkuP.LocationId
					ex.Responsible = SkuP.ResponsibleDemand
					ex.CreateDate = ConfigP.CurrentDate
					insertException(ex)
				} else {
					SkuP.AnnualForecast = float64(SkuP.AnnualForecastOverride.Float64)
				}
			} else {
				SkuP.AnnualForecast = (float64(1.0) + SkuP.Trend) * priorYearsDemand[2]
			}

			//Calculate weekly percentages for SKUs above a threshold
			var SkuWeeklyPcnt []float64

			if SkuP.AnnualForecast > ConfigP.CalcWklyPcntBySKU {
				//Replace profile shifted percentages
				SkuWeeklyPcnt = calculateSkuProfile(Demand)
				//TODO remove
				//fmt.Println("SKU wkly pcnt", SkuWeeklyPcnt)
			} else {
				SkuWeeklyPcnt = ProfileP.ShiftedWeeklyPcnt
			}

			//Calculate weekly forecasts
			var ForecastBase []float64 = calcWeeklyForecasts(SkuP.AnnualForecast, SkuWeeklyPcnt)

			//Calculate weekly forecasts, accumulate into monthly, quarterly, or half-year forecasts
			// and update the database
			forecastTimePeriod(ForecastBase, ForecastRowsP, SkuP)

			//Update sku for annual forecast, trend, last forecast date.
			updateSkuRow(SkuP)

			//Financial planning
			//Already in bucketed format
			updateInsertFinancialPlanningRow("DEMANDM1", Demand[104:156], SkuP)
			updateInsertFinancialPlanningRow("DEMANDM2", Demand[52:105], SkuP)
			updateInsertFinancialPlanningRow("SALESM1", Sales[104:156], SkuP)
			updateInsertFinancialPlanningRow("SALESM2", Sales[52:105], SkuP)

			//TODO remove - for checking
			//fmt.Println("graph format")
			//fmt.Println("dates", ConfigP.WeeklyFutureDates)
			//fmt.Println("base", ForecastBase)
			//fmt.Println("demand-m1", Demand[104:155])
			//fmt.Println("demand-m2", Demand[52:104])
			//fmt.Println("profile wkly pcnt", ProfileP.WeeklyPcnt)
			//fmt.Println("profile wkly pcnt-shifted", ProfileP.ShiftedWeeklyPcnt)

			//_, err = fmt.Fprint(w, "end of forecast")
			//if err != nil {
			//	w.WriteHeader(http.StatusInternalServerError)
			//}

			//If net change, add to the list
			//TODO remove and replace with net change logic, for now, replan all forecasted SKUs
			//TODO later when all SKUs are not replanned, change monitor-db-oh-s-01
			NetChgSupply = "SUPPLY"
			if NetChgSupply == "SUPPLY" {
				trigger := NetChangeTrigger{ConfigP.OrgId, itemId, locationId}
				exists := contains(triggersSupply, trigger)
				if exists == false {
					triggersSupply = append(triggersSupply, trigger)
				}
			}
			counter++
		}
	}

	//At the end of all rows, no need to break into smaller chunks.
	if len(triggersSupply) > 0 {
		err := publishTrigger("supply", triggersSupply)
		if err != nil {
			fmt.Println("error in triggers:", err)
		}
	}

	//end log
	insertBatchLog(ConfigP.OrgId, fmt.Sprintf("END %v", counter))

	//ack
	w.WriteHeader(204)
}

//Start of common section

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		err := errors.New(fmt.Sprintf("env variable not found: %s", k))
		insertErrorLog("0", "", "", err, 1)
	}
	return v
}

//End of common section
