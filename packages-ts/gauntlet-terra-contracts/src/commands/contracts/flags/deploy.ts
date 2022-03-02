import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { AccAddress } from '@terra-money/terra.js'
import { abstract, AbstractInstruction } from '../..'
import { CATEGORIES } from '../../../lib/constants'
import { CONTRACT_LIST, getContract } from '../../../lib/contracts'

type CommandInput = {
  raisingAccessController: string
  loweringAccessController: string
}

type ContractInput = {
  raising_access_controller: string
  lowering_access_controller: string
}

const makeCommandInput = async (flags: any, args: string[]): Promise<CommandInput> => {
  return {
    raisingAccessController: flags.raisingAccessController,
    loweringAccessController: flags.loweringAccessController,
  }
}

const makeContractInput = async (input: CommandInput): Promise<ContractInput> => {
  return {
    raising_access_controller: input.raisingAccessController,
    lowering_access_controller: input.loweringAccessController,
  }
}

const validateInput = (input: CommandInput): boolean => {
  if (!AccAddress.validate(input.raisingAccessController)) throw new Error(`Invalid raisingAccessController address`)
  if (!AccAddress.validate(input.loweringAccessController)) throw new Error(`Invalid loweringAccessController address`)

  return true
}

const deploy: AbstractInstruction<CommandInput, ContractInput> = {
  instruction: {
    category: CATEGORIES.FLAGS,
    contract: CONTRACT_LIST.FLAGS,
    function: 'deploy',
  },
  makeInput: makeCommandInput,
  validateInput: validateInput,
  makeContractInput: makeContractInput,
}

export default abstract.instructionToCommand(deploy)
