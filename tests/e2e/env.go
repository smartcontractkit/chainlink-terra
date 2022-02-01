package e2e

import "github.com/smartcontractkit/helmenv/environment"

// NewChainlinkTerraEnv returns a cluster config with LocalTerra node
func NewChainlinkTerraEnv(nodes int, stateful bool) *environment.Config {
	var db map[string]interface{}
	if stateful {
		db = map[string]interface{}{
			"stateful": true,
			"capacity": "2Gi",
		}
	} else {
		db = map[string]interface{}{
			"stateful": false,
		}
	}
	return &environment.Config{
		NamespacePrefix: "chainlink-terra",
		Charts: environment.Charts{
			"localterra": {
				Index: 1,
			},
			"mockserver-config": {
				Index: 2,
			},
			"mockserver": {
				Index: 3,
			},
			"chainlink": {
				Index: 4,
				Values: map[string]interface{}{
					"replicas": nodes,
					"chainlink": map[string]interface{}{
						"image": map[string]interface{}{
							"image":   "public.ecr.aws/z0b1w9r9/chainlink",
							"version": "candidate-develop-terr-qa-test-b64-decode-signers-3.233d4d167de902af193b8c83cc79e707d7e74541",
						},
					},
					"db": db,
					"env": map[string]interface{}{
						"EVM_ENABLED":                 "false",
						"EVM_RPC_ENABLED":             "false",
						"TERRA_ENABLED":               "true",
						"eth_disabled":                "true",
						"CHAINLINK_DEV":               "true",
						"USE_LEGACY_ETH_ENV_VARS":     "false",
						"FEATURE_OFFCHAIN_REPORTING2": "true",
						"feature_external_initiators": "true",
						"P2P_NETWORKING_STACK":        "V2",
						"P2PV2_LISTEN_ADDRESSES":      "0.0.0.0:6690",
						"P2PV2_DELTA_DIAL":            "5s",
						"P2PV2_DELTA_RECONCILE":       "5s",
						"p2p_listen_port":             "0",
					},
				},
			},
		},
	}
}
