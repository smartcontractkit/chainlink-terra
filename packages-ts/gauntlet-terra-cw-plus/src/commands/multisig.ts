import { Result } from '@chainlink/gauntlet-core'
import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import { TerraCommand, TransactionResponse } from '@chainlink/gauntlet-terra'
import { AccAddress, MsgExecuteContract, MsgSend } from '@terra-money/terra.js'
import { isDeepEqual } from '../lib/utils'
import { fetchProposalState, makeInspectionMessage } from './inspect'
import { Vote, Cw3WasmMsg, Action, State, Cw3BankMsg } from '../lib/types'

type ProposalAction = (
  signer: AccAddress,
  proposalId: number,
  message: MsgExecuteContract | MsgSend,
) => Promise<MsgExecuteContract>

export const wrapCommand = (command) => {
  return class Multisig extends TerraCommand {
    command: TerraCommand
    multisig: AccAddress

    static id = `${command.id}:multisig`

    constructor(flags, args) {
      super(flags, args)
    }

    buildCommand = async (flags, args): Promise<TerraCommand> => {
      if (!AccAddress.validate(process.env.CW3_FLEX_MULTISIG)) throw new Error(`Invalid Multisig wallet address`)
      if (!AccAddress.validate(process.env.CW4_GROUP)) throw new Error(`Invalid Multisig group address`)
      this.multisig = process.env.CW3_FLEX_MULTISIG as AccAddress

      const c = new command(flags, args) as TerraCommand
      await c.invokeMiddlewares(c, c.middlewares)
      this.command = c.buildCommand ? await c.buildCommand(flags, args) : c
      return this.command
    }

    makeRawTransaction = async (signer: AccAddress, state?: State) => {
      const message = await this.command.makeRawTransaction(this.multisig)

      const operations = {
        [Action.CREATE]: this.makeProposeTransaction,
        [Action.APPROVE]: this.makeAcceptTransaction,
        [Action.EXECUTE]: this.makeExecuteTransaction,
        [Action.NONE]: () => {
          throw new Error('No action needed')
        },
      }

      if (state.proposal.nextAction !== Action.CREATE) {
        this.require(
          await this.isSameProposal(state.proposal.data, [this.toMsg(message)]),
          'The transaction generated is different from the proposal provided',
        )
      }

      return operations[state.proposal.nextAction](signer, Number(this.flags.proposal), message)
    }

    isSameProposal = (proposalMsgs: (Cw3WasmMsg | Cw3BankMsg)[], generatedMsgs: (Cw3WasmMsg | Cw3BankMsg)[]) => {
      return isDeepEqual(proposalMsgs, generatedMsgs)
    }

    toMsg = (message: MsgSend | MsgExecuteContract): Cw3BankMsg | Cw3WasmMsg => {
      if (message instanceof MsgSend) return this.toBankMsg(message as MsgSend)
      if (message instanceof MsgExecuteContract) return this.toWasmMsg(message as MsgExecuteContract)
    }

    toBankMsg = (message: MsgSend): Cw3BankMsg => {
      return {
        bank: {
          send: {
            amount: message.amount.toArray().map((c) => c.toData()),
            to_address: message.to_address,
          },
        },
      }
    }

    toWasmMsg = (message: MsgExecuteContract): Cw3WasmMsg => {
      return {
        wasm: {
          execute: {
            contract_addr: message.contract,
            funds: message.coins.toArray().map((c) => c.toData()),
            msg: Buffer.from(JSON.stringify(message.execute_msg)).toString('base64'),
          },
        },
      }
    }

    makeProposeTransaction: ProposalAction = async (signer, _, message) => {
      logger.info('Generating data for creating new proposal')
      const proposeInput = {
        propose: {
          description: command.id,
          msgs: [this.toMsg(message)],
          title: command.id,
          // TODO: Set expiration time
          // latest: { at_height: 7970238 },
        },
      }
      return new MsgExecuteContract(signer, this.multisig, proposeInput)
    }

    makeAcceptTransaction: ProposalAction = async (signer, proposalId) => {
      logger.info(`Generating data for approving proposal ${proposalId}`)
      const approvalInput = {
        vote: {
          vote: Vote.YES,
          proposal_id: proposalId,
        },
      }
      return new MsgExecuteContract(signer, this.multisig, approvalInput)
    }

    makeExecuteTransaction: ProposalAction = async (signer, proposalId) => {
      logger.info(`Generating data for executing proposal ${proposalId}`)
      const executeInput = {
        execute: {
          proposal_id: proposalId,
        },
      }
      return new MsgExecuteContract(signer, this.multisig, executeInput)
    }

    fetchState = async (proposalId?: number): Promise<State> => {
      const query = this.provider.wasm.contractQuery.bind(this.provider.wasm)
      return fetchProposalState(query)(this.multisig, proposalId)
    }

    printPostInstructions = async (proposalId: number) => {
      const state = await this.fetchState(proposalId)
      if (!state.proposal.id) {
        logger.error(`Proposal ${proposalId} not found`)
        return
      }
      const approvalsLeft = state.multisig.threshold - state.proposal.approvers.length
      const messages = {
        passed: `The proposal reached the threshold and can be executed. Run the same command with the flag --proposal=${proposalId}`,
        open: `The proposal needs ${approvalsLeft} more approvals. Run the same command with the flag --proposal=${proposalId}`,
        pending: `The proposal needs ${approvalsLeft} more approvals. Run the same command with the flag --proposal=${proposalId}`,
        rejected: `The proposal has been rejected. No actions available`,
        executed: `The proposal has been executed. No more actions needed`,
      }
      logger.line()
      logger.info(`${messages[state.proposal.currentStatus]}`)
      logger.line()
    }

    execute = async () => {
      // TODO: Gauntlet core should initialize commands using `buildCommand` instead of new Command
      await this.buildCommand(this.flags, this.args)

      let proposalId = !!this.flags.proposal && Number(this.flags.proposal)
      const state = await this.fetchState(proposalId)
      logger.info(makeInspectionMessage(state))

      if (state.proposal.nextAction === Action.NONE) {
        await this.printPostInstructions(proposalId)
        return
      }
      const rawTx = await this.makeRawTransaction(this.wallet.key.accAddress, state)

      const actionMessage = {
        [Action.CREATE]: 'CREATING',
        [Action.APPROVE]: 'APPROVING',
        [Action.EXECUTE]: 'EXECUTING',
      }

      if (this.flags.execute) {
        await this.command.beforeExecute(this.multisig)

        await prompt(`Continue ${actionMessage[state.proposal.nextAction]} proposal?`)
        const tx = await this.signAndSend([rawTx])
        let response: Result<TransactionResponse> = {
          responses: [
            {
              tx,
              contract: this.multisig,
            },
          ],
          data: {
            proposalId,
          },
        }

        if (state.proposal.nextAction === Action.CREATE) {
          const proposalFromEvent = tx.events[0].wasm.proposal_id[0]
          logger.success(`New proposal created with ID: ${proposalFromEvent}`)
          proposalId = Number(proposalFromEvent)
        }

        if (state.proposal.nextAction === Action.EXECUTE && this.command.afterExecute) {
          const data = this.command.afterExecute(response)
          response = { ...response, data: { ...data } }
        }

        logger.success(`TX finished at ${tx.hash}`)
        await this.printPostInstructions(proposalId)

        return response
      }

      // TODO: Test raw message
      const msgData = Buffer.from(JSON.stringify(rawTx.execute_msg)).toString('base64')
      logger.line()
      logger.success(`Message generated succesfully for ${actionMessage[state.proposal.nextAction]} proposal`)
      logger.log()
      logger.log(msgData)
      logger.log()
      logger.line()

      return {
        responses: [
          {
            tx: {},
            contract: this.multisig,
          },
        ],
        data: {
          proposalId,
          message: msgData,
        },
      } as Result<TransactionResponse>
    }
  }
}
