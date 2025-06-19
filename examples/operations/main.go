package main

import (
	"fmt"

	"github.com/nitayStain/x-aio/internal/operations"
)

func main() {
	ops, err := operations.GetOperations()
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, op := range ops {
		fmt.Printf(
			"Name: %-35s | Type: %-7s | QueryID: %s\n",
			op.OperationName, op.OperationType, op.QueryID,
		)
	}
}
