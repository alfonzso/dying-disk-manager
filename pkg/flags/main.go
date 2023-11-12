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

func Parser() (string, *flag.FlagSet) {
	var filename string
	var verbose bool
	fff := flag.NewFlagSet("ddm", 1)

	fff.StringVar(&filename, "f", "", "<filename>")
	fff.BoolVar(&verbose, "v", false, "Verbose")

	fff.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n%s\n", AppDescription, AppUsage)
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	fff.Parse(os.Args[1:])

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	return filename, fff
}
