package observer

import (
	"fmt"
	"slices"

	"github.com/adhocore/gronx"
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	log "github.com/sirupsen/logrus"
)

func (d *DDMObserver) GetDiskStat(disk config.Disk) *DiskStat {
	idx := slices.IndexFunc(d.DiskStat, func(c DiskStat) bool { return c.Name == disk.Name })
	if idx == -1 {
		log.Debug("init getDiskStat ", disk.Name)
		diskStat := DiskStat{
			Name:   disk.Name,
			UUID:   disk.UUID,
			Active: true,
			Repair: Action{Status: Stopped},
			Mount:  Action{Status: Stopped},
			Test:   Action{Status: Stopped},
		}
		d.DiskStat = append(d.DiskStat, diskStat)
		return &diskStat
	}
	return &d.DiskStat[idx]
}

func (d *DDMObserver) IsCronDue(expr string) bool {
	gron := gronx.New()
	if !gron.IsValid(expr) {
		log.Warningf("Cron expression invalid: %s", expr)
		return true
	}
	due, err := gron.IsDue(expr)
	if err != nil {
		log.Warningf("Cron due failed: %s", expr)
		return true
	}
	return due
}

func New() *DDMObserver {
	d := &DDMObserver{DiskStat: []DiskStat{}}
	fmt.Printf("%#v\n", d)
	return d
}
