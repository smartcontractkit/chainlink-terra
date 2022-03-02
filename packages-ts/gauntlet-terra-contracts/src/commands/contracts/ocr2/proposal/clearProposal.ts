import { Result } from '@chainlink/gauntlet-core'
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
  if (!input.proposalId) throw new Error('A proposal ID is required. Provide it with --id flag')
  return true
}

const afterExecute = async (response: Result<TransactionResponse>) => {
  console.log(response.data)
  return
}

// yarn gauntlet ocr2:clear_proposal --network=bombay-testnet --id=7 terra14nrtuhrrhl2ldad7gln5uafgl8s2m25du98hlx
const instruction: AbstractInstruction<CommandInput, ContractInput> = {
  instruction: {
    category: CATEGORIES.OCR,
    contract: 'ocr2',
    function: 'clear_proposal',
  },
  makeInput: makeCommandInput,
  validateInput: validateInput,
  makeContractInput: makeContractInput,
  afterExecute,
}

export default abstract.instructionToCommand(instruction)
