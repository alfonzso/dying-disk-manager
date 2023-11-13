package main

import (
	"os"

	"github.com/alfonzso/dying-disk-manager/ddm"
	"github.com/alfonzso/dying-disk-manager/pkg/flags"
	"github.com/alfonzso/dying-disk-manager/pkg/input"
)

func main() {

	filename, flag := flags.Parser()
	cfg := input.Manager(filename, flag)

	os.Exit(ddm.Run(cfg))
}
