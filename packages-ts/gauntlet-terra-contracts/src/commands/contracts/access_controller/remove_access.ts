import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { bech32 } from 'bech32'
import { AbstractInstruction, instructionToCommand } from '../../abstract/wrapper'

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
  const { prefix: decodedPrefix } = bech32.decode(input.address) // throws error if checksum is invalid which will fail validation

  // verify address prefix
  if (decodedPrefix !== 'terra') {
    throw new Error(`Invalid address prefix (expecteed: 'terra')`)
  }

  return true
}

const removeAccess: AbstractInstruction<CommandInput, ContractInput> = {
  instruction: {
    contract: 'access_controller',
    function: 'remove_access',
  },
  makeInput: makeCommandInput,
  validateInput: validateInput,
  makeContractInput: makeContractInput,
}

export default instructionToCommand(removeAccess)
