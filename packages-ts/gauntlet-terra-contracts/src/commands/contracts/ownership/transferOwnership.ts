import { AccAddress } from '@terra-money/terra.js'
import { AbstractInstruction } from '../..'
import { CATEGORIES } from '../../../lib/constants'
import { CONTRACT_LIST } from '../../../lib/contracts'

type CommandInput = {
  to: string
}

type ContractInput = {
  to: string
}

const makeCommandInput = async (flags: any, args: string[]): Promise<CommandInput> => {
  return {
    to: flags.to,
  }
}

const makeContractInput = async (input: CommandInput): Promise<ContractInput> => {
  return {
    to: input.to,
  }
}

const validateInput = (input: CommandInput): boolean => {
  if (!AccAddress.validate(input.to)) {
    throw new Error(`Invalid proposed owner address!`)
  }

  return true
}

export const makeTransferOwnershipInstruction = (contractId: CONTRACT_LIST) => {
  const transferOwnershipInstruction: AbstractInstruction<CommandInput, ContractInput> = {
    instruction: {
      category: CATEGORIES.OWNERSHIP,
      contract: contractId,
      function: 'transfer_ownership',
    },
    makeInput: makeCommandInput,
    validateInput: validateInput,
    makeContractInput: makeContractInput,
  }

  return transferOwnershipInstruction
}
