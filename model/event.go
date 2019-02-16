package model

// Event represents an entire event, ie "Masters Men Age 30-39"
type Event struct {
	ID       int
	Start    string
	Name     string
	Distance int
	Bank     string
	Entries  []Entry `yaml:"entries,omitempty"`
	Races    []Race  `yaml:"races,omitempty"`
}
