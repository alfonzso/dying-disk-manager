package observer

import (
	"slices"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	log "github.com/sirupsen/logrus"
)

func (d *DDMObserver) GetDiskStat(disk config.Disk) *DiskStat {
	idx := slices.IndexFunc(d.DiskStat, func(c *DiskStat) bool { return c.Name == disk.Name })
	if idx == -1 {
		log.Debug("init getDiskStat ", disk.Name)
		diskStat := DiskStat{
			Name:   disk.Name,
			UUID:   disk.UUID,
			Active: true,
			Repair: Action{Name: "REPAIR", Status: Stopped},
			Mount:  Action{Name: "MOUNT", Status: Stopped},
			Test:   Action{Name: "TEST", Status: Stopped},
		}
		diskStat.Mount.ActionsToStop = []*Action{&diskStat.Test}
		diskStat.Repair.ActionsToStop = []*Action{&diskStat.Mount, &diskStat.Test}
		d.DiskStat = append(d.DiskStat, &diskStat)
		return &diskStat
	}
	return d.DiskStat[idx]
}

func New() *DDMObserver {
	return &DDMObserver{DiskStat: []*DiskStat{}}
}
