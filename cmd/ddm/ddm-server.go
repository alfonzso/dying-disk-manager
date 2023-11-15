package main

import (
	"os"

	"github.com/alfonzso/dying-disk-manager/ddm"
	"github.com/alfonzso/dying-disk-manager/pkg/communication"
	"github.com/alfonzso/dying-disk-manager/pkg/input"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
)

func main() {

	config := input.Manager()
	observer := observer.New()
	ddm := ddm.New(observer, config)

	communication.Socket(ddm)

	os.Exit(0)
}
