import { CATEGORIES } from '../../../lib/constants'
import { getRDD } from '../../../lib/rdd'
import { abstract, AbstractInstruction } from '../..'

type OnchainConfig = any
type CommandInput = {
  f: number
  proposalId: number
  signers: string[]
  transmitters: string[]
  payees: string[]
  onchainConfig: OnchainConfig
}

type ContractInput = {
  f: number
  id: number
  onchain_config: string
  signers: string[]
  transmitters: string[]
  payees: string[]
}

const makeCommandInput = async (flags: any, args: string[]): Promise<CommandInput> => {
  if (flags.input) return flags.input as CommandInput
  const rdd = getRDD(flags.rdd)
  const contract = args[0]
  const aggregator = rdd.contracts[contract]
  const aggregatorOperators: any[] = aggregator.oracles.map((o) => rdd.operators[o.operator])
  const signers = aggregatorOperators.map((o) => o.ocr2OnchainPublicKey[0].replace('ocr2on_terra_', ''))
  const transmitters = aggregatorOperators.map((o) => o.ocrNodeAddress[0])
  const payees = aggregatorOperators.map((o) => o.adminAddress)

  return {
    f: aggregator.config.f,
    proposalId: flags.proposalId,
    signers,
    transmitters,
    payees,
    onchainConfig: '',
  }
}

const makeContractInput = async (input: CommandInput): Promise<ContractInput> => {
  const signers = input.signers.map((s) => Buffer.from(s, 'hex').toString('base64'))

  return {
    f: Number(input.f),
    id: input.proposalId,
    onchain_config: input.onchainConfig,
    signers: signers,
    transmitters: input.transmitters,
    payees: input.payees,
  }
}

const validateInput = (input: CommandInput): boolean => {
  if (3 * input.f >= input.signers.length)
    throw new Error(`Signers length needs to be higher than 3 * f (${3 * input.f}). Currently ${input.signers.length}`)

  if (input.signers.length !== input.transmitters.length)
    throw new Error(`Signers and Trasmitters length are different`)

  if (input.transmitters.length !== input.payees.length) throw new Error(`Trasmitters and Payees length are different`)

  return true
}

// yarn gauntlet ocr2:propose_config --network=bombay-testnet --proposalId=4 --rdd=../reference-data-directory/directory-terra-mainnet.json terra14nrtuhrrhl2ldad7gln5uafgl8s2m25du98hlx
const instruction: AbstractInstruction<CommandInput, ContractInput> = {
  instruction: {
    contract: 'ocr2',
    function: 'propose_config',
    category: CATEGORIES.OCR,
  },
  makeInput: makeCommandInput,
  validateInput: validateInput,
  makeContractInput: makeContractInput,
}

export default abstract.instructionToCommand(instruction)
