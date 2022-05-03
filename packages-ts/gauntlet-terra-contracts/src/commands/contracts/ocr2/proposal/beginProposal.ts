import { Result } from '@chainlink/gauntlet-core'
import { logger } from '@chainlink/gauntlet-core/dist/utils'
import { TransactionResponse } from '@chainlink/gauntlet-terra'
import { CATEGORIES } from '../../../../lib/constants'
import { AbstractInstruction, instructionToCommand } from '../../../abstract/executionWrapper'

type CommandInput = {}

type ContractInput = {}

const makeCommandInput = async (flags: any, args: string[]): Promise<CommandInput> => {
  return {}
}

const makeContractInput = async (input: CommandInput): Promise<ContractInput> => {
  return {}
}

const validateInput = (input: CommandInput): boolean => {
  return true
}

const afterExecute = (context) => async (
  response: Result<TransactionResponse>,
): Promise<{ proposalId: string } | undefined> => {
  const events = response.responses[0].tx.events
  if (!events) {
    logger.error('No events found. Config Proposal ID could not be retrieved')
    return
  }

  try {
    // Non-Multisig response parsing
    var proposalId = events.reduce((prev, curr) => {
      return curr.wasm.contract_address[0] == context.contract ? curr.wasm.proposal_id[0] : prev
    }, null)

    // Multisig response parsing
    proposalId =
      proposalId ||
      events[0].wasm.contract_address.reduce((prev, curr, i) => {
        return curr == context.contract ? events[0].wasm.proposal_id[i] : prev
      }, null)

    if (!proposalId) {
      throw new Error('ProposalId for the given contract does not exist inside events')
    }

    logger.success(`New config proposal created with Config Proposal ID: ${proposalId}`)
    return {
      proposalId,
    }
  } catch (e) {
    logger.error('Config Proposal ID not found inside events')
    return
  }
}

const instruction: AbstractInstruction<CommandInput, ContractInput> = {
  examples: ['yarn ocr2:begin_proposal --network=<NETWORK> <CONTRACT_ADDRESS>'],
  instruction: {
    contract: 'ocr2',
    function: 'begin_proposal',
    category: CATEGORIES.OCR,
  },
  makeInput: makeCommandInput,
  validateInput: validateInput,
  makeContractInput: makeContractInput,
  afterExecute,
}

export default instructionToCommand(instruction)
