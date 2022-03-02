import { Result } from '@chainlink/gauntlet-core'
import { logger } from '@chainlink/gauntlet-core/dist/utils'
import { TransactionResponse } from '@chainlink/gauntlet-terra'
import { CATEGORIES } from '../../../../lib/constants'
import { abstract, AbstractInstruction } from '../../..'

type CommandInput = {
  proposalId: string
}

type ContractInput = {
  id: string
}

const makeCommandInput = async (flags: any, args: string[]): Promise<CommandInput> => {
  if (flags.input) return flags.input as CommandInput
  return {
    proposalId: flags.proposalId,
  }
}

const makeContractInput = async (input: CommandInput): Promise<ContractInput> => {
  return {
    id: input.proposalId,
  }
}

const validateInput = (input: CommandInput): boolean => {
  if (!input.proposalId) throw new Error('A proposal ID is required. Provide it with --proposalId flag')
  return true
}

const afterExecute = (response: Result<TransactionResponse>): { proposalId: string; digest: string } | undefined => {
  const events = response.responses[0].tx.events
  if (!events) {
    logger.error('Could not retrieve events from tx')
    return
  }

  const proposalId = events[0].wasm.proposal_id[0]
  const digest = events[0].wasm.digest[0]
  logger.success(`Proposal ${proposalId} finalized`)
  logger.line()
  logger.info('Important: Save the proposal DIGEST to accept the proposal in the future:')
  logger.info(digest)
  logger.line()
  return {
    proposalId,
    digest,
  }
}

// yarn gauntlet ocr2:finalize_proposal --network=bombay-testnet --proposalId=4 terra14nrtuhrrhl2ldad7gln5uafgl8s2m25du98hlx
const instruction: AbstractInstruction<CommandInput, ContractInput> = {
  instruction: {
    contract: 'ocr2',
    function: 'finalize_proposal',
    category: CATEGORIES.OCR,
  },
  makeInput: makeCommandInput,
  validateInput: validateInput,
  makeContractInput: makeContractInput,
  afterExecute,
}

export default abstract.instructionToCommand(instruction)
