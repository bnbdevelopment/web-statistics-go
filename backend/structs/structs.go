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
	CountryCode *string  `gorm:"size:2"`             // ISO 3166-1 alpha-2 (e.g., "US", "GB")
	CountryName *string  `gorm:"size:255"`           // Full country name
	City        *string  `gorm:"size:255"`           // City name
	Region      *string  `gorm:"size:255"`           // Region/State
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

type LocationQueryResult struct {
	City      string  `json:"City" gorm:"size:255"`
	Latitude  float64 `json:"Latitude" gorm:"type:decimal(10,8)"`
	Longitude float64 `json:"Longitude" gorm:"type:decimal(11,8)"`
	UserCount int     `json:"UserCount" gorm:"column:user_count"`
}

type LandingPageStat struct {
	Page       string  `json:"page"`
	Sessions   int     `json:"sessions"`
	Bounces    int     `json:"bounces"`
	BounceRate float64 `json:"bounceRate"`
}

type ExitPageStat struct {
	Page     string  `json:"page"`
	Exits    int     `json:"exits"`
	ExitRate float64 `json:"exitRate"`
}

type BounceRateResponse struct {
	BounceRate float64 `json:"bounceRate"`
}

type AvgTimeResponse struct {
	AvgTimeSpent float64 `json:"avgTimeSpent"`
}

type CohortRow struct {
	CohortWeek time.Time `json:"cohort_week"`

	WeekNumber int `json:"week_number"`

	UserCount int `json:"user_count"`
}

type CohortData struct {
	CohortDate string `json:"cohort_date"`

	TotalUsers int `json:"total_users"`

	RetentionData []float64 `json:"retention_data"`
}

type FlowResult struct {
	SourcePage string `json:"source_page"`

	TargetPage string `json:"target_page"`

	FlowCount int `json:"flow_count"`
}

type SankeyNode struct {
	Name string `json:"name"`
}

type SankeyLink struct {
	Source int `json:"source"`

	Target int `json:"target"`

	Value int `json:"value"`
}

type SankeyData struct {
	Nodes []SankeyNode `json:"nodes"`

	Links []SankeyLink `json:"links"`
}

type FunnelStepStat struct {
	Step       string  `json:"step"`
	Users      int     `json:"users"`
	Conversion float64 `json:"conversion"`
	Dropoff    float64 `json:"dropoff"`
}

type FunnelResponse struct {
	Steps         []FunnelStepStat `json:"steps"`
	FirstStepSize int              `json:"firstStepSize"`
	Overall       float64          `json:"overallConversion"`
}

type EngagementSegment struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
}

type EngagementLead struct {
	SessionID string  `json:"sessionId"`
	Score     float64 `json:"score"`
	Duration  float64 `json:"duration"`
	Pages     int     `json:"pages"`
}

type EngagementResponse struct {
	Segments    []EngagementSegment `json:"segments"`
	TopSessions []EngagementLead    `json:"topSessions"`
	Average     float64             `json:"averageScore"`
}

type GeoTemporalBucket struct {
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Hour      int     `json:"hour"`
	Count     int     `json:"count"`
}

type ArchetypeAlert struct {
	Level     string  `json:"level"`
	Message   string  `json:"message"`
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
}
