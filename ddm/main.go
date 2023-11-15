package ddm

import (
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
)

type DDMData struct {
	*observer.DDMObserver
	*config.DDMConfig
}

func New(o *observer.DDMObserver, c *config.DDMConfig) *DDMData {
	return &DDMData{o, c}
}
