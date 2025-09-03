package main

import (
	"ldv/tinvest/operations"
	"log"
)

func addOperationsBySecurity(token string, accDetail *AccDetail) {
	var opers operations.Opers
	for _, sec := range accDetail.Pos.Securities {
		opers = operations.GetOperations(token, accDetail.Account.Id, sec.Figi)
		beginQuantity := 0
		lenOpers := len(opers.Operations) - 1
		zeroIndex := lenOpers
		for i := lenOpers; i >=0; i-- {
			switch opers.Operations[i].OperationType {
			case "OPERATION_TYPE_BUY":
				for _, trade := range opers.Operations[i].Trades {
					beginQuantity += int(trade.Quality)
				}
			case "OPERATION_TYPE_SELL":
				for _, trade := range opers.Operations[i].Trades {
					beginQuantity -= int(trade.Quality)
				}
			}
			if beginQuantity == 0 {
				zeroIndex = lenOpers - i
			}
		}
		log.Printf("%v, zeroIndex = %d\n", sec.Ticker, zeroIndex)
	}

}
