import { TerraABI } from './schema'

export type Contract = {
  id: string // ContractList
  abi: TerraABI
  bytecode: string
}

//export type Contracts = Record<ContractId, Contract>
export interface ContractGetter<ContractList> {
  (id: ContractList, version: string): Promise<Contract>
}
