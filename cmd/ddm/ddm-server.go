package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alfonzso/dying-disk-manager/ddm"
	// "github.com/alfonzso/dying-disk-manager/pkg/communication"
	"github.com/alfonzso/dying-disk-manager/pkg/common"
	"github.com/alfonzso/dying-disk-manager/pkg/communication"
	"github.com/alfonzso/dying-disk-manager/pkg/input"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	// log "github.com/sirupsen/logrus"
)

func main() {

	common.NewLogrus()

	config := input.Manager()
	observer := observer.New()
	ddm := ddm.New(observer, config)

	ddm.BeforeMount()
	ddm.Mount()

	go ddm.Threading()
	communication.Socket(ddm)
	for {
		// sleeping
		for _, job := range ddm.Scheduler.Jobs() {
			nexTrun, _ := job.NextRun()
			fmt.Println(job.Name(), nexTrun, job.Tags())
		}
		// jobs := ddm.Scheduler.Jobs()
		time.Sleep(15 * time.Second)
	}

	os.Exit(0)
}
