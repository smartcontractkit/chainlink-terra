import AbstractCommand, { makeAbstractCommand } from '.'
import { Result } from '@chainlink/gauntlet-core'
import { TerraCommand, TransactionResponse } from '@chainlink/gauntlet-terra'
import { AccAddress } from '@terra-money/terra.js'
import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import { Query } from './inspectionWrapper'

export type BeforeExecutionContext<Input, ContractInput> = {
  input: Input
  contractInput: ContractInput
  id: string
  contract: string
  query: Query
  flags: any
}

export type BeforeExecute<Input, ContractInput> = (
  context: BeforeExecutionContext<Input, ContractInput>,
) => (signer: AccAddress) => Promise<void>

export interface AbstractInstruction<Input, ContractInput> {
  examples?: string[]
  instruction: {
    category: string
    contract: string
    function: string
  }
  makeInput: (flags: any, args: string[]) => Promise<Input>
  validateInput: (input: Input) => boolean
  makeContractInput: (input: Input) => Promise<ContractInput>
  beforeExecute?: BeforeExecute<Input, ContractInput>
  afterExecute?: (response: Result<TransactionResponse>) => any
}

export const instructionToCommand = <Input, ContractInput>(instruction: AbstractInstruction<Input, ContractInput>) => {
  const id = `${instruction.instruction.contract}:${instruction.instruction.function}`
  const category = `${instruction.instruction.category}`
  const examples = instruction.examples || []
  return class Command extends TerraCommand {
    static id = id
    static category = category
    static examples = examples

    command: AbstractCommand

    constructor(flags, args) {
      super(flags, args)
    }

    defaultBeforeExecute = (context: BeforeExecutionContext<Input, ContractInput>) => async () => {
      logger.loading(`Executing ${context.id} from contract ${context.contract}`)
      logger.log('Input Params:', context.contractInput)
      await prompt(`Continue?`)
    }

    afterExecute = instruction.afterExecute || this.afterExecute

    buildCommand = async (flags, args): Promise<TerraCommand> => {
      const input = await instruction.makeInput(flags, args)
      if (!instruction.validateInput(input)) {
        throw new Error(`Invalid input params:  ${JSON.stringify(input)}`)
      }
      const query: Query = this.provider.wasm.contractQuery.bind(this.provider.wasm)
      const contractInput = await instruction.makeContractInput(input)
      const beforeExecutionContext: BeforeExecutionContext<Input, ContractInput> = {
        input,
        contractInput,
        id,
        contract: this.args[0],
        query,
        flags,
      }
      this.beforeExecute = instruction.beforeExecute
        ? instruction.beforeExecute(beforeExecutionContext)
        : this.defaultBeforeExecute(beforeExecutionContext)

      const abstractCommand = await makeAbstractCommand(id, this.flags, this.args, contractInput)
      await abstractCommand.invokeMiddlewares(abstractCommand, abstractCommand.middlewares)
      this.command = abstractCommand

      return this
    }

    makeRawTransaction = async (signer: AccAddress) => {
      return this.command.makeRawTransaction(signer)
    }

    execute = async (): Promise<Result<TransactionResponse>> => {
      // TODO: Command should be built from gauntet-core
      await this.buildCommand(this.flags, this.args)
      //
      await this.beforeExecute(this.wallet.key.accAddress)
      let response = await this.command.execute()
      const data = this.afterExecute(response)
      if (data) {
        response = { ...response, data: { ...data } }
      }
      return response
    }
  }
}
