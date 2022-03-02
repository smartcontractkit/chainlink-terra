import { AccAddress } from '@terra-money/terra.js'
import { abstract, AbstractInstruction } from '../..'
import { CATEGORIES } from '../../../lib/constants'
import { CONTRACT_LIST, getContract } from '../../../lib/contracts'

type CommandInput = {
  address: string
}

type ContractInput = {
  address: string
}

const makeCommandInput = async (flags: any, args: string[]): Promise<CommandInput> => {
  return {
    address: flags.address,
  }
}

const makeContractInput = async (input: CommandInput): Promise<ContractInput> => {
  return {
    address: input.address,
  }
}

const validateInput = (input: CommandInput): boolean => {
  if (!AccAddress.validate(input.address)) {
    throw new Error(`Invalid address`)
  }

  return true
}

const addAccess: AbstractInstruction<CommandInput, ContractInput> = {
  instruction: {
    category: CATEGORIES.ACCESS_CONTROLLER,
    contract: 'access_controller',
    function: 'add_access',
  },
  makeInput: makeCommandInput,
  validateInput: validateInput,
  makeContractInput: makeContractInput,
}

export default abstract.instructionToCommand(addAccess)
