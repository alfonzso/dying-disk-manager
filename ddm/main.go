package ddm

import (
	"slices"
	"time"

	"github.com/alfonzso/dying-disk-manager/pkg/common"
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	"github.com/go-co-op/gocron/v2"
	log "github.com/sirupsen/logrus"
)

type DDMData struct {
	Scheduler gocron.Scheduler
	*linux.Linux
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
	log.Debug("==> Threads are started")
	for {
		for _, disk := range ddmData.Disks {
			ddmData.setupTestThread(disk)
			ddmData.setupMountThread(disk)
			ddmData.setupRepairThread(disk)
		}
		time.Sleep(10 * time.Second)
	}
}

// func (ddmData *DDMData) RepairIsOn(){
func RepairIsOn(actionName string, diskStat *observer.DiskStat) {
	if !diskStat.Repair.IsRunning() {
		return
	}
	// if diskStat.Mount.IsRunning() {
	// 	diskStat.Mount.SetToStop()
	// }
	log.Debugf("[%s] %s -> Repair is ON", diskStat.Name, actionName)
}

func (ddmData *DDMData) CurrentActionsJobNotRunning(actionName, diskName, uuid string) bool {
	idx := slices.IndexFunc(ddmData.Scheduler.Jobs(), func(c gocron.Job) bool { return c.Name() == actionName })
	if idx == -1 {
		return true
	}
	currentJob := ddmData.Scheduler.Jobs()[idx]
	idx = slices.IndexFunc(currentJob.Tags(), func(tag string) bool { return tag == uuid })
	return idx == -1
}

func (ddmData *DDMData) CurrentActionsJobRunning(actionName, diskName, uuid string) bool {

	// common.Map(ddmData.Scheduler.Jobs(), func(c gocron.Job, idx int) gocron.Job {
	// 	if c.Name() == actionName {
	// 		return c
	// 	}
	// 	return nil
	// })

	jobs := common.Filter(ddmData.Scheduler.Jobs(), func(c gocron.Job) bool {
		return c.Name() == actionName && slices.Contains(c.Tags(), uuid)
	})
	return len(jobs) > 0
	// jobs := common.Filter(jobs, func(c gocron.Job) bool { return c.Name() == actionName })

	// idx := slices.IndexFunc(ddmData.Scheduler.Jobs(), func(c gocron.Job) bool { return c.Name() == actionName })
	// if idx == -1 {
	// 	return false
	// }
	// log.Debugf("[%s - %s] job found: %d ", diskName, uuid, idx)
	// currentJob := ddmData.Scheduler.Jobs()[idx]
	// idx = slices.IndexFunc(currentJob.Tags(), func(tag string) bool { return tag == uuid })
	// if idx != -1 {
	// 	log.Debugf("[%s - %s] job with tag found: %s ", diskName, uuid, currentJob.Tags()[idx])
	// }
	// return idx != -1
}

func IsInActiveOrDisabled(actionName string, diskStat *observer.DiskStat, action observer.Action) bool {
	if diskStat.IsInActive() || action.DisabledByAction {
		log.Warningf("[%s] %sThread => Disk deactivated => active: %t, disabledBy: %t",
			diskStat.Name, actionName, diskStat.Active, action.DisabledByAction,
		)
		action.SetToIddle()
		return true
	}
	return false
}

func WaitForThreadToBeIddle(msg string, as []*observer.Action) {
	for {
		iddleList := []bool{}
		for _, diskAs := range as {
			diskAs.DisabledByAction = true
			if diskAs.Status == observer.Iddle {
				iddleList = append(iddleList, true)
			}
		}
		if len(iddleList) == len(as) {
			return
		}
		log.Debugf("[%s] WaitForThreads", msg)
		time.Sleep(5 * time.Second)
	}
}

func StartThreads(as []*observer.Action) {
	for _, diskAs := range as {
		diskAs.DisabledByAction = false
		// diskAs.SetToStop()
	}
}

func (ddmData *DDMData) SetupCron(
	taskName string,
	function any,
	disk config.Disk,
	actions []*observer.Action,
	// diskStat *observer.DiskStat,
	cron string,
) (int, error) {
	_, err := ddmData.Scheduler.NewJob(
		gocron.CronJob(
			cron, false,
		),
		gocron.NewTask(
			function,
			disk,
			actions,
			// diskStat,
		),
		gocron.WithName(taskName),
		gocron.WithTags(disk.UUID),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)

	if err != nil {
		log.Errorf("[%s] Cron job failed =>\n%v\n", disk.Name, err)
		return 1, err
	}

	log.Debugf("[%s - %s] Cron expr: %s", taskName, disk.Name, cron)
	// ddmData.Scheduler.Start()
	return 0, nil
}

func New(o *observer.DDMObserver, c *config.DDMConfig) *DDMData {
	s, err := gocron.NewScheduler()
	linux := &linux.Linux{Exec: linux.NewExecCommand()}
	if err != nil {
		log.Error("Cron scheduler failed to setup")
	}
	s.Start()
	return &DDMData{s, linux, o, c}
}
