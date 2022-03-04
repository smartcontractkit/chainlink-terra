import { Result, WriteCommand } from '@chainlink/gauntlet-core'
import { logger } from '@chainlink/gauntlet-core/dist/utils'
import { EventsByType, MsgStoreCode, AccAddress, TxLog } from '@terra-money/terra.js'
import { SignMode } from '@terra-money/terra.proto/cosmos/tx/signing/v1beta1/signing'

import { withProvider, withWallet, withCodeIds, withNetwork } from '../middlewares'
import {
  BlockTxBroadcastResult,
  LCDClient,
  MsgExecuteContract,
  MsgInstantiateContract,
  TxError,
  Wallet,
} from '@terra-money/terra.js'
import { TransactionResponse } from '../types'
import { LedgerKey } from '../ledgerKey'

type CodeIds = Record<string, number>

export default abstract class TerraCommand extends WriteCommand<TransactionResponse> {
  wallet: Wallet
  provider: LCDClient
  contracts: string[]
  public codeIds: CodeIds
  abstract execute: () => Promise<Result<TransactionResponse>>
  abstract makeRawTransaction: (signer: AccAddress) => Promise<MsgExecuteContract[]>
  afterExecute?: (response: Result<TransactionResponse>) => any

  constructor(flags, args) {
    super(flags, args)
    this.use(withNetwork, withProvider, withWallet, withCodeIds)
  }

  parseResponseValue(receipt: any, eventType: string, attributeType: string) {
    try {
      const parsed = JSON.parse(receipt?.raw_log)
      const event = parsed[0].events.filter((event) => event.type === eventType)[0]
      if (event) {
        const attribute = event.attributes.filter((attr) => attr.key === attributeType)[0]
        return attribute.value
      }
    } catch (e) {
      logger.log('Error parsing response', e.message)
      return undefined
    }
  }

  makeEventsFromLogs = (logs: TxLog.Data[]): EventsByType[] => {
    if (!logs) return []
    return logs.map((log) => TxLog.fromData(log).eventsByType)
  }

  // TODO: need to add type of tx, address is parsed only for intantiation
  wrapResponse = (tx: BlockTxBroadcastResult): TransactionResponse => ({
    hash: tx.txhash,
    address: this.parseResponseValue(tx, 'instantiate_contract', 'contract_address'),
    wait: () => ({
      success: tx.logs.length > 0 && !(tx as TxError)?.code,
    }),
    tx,
    events: this.makeEventsFromLogs(tx.logs),
  })

  async query(address, input, params?): Promise<any> {
    return await this.provider.wasm.contractQuery(address, input, params)
  }

  signAndSend = async (messages: MsgExecuteContract[]): Promise<TransactionResponse> => {
    try {
      logger.loading('Signing transaction...')
      const tx = await this.wallet.createAndSignTx({
        msgs: messages,
        ...(this.wallet.key instanceof LedgerKey && {
          signMode: SignMode.SIGN_MODE_LEGACY_AMINO_JSON,
        }),
      })

      logger.loading('Sending transaction...')
      const res = await this.provider.tx.broadcast(tx)
      return this.wrapResponse(res)
    } catch (e) {
      const message = e?.response?.data?.message || e.message
      throw new Error(message)
    }
  }

  async call(address, input) {
    const msg = new MsgExecuteContract(this.wallet.key.accAddress, address, input)

    const tx = await this.wallet.createAndSignTx({
      msgs: [msg],
      ...(this.wallet.key instanceof LedgerKey && {
        signMode: SignMode.SIGN_MODE_LEGACY_AMINO_JSON,
      }),
    })

    const res = await this.provider.tx.broadcast(tx)
    return this.wrapResponse(res)
  }

  async callBatch(address, inputs) {
    const msgs: MsgExecuteContract[] = inputs.map(
      (input) => new MsgExecuteContract(this.wallet.key.accAddress, address, input),
    )

    const tx = await this.wallet.createAndSignTx({
      msgs: msgs,
      ...(this.wallet.key instanceof LedgerKey && {
        signMode: SignMode.SIGN_MODE_LEGACY_AMINO_JSON,
      }),
    })

    const res = await this.provider.tx.broadcast(tx)
    return this.wrapResponse(res)
  }

  async deploy(codeId, instantiateMsg, migrationContract = undefined) {
    const instantiate = new MsgInstantiateContract(
      this.wallet.key.accAddress,
      migrationContract,
      codeId,
      instantiateMsg,
    )
    const instantiateTx = await this.wallet.createAndSignTx({
      msgs: [instantiate],
      memo: 'Instantiating',
      ...(this.wallet.key instanceof LedgerKey && {
        signMode: SignMode.SIGN_MODE_LEGACY_AMINO_JSON,
      }),
    })
    logger.loading(`Deploying contract...`)
    const res = await this.provider.tx.broadcast(instantiateTx)

    return this.wrapResponse(res)
  }

  async upload(wasm, contractName) {
    const code = new MsgStoreCode(this.wallet.key.accAddress, wasm)

    const tx = await this.wallet.createAndSignTx({
      msgs: [code],
      memo: `Storing ${contractName}`,
      ...(this.wallet.key instanceof LedgerKey && {
        signMode: SignMode.SIGN_MODE_LEGACY_AMINO_JSON,
      }),
    })

    logger.loading(`Uploading ${contractName} contract code...`)
    const res = await this.provider.tx.broadcast(tx)

    return this.wrapResponse(res)
  }
}
