package structs

// TrafficByDay represents the traffic volume for a specific day of the week.
type TrafficByDay struct {
	Day   string  `json:"day"`
	Count float64 `json:"count"`
}

// TrafficByHour represents the average traffic volume for a specific hour of the day.
type TrafficByHour struct {
	Hour  int     `json:"hour"`
	Count float64 `json:"count"`
}
