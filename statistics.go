package main

import (
	"strconv"
)

func stringsToInts(strs []string) ([]int, error) {
	var err error
	ints := make([]int, len(strs))
	for i, s := range strs {
		ints[i], err = strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
	}
	return ints, nil
}

func getVisitTotal(visit *Visit, paid int) int {

	visit.FinalBill.Foods = totalItemsFromTransactions(getAllTransactionsFromVisit(visit.BraceletID, visit.InvoiceID, paid, ItemTypes.food))

	visit.FinalBill.Drinks = totalItemsFromTransactions(getAllTransactionsFromVisit(visit.BraceletID, visit.InvoiceID, paid, ItemTypes.drink))

	visit.FinalBill.Miscs = totalItemsFromTransactions(getAllTransactionsFromVisit(visit.BraceletID, visit.InvoiceID, paid, ItemTypes.misc))

	totalFood := totalFromItems(visit.FinalBill.Foods)
	totalDrink := totalFromItems(visit.FinalBill.Drinks)
	totalMisc := totalFromItems(visit.FinalBill.Miscs)

	// add them up
	return totalFood + totalDrink + totalMisc
}

func totalFromItems(items []*Item) int {
	total := 0
	for _, tot := range items {
		total += tot.Total
	}
	return total
}

func totalItemsFromTransactions(txs []*Transaction) []*Item {
	someItems := make(map[string]int)
	someAmounts := make(map[string]int)

	for _, tx := range txs {
		someItems[tx.Name] += tx.Total
		someAmounts[tx.Name] += tx.Amount
	}

	var theseItems []*Item
	for name, total := range someItems {
		theseItems = append(theseItems, &Item{
			Name:   name,
			Amount: someAmounts[name],
			Total:  total,
		})
	}
	return theseItems
}

func getMenuSummaryFromDay(stats *Statistics) *BillOrMenu {
	deezFoods := totalItemsFromTransactions(stats.Foods)
	deezDrinks := totalItemsFromTransactions(stats.Drinks)
	deezMiscs := totalItemsFromTransactions(stats.Miscs)

	return &BillOrMenu{
		Foods:  deezFoods,
		Drinks: deezDrinks,
		Miscs:  deezMiscs,
	}
}

func calculateDailyTotals(txns []*Transaction) int {
	var total int
	for _, tx := range txns {
		total += tx.Total
	}
	return total
}
