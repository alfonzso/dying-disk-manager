package flags

import (
	"flag"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

var AppDescription = "Dying disk manager: Monitor your mounts and repair your disks\n"
var AppName = strings.Replace(os.Args[0], "./", "", -1)
var AppUsage = fmt.Sprintf(
	`Usage:
    %s [ OPTIONS ]
`, AppName,
)

func Parser() (filename string) {

	var verbose bool
	// flag.StringVar(&filename, "f", "", "<filename>")
	flag.BoolVar(&verbose, "v", false, "Verbose")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n%s\n", AppDescription, AppUsage)
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	return filename
}
