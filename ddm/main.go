package ddm

import (
	"time"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	"github.com/go-co-op/gocron/v2"
	log "github.com/sirupsen/logrus"
)

type DDMData struct {
	Scheduler gocron.Scheduler
	*observer.DDMObserver
	*config.DDMConfig
}

func GetCronExpr(diskCron string, commonCron string) string {
	cron := diskCron
	if len(cron) == 0 {
		cron = commonCron
	}
	return cron
}

func (ddmData *DDMData) Threading() {
	log.Debug("Thread => Test is started")
	for {
		ddmData.setupTestThread()
		ddmData.setupMountThread()
		time.Sleep(30 * time.Second)
	}
}

func (ddmData *DDMData) SetupCron(
	taskName string,
	function any,
	disk config.Disk,
	diskStat *observer.DiskStat,
	cron string,
) (int, error) {
	_, err := ddmData.Scheduler.NewJob(
		gocron.CronJob(
			cron, false,
		),
		gocron.NewTask(
			function,
			disk,
			diskStat,
		),
		gocron.WithName(taskName),
		gocron.WithTags(disk.UUID),
	)

	if err != nil {
		log.Errorf("[%s] Cron job failed =>\n%v\n", disk.Name, err)
		return 1, err
	}

	log.Debugf("[%s] Cron staring with expr... %s", disk.Name, cron)
	ddmData.Scheduler.Start()
	return 0, nil
}

func New(o *observer.DDMObserver, c *config.DDMConfig) *DDMData {
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Error("Cron scheduler failed to setup")
	}
	return &DDMData{s, o, c}
}
