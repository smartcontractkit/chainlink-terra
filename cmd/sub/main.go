package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pkgClient "github.com/smartcontractkit/chainlink-terra/pkg/terra/client"
	"github.com/smartcontractkit/chainlink/core/logger"
)

func main() {
	coreLog, closeLggr := logger.NewLogger()
	defer func() {
		if err := closeLggr(); err != nil {
			log.Println(fmt.Sprintf("Error while closing Logger: %v", err))
		}
	}()

	client, err := pkgClient.NewClient(
		"1",
		"blabla",
		10*time.Second, // read timeout
		coreLog,
	)
	if err != nil {
		panic(err)
	}

	if err := client.Start(); err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Stop(); err != nil {
			panic(err)
		}
	}()

	contractAddressBech32 := "terra10kc4n52rk4xqny3hdew3ggjfk9r420pqxs9ylf"

	ctx := context.Background()
	query := fmt.Sprintf(`wasm-new_transmission.contract_address='%s'`, contractAddressBech32)
	resultsCh, err := client.Subscribe(ctx, query, 0)
	if err != nil {
		panic(err)
	}
	defer client.Unsubscribe(ctx, query)

	for {
		select {
		case result := <-resultsCh:
			fmt.Println(">>>>", result)
		}
	}
}
