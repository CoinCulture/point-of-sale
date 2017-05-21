package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB
var err error
var activeSession bool

// ----------- main pages, rendered from templates ------------

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// for the colour feature on Add Items page
func isLockerActive(w http.ResponseWriter, r *http.Request) {
	getBraceletID := strings.Split(r.URL.String(), "?")

	var bID int
	if len(getBraceletID) == 2 { // need better sanity check
		if getBraceletID[1] == "" {
			bID = -1
		} else {
			bID, err = strconv.Atoi(getBraceletID[1])
			if err != nil {
				panic(err)
			}
		}

	}
	activeSession, _ = doesSessionAlreadyExist(bID)
	w.Write([]byte(strconv.FormatBool(activeSession)))
}

// funcs in order as display on index.html
func newSession(w http.ResponseWriter, r *http.Request) {

	numberOfVisits := strings.Split(r.URL.String(), "?")

	var lastVisits []*Visit
	if len(numberOfVisits) == 2 { // need better sanity check
		ugh, err := strconv.Atoi(numberOfVisits[1])
		if err != nil {
			panic(err)
		}
		lastVisits = getLastOpenedVisitInfo(ugh)
	}

	lastSeshBytes, err := renderTemplate("lastSesh", lastSessionTemplate, lastVisits)
	if err != nil {
		panic(err)
	}

	menu := new(BillOrMenu)
	menu.Miscs = getActiveItems("misc")

	miscBytes, err := renderTemplate("miscItems", miscTemplate, menu)
	if err != nil {
		panic(err)
	}

	visits := getActiveVisits(true)
	for _, visit := range visits {
		visit.FinalBill.Total = getVisitTotal(visit, 0)
	}

	renderAndWrite(w, "newSesh", newSessionTemplate(string(lastSeshBytes), string(miscBytes)), visits)
}

func getLastOpenedVisitInfo(number int) []*Visit {
	var lastOpenedVisits []*Visit

	db.Order("invoice_id desc").Limit(number).Find(&lastOpenedVisits)
	for _, visit := range lastOpenedVisits {
		visit.Total = getVisitTotal(visit, 1) // won't show pay_at_end, that's fine
	}

	return lastOpenedVisits

}

func addItems(w http.ResponseWriter, r *http.Request) {

	menu := new(BillOrMenu)
	numberOfRows := strings.Split(r.URL.String(), "?")

	if len(numberOfRows) == 2 { // need better sanity check
		menu.LatestTransactions = getLastTransactions(numberOfRows[1])
	}

	// fill in the struct
	menu.Foods = getActiveItems("food")
	menu.Drinks = getActiveItems("drink")
	menu.Miscs = getActiveItems("misc")

	renderAndWrite(w, "tMenu", menuTemplate, menu)
}

func closeSession(w http.ResponseWriter, r *http.Request) {

	visits := getActiveVisits(true)
	for _, visit := range visits {
		visit.FinalBill.Total = getVisitTotal(visit, 0)
	}

	renderAndWrite(w, "invoice", displayInvoiceTemplate, visits)
}

func adminPage(w http.ResponseWriter, r *http.Request) {
	// has no templating to render, just routes via submit buttons
	w.Write([]byte(adminPageTemplate))
}

func selectTodaysMenuPage(w http.ResponseWriter, r *http.Request) {

	allFoods := getActiveItems("allfood")

	currentlyActiveFoods := getActiveItems("food")

	for _, food := range allFoods {
		for _, activeFood := range currentlyActiveFoods {
			if food.Name == activeFood.Name {
				food.IsActive = "checked" // set the checkbox
			}
		}
	}

	renderAndWrite(w, "selectMenu", selectTodaysMenuPageTemplate, allFoods)
}

func displayBill(w http.ResponseWriter, r *http.Request) {
	braceletID, err := parseBraceletID(r)
	if err != nil {
		writeError(w, err.Error(), err)
		return
	}

	visit, err := getVisitFromBraceletID(braceletID)
	if err != nil {
		writeError(w, ErrSessionDoesNotExist, nil)
		return
	}

	// assemble the bill
	visit.FinalBill.Total = getVisitTotal(visit, 0)

	renderAndWrite(w, "finalBill", finalBillTemplate, visit)
}

func statisticsPage(w http.ResponseWriter, r *http.Request) {

	stats := new(Statistics)

	stats.Foods = getTransactionsByTypeFinalBill("food")
	stats.Drinks = getTransactionsByTypeFinalBill("drink")
	stats.Miscs = getTransactionsByTypeFinalBill("misc")

	stats.TotalFood = calculateDailyTotals(stats.Foods)
	stats.TotalDrink = calculateDailyTotals(stats.Drinks)
	stats.TotalMisc = calculateDailyTotals(stats.Miscs)

	stats.GrandTotal = stats.TotalFood + stats.TotalDrink + stats.TotalMisc

	stats.Menu = getMenuSummaryFromDay(stats)

	var totalVisits int = 0
	for _, find := range stats.Menu.Miscs {
		if find.Name == "adult admission" {
			totalVisits += find.Amount
		}
		if find.Name == "child admission" {
			totalVisits += find.Amount
		}
		if find.Name == "punch a pass" {
			totalVisits += find.Amount
		}
	}

	stats.NumberOfVisits = totalVisits

	renderAndWrite(w, "stats", statisticsPageTemplate, stats)
}

func reopenOrDeleteSessionPage(w http.ResponseWriter, r *http.Request) {
	visits := getActiveVisits(false)
	for _, visit := range visits {
		visit.FinalBill.Total = getVisitTotal(visit, 1)
	}

	renderAndWrite(w, "reopen", reopenOrDeleteSessionPageTemplate, visits)
}

func renderAndWrite(w http.ResponseWriter, name string, template string, data interface{}) {
	templateBytes, err := renderTemplate(name, template, data)
	if err != nil {
		panic(err)
	}
	w.Write(templateBytes)
}

func insertNewItemsPage(w http.ResponseWriter, r *http.Request) {
	// do a simple insert first
	w.Write([]byte(insertNewItemsTemplate))
}

// ---------- actions to take from main pages ------------------

func parseBraceletID(r *http.Request) (int, error) {
	idString := r.FormValue("bracelet_id")
	if idString == "" {
		return -1, errors.New(ErrNoBraceletNumberEntered)
	}
	idInt, err := strconv.Atoi(idString)
	if err != nil {
		return -1, errors.New(ErrBraceletNumberInvalid)
	}

	// zero is a bracelet number for buying items (no entry)
	// the max # of lockers should be in a config file
	if idInt > 100 || idInt < 0 {
		return -1, errors.New(ErrBraceletNumberInvalid)
	}
	return idInt, nil
}

func initializeSession(w http.ResponseWriter, r *http.Request) {

	visit := new(Visit)

	braceletID, err := parseBraceletID(r)
	if err != nil {
		writeError(w, err.Error(), nil)
		return
	}

	visit.Date = DateString()
	visit.BraceletID = braceletID

	admissionType := r.FormValue("payment_method") // uses radio buttons
	if admissionType == "" {
		writeError(w, ErrNoAdmissionTypeSelected, nil)
		return
	}

	activeSession, _ = doesSessionAlreadyExist(visit.BraceletID) // throw out invoiceID, it shouldn't exist yet...
	if activeSession {
		writeError(w, ErrSessionAlreadyExists, nil)
		return
	}

	visit.EntryTime = CurrentTime()
	visit.Active = 1

	db.NewRecord(visit)
	db.Create(&visit)

	// initialize a visit, generates an invoiceID
	fmt.Println(visit)
	visit, err = getVisitFromBraceletID(braceletID)
	if err != nil {
		writeError(w, ErrSessionDoesNotExist, nil)
		return
	}

	var paid int // 0 is unpaid, 1 is paid
	var entryThing string
	var punch bool
	var adult bool

	switch admissionType {
	case EntryTypes.General:
		paid = 1
		entryThing = "entry"
		punch = false
		adult = true
	case EntryTypes.Punch:
		paid = 1
		entryThing = "entry"
		punch = true
		adult = false
	case EntryTypes.PayAtEnd:
		paid = 0
		entryThing = "not_entry"
		punch = false
		adult = true
	case EntryTypes.ItemsOnly:
		if visit.BraceletID != 0 { // this should throw if 0 is used for !ItemsOnly (ideally)
			writeError(w, ErrItemsOnlyWithLockerZero, nil)
			return
		}

		paid = 0                 // still need to close bill
		entryThing = "not_entry" // this matter WRT to re-opening sessions & needs to be fixed
		punch = false
		adult = false // don't insert a visit
		// flow through all the crap & insert any misc items
	}

	// easier to hard code than to deduplication for now
	if punch {
		transaction := Transaction{
			InvoiceID:   visit.InvoiceID,
			BraceletID:  visit.BraceletID,
			Name:        "punch a pass",
			Amount:      1,
			Price:       0,
			Total:       0,
			TimeOrdered: CurrentTime(),
			Notes:       entryThing,
			Type:        "misc",
			Paid:        paid,
		}
		db.NewRecord(transaction)
		db.Create(&transaction)
	}

	if adult {
		transaction := Transaction{
			InvoiceID:   visit.InvoiceID,
			BraceletID:  visit.BraceletID,
			Name:        "adult admission",
			Amount:      1,
			Price:       5,
			Total:       5,
			TimeOrdered: CurrentTime(),
			Notes:       entryThing,
			Type:        "misc",
			Paid:        paid,
		}
		db.NewRecord(transaction)
		db.Create(&transaction)
	}

	// add additional items if bought (including kids, a pass, or extra people)
	items := getActiveItems("misc")

	// TODO deduplicate with addItemsToSession
	for _, activeItems := range items {
		numberOrdered := r.FormValue(activeItems.Name)
		// these are required to prevent a string parsing error
		if numberOrdered != "" && numberOrdered != "0" {
			num0, err := strconv.Atoi(numberOrdered)
			if err != nil {
				writeError(w, err.Error(), err)
				return
			}

			if num0 < 0 { // check if at least that many items are already in the invoice for this user, else throw error
				txnsByName := getTransactionsByTypeNameID("misc", activeItems.Name, visit.BraceletID)
				amt := 0
				for _, txn := range txnsByName {
					amt = amt + txn.Amount
				}

				if -num0 > amt { // note the -num0
					writeError(w, "can't minus items that haven't been ordered", nil)
					return
				}
			}

			total := (num0 * activeItems.Price)
			transaction := Transaction{
				InvoiceID:   visit.InvoiceID,
				BraceletID:  visit.BraceletID,
				Name:        activeItems.Name,
				Amount:      num0,
				Price:       activeItems.Price,
				Total:       total,
				TimeOrdered: CurrentTime(),
				Notes:       entryThing,
				Type:        "misc",
				Paid:        paid,
			}
			db.NewRecord(transaction)
			db.Create(&transaction)
			if err != nil {
				writeError(w, ErrWithSQLquery, err)
				return
			}
		}
	}

	http.Redirect(w, r, "/newSession?1", 301) // ?1 to display the most recently opened session
}

func addItemsToASession(w http.ResponseWriter, r *http.Request) {
	braceletID, err := parseBraceletID(r)
	if err != nil {
		writeError(w, err.Error(), err)
		return
	}

	visit, err := getVisitFromBraceletID(braceletID)
	if err != nil {
		writeError(w, ErrSessionDoesNotExist, nil)
		return
	}

	// [zr] note: this logic may be ~ confusing but deduplicates a ton of code
	i := 0
	categories := []string{ItemTypes.Food, ItemTypes.Drink, ItemTypes.Misc}
	for _, category := range categories {
		items := getActiveItems(category)

		for _, activeItems := range items {
			numberOrdered := r.FormValue(activeItems.Name)
			// these are required to prevent a string parsing error
			if numberOrdered != "" && numberOrdered != "0" {
				// notes  hasn't been implemented
				notes := r.FormValue(fmt.Sprintf("%s%s", activeItems.Name, "Notes"))
				num0, err := strconv.Atoi(numberOrdered)
				if err != nil {
					writeError(w, err.Error(), err)
					return
				}

				if num0 < 0 { // check if at least that many items are already in the invoice for this user, else throw error
					txnsByName := getTransactionsByTypeNameID(category, activeItems.Name, visit.BraceletID)
					amt := 0
					for _, txn := range txnsByName {
						amt = amt + txn.Amount
					}

					if -num0 > amt { // note the -num0
						writeError(w, "can't minus items that haven't been ordered", nil)
						return
					}
				}

				total := (num0 * activeItems.Price)
				transaction := Transaction{
					InvoiceID:   visit.InvoiceID,
					BraceletID:  visit.BraceletID,
					Name:        activeItems.Name,
					Amount:      num0,
					Price:       activeItems.Price,
					Total:       total,
					TimeOrdered: CurrentTime(),
					Notes:       notes,
					Type:        category,
					Paid:        0,
				}
				db.NewRecord(transaction)
				db.Create(&transaction)
				if err != nil {
					writeError(w, ErrWithSQLquery, err)
					return
				}
				// print the food (but not voids)
				if category == ItemTypes.Food && num0 > 0 {
					// there's a better way to handle not printing certain foods
					// this should be set via a config file
					// same with the buzzer
					if (activeItems.Name != "noChitNeeded") && (activeItems.Name != "notForTheCook") {
						if err := printTheChit(braceletID, numberOrdered, activeItems.Name); err != nil {
							// ignore error
						}
						if activeItems.Name != "dontBuzzThisFood" {
							if err := activateBuzzer(); err != nil {
								// ignore error
							}
						}
					}
				}
				i++
			}
		}
	}

	addItemsWithTotalRows := fmt.Sprintf("/addItems?%s", strconv.Itoa(i)) // total # of orders just inserted, to be displayed
	http.Redirect(w, r, addItemsWithTotalRows, 301)
}

func closeBill(w http.ResponseWriter, r *http.Request) {
	visit := new(Visit)
	// hacky
	getFields := strings.Split(r.URL.String(), "?")[1]
	ints, err := stringsToInts(strings.Split(getFields, "&"))
	if err != nil {
		panic(err)
	}
	visit.BraceletID, visit.InvoiceID, visit.Total = ints[0], ints[1], ints[2] // TODO fix total. see below

	var t = CurrentTime()
	visit.ExitTime = &t

	// total is calculated from bill displayed & doesn't include items bought up front, which it should
	//_, err = db.Exec("UPDATE visits SET exit_time=?,total=?,active=0 WHERE bracelet_id=? AND invoice_id=?", visit.ExitTime, visit.Total, visit.BraceletID, visit.InvoiceID)
	if err := db.Exec("UPDATE visits SET exit_time = ?, active = ? WHERE bracelet_id = ? AND invoice_id = ?", visit.ExitTime, 0, visit.BraceletID, visit.InvoiceID).Error; err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	if err := db.Exec("UPDATE transactions SET paid = ? WHERE bracelet_id = ? AND invoice_id = ?", 1, visit.BraceletID, visit.InvoiceID).Error; err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	http.Redirect(w, r, "/closeSession", 301)
}

func closeDay(w http.ResponseWriter, r *http.Request) {
	panic("unable to close day - need to fix the queries")
	/*
		if len(getActiveVisits(true)) != 0 {
			writeError(w, "cannot close day with sessions still open", nil)
			return
		}

		// ugly hack to get date from first visit, otherwise getFormattedDate()
		// will use the following day, if a day is closed after midnight
		// TODO struct Day
		var date string
		// nested because deleting a session can remove an entry with its invoice id
		if err := db.QueryRow("SELECT date FROM visits WHERE invoice_id=5").Scan(&date); err != nil {
			if err := db.QueryRow("SELECT date FROM visits WHERE invoice_id=10").Scan(&date); err != nil {
				if err := db.QueryRow("SELECT date FROM visits WHERE invoice_id=15").Scan(&date); err != nil {
					panic(err)
				}
			}
		}

		sqlDateFormat := strings.Replace(date, "-", "_", 2)
		sqlDateFormat = strings.Split(sqlDateFormat, "T")[0] // awful

		createQueryVisit := fmt.Sprintf("CREATE TABLE visits_%s LIKE visits", sqlDateFormat)
		insertQueryVisit := fmt.Sprintf("INSERT visits_%s SELECT * FROM visits", sqlDateFormat)

		createQueryTransactions := fmt.Sprintf("CREATE TABLE transactions_%s LIKE transactions", sqlDateFormat)
		insertQueryTransactions := fmt.Sprintf("INSERT transactions_%s SELECT * FROM transactions", sqlDateFormat)

		_, err = db.Query(createQueryVisit)
		if err != nil {
			panic(err)
		}
		_, err = db.Query(insertQueryVisit)
		if err != nil {
			panic(err)
		}
		_, err = db.Query("TRUNCATE visits")
		if err != nil {
			panic(err)
		}

		_, err = db.Query(createQueryTransactions)
		if err != nil {
			panic(err)
		}
		_, err = db.Query(insertQueryTransactions)
		if err != nil {
			panic(err)
		}
		_, err = db.Query("TRUNCATE transactions")
		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, "/", 301)
	*/
}

// flip the bracelet_id from active = 0 to 1
func reopenSession(w http.ResponseWriter, r *http.Request) {
	braceletID, err := parseBraceletID(r)
	if err != nil {
		writeError(w, err.Error(), nil)
		return
	}

	// can't reopen an already open session
	sessionOpen, _ := doesSessionAlreadyExist(braceletID)
	if sessionOpen {
		writeError(w, ErrSessionAlreadyExists, nil)
		return
	}

	if err := db.Exec("UPDATE visits SET exit_time = NULL, total = NULL, active = ? WHERE bracelet_id = ? ORDER BY invoice_id DESC LIMIT 1", 1, braceletID).Error; err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	// to re-open items that were marked as paid
	visit, err := getVisitFromBraceletID(braceletID)
	if err != nil {
		writeError(w, ErrSessionDoesNotExist, nil)
		fmt.Println("Using visit variable %s", visit)
		return
	}

	// set paid=0, notes<>'entry' is != and sort of hacky but needed for this feature
	// TODO get rid of this hack !
	if err := db.Exec("UPDATE transactions SET paid = ? WHERE notes<>'entry' AND bracelet_id = ? AND invoice_id = ?", 0, 1, braceletID).Error; err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	http.Redirect(w, r, "/reopenOrDeleteSessionPage", 301)
}

func deleteSession(w http.ResponseWriter, r *http.Request) {

	braceletID, err := parseBraceletID(r)
	if err != nil {
		writeError(w, err.Error(), nil)
		return
	}

	// can't delete an already open session
	sessionOpen, _ := doesSessionAlreadyExist(braceletID)
	if sessionOpen {
		writeError(w, ErrSessionAlreadyExists, nil)
		return
	}

	var visit Visit
	if err := db.Where("bracelet_id = ? AND active = ?", braceletID, 0).Last(&visit).Error; err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	if err := db.Where("bracelet_id = ? AND invoice_id = ?", braceletID, visit.InvoiceID).Delete(Transaction{}).Error; err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}
	if err := db.Where("bracelet_id = ? AND invoice_id = ?", braceletID, visit.InvoiceID).Delete(Visit{}).Error; err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	http.Redirect(w, r, "/reopenOrDeleteSessionPage", 301)
}

func selectTodaysMenu(w http.ResponseWriter, r *http.Request) {
	if err := db.Table("items").Where("item_type = ?", "food").Updates(Item{Active: 0}).Error; err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	allFoods := getActiveItems("allfood") // not really all "active" ones but w/e we can rename function or something
	for _, foodItem := range allFoods {
		if r.FormValue(foodItem.Name) == "on" {
			if err := db.Table("items").Where("item_type = ? AND name = ?", "food", foodItem.Name).Updates(Item{Active: 1}).Error; err != nil {
				writeError(w, ErrWithSQLquery, err)
				return
			}
		}
	}
	http.Redirect(w, r, "/addItems", 301)
}

// for adding new items on the fly, has some UX bugs
func insertNewItems(w http.ResponseWriter, r *http.Request) {
	item := Item{
		Name:   r.FormValue("name"),
		Type:   r.FormValue("type"),
		Notes:  "todo",
		Active: 1,
	}
	item.Price, err = strconv.Atoi(r.FormValue("price"))
	if err != nil {
		writeError(w, err.Error(), err)
		return
	}
	switch item.Type {
	case "food", "drink", "misc":
		db.NewRecord(item)
		db.Create(&item)
	default:
		writeError(w, "type of item entered not available; select from: food,drink,misc", nil)
		return
	}

	// to show that the new item was added
	// note: is set to active=1 by default
	http.Redirect(w, r, "/addItems", 301)
}

func main() {
	password := flag.String("password", "", "password for mysql db")
	flag.Parse()

	cfg := mysql.Config{
		User:   "root",
		DBName: "myBusiness",
		Passwd: *password,
		Params: map[string]string{
			"parseTime": "true",
			"loc":       "EST",
			"charset":   "utf8",
		},
	}

	db, err = gorm.Open("mysql", cfg.FormatDSN())

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	// file serving endpoints
	http.HandleFunc("/", homePage)
	http.HandleFunc("/newSession", newSession)
	http.HandleFunc("/addItems", addItems)
	http.HandleFunc("/closeSession", closeSession)
	http.HandleFunc("/adminPage", adminPage)

	// places they route to
	http.HandleFunc("/initializeSession", initializeSession)
	http.HandleFunc("/addItemsToASession", addItemsToASession)
	http.HandleFunc("/displayBill", displayBill)
	http.HandleFunc("/closeBill", closeBill)
	http.HandleFunc("/reopenSession", reopenSession)
	http.HandleFunc("/deleteSession", deleteSession)
	http.HandleFunc("/closeDay", closeDay)
	http.HandleFunc("/selectTodaysMenu", selectTodaysMenu)
	http.HandleFunc("/insertNewItems", insertNewItems)

	// admin options
	http.HandleFunc("/selectTodaysMenuPage", selectTodaysMenuPage)
	http.HandleFunc("/statisticsPage", statisticsPage)
	http.HandleFunc("/reopenOrDeleteSessionPage", reopenOrDeleteSessionPage)
	http.HandleFunc("/insertNewItemsPage", insertNewItemsPage)

	// for js
	http.HandleFunc("/isLockerActive", isLockerActive)

	fmt.Println("POS app started, listening on port 8081")
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		panic(err.Error())
	}
}
