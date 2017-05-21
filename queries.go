package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// TODO deal with panics in here...

func getVisitFromBraceletID(braceletID int) (*Visit, error) {
	var visit Visit
	if err := db.Where("bracelet_id = ? AND active = ?", braceletID, 1).Find(&visit).Error; err != nil {
		return nil, err
	}
	return &visit, nil
}

func doesSessionAlreadyExist(braceletID int) (bool, int) {
	var visit Visit
	if err := db.Where("bracelet_id = ? AND active = ?", braceletID, 1).Find(&visit).Error; err != nil {
		// bracelet ID does not currently exist in open visits
		return false, 0
	}
	return true, visit.InvoiceID
}

func getActiveVisits(active bool) []*Visit {
	var oneOrZero int
	if active {
		oneOrZero = 1
	}
	var visits []*Visit
	if err := db.Where("active = ?", oneOrZero).Find(&visits).Error; err != nil {
		panic(err)
	}
	return visits
}

func getTransactionsByTypeFinalBill(txType string) []*Transaction {
	var transactions []*Transaction
	if err := db.Where("paid = ? AND type = ?", 1, txType).Find(&transactions).Error; err != nil {
		panic(err)
	}
	return transactions
}

func getTransactionsByTypeName(txType, itemName string) []*Transaction {
	var transactions []*Transaction
	if err := db.Where("paid = ? AND type = ? AND name = ?", 0, txType, itemName).Find(&transactions).Error; err != nil {
		panic(err)
	}
	return transactions
}

func getTransactionsByTypeNameID(txType, itemName string, braceletID int) []*Transaction {
	var transactions []*Transaction
	if err := db.Where("paid = ? AND type = ? AND name = ? AND bracelet_id = ?", 0, txType, itemName, braceletID).Find(&transactions).Error; err != nil {
		panic(err)
	}
	return transactions
}

func getActiveItems(itemType string) []*Item {
	var items []*Item

	switch itemType {
	case "food", "misc", "drink":
		if err := db.Where("active = ? AND item_type = ?", 1, itemType).Order("name").Find(&items).Error; err != nil {
			panic(err)
		}
	case "allfood":
		if err := db.Where("item_type = ?", "food").Order("name").Find(&items).Error; err != nil {
			panic(err)
		}
	default:
		panic(fmt.Errorf("wrong item type provided"))
	}
	return items
}

func getLastTransactions(limit string) []*Transaction {
	var partialInvoice []*Transaction
	if err := db.Where("paid = ?", 0).Order("id desc").Limit(limit).Find(&partialInvoice).Error; err != nil {
		panic(err)
	}
	return partialInvoice
}

func getAllTransactionsFromVisit(braceletID, invoiceID, paid int, typeParam string) []*Transaction {
	var partialInvoice []*Transaction
	if err := db.Where("paid = ? AND bracelet_id = ? AND invoice_id = ? AND type = ?", paid, braceletID, invoiceID, typeParam).Order("id desc").Find(&partialInvoice).Error; err != nil {
		panic(err)
	}
	return partialInvoice
}
