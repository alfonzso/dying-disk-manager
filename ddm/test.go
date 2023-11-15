package ddm

import (
	"time"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"

	"github.com/go-co-op/gocron/v2"

	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) ThreadTest() {
	log.Debug("Thread => Test is started")
	for {
		ddmData.setupTestThread()
	}
}

func (ddmData *DDMData) isTestCanBeRun(disk config.Disk, diskStat *observer.DiskStat) bool {
	diskTest := disk.Test
	commonTest := ddmData.Common.Test
	return (diskTest.Enabled || commonTest.Enabled) && !diskStat.TestThreadIsRunning
}

func (ddmData *DDMData) setupTestThread() {
	for _, disk := range ddmData.Disks {
		diskStat := ddmData.GetDiskStat(disk)
		if ddmData.isTestCanBeRun(disk, diskStat) {
			// log.Debugf("[%s] Setup cron starting .. %t ..\n", diskStat.Name, diskStat.TestThreadIsRunning)
			// diskStat.TestThreadIsRunning = true
			go ddmData.setupCron(disk, diskStat, GetCronExpr(disk.Test.Cron, ddmData.Common.Test.Cron))
		}
	}
	time.Sleep(5 * time.Second)
}

func (ddmData *DDMData) setupCron(disk config.Disk, diskStat *observer.DiskStat, cron string) (int, error) {
	_, err := ddmData.Scheduler.NewJob(
		gocron.CronJob(
			cron, false,
		),
		gocron.NewTask(
			Test,
			disk,
			diskStat,
		),
	)

	if err != nil {
		log.Errorf("[%s] Cron job failed =>\n%v\n", disk.Name, err)
		return 1, err
	}

	diskStat.TestThreadIsRunning = true
	log.Debugf("[%s] Cron staring with expr... %s", time.Now().Format("2006-01-02 15:04:05"), cron)
	ddmData.Scheduler.Start()
	return 0, nil
}

func Test(disk config.Disk, diskStat *observer.DiskStat) (int, error) {
	// log.Debug("Test ...", time.Now().Format("2006-01-02 15:04:05"))
	if diskStat.Active {
		log.Debug("Testing disk", diskStat.Name)
	}
	return 0, nil
}
