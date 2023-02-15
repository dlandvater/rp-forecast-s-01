package main

func bucketSalesHistory(skuP *SKU, SalesHistoryRowsP []SalesHistory) ([]float32, []float32, []string) {

	var Sales = make([]float32, 156)
	var Demand = make([]float32, 156)
	var Symbol = make([]string, 156)
	var WeeklyPastDates = ConfigP.WeeklyPastDates

	//Prevent array out of bounds, since looks ahead one element
	var lastElement = len(WeeklyPastDates) - 1
	for i := 0; i < len(SalesHistoryRowsP); i++ {
		//Ignore dates earlier than the start of the history dates
		if SalesHistoryRowsP[i].StartDate.Before(WeeklyPastDates[0]) {
			continue
		}
		for w := 0; w < len(WeeklyPastDates)-1; w++ {
			//Conditions:
			//Equal to one of the weekly dates
			//After one weekly date and before the next
			//Equal to or after the last weekly date
			if (SalesHistoryRowsP[i].StartDate.Equal(WeeklyPastDates[w])) ||
				((SalesHistoryRowsP[i].StartDate.After(WeeklyPastDates[w])) && (SalesHistoryRowsP[i].StartDate.Before(WeeklyPastDates[w+1]))) ||
				((SalesHistoryRowsP[i].StartDate.Equal(WeeklyPastDates[lastElement])) || (SalesHistoryRowsP[i].StartDate.After(WeeklyPastDates[lastElement]))) {

				//Index for last element
				if (SalesHistoryRowsP[i].StartDate.Equal(WeeklyPastDates[lastElement])) || (SalesHistoryRowsP[i].StartDate.After(WeeklyPastDates[lastElement])) {
					w = lastElement
				}

				//Accumulate sales
				Sales[w] = Sales[w] + SalesHistoryRowsP[i].SaleQty

				//Check for adjusted demand
				if SalesHistoryRowsP[i].AdjustedQty.Valid {
					Demand[w] = Demand[w] + float32(SalesHistoryRowsP[i].AdjustedQty.Float64)
				} else {
					//Ignore adjusted quantity
					Demand[w] = Demand[w] + SalesHistoryRowsP[i].SaleQty
				}

				if SalesHistoryRowsP[i].Promotion.Valid &&
					SalesHistoryRowsP[i].Promotion.String == "P" {
					Symbol[w] = Symbol[w] + SalesHistoryRowsP[i].Promotion.String
				}

				if SalesHistoryRowsP[i].AdjustedQty.Valid &&
					SalesHistoryRowsP[i].AbnormalDemand.String == "A" {
					Symbol[w] = Symbol[w] + SalesHistoryRowsP[i].AbnormalDemand.String
				}

				//Found the date, move to the next sales history row
				break
			}
		}
	}
	return Sales, Demand, Symbol
}
