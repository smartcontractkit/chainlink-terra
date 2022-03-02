import { CATEGORIES } from '../../../lib/constants'
import { isValidAddress } from '../../../lib/utils'
import { abstract, AbstractInstruction } from '../../'

type CW4_GROUP_Member = {
  addr: string
  weight: number
}

type CommandInput = {
  add: string[]
  remove: string[]
}

type ContractInput = {
  add: CW4_GROUP_Member[]
  remove: string[]
}

const makeCommandInput = async (flags: any, args: any[]): Promise<CommandInput> => {
  return {
    add: flags.add?.split(',') || [],
    remove: flags.remove?.split(',') || [],
  } as CommandInput
}

const validateInput = (input: CommandInput): boolean => {
  if (!input.add.every((addr) => isValidAddress(addr))) {
    throw new Error("One of provided 'add' addresses is not valid!")
  }

  if (!input.remove.every((addr) => isValidAddress(addr))) {
    throw new Error("One of provided 'remove' addresses of not valid!")
  }

  if (input.add.length === 0 && input.remove.length === 0) {
    throw new Error("You must specify 'add' or 'remove' addresses!")
  }

  return true
}

const makeContractInput = async (input: CommandInput): Promise<ContractInput> => {
  const membersToAdd = input.add.map((addr: string) => {
    return {
      addr,
      weight: 1,
    } as CW4_GROUP_Member
  })

  return {
    add: membersToAdd,
    remove: input.remove,
  } as ContractInput
}

const createUpdateMembersInstruction: AbstractInstruction<CommandInput, ContractInput> = {
  examples: [
    'yarn gauntlet cw4_group:update_members --add=<ADDRESS1_TO_ADD>,<ADDRESS2_TO_ADD> --remove=<ADDRESS3_TO_REMOVE> <CONTRACT_ADDRESS>',
    'yarn gauntlet cw4_group:update_members --add=<ADDRESS1_TO_ADD> <CONTRACT_ADDRESS>',
    'yarn gauntlet cw4_group:update_members --remove=<ADDRESS1_TO_REMOVE> <CONTRACT_ADDRESS>',
  ],
  instruction: {
    category: CATEGORIES.MULTISIG,
    contract: 'cw4_group',
    function: 'update_members',
  },
  makeInput: makeCommandInput,
  validateInput,
  makeContractInput,
}

export const UpdateMembers = abstract.instructionToCommand(createUpdateMembersInstruction)
