package cronjob

import (
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"time"
	"wordpress-go-next/backend/pkg/logger"
)

var scheduler gocron.Scheduler
var activeTasks map[string]string

func init() {
	s, err := gocron.NewScheduler()
	if err != nil {
		logger.Errorf("cronjob init error: %v", err)
	}

	scheduler = s
	activeTasks = make(map[string]string)

	go Start()
}

func AddJob(jobDefinition gocron.JobDefinition, taskFunc func(), jobOption gocron.JobOption) (string, error) {
	j, err := scheduler.NewJob(
		jobDefinition,
		gocron.NewTask(taskFunc),
		jobOption,
	)
	if err != nil {
		logger.Errorf("cronjob add job error: %v", err)
		return "", err
	}

	return j.ID().String(), nil
}

func Start() {
	scheduler.Start()
}

func Shutdown() error {
	return scheduler.Shutdown()
}

func MonitorDatabaseTaskChange() {
	_, err := AddJob(gocron.DurationJob(1*time.Minute), func() {

		fmt.Printf("current task: %v\n", activeTasks)
	}, gocron.WithName("system"))

	if err != nil {
		fmt.Printf("MonitorDatabaseTaskChange error: %v", err)
	}
}
