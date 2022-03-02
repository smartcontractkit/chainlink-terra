import { Result } from '@chainlink/gauntlet-core'
import { logger } from '@chainlink/gauntlet-core/dist/utils'
import { TransactionResponse } from '@chainlink/gauntlet-terra'
import { CATEGORIES } from '../../../../lib/constants'
import { abstract, AbstractInstruction } from '../../..'

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

const afterExecute = (response: Result<TransactionResponse>): { proposalId: string } | undefined => {
  const events = response.responses[0].tx.events
  if (!events) {
    logger.error('No events found. Proposal ID could not be retrieved')
    return
  }

  try {
    const proposalId = events[0].wasm.proposal_id[0]
    logger.success(`New config proposal created with Proposal ID: ${proposalId}`)
    return {
      proposalId,
    }
  } catch (e) {
    logger.error('Proposal ID not found inside events')
    return
  }
}

const instruction: AbstractInstruction<CommandInput, ContractInput> = {
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

export default abstract.instructionToCommand(instruction)
