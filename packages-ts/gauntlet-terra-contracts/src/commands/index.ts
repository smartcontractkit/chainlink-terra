import { getContract, CONTRACT_LIST } from '../lib/contracts'
import {
  AbstractTools,
  AbstractInstructionTemplate,
  InspectInstructionTemplate,
} from '@chainlink/gauntlet-terra/dist/commands/abstract'
const enumKeys = Object.keys(CONTRACT_LIST)
const contractRecords: Record<string, CONTRACT_LIST> = enumKeys.reduce(
  // create map from strings to enum keys
  (rec: Record<string, CONTRACT_LIST>, k: CONTRACT_LIST) => ((rec[CONTRACT_LIST[k]] = k), rec),
  {},
)
export const abstract: AbstractTools<CONTRACT_LIST> = new AbstractTools<CONTRACT_LIST>(contractRecords, getContract)
export type AbstractInstruction<CommandInput, ContractInput> = AbstractInstructionTemplate<
  CommandInput,
  ContractInput,
  CONTRACT_LIST
>
export type InspectInstruction<CommandInput, ContractInput> = InspectInstructionTemplate<
  CommandInput,
  ContractInput,
  CONTRACT_LIST
>

import Upload from './tooling/upload'
import TransferLink from './contracts/link/transfer'
import DeployLink from './contracts/link/deploy'
import OCR2 from './contracts/ocr2'
import AccessController from './contracts/access_controller'
import Flags from './contracts/flags'
import Proxy_OCR2 from './contracts/proxy_ocr2'
import DeviationFlaggingValidator from './contracts/deviation_flagging_validator'
import Multisig from './contracts/multisig'
import CW4_GROUP from './contracts/cw4_group'

export default [
  Upload,
  DeployLink,
  TransferLink,
  ...OCR2,
  ...AccessController,
  ...Flags,
  ...DeviationFlaggingValidator,
  ...Proxy_OCR2,
  ...Multisig,
  ...CW4_GROUP,
]
