package job

import (
	"os"
	"time"

	"github.com/daodao97/xgo/xapp"
	"github.com/daodao97/xgo/xcron"
	"github.com/daodao97/xgo/xredis"
)

func NewCronServer() xapp.NewServer {
	return func() xapp.Server {
		if os.Getenv("DISABLE_CRON") == "true" {
			return xcron.New2()
		}
		return xcron.New2(
			xcron.WithJobs(
				xcron.Job{
					Name:           "cron_exmaple",
					Spec:           "* * * * * *",
					Func:           ExmapleCron,
					Immediate:      false,
					EnableDistLock: true,
					LockTimeout:    5 * time.Minute,
					LockRetryDelay: 500 * time.Millisecond,
				},
			),
			xcron.WithRdb(xredis.Get()),
		)
	}
}

func ExmapleCron() {

}
