import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { getRDD } from '../../../lib/rdd'
import { abstract, AbstractInstruction } from '../..'
import { CATEGORIES } from '../../../lib/constants'

type CommandInput = {
  recommendedGasPriceMicro: string
  observationPaymentGjuels: number
  transmissionPaymentGjuels: number
}

type ContractInput = {
  config: {
    recommended_gas_price_micro: string
    observation_payment_gjuels: number
    transmission_payment_gjuels: number
  }
}

const makeCommandInput = async (flags: any, args: string[]): Promise<CommandInput> => {
  if (flags.input) return flags.input as CommandInput
  const rdd = getRDD(flags.rdd)
  const contract = args[0]
  const billingInfo = rdd.contracts[contract]?.billing
  return {
    observationPaymentGjuels: billingInfo.observationPaymentGjuels,
    transmissionPaymentGjuels: billingInfo.transmissionPaymentGjuels,
    recommendedGasPriceMicro: billingInfo.recommendedGasPriceMicro,
  }
}

const makeContractInput = async (input: CommandInput): Promise<ContractInput> => {
  return {
    config: {
      observation_payment_gjuels: new BN(input.observationPaymentGjuels).toNumber(),
      transmission_payment_gjuels: new BN(input.transmissionPaymentGjuels).toNumber(),
      recommended_gas_price_micro: input.recommendedGasPriceMicro,
    },
  }
}

const validateInput = (input: CommandInput): boolean => {
  let observationPayment: BN
  let transmissionPayment: BN

  const gasPrice: number = Number(input.recommendedGasPriceMicro) // parse as float64
  if (!isFinite(gasPrice)) {
    throw new Error(`recommendedGasPriceMicro=${input.recommendedGasPriceMicro} is not a valid floating point number.`)
  }

  if (gasPrice < 0.0) {
    throw new Error(`recommendedGasPriceMicro=${input.recommendedGasPriceMicro} cannot be negative`)
  }

  try {
    observationPayment = new BN(input.observationPaymentGjuels)
    transmissionPayment = new BN(input.transmissionPaymentGjuels) // parse as integers
  } catch {
    throw new Error(
      `observationPaymentGjuels=${input.observationPaymentGjuels} and ` +
        `transmissionPaymentGjuels=${input.transmissionPaymentGjuels} must both be integers`,
    )
  }
  if (observationPayment.isNeg() || transmissionPayment.isNeg()) {
    throw new Error(
      `observationPaymentGjuels=${input.observationPaymentGjuels} and ` +
        `transmissionPaymentGjuels=${input.transmissionPaymentGjuels} cannot be negative`,
    )
  }
  return true
}

const setBillingInstruction: AbstractInstruction<CommandInput, ContractInput> = {
  instruction: {
    category: CATEGORIES.OCR,
    contract: 'ocr2',
    function: 'set_billing',
  },
  makeInput: makeCommandInput,
  validateInput: validateInput,
  makeContractInput: makeContractInput,
}

export default abstract.instructionToCommand(setBillingInstruction)
