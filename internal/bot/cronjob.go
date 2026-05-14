package bot

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/taufiq30s/chisa/utils"
)

func (chisa *Bot) CreateJobs() {
	schedule, err := gocron.NewScheduler()
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to create scheduler: %s", err)
	}

	// Update Scam Dataset
	_, err = schedule.NewJob(
		gocron.DailyJob(CronDailyJobCount, gocron.NewAtTimes(gocron.NewAtTime(CronHourMidnight, CronMinuteZero, CronSecondZero))),
		gocron.NewTask(chisa.updateScamDataset),
	)
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to create Job for Update Scam Dataset: %v\n", err)
	}

	chisa.scheduler = schedule
	schedule.Start()
	utils.InfoLog.Println("Cron Job created")
}
