import { inspection, logger } from '@chainlink/gauntlet-core/dist/utils'
import { CONTRACT_LIST } from '../../../../lib/contracts'
import { CATEGORIES } from '../../../../lib/constants'
import { getRDD } from '../../../../lib/rdd'
import { abstract, InspectInstruction } from '../../..'

const MIN_LINK_AVAILABLE = '100'

type ContractExpectedInfo = {
  digest: string
  description: string
  decimals: string | number
  transmitters: string[]
  billingAccessController: string
  requesterAccessController: string
  link: string
  linkAvailable: string
  billing: {
    observationPaymentGjuels: string
    recommendedGasPriceMicro: string
    transmissionPaymentGjuels: string
  }
}

const makeInput = async (flags: any, args: string[]): Promise<ContractExpectedInfo> => {
  if (flags.input) return flags.input as ContractExpectedInfo
  const rdd = getRDD(flags.rdd)
  const contract = args[0]
  const info = rdd.contracts[contract]
  const aggregatorOperators: string[] = info.oracles.map((o) => o.operator)
  const transmitters = aggregatorOperators.map((o) => rdd.operators[o].ocrNodeAddress[0])
  const billingAccessController = flags.billingAccessController || process.env.BILLING_ACCESS_CONTROLLER
  const requesterAccessController = flags.requesterAccessController || process.env.REQUESTER_ACCESS_CONTROLLER
  const link = flags.link || process.env.LINK

  return {
    digest: flags.digest,
    description: info.name,
    decimals: info.decimals,
    transmitters,
    billingAccessController,
    requesterAccessController,
    link,
    linkAvailable: MIN_LINK_AVAILABLE,
    billing: {
      observationPaymentGjuels: info.billing.observationPaymentGjuels,
      recommendedGasPriceMicro: info.billing.recommendedGasPriceMicro,
      transmissionPaymentGjuels: info.billing.transmissionPaymentGjuels,
    },
  }
}

const makeInspectionData = () => async (input: ContractExpectedInfo): Promise<ContractExpectedInfo> => input

const makeOnchainData = () => (instructionsData: any[]): ContractExpectedInfo => {
  const latestConfigDetails = instructionsData[0]
  const description = instructionsData[1]
  const transmitters = instructionsData[2]
  const decimals = instructionsData[3]
  const billing = instructionsData[4]
  const billingAC = instructionsData[5]
  const requesterAC = instructionsData[6]
  const link = instructionsData[7]
  const linkAvailable = instructionsData[8]

  return {
    description,
    decimals,
    transmitters: transmitters.addresses,
    billingAccessController: billingAC,
    requesterAccessController: requesterAC,
    link,
    linkAvailable: linkAvailable.amount,
    digest: Buffer.from(latestConfigDetails.config_digest).toString('hex'),
    billing: {
      observationPaymentGjuels: billing.observation_payment_gjuels,
      transmissionPaymentGjuels: billing.transmission_payment_gjuels,
      recommendedGasPriceMicro: billing.recommended_gas_price_micro,
    },
  }
}

const inspect = (expected: ContractExpectedInfo, onchainData: ContractExpectedInfo): boolean => {
  const inspections: inspection.Inspection[] = [
    inspection.makeInspection(onchainData.description, expected.description, 'Description'),
    inspection.makeInspection(onchainData.decimals, expected.decimals, 'Decimals'),
    inspection.makeInspection(onchainData.digest, expected.digest, 'Offchain config digest'),
    inspection.makeInspection(onchainData.transmitters, expected.transmitters, 'Transmitters'),
    inspection.makeInspection(
      onchainData.billingAccessController,
      expected.billingAccessController,
      'Billing Access Controller',
    ),
    inspection.makeInspection(
      onchainData.requesterAccessController,
      expected.requesterAccessController,
      'Requester Access Controller',
    ),
    inspection.makeInspection(onchainData.link, expected.link, 'LINK'),
    inspection.makeInspection(
      onchainData.billing.observationPaymentGjuels,
      expected.billing.observationPaymentGjuels,
      'Observation Payment',
    ),
    inspection.makeInspection(
      onchainData.billing.recommendedGasPriceMicro,
      expected.billing.recommendedGasPriceMicro,
      'Recommended Gas Price',
    ),
    inspection.makeInspection(
      onchainData.billing.transmissionPaymentGjuels,
      expected.billing.transmissionPaymentGjuels,
      'Transmission Payment',
    ),
  ]
  logger.line()
  logger.info('Inspection results:')
  logger.info(`LINK Available: ${onchainData.linkAvailable}`)
  return inspection.inspect(inspections)
}

const instruction: InspectInstruction<any, ContractExpectedInfo> = {
  command: {
    category: CATEGORIES.OCR,
    contract: CONTRACT_LIST.OCR_2,
    id: 'inspect',
  },
  instructions: [
    {
      contract: 'ocr2',
      function: 'latest_config_details',
    },
    {
      contract: 'ocr2',
      function: 'description',
    },
    {
      contract: 'ocr2',
      function: 'transmitters',
    },
    {
      contract: 'ocr2',
      function: 'decimals',
    },
    {
      contract: 'ocr2',
      function: 'billing',
    },
    {
      contract: 'ocr2',
      function: 'billing_access_controller',
    },
    {
      contract: 'ocr2',
      function: 'requester_access_controller',
    },
    {
      contract: 'ocr2',
      function: 'link_token',
    },
    {
      contract: 'ocr2',
      function: 'link_available_for_payment',
    },
  ],
  makeInput,
  makeInspectionData,
  makeOnchainData,
  inspect,
}

export default abstract.instructionToInspectCommand<any, ContractExpectedInfo>(instruction)
