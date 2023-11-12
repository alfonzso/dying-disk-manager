package main

import (
	"os"

	"github.com/alfonzso/dying-disk-manager/ddm"
	"github.com/alfonzso/dying-disk-manager/pkg/flags"
	"github.com/alfonzso/dying-disk-manager/pkg/input"
)

func main() {

	filename := flags.Parser()
	composeFileContentAsString := input.Manager(filename)

	os.Exit(ddm.Run(composeFileContentAsString))
}
