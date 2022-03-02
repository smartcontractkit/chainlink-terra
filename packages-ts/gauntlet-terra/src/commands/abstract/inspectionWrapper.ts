import { AbstractTools } from '.'
import { Result } from '@chainlink/gauntlet-core'
import { TerraCommand, TransactionResponse } from '../..'
import { logger } from '@chainlink/gauntlet-core/dist/utils'
import { Contract } from '../../lib/contracts'
import { APIParams } from '@terra-money/terra.js/dist/client/lcd/APIRequester'

export type Query = (contractAddress: string, query: any, params?: APIParams) => Promise<any>

/**
 * Inspection commands need to match this interface
 * command: {
 *   contract: Contract related to the inspection
 *   id: Name of the command the user will execute
 * }
 * instructions: instruction[] Set of abstract query commands the inspection command will run
 * makeInput: Receives flags and args. Should return the input the underneath commands
 * makeInspectionInput: Transforms input into a comparable format
 * makeOnchainData: Parses every instruction command result to match the same interface the Inspection command expects
 * inspect: Compares both expected and onchain data.
 */
export interface InspectInstructionTemplate<CommandInput, ContractExpectedInfo, ContractList extends string> {
  command: {
    category: string // CATEGORIES
    contract: ContractList
    id: 'inspect'
  }
  instructions: {
    contract: string
    function: string
  }[]
  makeInput: (flags: any, args: string[]) => Promise<ContractExpectedInfo>
  makeInspectionData: (query: Query) => (input: ContractExpectedInfo) => Promise<ContractExpectedInfo>
  makeOnchainData: (query: Query) => (instructionsData: any[]) => ContractExpectedInfo
  inspect: (expected: ContractExpectedInfo, data: ContractExpectedInfo) => boolean
}

export const instructionToInspectCommand = <CommandInput, ContractExpectedInfo, ContractList extends string>(
  abstract: AbstractTools<ContractList>,
  inspectInstruction: InspectInstructionTemplate<CommandInput, ContractExpectedInfo, any>,
) => {
  const id = `${inspectInstruction.command.contract}:${inspectInstruction.command.id}`
  return class Command extends TerraCommand {
    static id = id
    static abstract: AbstractTools<ContractList> = abstract

    constructor(flags, args) {
      super(flags, args)
    }

    makeRawTransaction = () => {
      throw new Error('Inspection command does not involve any transaction')
    }

    execute = async (): Promise<Result<TransactionResponse>> => {
      const input = await inspectInstruction.makeInput(this.flags, this.args)
      const commands = await Promise.all(
        inspectInstruction.instructions.map((instruction) =>
          Command.abstract.makeAbstractCommand(
            `${instruction.contract}:${instruction.function}`,
            this.flags,
            this.args,
            input,
          ),
        ),
      )

      logger.loading('Fetching contract information...')
      const data = await Promise.all(
        commands.map(async (command) => {
          await command.invokeMiddlewares(command, command.middlewares)
          const { data } = await command.execute()
          return data
        }),
      )

      const query: Query = this.provider.wasm.contractQuery.bind(this.provider.wasm)
      const onchainData = inspectInstruction.makeOnchainData(query)(data)
      const inspectData = await inspectInstruction.makeInspectionData(query)(input)
      const inspection = inspectInstruction.inspect(inspectData, onchainData)
      return {
        data: inspection,
        responses: [
          {
            tx: {
              hash: '',
              wait: () => ({ success: inspection }),
            },
            contract: this.args[0],
          },
        ],
      }
    }
  }
}
