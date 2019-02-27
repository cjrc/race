// Copyright © 2019 CJRC, Inc <greg@jrc.us>
//

package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/cjrc/race/model"
	"github.com/jinzhu/gorm"
	"github.com/spf13/cobra"
)

// scheduleCmd represents the schedule command
var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Create races based on events and entries",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("schedule called")
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
}

// TODO: This is still the old scheduling system
// Need to update to use the new database system, not GORM

// ScheduleEvents schedules the specified events to run together at the same time, same bank
func ScheduleEvents(db *gorm.DB, events []model.Event, bank string, start time.Time) {
	var entries []model.Entry
	var name string

	if len(events) > 1 {
		name = "Events " + strconv.FormatUint(uint64(events[0].ID), 10)
		for _, e := range events[1:] {
			name += ", " + strconv.FormatUint(uint64(e.ID), 10)
		}
	} else {
		name = strconv.FormatUint(uint64(events[0].ID), 10) + " " + events[0].Name
	}

	for _, e := range events {
		entries = append(entries, e.Entries...)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Seed < entries[j].Seed
	})

	nraces := len(entries)/(C.NLanes-2) + 1
	races := make([]model.Race, nraces)

	fmt.Printf("Creating %d races for event %d. There are %d entries.\n",
		nraces, events[0].ID, len(entries))

	for i := range races {
		races[i].BoatType = 0
		races[i].Distance = events[0].Distance
		races[i].EnableStrokeData = false
		races[i].SplitDistance = 500
		races[i].SplitTime = 120
		races[i].Name = name
		races[i].StartTime = start.Add(time.Minute * (C.RaceDuration * time.Duration(i)))
		races[i].Bank = bank

		db.Create(&races[i])

		for j := 0; j < C.NLanes; j++ {
			if len(entries) > 0 {
				db.Model(&entries[0]).Update("lane", C.SeedOrder[j])
				db.Model(&entries[0]).Update("race_id", races[i].ID)
				entries = entries[1:]
			}
		}
	}
}

// CreateSchedule returns a new Schedule populated with Events and Entries
func CreateSchedule(db *gorm.DB) {
	var events []model.Event
	db.Preload("Entries").Find(&events)
	events = append([]model.Event{model.Event{}}, events...)

	// Masters, Veteran, and Open Men
	var e []model.Event
	e = append(e, events[1], events[3], events[5], events[7])
	ScheduleEvents(db, e, e[0].Bank, e[0].startTime)

	// Masters, Veteran, and Open Women
	e = nil
	e = append(e, events[2], events[4], events[6], events[8], events[11], events[13])
	ScheduleEvents(db, e, e[0].Bank, e[0].startTime)

	// Col Varsity and Nov Men
	e = nil
	e = append(e, events[10], events[12])
	ScheduleEvents(db, e, e[0].Bank, e[0].startTime)

	// Col Coxswains
	e = nil
	e = append(e, events[14], events[15])
	ScheduleEvents(db, e, e[0].Bank, e[0].startTime)

	// The rest of the events
	for i := 16; i <= 22; i++ {
		e = nil
		e = append(e, events[i])
		ScheduleEvents(db, e, e[0].Bank, e[0].startTime)
	}
}

// FindLane Returns the first lane available for an entry in RACE or an error
// if none are available
func FindLane(race Race) (int, error) {

	used := make([]bool, C.NLanes)

	for _, e := range race.Entries {
		if e.Lane == 0 {
			continue
		}
		used[e.Lane-1] = true
	}

	for i := 0; i < C.NLanes; i++ {
		if used[C.SeedOrder[i]-1] == false {
			return C.SeedOrder[i], nil
		}
	}

	return -1, fmt.Errorf("There are no free lanes in race %d", race.ID)

}

//AssignLanes will assign a lane to any entry that doesn't have one
func AssignLanes(db *gorm.DB) {
	var entries []Entry
	db.Where("lane = 0 AND NOT race_id =0").Find(&entries)

	for _, e := range entries {
		var race Race
		db.Preload("Entries").Find(&race, e.RaceID)
		if e.RaceID == 0 {
			continue
		}
		lane, err := FindLane(race)
		if err != nil {
			fmt.Printf("ERROR: %s", err)
			continue
		}

		if verifyPrintf("Assign %s (%d) to race %d, lane %d?",
			e.BoatName, e.ID, race.ID, lane) {
			e.RaceID = int(race.ID)
			e.Lane = lane
			db.Save(&e)
			return
		}
	}
}

// AssignEntryToRace will try and fit an entry into an empty race
func AssignEntryToRace(entry Entry, db *gorm.DB) {
	var races []Race
	db.Preload("Entries").Find(&races)

	for _, race := range races {
		for _, e := range race.Entries {
			if e.EventID == entry.EventID {
				lane, err := FindLane(race)
				if err == nil {
					if verifyPrintf("Assign %s (%d) to race %d, lane %d?",
						entry.BoatName, entry.ID, race.ID, lane) {
						entry.RaceID = int(race.ID)
						entry.Lane = lane
						db.Save(&entry)
						return
					}
				}
				break
			}
		}
	}
}

// AssignRaces will put entries that are without a race into a race, if possible
func AssignRaces(db *gorm.DB) {
	var entries []Entry
	db.Where("race_id = 0 and scratched IS NOT true").Find(&entries)

	for _, entry := range entries {
		AssignEntryToRace(entry, db)
	}
}
