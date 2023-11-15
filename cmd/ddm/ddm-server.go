package main

import (
	"os"

	"github.com/alfonzso/dying-disk-manager/ddm"
	// "github.com/alfonzso/dying-disk-manager/pkg/communication"
	"github.com/alfonzso/dying-disk-manager/pkg/input"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	// log "github.com/sirupsen/logrus"
)

func main() {

	config := input.Manager()
	observer := observer.New()
	ddm := ddm.New(observer, config)

	ddm.Mount()

	go ddm.ThreadTest()
	// communication.Socket(ddm)
	for {
		// sleeping
	}

	os.Exit(0)
}
