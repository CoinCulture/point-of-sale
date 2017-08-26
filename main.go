package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB
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

	query := fmt.Sprintf("SELECT bracelet_id, invoice_id FROM visits ORDER BY invoice_id DESC LIMIT %v", number)
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var lastOpenedVisits []*Visit

	for rows.Next() {
		last := new(Visit)
		err = rows.Scan(&last.BraceletID, &last.InvoiceID)
		if err != nil {
			panic(err)
		}
		last.Total = getVisitTotal(last, 1) // won't show pay_at_end, that's fine
		lastOpenedVisits = append(lastOpenedVisits, last)
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

	// initialize a visit, generates an invoiceID
	_, err = db.Exec("INSERT INTO visits(date, bracelet_id, entry_time, active) values (?, ?, ?, ?)", visit.Date, visit.BraceletID, CurrentTime(), "1")
	if err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

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
		_, err = db.Exec("INSERT INTO transactions(invoice_id, bracelet_id, name, amount, price, total, time_ordered, notes, type, paid) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", visit.InvoiceID, visit.BraceletID, "punch a pass", 1, 0, 0, CurrentTime(), entryThing, "misc", paid)
	}

	if adult {
		_, err = db.Exec("INSERT INTO transactions(invoice_id, bracelet_id, name, amount, price, total, time_ordered, notes, type, paid) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", visit.InvoiceID, visit.BraceletID, "adult admission", 1, 5, 5, CurrentTime(), entryThing, "misc", paid)
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
			_, err = db.Exec("INSERT INTO transactions(invoice_id, bracelet_id, name, amount, price, total, time_ordered, notes, type, paid) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", visit.InvoiceID, visit.BraceletID, activeItems.Name, numberOrdered, activeItems.Price, total, CurrentTime(), entryThing, "misc", paid)
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

	// if the quick "Add Session" box was checked, an non-existing session
	// can be initialized on the fly
	if r.FormValue("quick_add") == "quick_add" {
		// initialize a visit, generates an invoiceID (duplicate with initliazing a session
		_, err = db.Exec("INSERT INTO visits(date, bracelet_id, entry_time, active) values (?, ?, ?, ?)", DateString(), braceletID, CurrentTime(), "1")
		if err != nil {
			writeError(w, ErrWithSQLquery, err)
			return
		}
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
				_, err = db.Exec("INSERT INTO transactions(invoice_id, bracelet_id, name, amount, price, total, time_ordered, notes, type, paid) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", visit.InvoiceID, visit.BraceletID, activeItems.Name, numberOrdered, activeItems.Price, total, CurrentTime(), notes, category, 0) // <= category is clutch here
				if err != nil {
					writeError(w, ErrWithSQLquery, err)
					return
				}
				// print the food (but not voids)
				if category == ItemTypes.Food && num0 > 0 {
					dontPrint, dontNotify := handleFoodExceptions()

					if !dontPrint[activeItems.Name] {
						
						if err := printTheChit(braceletID, numberOrdered, activeItems.Name); err != nil {
							fmt.Printf("printer error:\n%v", err)
						}
					}

					if !dontNotify[activeItems.Name] {
						if err := activateNotification(); err != nil {
							fmt.Printf("notification error:\n%v", err)
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

	visit.ExitTime = CurrentTime()

	_, err = db.Exec("UPDATE visits SET exit_time=?,active=0 WHERE bracelet_id=? AND invoice_id=?", visit.ExitTime, visit.BraceletID, visit.InvoiceID)
	// total is calculated from bill displayed & doesn't include items bought up front, which it should
	//_, err = db.Exec("UPDATE visits SET exit_time=?,total=?,active=0 WHERE bracelet_id=? AND invoice_id=?", visit.ExitTime, visit.Total, visit.BraceletID, visit.InvoiceID)
	if err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	_, err = db.Exec("UPDATE transactions SET paid=1 WHERE bracelet_id=? AND invoice_id=?", visit.BraceletID, visit.InvoiceID)
	if err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	http.Redirect(w, r, "/closeSession", 301)
}

func closeDay(w http.ResponseWriter, r *http.Request) {

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

	_, err = db.Exec("UPDATE visits SET exit_time=NULL,total=NULL,active=1 WHERE bracelet_id=? ORDER BY invoice_id DESC LIMIT 1", braceletID)
	if err != nil {
		if err != nil {
			panic(err)
		}
	}

	// to re-open items that were marked as paid
	visit, err := getVisitFromBraceletID(braceletID)
	if err != nil {
		writeError(w, ErrSessionDoesNotExist, nil)
		return
	}

	// set paid=0, notes<>'entry' is != and sort of hacky but needed for this feature
	// TODO get rid of this hack !
	_, err = db.Exec("UPDATE transactions SET paid=0 WHERE notes<>'entry' AND bracelet_id=? AND invoice_id=?", visit.BraceletID, visit.InvoiceID)
	if err != nil {
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

	var invoiceID int
	if err := db.QueryRow("SELECT invoice_id FROM visits WHERE bracelet_id=? AND active=0 ORDER BY invoice_id DESC LIMIT 1", braceletID).Scan(&invoiceID); err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	_, err = db.Exec("DELETE from transactions WHERE bracelet_id=? AND invoice_id=?", braceletID, invoiceID)
	if err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	_, err = db.Exec("DELETE from visits WHERE bracelet_id=? AND invoice_id=?", braceletID, invoiceID)
	if err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	http.Redirect(w, r, "/reopenOrDeleteSessionPage", 301)
}

func selectTodaysMenu(w http.ResponseWriter, r *http.Request) {
	_, err := db.Exec("UPDATE items SET active=0 WHERE item_type='food'")
	if err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}

	allFoods := getActiveItems("allfood") // not really all "active" ones but w/e we can rename function or something
	for _, foodItem := range allFoods {
		if r.FormValue(foodItem.Name) == "on" {

			_, err := db.Exec("UPDATE items SET active = 1 WHERE name=? AND item_type='food'", foodItem.Name)
			if err != nil {
				writeError(w, ErrWithSQLquery, err)
				return
			}
		}
	}
	http.Redirect(w, r, "/addItems", 301)
}

// for adding new items on the fly, has some UX bugs
func insertNewItems(w http.ResponseWriter, r *http.Request) {
	item := new(Item)
	item.Name = r.FormValue("name")
	item.Type = r.FormValue("type")
	item.Price, err = strconv.Atoi(r.FormValue("price"))
	if err != nil {
		writeError(w, err.Error(), err)
		return
	}
	var query string
	queryString := "INSERT into items (name,price,notes,item_type,active) VALUES ('%s', %v, '%s', '%s', 1)"
	switch item.Type {
	case "food", "drink", "misc":
		query = fmt.Sprintf(queryString, item.Name, item.Price, "todo", item.Type)
	default:
		writeError(w, "type of item entered not available; select from: food,drink,misc", nil)
		return
	}
	_, err = db.Exec(query)
	if err != nil {
		writeError(w, ErrWithSQLquery, err)
		return
	}
	// to show that the new item was added
	// note: is set to active=1 by default
	http.Redirect(w, r, "/addItems", 301)
}

func main() {
	password := flag.String("password", "", "password for mysql db")
	flag.Parse()

	if *password == "" {
		fmt.Println("[NOTICE] No password provided")
	}

	config := LoadConfig()

	cfg := mysql.Config{
		User:   config.db.user,
		DBName: config.db.databaseName,
		Passwd: *password,
		Net:    config.db.net,
		Addr:   config.db.getAddr(),
		Params: map[string]string{
			"parseTime": "true",
			"loc":       "EST",
		},
	}

	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

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

	fmt.Printf("POS app started, listening on port %v\n", config.http.port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", config.http.port), nil)
	if err != nil {
		panic(err.Error())
	}
}
