package config

type DDMConfig struct {
	Common Common  `yaml:"common"`
	Disks  []Disks `yaml:"disks"`
}

type Common struct {
	Mount  Mount  `yaml:"mount"`
	Test   Test   `yaml:"test"`
	Repair Repair `yaml:"repair"`
}

type Disks struct {
	Name   string         `yaml:"name"`
	UUID   string         `yaml:"uuid"`
	Mount  ExtendedMount  `yaml:"mount"`
	Test   Test           `yaml:"test"`
	Repair ExtendedRepair `yaml:"repair"`
}

type Mount struct {
	Enabled       bool          `yaml:"enabled"`
	PeriodicCheck PeriodicCheck `yaml:"periodicCheck"`
}

type ExtendedMount struct {
	Mount `yaml:",inline"`
	Path  string `yaml:"path"`
}

type Repair struct {
	Enabled       bool   `yaml:"enabled"`
	Command       string `yaml:"command"`
	CommandBefore string `yaml:"commandBefore"`
	CommandAfter  string `yaml:"commandAfter"`
}

type ExtendedRepair struct {
	DiskNumber int `yaml:"diskNumber"`
	Repair     `yaml:",inline"`
}

type Test struct {
	Enabled bool   `yaml:"enabled"`
	Cron    string `yaml:"cron"`
}

type PeriodicCheck struct {
	Enabled bool   `yaml:"enabled"`
	Cron    string `yaml:"cron"`
}
