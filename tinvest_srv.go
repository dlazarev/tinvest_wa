package main

import (
	"ldv/tinvest"
	"ldv/tinvest/marketdataservice"
	"ldv/tinvest/operations"
	"log"
)

func addOperationsBySecurity(token string, accDetail *AccDetail) {
	var opers operations.Opers
	for i := 0; i < len(accDetail.Pos.Securities); i++ {
		opers = operations.GetOperations(token, accDetail.Account.Id, accDetail.Pos.Securities[i].Figi)
		totalSum := 0.0
		totalQuantity := 0
		for i, oper := range opers.Operations {
			switch oper.OperationType {
			case "OPERATION_TYPE_BUY", "OPERATION_TYPE_BUY_CARD":
				for _, trade := range opers.Operations[i].Trades {
					totalQuantity += int(trade.Quality)
					totalSum += float64(trade.Price.Sum() * float64(trade.Quality))
				}
			}
		}
		if totalQuantity != 0 {
			accDetail.Pos.Securities[i].WeightedAveragePrice = tinvest.SumFloat(totalSum / float64(totalQuantity))
		}
		//		log.Printf("%v, zeroIndex = %d\n", sec.Ticker, zeroIndex)
	}

}

func getActualPrices(token string, accDetail *AccDetail) {
	var figies []string

	for _, sec := range accDetail.Pos.Securities {
		figies = append(figies, sec.Figi)
	}

	lastPrices := marketdataservice.GetLastPrices(token, figies)
	log.Println(lastPrices)
}
