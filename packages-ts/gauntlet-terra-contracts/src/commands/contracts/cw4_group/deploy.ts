import { CATEGORIES } from '../../../lib/constants'
import { isValidAddress } from '../../../lib/utils'
import { CONTRACT_LIST, getContract } from '../../../lib/contracts'
import { abstract, AbstractInstruction } from '../..'

type CommandInput = {
  owners: string[]
  admin: string
}

type ContractInput = {
  members: {
    addr: string
    weight: number
  }[]
  admin: string
}

const makeCommandInput = async (flags: any, args: any[]): Promise<CommandInput> => {
  if (flags.input) return flags.input as CommandInput
  return {
    owners: args,
    admin: flags.admin,
  }
}
const validateInput = (input: CommandInput): boolean => {
  if (input.owners.length === 0) {
    throw new Error(`You must specify at least one group member (wallet owner)`)
  }
  const areValidOwners = input.owners.filter((owner) => !isValidAddress(owner)).length === 0
  if (!areValidOwners) throw new Error('Owners are not valid')
  if (!isValidAddress(input.admin)) throw new Error('Admin is not valid')
  return true
}

const makeContractInput = async (input: CommandInput): Promise<ContractInput> => {
  return {
    members: input.owners.map((owner) => ({
      addr: owner,
      // Same weight for every owner
      weight: 1,
    })),
    admin: input.admin,
  }
}

const createGroupInstruction: AbstractInstruction<CommandInput, ContractInput> = {
  examples: ['yarn gauntlet cw4_group:deploy --network=bombay-testnet --admin=<ADMIN_ADDRESS> <OWNERS_LIST>'],
  instruction: {
    category: CATEGORIES.MULTISIG,
    contract: 'cw4_group',
    function: 'deploy',
  },
  makeInput: makeCommandInput,
  validateInput,
  makeContractInput,
}

export const CreateGroup = abstract.instructionToCommand(createGroupInstruction)
