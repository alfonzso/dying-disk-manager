package main

import (
	"os"
	"time"

	"github.com/alfonzso/dying-disk-manager/ddm"
	"github.com/alfonzso/dying-disk-manager/pkg/common"
	"github.com/alfonzso/dying-disk-manager/pkg/communication"
	"github.com/alfonzso/dying-disk-manager/pkg/flags"
	"github.com/alfonzso/dying-disk-manager/pkg/input"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"

	log "github.com/sirupsen/logrus"
)

func main() {
	filename, flag := flags.Parser()

	common.NewLogrus()

	config := input.Manager(filename, flag)
	observer := observer.New()
	ddm := ddm.New(observer, config)

	for _, command := range []string{"mount", "umount", "lsblk", "fsck"} {
		if ddm.Exec.CheckCommandAvailability(command).IsFailed() {
			log.Errorf("Command not installed: %s", command)
			os.Exit(1)
		}
	}

	ddm.BeforeMount()
	ddm.Mount()

	go ddm.Threading()
	communication.Socket(ddm)
	for {
		time.Sleep(15 * time.Second)
	}

	os.Exit(0)
}
