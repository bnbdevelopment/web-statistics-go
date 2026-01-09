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
}

type QueryResult struct {
	traffic int `gorm:"column:traffic"`
}

type ActiveUsers struct {
	NumberOfUsers int    `gorm:"column:number_of_users"`
	Page          string `gorm:"size:255"`
}
