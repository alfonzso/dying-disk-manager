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

	fmt.Println("Test string to test everything")
	fmt.Println("Test string to test everything 1")

	common.NewLogrus()

	config := input.Manager()
	observer := observer.New()
	ddm := ddm.New(observer, config)

	ddm.BeforeMount()
	ddm.Mount()

	go ddm.Threading()
	communication.Socket(ddm)
	for {
		time.Sleep(15 * time.Second)
	}

	os.Exit(0)
}
