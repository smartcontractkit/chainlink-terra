package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	opsCore "github.com/smartcontractkit/chainlink-relay/ops"
	gauntlet "github.com/smartcontractkit/chainlink-terra/ops/deployer/gauntlet"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// initiate `gauntlet` usage
		terra, err := gauntlet.New(ctx)
		if err != nil {
			return err
		}

		// start creating environment and use deployer interface for deploying contracts
		if err := opsCore.New(ctx, &terra, ObservationSource, JuelsSource, RelayConfig); err != nil {
			return err
		}

		return nil
	})
}

func RelayConfig(ctx *pulumi.Context, addresses map[int]string) (map[string]string, error) {
	return map[string]string{
		"nodeType":      config.Require(ctx, "CL-NODE_TYPE"),
		"tendermintURL": config.Require(ctx, "CL-TENDERMINT_URL"),
		"fcdURL":        config.Require(ctx, "CL-FCD_URL"),
		"chainID":       config.Require(ctx, "CL-RELAY_CHAINID"),
	}, nil
}

func ObservationSource(priceAdapter string) string {
	return fmt.Sprintf(`
	 ea  [type=bridge name=%s requestData=<{"data":{"from":"LINK", "to":"USD"}}>]
	 parse [type="jsonparse" path="result"]
	 multiply [type="multiply" times=100000000]

	 ea -> parse -> multiply
	 `,
		priceAdapter)
}

func JuelsSource(priceAdapter string) string {
	return fmt.Sprintf(`
	 link2usd [type=bridge name=%s requestData=<{"data":{"from":"LINK", "to":"USD"}}>]
	 parseL [type="jsonparse" path="result"]

	 luna2usd [type=bridge name=%s requestData=<{"data":{"from":"LUNA", "to":"USD"}}>]
	 parseT [type="jsonparse" path="result"]

	 // parseL (dollars/LINK)
	 // parseT (dollars/LUNA)
	 // parseT / parseL = LINK/LUNA
	 divide [type="divide" input="$(parseT)" divisor="$(parseL)" precision="18"]
   scale [type="multiply" times=1000000000000000000]

	 link2usd -> parseL -> divide
	 luna2usd -> parseT -> divide
	 divide -> scale
	 `,
		priceAdapter, priceAdapter)
}
