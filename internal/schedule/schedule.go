package schedule

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"time"
)

func StartScheduleTask(repos *repositories.Repos) {
	ticker := time.NewTicker(time.Duration(2) * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				scheduledTask(repos)
			}
		}
	}()
}

func scheduledTask(repos *repositories.Repos) {
	fmt.Printf("%v SCHEDULED TASK\n", time.Now())
}