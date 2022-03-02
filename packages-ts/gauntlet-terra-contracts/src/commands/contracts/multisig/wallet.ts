import { logger } from '@chainlink/gauntlet-core/dist/utils'
import { CATEGORIES } from '../../../lib/constants'
import { isValidAddress } from '../../../lib/utils'
import { abstract, AbstractInstruction } from '../..'

// 7 days
const DEFAULT_MULTISIG_EXPIRATION_TIME_IN_SECS = 7 * 24 * 60 * 60

type Duration = {
  time: number // length of time in seconds
}

type CommandInput = {
  group: string
  threshold: number
  votingPeriod: {
    time: number
  }
}

type ContractInput = {
  group_addr: string
  max_voting_period: Duration
  threshold: Threshold
}

type Threshold = {
  absolute_count?: {
    weight: number
  }
  absolute_percentage?: {
    percentage: number
  }
  threshold_quorum?: {
    threshold: number
    quorum: number
  }
}

const makeCommandInput = async (flags: any): Promise<CommandInput> => {
  if (flags.input) return flags.input as CommandInput
  return {
    group: flags.group,
    threshold: Number(flags.threshold),
    votingPeriod: {
      time: Number(flags.time) || DEFAULT_MULTISIG_EXPIRATION_TIME_IN_SECS,
    },
  }
}

const validateInput = (input: CommandInput): boolean => {
  // TODO: Add time validation
  const isValidTime = (a: any) => {
    if (!a) return false
    if (Number(a) <= 0) return false
    return true
  }
  if (!isValidAddress(input.group)) {
    throw new Error(`group ${input.group} is not a valid terra address`)
  }

  if (input.threshold === 0) {
    throw new Error(`Threshold ${input.threshold} is invalid. Should be higher than zero`)
  }

  if (!isValidTime(input.votingPeriod.time)) {
    throw new Error(`Voting period time ${input.votingPeriod.time} is not a valid time`)
  }

  return true
}

const makeContractInput = async (input: CommandInput): Promise<ContractInput> => {
  return {
    group_addr: input.group,
    max_voting_period: {
      time: input.votingPeriod.time,
    },
    threshold: {
      absolute_count: {
        weight: input.threshold,
      },
    },
  }
}

const createWalletInstruction: AbstractInstruction<CommandInput, ContractInput> = {
  examples: [
    'yarn gauntlet cw3_flex_multisig:deploy --network=bombay-testnet --group=<GROUP_ADDRESS> --threshold=<THRESHOLD> (--time=<EXPIRATION_TIME_IN_SECS>)',
  ],
  instruction: {
    category: CATEGORIES.MULTISIG,
    contract: 'cw3_flex_multisig',
    function: 'deploy',
  },
  makeInput: makeCommandInput,
  validateInput,
  makeContractInput: makeContractInput,
}

export const CreateWallet = abstract.instructionToCommand(createWalletInstruction)
