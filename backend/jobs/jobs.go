package jobs

import (
	"fmt"
	"statistics/database"
	"statistics/statistics"
	"statistics/structs"

	"github.com/robfig/cron"
)

func activeUsersSnapShot() {

	pages := []string{"irodalomerettsegi.hu"}

	for _, page := range pages {
		count := statistics.ActiveUsers(page)
		record := structs.ActiveUsers{
			Page:          page,
			NumberOfUsers: int(count),
		}
		err := database.Session.Create(&record).Error
		if err != nil {
			fmt.Println("Error inserting active users data:", err)
		} else {
			fmt.Println("Active users data inserted successfully for page:", page)
		}
	}
}

func InitCronScheduler() *cron.Cron {
	c := cron.New()

	c.AddFunc("@every 00h01m00s", activeUsersSnapShot)

	c.Start()
	fmt.Println("Cron scheduler initialized")
	return c
}
