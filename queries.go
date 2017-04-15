package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// TODO deal with panics in here...

func getVisitFromBraceletID(braceletID int) (*Visit, error) {
	visit := new(Visit)
	rows := db.QueryRow("SELECT bracelet_id, invoice_id, entry_time FROM visits WHERE bracelet_id=? AND active=1", braceletID)
	err := rows.Scan(&visit.BraceletID, &visit.InvoiceID, &visit.EntryTime)
	if err != nil {
		return nil, err
	}
	return visit, nil
}

func doesSessionAlreadyExist(braceletID int) (bool, string) {
	var invoiceID string
	err := db.QueryRow("SELECT invoice_id FROM visits WHERE bracelet_id=? AND active=1", braceletID).Scan(&invoiceID)

	if err == sql.ErrNoRows {
		// bracelet ID does not currently exist in open visits
		return false, ""
	} else {
		// bracelet ID does exist - do not create new session
		return true, invoiceID
	}
}

func getActiveVisits(active bool) []*Visit {
	var oneOrZero int
	if active {
		oneOrZero = 1
	}
	rows, err := db.Query("SELECT bracelet_id, entry_time, invoice_id FROM visits WHERE active=?", oneOrZero)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var ActiveVisits []*Visit
	for rows.Next() {
		activeVisit := new(Visit)
		err = rows.Scan(&activeVisit.BraceletID, &activeVisit.EntryTime, &activeVisit.InvoiceID)
		if err != nil {
			panic(err)
		}
		ActiveVisits = append(ActiveVisits, activeVisit)
	}

	return ActiveVisits
}

func transactionsFromQuery(query string) []*Transaction {
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var Txns []*Transaction
	for rows.Next() {
		txn := new(Transaction)
		err = rows.Scan(&txn.Name, &txn.Amount, &txn.Total)
		if err != nil {
			panic(err)
		}
		Txns = append(Txns, txn)
	}
	return Txns
}

func getTransactionsByTypeFinalBill(txType string) []*Transaction {
	queryType := "SELECT name, amount, total FROM transactions WHERE paid=1 AND type='%s'" // single quotes essential
	query := fmt.Sprintf(queryType, txType)
	return transactionsFromQuery(query)
}

func getTransactionsByTypeName(txType, itemName string) []*Transaction {
	queryName := "SELECT name, amount, total FROM transactions WHERE paid=0 AND type='%s' AND name='%s'"
	query := fmt.Sprintf(queryName, txType, itemName)
	return transactionsFromQuery(query)
}

func getTransactionsByTypeNameID(txType, itemName string, braceletID int) []*Transaction {
	queryNameAndID := "SELECT name, amount, total FROM transactions WHERE paid=0 AND type='%s' AND name='%s' AND bracelet_id='%v'"
	query := fmt.Sprintf(queryNameAndID, txType, itemName, braceletID)
	return transactionsFromQuery(query)
}

func getActiveItems(itemType string) []*Item {
	var query string
	var queryString = "SELECT name, price FROM items WHERE active=1 AND item_type='%s' ORDER by name"

	switch itemType {
	case "food", "misc", "drink":
		query = fmt.Sprintf(queryString, itemType)
	case "allfood":
		query = "SELECT name, price FROM items WHERE item_type='food' ORDER by name"
	default:
		panic(fmt.Errorf("wrong item type provided"))
	}

	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var Items []*Item

	for rows.Next() {
		item := new(Item)
		err = rows.Scan(&item.Name, &item.Price)
		if err != nil {
			panic(err)
		}
		Items = append(Items, item)
	}
	return Items
}

func getLastTransactions(last string) []*Transaction {
	rows, err := db.Query("SELECT name, amount, bracelet_id FROM (SELECT * FROM transactions WHERE paid=0 ORDER BY id DESC LIMIT ?) sub ORDER BY id ASC", last)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var PartialInvoice []*Transaction

	for rows.Next() {
		i := new(Transaction)
		err = rows.Scan(&i.Name, &i.Amount, &i.BraceletID)
		if err != nil {
			panic(err)
		}
		PartialInvoice = append(PartialInvoice, i)
	}
	return PartialInvoice
}

func getAllTransactionsFromVisit(braceletID, invoiceID, paid int, tYpe string) []*Transaction {
	rows, err := db.Query("SELECT name, amount, total FROM transactions WHERE paid=? AND bracelet_id=? AND invoice_id=? AND type=?", paid, braceletID, invoiceID, tYpe)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var PartialInvoice []*Transaction

	for rows.Next() {
		i := new(Transaction)
		err = rows.Scan(&i.Name, &i.Amount, &i.Total)
		if err != nil {
			panic(err)
		}
		PartialInvoice = append(PartialInvoice, i)
	}
	return PartialInvoice
}
