package ddm

import (
	"time"

	"github.com/alfonzso/dying-disk-manager/pkg/observer"

	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) setupTestThread(diskData DiskData) {
	if RepairIsOn(diskData.Test.Name, diskData) {
		return
	}

	if !(diskData.conf.Test.Enabled || ddmData.Common.Test.Enabled) {
		return
	}

	go ddmData.startTestAndWaitTillAlive(diskData)
}

func (ddmData *DDMData) startTestAndWaitTillAlive(diskData DiskData) {
	if ddmData.ActionsJobRunning(diskData.Test.Name, diskData.UUID, diskData.Test.Cron) {
		return
	}

	ddmData.SetupCron(
		diskData.Test.Name,
		ddmData.diskTest,
		diskData,
		diskData.Test.Cron,
	)

	diskData.Test.SetToRun()
}

func (ddmData *DDMData) diskTest(diskData DiskData) (int, error) {
	res, err := ddmData.testWrapper(diskData)
	go func() {
		times := ddmData.GetJobNextRun(diskData.Test.Name, diskData.UUID)
		time.Sleep(times)
		diskData.Test.HealthCheck = observer.None
	}()
	return res, err
}

func (ddmData *DDMData) testWrapper(diskData DiskData) (int, error) {
	diskData.Test.SetToRun()
	diskData.Test.HealthCheck = observer.OK

	if IsInActiveOrDisabled(diskData.Test.Name, diskData, diskData.Test) {
		diskData.Test.SetToIddle()
		return 0, nil
	}

	if ddmData.Exec.RunDryFsck(diskData.UUID).IsFailed() {
		log.Debugf("[%s] Fsck failed, triggering repair", diskData.Name)
		diskData.Active = false
		diskData.Repair.SetToRun()
	}

	if ddmData.Exec.CheckMountStatus(diskData.UUID, diskData.conf.Mount.Path).IsMounted() {
		if ddmData.Exec.WriteIntoDisk(diskData.conf.Mount.Path).IsFailed() {
			log.Debugf("[%s] Write to disk failed, triggering repair", diskData.Name)
			diskData.Active = false
			diskData.Repair.SetToRun()
		}
	}

	diskData.Test.SetToIddle()
	return 0, nil
}
