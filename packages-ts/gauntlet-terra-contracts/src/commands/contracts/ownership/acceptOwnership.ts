import { AbstractInstruction } from '../..'
import { CATEGORIES } from '../../../lib/constants'
import { CONTRACT_LIST } from '../../../lib/contracts'

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

export const makeAcceptOwnershipInstruction = (contractId: CONTRACT_LIST) => {
  const acceptOwnershipInstruction: AbstractInstruction<CommandInput, ContractInput> = {
    instruction: {
      category: CATEGORIES.OWNERSHIP,
      contract: contractId,
      function: 'accept_ownership',
    },
    makeInput: makeCommandInput,
    validateInput: validateInput,
    makeContractInput: makeContractInput,
  }

  return acceptOwnershipInstruction
}
