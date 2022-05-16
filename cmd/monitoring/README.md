# Terra on-chain monitor

## Example of running the monitor locally

```bash
TERRA_TENDERMINT_URL="<some terra url>" \
TERRA_FCD_URL="https://fcd.terra.dev/" \
TERRA_NETWORK_NAME="terra-devnet" \
TERRA_NETWORK_ID="terra-devnet" \
TERRA_CHAIN_ID="1" \
TERRA_READ_TIMEOUT="15s" \
TERRA_POLL_INTERVAL="5s" \
TERRA_LINK_TOKEN_ADDRESS="terra1eq0xqc88ceuvw2ztz2a08200he8lrgvnplrcst" \
KAFKA_BROKERS="localhost:29092" \
KAFKA_CLIENT_ID=“terra” \
KAFKA_SECURITY_PROTOCOL="PLAINTEXT" \
KAFKA_SASL_MECHANISM="PLAIN" \
KAFKA_SASL_USERNAME="" \
KAFKA_SASL_PASSWORD="" \
KAFKA_TRANSMISSION_TOPIC="transmission_topic" \
KAFKA_CONFIG_SET_SIMPLIFIED_TOPIC="config_set_simplified" \
SCHEMA_REGISTRY_URL="http://localhost:8989" \
SCHEMA_REGISTRY_USERNAME="" \
SCHEMA_REGISTRY_PASSWORD="" \
HTTP_ADDRESS="localhost:3000" \
FEEDS_URL="http://localhost:4000/feeds.json" \
go run ./cmd/monitoring/main.go
```

## Example of feed configurations returned by weiwatchers.com

```json
[
  {
    "name": "UST/USD",
    "path": "ust-usd",
    "symbol": "$",
    "heartbeat": 50,
    "contract_type": "",
    "status": "live",
    "contract_address_bech32": "terra1dre5vgujqex83kc4kw3jr6fc6z8erdsqxlsvhk" ,
    "multiply": "100000000"
  }
]
```

```
# Terra

## Events

### OCR2 contract

```
- set_link_token
    - old_link_token: string
    - new_link_token: string
- receive_funds
    - sender: Address
    - amount: uint64
- set_config
    - previous_config_block_number: string (uint64)
    - latest_config_digest: string (hex)
    - config_count: string (int?)
    - signers: []string (hex encoded pub keys)
    - transmitters: []string (ocr2types.Account?!)
    - payees: []string (ocr2types.PublicKey)
    - f:string (uint8)
    - onchain_config: string (base64)
    - offchain_config_version: string (int)
    - offchain_config: string (base64)
- round_requested
    - requester: string
    - config_digest: string (hex)
    - round: string (int)
    - epoch: string (int)
- transmitted
    - config_digest: string (hex)
    - epoch: string (int)
```




        new_transmission:
                    - aggregator_round_id: string (uint64)
                - answer: string (int)
                - transmitter: string (hex?)
                    observations_timestamp:
                ),
                attr("observers", hex::encode(report.observers)),
                attr("juels_per_fee_coin", report.juels_per_fee_coin.to_string()),
                attr("config_digest", hex::encode(config_digest)),
                attr("epoch", config.epoch.to_string()),
                attr("round", config.round.to_string()),
                attr("reimbursement", reimbursement.to_string()),
            ])
            .add_attributes(observations),


            Event::new("set_link_token")
                .add_attribute("old_link_token", old_link_token.0)
                .add_attribute("new_link_token", config.link_token.0),

        Event::new("set_billing")
            .add_attribute(
                "recommended_gas_price_micro",
                config.billing.recommended_gas_price_micro.to_string(),
            )
            .add_attribute(
                "observation_payment_gjuels",
                config.billing.observation_payment_gjuels.to_string(),
            )
            .add_attribute(
                "transmission_payment_gjuels",
                config.billing.transmission_payment_gjuels.to_string(),
            ),


        Event::new("oracle_paid")
            .add_attribute("transmitter", &transmitter)
            .add_attribute("payee", &payee)
            .add_attribute("amount", amount.to_string())
            .add_attribute("link_token", config.link_token.0),


            Event::new("payeeship_transfer_requested")
                .add_attribute("transmitter", &transmitter)
                .add_attribute("current", current_payee.unwrap().as_str())
                .add_attribute("proposed", &proposed),

        Event::new("payeeship_transferred")
            .add_attribute("transmitter", &transmitter)
            .add_attribute(
                "previous",
                current_payee.as_ref().map(|p| p.as_str()).unwrap_or(""),
            )
            .add_attribute("current", &info.sender),

### Proxy-OCR2

                Event::new("propose_contract")
                    .add_attribute("current_address", current_address)
                    .add_attribute("proposed_address", address),

                Event::new("confirm_contract")
                    .add_attribute("old_address", old_address)
                    .add_attribute("new_address", current_phase.contract_address),
```
