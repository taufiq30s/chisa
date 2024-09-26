package cronjob

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/taufiq30s/chisa/utils"
)

func CreateJobs() {
	schedule, err := gocron.NewScheduler()
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to create scheduler: %s", err)
	}
	defer schedule.Start()

	// Update Scam Dataset
	_, err = schedule.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(0, 0, 0))),
		gocron.NewTask(updateScamDataset),
	)
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to create Job: %v", err)
	}
}
