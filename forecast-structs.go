package main

import (
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
	ProfileId              sql.NullString
	RouteId                sql.NullString
	AnnualForecast         float32
	AnnualForecastOverride sql.NullFloat64
	Trend                  float32
	TrendOverride          sql.NullFloat64
	StartDate              sql.NullTime
	EndDate                sql.NullTime
	LastReplenishmentDate  sql.NullTime
	IsHighlySeasonal       string //empty, S
	SeasonStartDate        sql.NullTime
	SeasonEndDate          sql.NullTime
	Cost                   float32
	inSeason               bool
}

type SalesHistory struct {
	RowId          string
	OrgId          string
	ItemId         string
	LocationId     string
	PostalCode     sql.NullString
	StartDate      time.Time
	SaleQty        float32
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
	WeeklyPcnt        []float32
	ShiftedWeeklyPcnt []float32
}

type ProfileDatabase struct {
	RowId        string
	OrgId        string
	ProfileId    string
	Responsible  string
	WeeklyPcnt1  float32
	WeeklyPcnt2  float32
	WeeklyPcnt3  float32
	WeeklyPcnt4  float32
	WeeklyPcnt5  float32
	WeeklyPcnt6  float32
	WeeklyPcnt7  float32
	WeeklyPcnt8  float32
	WeeklyPcnt9  float32
	WeeklyPcnt10 float32
	WeeklyPcnt11 float32
	WeeklyPcnt12 float32
	WeeklyPcnt13 float32
	WeeklyPcnt14 float32
	WeeklyPcnt15 float32
	WeeklyPcnt16 float32
	WeeklyPcnt17 float32
	WeeklyPcnt18 float32
	WeeklyPcnt19 float32
	WeeklyPcnt20 float32
	WeeklyPcnt21 float32
	WeeklyPcnt22 float32
	WeeklyPcnt23 float32
	WeeklyPcnt24 float32
	WeeklyPcnt25 float32
	WeeklyPcnt26 float32
	WeeklyPcnt27 float32
	WeeklyPcnt28 float32
	WeeklyPcnt29 float32
	WeeklyPcnt30 float32
	WeeklyPcnt31 float32
	WeeklyPcnt32 float32
	WeeklyPcnt33 float32
	WeeklyPcnt34 float32
	WeeklyPcnt35 float32
	WeeklyPcnt36 float32
	WeeklyPcnt37 float32
	WeeklyPcnt38 float32
	WeeklyPcnt39 float32
	WeeklyPcnt40 float32
	WeeklyPcnt41 float32
	WeeklyPcnt42 float32
	WeeklyPcnt43 float32
	WeeklyPcnt44 float32
	WeeklyPcnt45 float32
	WeeklyPcnt46 float32
	WeeklyPcnt47 float32
	WeeklyPcnt48 float32
	WeeklyPcnt49 float32
	WeeklyPcnt50 float32
	WeeklyPcnt51 float32
	WeeklyPcnt52 float32
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
	Quantity     float32
	Sold         float32
	SellingPrice sql.NullFloat64
}
