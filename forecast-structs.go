package main

import (
	"cloud.google.com/go/spanner"
	"database/sql"
	"time"
)

type TriggerFcst struct {
	ItemId     string
	LocationId string
}

type RowError struct {
	RowId        uint64
	ColumnErrors []ColumnError
}

type ColumnError struct {
	ColumnName string
	Message    string
}

type SKU struct {
	RowId                  string
	OrgId                  string
	ItemId                 string
	LocationId             string
	ResponsibleDemand      string
	ResponsibleSupply      string
	CategoryId             string
	ProfileId              spanner.NullString
	RouteId                spanner.NullString
	AnnualForecast         float64
	AnnualForecastOverride spanner.NullFloat64
	Trend                  float64
	TrendOverride          spanner.NullFloat64
	StartDate              spanner.NullTime
	EndDate                spanner.NullTime
	LastReplenishmentDate  spanner.NullTime
	IsHighlySeasonal       spanner.NullString //empty, S
	SeasonStartDate        spanner.NullTime
	SeasonEndDate          spanner.NullTime
	Cost                   float64
	inSeason               bool
}

type SalesHistory struct {
	RowId          string
	OrgId          string
	ItemId         string
	LocationId     string
	PostalCode     sql.NullString
	StartDate      time.Time
	SaleQty        float64
	Promotion      sql.NullString
	AbnormalDemand sql.NullString
	AdjustedQty    sql.NullFloat64
}

type Category struct {
	RowId          string
	OrgId          string
	CategoryId     string
	LocationId     string
	Responsible    string
	ProfileId      string
	ProfileDefault string
}

type Profile struct {
	RowId             string
	OrgId             string
	ProfileId         string
	Responsible       string
	WeeklyPcnt        []float64
	ShiftedWeeklyPcnt []float64
}

type ProfileDatabase struct {
	RowId        string
	OrgId        string
	ProfileId    string
	Responsible  string
	WeeklyPcnt1  float64
	WeeklyPcnt2  float64
	WeeklyPcnt3  float64
	WeeklyPcnt4  float64
	WeeklyPcnt5  float64
	WeeklyPcnt6  float64
	WeeklyPcnt7  float64
	WeeklyPcnt8  float64
	WeeklyPcnt9  float64
	WeeklyPcnt10 float64
	WeeklyPcnt11 float64
	WeeklyPcnt12 float64
	WeeklyPcnt13 float64
	WeeklyPcnt14 float64
	WeeklyPcnt15 float64
	WeeklyPcnt16 float64
	WeeklyPcnt17 float64
	WeeklyPcnt18 float64
	WeeklyPcnt19 float64
	WeeklyPcnt20 float64
	WeeklyPcnt21 float64
	WeeklyPcnt22 float64
	WeeklyPcnt23 float64
	WeeklyPcnt24 float64
	WeeklyPcnt25 float64
	WeeklyPcnt26 float64
	WeeklyPcnt27 float64
	WeeklyPcnt28 float64
	WeeklyPcnt29 float64
	WeeklyPcnt30 float64
	WeeklyPcnt31 float64
	WeeklyPcnt32 float64
	WeeklyPcnt33 float64
	WeeklyPcnt34 float64
	WeeklyPcnt35 float64
	WeeklyPcnt36 float64
	WeeklyPcnt37 float64
	WeeklyPcnt38 float64
	WeeklyPcnt39 float64
	WeeklyPcnt40 float64
	WeeklyPcnt41 float64
	WeeklyPcnt42 float64
	WeeklyPcnt43 float64
	WeeklyPcnt44 float64
	WeeklyPcnt45 float64
	WeeklyPcnt46 float64
	WeeklyPcnt47 float64
	WeeklyPcnt48 float64
	WeeklyPcnt49 float64
	WeeklyPcnt50 float64
	WeeklyPcnt51 float64
	WeeklyPcnt52 float64
}

type ForecastBaseline struct {
	RowId        string
	OrgId        string
	ItemId       string
	LocationId   string
	Type         string //B
	Days         int
	StartDate    time.Time
	EndDate      time.Time
	Quantity     float64
	Sold         float64
	SellingPrice sql.NullFloat64
}
