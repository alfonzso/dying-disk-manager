package input

import (
	"flag"
	"fmt"
	"os"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	log "github.com/sirupsen/logrus"
)

func Manager(filename string, flag *flag.FlagSet) *config.DDMConfig {

	f, err := os.Open("config.yaml")
	if err == nil {
		log.Debug("Found default config file: config.yaml")
		defer f.Close()
		if read, err := config.ReadConf("config.yaml"); err == nil {
			return read
		}
	}

	if filename != "" {
		log.Debug("filename: ", filename)
		f, err := os.Open(filename)
		if err != nil {
			fmt.Println("[ ERROR ] cannot open file: err:", err)
			os.Exit(1)
		}
		defer f.Close()
		if read, err := config.ReadConf(filename); err == nil {
			return read
		}
	}

	fmt.Println("[ ERROR ] No config file given")
	flag.Usage()
	os.Exit(1)
	return nil

}
