package bot

import (
	"fmt"

	"github.com/go-co-op/gocron/v2"
	"github.com/taufiq30s/chisa/utils"
)

func (chisa *Bot) CreateJobs() {
	schedule, err := gocron.NewScheduler()
	if err != nil {
		fmt.Printf("Failed to create scheduler: %s", err)
		utils.ErrorLog.Fatalf("Failed to create scheduler: %s", err)
	}
	defer schedule.Start()
	defer fmt.Println("Cron Job created")
	defer utils.InfoLog.Println("Cron Job created")

	// Update Scam Dataset
	_, err = schedule.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(0, 0, 0))),
		gocron.NewTask(chisa.updateScamDataset),
	)
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to create Job: %v", err)
	}
}
