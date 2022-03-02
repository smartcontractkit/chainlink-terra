import { Result } from '@chainlink/gauntlet-core'
import { logger } from '@chainlink/gauntlet-core/dist/utils'
import { TransactionResponse } from '@chainlink/gauntlet-terra'
import { CATEGORIES } from '../../../../lib/constants'
import { abstract, AbstractInstruction } from '../../..'

type CommandInput = {
  proposalId: string
  digest: string
}

type ContractInput = {
  id: string
  digest: string
}

const makeCommandInput = async (flags: any, args: string[]): Promise<CommandInput> => {
  if (flags.input) return flags.input as CommandInput
  return {
    proposalId: flags.proposalId,
    digest: flags.digest,
  }
}

const makeContractInput = async (input: CommandInput): Promise<ContractInput> => {
  return {
    id: input.proposalId,
    digest: Buffer.from(input.digest, 'hex').toString('base64'),
  }
}

const validateInput = (input: CommandInput): boolean => {
  if (!input.proposalId) throw new Error('A proposal ID is required. Provide it with --proposalId flag')
  return true
}

const afterExecute = (response: Result<TransactionResponse>) => {
  const events = response.responses[0].tx.events
  if (!events) {
    logger.error('Could not retrieve events from tx')
    return
  }

  const digest = events[0]['wasm-set_config'].latest_config_digest[0]
  logger.success(`Proposal accepted`)
  logger.line()
  logger.info('Important: To inspect the aggregator, save the following DIGEST:')
  logger.info(digest)
  logger.line()
  return {
    digest,
  }
}

// yarn gauntlet ocr2:accept_proposal --network=bombay-testnet --id=4 --digest=71e6969c14c3e0cd47d75da229dbd2f76fd0f3c17e05635f78ac755a99897a2f terra14nrtuhrrhl2ldad7gln5uafgl8s2m25du98hlx
const instruction: AbstractInstruction<CommandInput, ContractInput> = {
  instruction: {
    contract: 'ocr2',
    function: 'accept_proposal',
    category: CATEGORIES.OCR,
  },
  makeInput: makeCommandInput,
  validateInput: validateInput,
  makeContractInput: makeContractInput,
  afterExecute,
}

export default abstract.instructionToCommand(instruction)
