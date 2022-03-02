import { CATEGORIES } from '../../../lib/constants'
import { CONTRACT_LIST, getContract } from '../../../lib/contracts'
import { isValidAddress } from '../../../lib/utils'
import { abstract, AbstractInstruction } from '../..'

type CommandInput = {
  admin: string
}

type ContractInput = {
  admin: string
}

const makeCommandInput = async (flags: any, args: any[]): Promise<CommandInput> => {
  return {
    admin: flags.admin,
  } as CommandInput
}

const validateInput = (input: CommandInput): boolean => {
  if (!isValidAddress(input.admin)) {
    throw new Error('Admin address is not valid!')
  }

  return true
}

const makeContractInput = async (input: CommandInput): Promise<ContractInput> => {
  return {
    admin: input.admin,
  } as ContractInput
}

const createUpdateAdminInstruction: AbstractInstruction<CommandInput, ContractInput> = {
  examples: ['yarn gauntlet cw4_group:update_admin --admin=<NEW_ADMIN_ADDRESS> <CONTRACT_ADDRESS>'],
  instruction: {
    category: CATEGORIES.MULTISIG,
    contract: 'cw4_group',
    function: 'update_admin',
  },
  makeInput: makeCommandInput,
  validateInput,
  makeContractInput,
}

export const UpdateAdmin = abstract.instructionToCommand(createUpdateAdminInstruction)
