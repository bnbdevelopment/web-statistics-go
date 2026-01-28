package structs

import (
	"time"
)

type WebMetric struct {
	Id        uint      `gorm:"primaryKey"`
	Timestamp time.Time `gorm:"type:timestamp with time zone"`
	Page      string    `gorm:"size:255"`
	Site      string    `gorm:"size:255"`
	Ip        string    `gorm:"size:255"`
	SessionId string    `gorm:"size:255"`

	// Geolocation fields (nullable for graceful degradation)
	CountryCode *string  `gorm:"size:2"`              // ISO 3166-1 alpha-2 (e.g., "US", "GB")
	CountryName *string  `gorm:"size:255"`            // Full country name
	City        *string  `gorm:"size:255"`            // City name
	Region      *string  `gorm:"size:255"`            // Region/State
	Latitude    *float64 `gorm:"type:decimal(10,8)"` // Coordinate precision
	Longitude   *float64 `gorm:"type:decimal(11,8)"` // Coordinate precision
}

type QueryResult struct {
	traffic int `gorm:"column:traffic"`
}

type ActiveUsers struct {
	NumberOfUsers int    `gorm:"column:number_of_users"`
	Page          string `gorm:"size:255"`
}


type LocationQueryResult struct{
	City string `gorm:"size:255"`
	Latitude float64 `gorm:"type:decimal(10,8)"`
	Longitude float64 `gorm:"type:decimal(11,8)"`
	UserCount int `gorm:"column:user_count"`
}

type BounceRateResponse struct {
	BounceRate float64 `json:"bounceRate"`
}

type AvgTimeResponse struct {

	AvgTimeSpent float64 `json:"avgTimeSpent"`

}



type CohortRow struct {

	CohortWeek time.Time `json:"cohort_week"`

	WeekNumber int       `json:"week_number"`

	UserCount  int       `json:"user_count"`

}



type CohortData struct {

	CohortDate    string    `json:"cohort_date"`

	TotalUsers    int       `json:"total_users"`

	RetentionData []float64 `json:"retention_data"`

}
