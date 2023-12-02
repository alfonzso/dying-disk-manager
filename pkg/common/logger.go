package common

import (
	"github.com/sirupsen/logrus"
)

func NewLogrus() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	logrus.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
}
