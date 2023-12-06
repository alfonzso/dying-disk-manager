package ddm

import (
	"slices"
	"time"

	"github.com/alfonzso/dying-disk-manager/pkg/common"
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	"github.com/go-co-op/gocron/v2"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

// type CronAction struct {
// 	Mount string
// 	Test  string
// }
// type DDMScheduler struct {
// 	gocron.Scheduler
// 	CronAction
// }
type DDMData struct {
	Scheduler gocron.Scheduler
	// *DDMScheduler
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

func (ddmData *DDMData) GetDiskStat(disk config.Disk) *observer.DiskStat {
	idx := slices.IndexFunc(ddmData.DiskStat, func(c *observer.DiskStat) bool { return c.Name == disk.Name })
	if idx == -1 {
		log.Debug("init getDiskStat ", disk.Name)
		mount := GetCronExpr(disk.Mount.PeriodicCheck.Cron, ddmData.Common.Mount.PeriodicCheck.Cron)
		test := GetCronExpr(disk.Test.Cron, ddmData.Common.Test.Cron)
		diskStat := observer.DiskStat{
			Name:   disk.Name,
			UUID:   disk.UUID,
			Active: true,
			Repair: observer.Action{Name: "REPAIR", Status: observer.Stopped},
			Mount:  observer.Action{Name: "MOUNT", Status: observer.Stopped, Cron: mount},
			Test:   observer.Action{Name: "TEST", Status: observer.Stopped, Cron: test},
		}
		diskStat.Mount.ActionsToStop = []*observer.Action{&diskStat.Test}
		diskStat.Repair.ActionsToStop = []*observer.Action{&diskStat.Mount, &diskStat.Test}
		ddmData.DiskStat = append(ddmData.DiskStat, &diskStat)
		return &diskStat
	}
	return ddmData.DiskStat[idx]
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

func RepairIsOn(actionName string, diskStat *observer.DiskStat) bool {
	if !diskStat.Repair.IsRunning() {
		return false
	}
	log.Debugf("[%s] %s -> Repair is ON", diskStat.Name, actionName)
	return true
}

func (ddmData *DDMData) FindAJobByNameAndUUID(actionName, uuid string) []gocron.Job {
	jobs := common.Filter(ddmData.Scheduler.Jobs(), func(c gocron.Job) bool {
		return c.Name() == actionName && slices.Contains(c.Tags(), uuid)
	})
	return jobs
}

func (ddmData *DDMData) ActionsJobRunning(actionName, uuid, cronExpr string) bool {
	jobs := ddmData.FindAJobByNameAndUUID(actionName, uuid)

	cronSchedule, _ := cron.ParseStandard(cronExpr)
	for _, v := range jobs {
		next, err := v.NextRun()
		if err != nil {
			continue
		}

		nextEpoh := next.Unix()
		trueNextEpoh := cronSchedule.Next(time.Now()).Unix() - int64((5 * time.Minute).Seconds())

		if nextEpoh < trueNextEpoh {
			log.Warnf("jobEpoch: %d, parsedEpoch: %d", nextEpoh, trueNextEpoh)
			ddmData.Scheduler.RemoveByTags(uuid)
		}
	}
	return len(jobs) > 0
}

func (ddmData *DDMData) GetJobNextRun(actionName, uuid string) time.Duration {
	jobs := ddmData.FindAJobByNameAndUUID(actionName, uuid)
	nextRuns := common.Map(jobs, func(job gocron.Job, idx int) time.Time {
		res, _ := job.NextRun()
		return res
	})
	if len(nextRuns) == 0 {
		return (5 * time.Second)
	}
	return time.Until(nextRuns[0]) - (5 * time.Second)
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
	}
}

func (ddmData *DDMData) SetupCron(
	taskName string,
	function any,
	disk config.Disk,
	cron string,
) (int, error) {
	_, err := ddmData.Scheduler.NewJob(
		gocron.CronJob(
			cron, false,
		),
		gocron.NewTask(
			function,
			disk,
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
	// // GetCronExpr(disk.Mount.PeriodicCheck.Cron, ddmData.Common.Mount.PeriodicCheck.Cron)
	// mount := c.Common.Mount.PeriodicCheck.Cron
	// test := c.Common.Test.Cron
	// return &DDMData{&DDMScheduler{Scheduler: s,CronAction: struct{  mount; Test: test}}, linux, o, c}
	// // return &DDMData{&DDMScheduler{Scheduler: s, CronAction: CronAction{mount, test}}, linux, o, c}
}
