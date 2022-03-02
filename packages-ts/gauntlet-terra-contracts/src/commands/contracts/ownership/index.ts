import { makeTransferOwnershipInstruction } from './transferOwnership'
import { makeAcceptOwnershipInstruction } from './acceptOwnership'
import { abstract } from '../..'
import { CONTRACT_LIST } from '../../../lib/contracts'

export const makeTransferOwnershipCommand = (contractId: CONTRACT_LIST) => {
  return abstract.instructionToCommand(makeTransferOwnershipInstruction(contractId))
}

export const makeAcceptOwnershipCommand = (contractId: CONTRACT_LIST) => {
  return abstract.instructionToCommand(makeAcceptOwnershipInstruction(contractId))
}
