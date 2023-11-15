package ddm

import (
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

func New(o *observer.DDMObserver, c *config.DDMConfig) *DDMData {
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Error("Cron scheduler failed to setup")
	}
	return &DDMData{s, o, c}
}
