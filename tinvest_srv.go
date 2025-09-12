package main

import (
	"ldv/tinvest"
	"ldv/tinvest/marketdataservice"
	"ldv/tinvest/operations"
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

	for i := 0; i < len(accDetail.Pos.Securities); i++ {
		accDetail.Pos.Securities[i].LastPrice = findLastPriceByFigi(accDetail.Pos.Securities[i].Figi, &lastPrices)
		if accDetail.Pos.Securities[i].InstrumentType == "bond" {
			accDetail.Pos.Securities[i].LastPrice = accDetail.Pos.Securities[i].LastPrice / 100.0 * tinvest.SumFloat(accDetail.Pos.Securities[i].InstrumentDesc.Nominal.Sum())
		}
	}
}

func findLastPriceByFigi(figi string, lp *marketdataservice.Prices) tinvest.SumFloat {
	for _, price := range lp.LastPrices {
		if price.Figi == figi {
			return tinvest.SumFloat(price.Price.Sum())
		}
	}
	return 0.0
}
