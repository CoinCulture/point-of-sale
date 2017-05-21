package main

import (
	"fmt"
	"strings"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

//--------------------------------------------------
// date/time

// tz := "America/New_York"
var _timezone string = "EST"
var _location, _ = time.LoadLocation(_timezone)

func CurrentTime() time.Time {
	now := time.Now().In(_location)
	return now //now.Add(-5 * time.Hour) //location not working...
}

func DateString() string {
	return strings.Split(fmt.Sprintf("%s", CurrentTime()), " ")[0] // zero grabs the sql formatted date
}

//--------------------------------------------------
// day, visits, sessions ...

type Day struct {
	Start time.Time
}

// memdb or sqldb
type Visits interface {
	All() []*Visit
}

func NewDay() *Day {
	return &Day{
		Start: CurrentTime(), // what if its not tomorrow yet ?!
	}
}

type EntryType struct {
	General   string
	Punch     string
	PayAtEnd  string
	ItemsOnly string
}

var EntryTypes = EntryType{
	General:   "general",
	Punch:     "punch_a_pass",
	PayAtEnd:  "pay_at_end",
	ItemsOnly: "items_only",
}

type ItemType struct {
	Food  string
	Drink string
	Misc  string
}

var ItemTypes = ItemType{
	Food:  "food",
	Drink: "drink",
	Misc:  "misc",
}

type Visit struct {
	Date       string
	BraceletID int
	EntryTime  time.Time
	ExitTime   *time.Time
	Total      int
	Active     int
	InvoiceID  int `gorm:"primary_key"`

	FinalBill BillOrMenu

	ExitTimeHack  string `gorm:"-"` // Ignore this field // TODO time.Time
	AdmissionType string `gorm:"-"` // Ignore this field // should be EntryTypes from above
	Kids          int    `gorm:"-"` // Ignore this field
}

type BillOrMenu struct {
	Foods  []*Item
	Drinks []*Item
	Miscs  []*Item

	LatestTransactions []*Transaction

	Total int
}

type Item struct {
	Name  string
	Price int

	// to set checkboxes for the menu
	IsActive string

	Active int

	// used for stats & inserting
	Type   string // TODO work in with ItemType struct
	Amount int
	Total  int

	Notes string
}

type Statistics struct {
	Visits          []*Visit
	TotalVisits     int
	MeanVisitLength int // TODO time.Time

	// should be []*Items?
	Foods     []*Transaction
	TotalFood int

	Drinks     []*Transaction
	TotalDrink int

	Miscs     []*Transaction
	TotalMisc int

	GrandTotal     int
	NumberOfVisits int

	Menu *BillOrMenu
}

type Transaction struct {
	ID          int `gorm:"primary_key"`
	InvoiceID   int
	BraceletID  int
	Name        string
	Amount      int
	Price       int
	Total       int
	Type        string
	Notes       string
	Paid        int
	TimeOrdered time.Time
}
