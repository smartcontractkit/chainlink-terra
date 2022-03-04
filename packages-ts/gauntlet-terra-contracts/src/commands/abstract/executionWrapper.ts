import AbstractCommand, { makeAbstractCommand } from '.'
import { Result } from '@chainlink/gauntlet-core'
import { TerraCommand, TransactionResponse } from '@chainlink/gauntlet-terra'
import { AccAddress, MsgExecuteContract } from '@terra-money/terra.js'

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
  afterExecute?: (response: Result<TransactionResponse>) => any
}

export const instructionToCommand = (instruction: AbstractInstruction<any, any>) => {
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

    afterExecute = instruction.afterExecute

    buildCommand = async (): Promise<TerraCommand> => {
      const commandInput = await instruction.makeInput(this.flags, this.args)
      if (!instruction.validateInput(commandInput)) {
        throw new Error(`Invalid input params:  ${JSON.stringify(commandInput)}`)
      }
      const input = await instruction.makeContractInput(commandInput)
      const abstractCommand = await makeAbstractCommand(id, this.flags, this.args, input)
      await abstractCommand.invokeMiddlewares(abstractCommand, abstractCommand.middlewares)
      return abstractCommand
    }

    makeRawTransaction = async (signer: AccAddress): Promise<MsgExecuteContract[]> => {
      const command = await this.buildCommand()
      return command.makeRawTransaction(signer)
    }

    execute = async (): Promise<Result<TransactionResponse>> => {
      const command = await this.buildCommand()
      let response = await command.execute()
      if (this.afterExecute) {
        const data = this.afterExecute(response)
        response = { ...response, data: { ...data } }
      }
      return response
    }
  }
}
