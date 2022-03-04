import { Result } from '@chainlink/gauntlet-core'
import { AccAddress, MsgExecuteContract } from '@terra-money/terra.js'
import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import { TransactionResponse, TerraCommand } from '@chainlink/gauntlet-terra'
import { Contract, CONTRACT_LIST, getContract, TerraABI, TERRA_OPERATIONS } from '../../lib/contracts'
import { DEFAULT_RELEASE_VERSION } from '../../lib/constants'
import schema from '../../lib/schema'

export interface AbstractOpts {
  contract: Contract
  function: string
  action: TERRA_OPERATIONS.DEPLOY | TERRA_OPERATIONS.EXECUTE | TERRA_OPERATIONS.QUERY | 'help'
}

export interface AbstractParams {
  [param: string]: any
}

export const makeAbstractCommand = async (
  instruction: string,
  flags: any,
  args: string[],
  input?: any,
): Promise<TerraCommand> => {
  const commandOpts = await parseInstruction(instruction, flags.version)
  const params = parseParams(commandOpts, input || flags)
  return new AbstractCommand(flags, args, commandOpts, params)
}

export const parseInstruction = async (instruction: string, inputVersion: string): Promise<AbstractOpts> => {
  const isValidContract = (contractName: string): boolean => {
    // Validate that we have this contract available
    return Object.values(CONTRACT_LIST).includes(contractName as CONTRACT_LIST)
  }

  const isValidFunction = (abi: TerraABI, functionName: string): boolean => {
    // Check against ABI if method exists
    const availableFunctions = [
      ...(abi.query.oneOf || abi.query.anyOf || []),
      ...(abi.execute.oneOf || abi.execute.anyOf || []),
    ].reduce((agg, prop) => {
      if (prop?.required && prop.required.length > 0) return [...agg, ...prop.required]
      if (prop?.enum && prop.enum.length > 0) return [...agg, ...prop.enum]
      return [...agg]
    }, [])
    logger.debug(`Available functions on this contract: ${availableFunctions}`)
    return availableFunctions.includes(functionName)
  }

  const isQueryFunction = (abi: TerraABI, functionName: string) => {
    const functionList = abi.query.oneOf || abi.query.anyOf
    return functionList.find((queryAbi: any) => {
      if (queryAbi.enum) return queryAbi.enum.includes(functionName)
      if (queryAbi.required) return queryAbi.required.includes(functionName)
      return false
    })
  }

  const command = instruction.split(':')
  if (!command.length || command.length > 2) throw new Error(`Abstract: Contract ${command[0]} not found`)

  const version = inputVersion ? inputVersion : DEFAULT_RELEASE_VERSION
  const contractByteCode = await getContract(command[0] as CONTRACT_LIST, version)
  const contract = isValidContract(command[0]) && contractByteCode
  if (!contract) throw new Error(`Abstract: Contract ${command[0]} not found`)

  if (command[1] === 'help') {
    return {
      contract,
      function: 'help',
      action: 'help',
    }
  }

  if (command[1] === 'deploy') {
    return {
      contract,
      function: TERRA_OPERATIONS.DEPLOY,
      action: TERRA_OPERATIONS.DEPLOY,
    }
  }

  const functionName = isValidFunction(contract.abi, command[1]) && command[1]
  if (!functionName) throw new Error(`Abstract: Function ${command[1]} for contract ${contract.id} not found`)

  return {
    contract,
    function: functionName,
    action: isQueryFunction(contract.abi, functionName) ? TERRA_OPERATIONS.QUERY : TERRA_OPERATIONS.EXECUTE,
  }
}

export const parseParams = (commandOpts: AbstractOpts, params: any): AbstractParams => {
  if (commandOpts.action === 'help') return params
  const abiSchema = commandOpts.contract.abi[commandOpts.action]
  const validate = schema.compile(abiSchema)
  const schemas = [...(abiSchema.oneOf || []), ...(abiSchema.anyOf || [])]
  // isEnum means the function doesn't receive any parameter
  const isEnumData = schemas.some((subSchema) => !!subSchema.enum && subSchema.enum.includes(commandOpts.function))
  const data = isEnumData
    ? commandOpts.function
    : commandOpts.action === TERRA_OPERATIONS.DEPLOY
    ? params
    : {
        [commandOpts.function]: params,
      }

  if (!validate(data)) {
    logger.log(validate.errors)
    throw new Error(`Error validating parameters for function ${commandOpts.function}`)
  }

  return data
}

type AbstractExecute = (params: any, address?: string) => Promise<Result<TransactionResponse>>
export default class AbstractCommand extends TerraCommand {
  opts: AbstractOpts
  params: AbstractParams

  constructor(flags, args, opts, params) {
    super(flags, args)

    this.opts = opts
    this.params = params

    if ([...TERRA_OPERATIONS.EXECUTE, ...TERRA_OPERATIONS.QUERY].includes(this.opts.action)) {
      this.require(args[0], 'Provide a valid contract address')
    }

    this.contracts = [this.opts.contract.id]
  }

  makeRawTransaction = async (signer: AccAddress): Promise<MsgExecuteContract[]> => {
    const address = this.args[0]
    return [new MsgExecuteContract(signer, address, this.params)]
  }

  abstractDeploy: AbstractExecute = async (params: any) => {
    logger.loading(`Deploying contract ${this.opts.contract.id}`)
    const codeId = this.codeIds[this.opts.contract.id]
    this.require(!!codeId, `Code Id for contract ${this.opts.contract.id} not found`)
    const deploy = await this.deploy(codeId, params)
    logger.success(`Deployed ${this.opts.contract.id} to ${deploy.address}`)
    return {
      responses: [
        {
          tx: deploy,
          contract: deploy.address,
        },
      ],
    } as Result<TransactionResponse>
  }

  abstractExecute: AbstractExecute = async (params: any, address: string) => {
    logger.loading(`Executing ${this.opts.function} from contract ${this.opts.contract.id} at ${address}`)
    logger.log('Input Params:', params)
    await prompt(`Continue?`)
    const tx = await this.call(address, params)
    logger.success(`Execution finished at tx ${tx.hash}`)
    return {
      responses: [
        {
          tx,
          contract: address,
        },
      ],
    } as Result<TransactionResponse>
  }

  abstractQuery: AbstractExecute = async (params: any, address: string) => {
    logger.loading(`Calling ${this.opts.function} from contract ${this.opts.contract.id} at ${address}`)
    const result = await this.query(address, params)
    logger.debug(`Query finished with result: ${JSON.stringify(result)}`)
    return {
      data: result,
      responses: [
        {
          contract: address,
        },
      ],
    } as Result<TransactionResponse>
  }

  abstractHelp: AbstractExecute = async () => {
    // TODO: Get functions per operation with their own needed parameters. AJV doesn't offer a way to do this
    // const queryFunctions = getSchemaFunctions(this.opts.contract.abi.query)
    // const deployFunction = getSchemaFunctions(this.opts.contract.abi.instantiate)
    // const executeFunctions = getSchemaFunctions(this.opts.contract.abi.execute)
    return {
      responses: [],
    }
  }

  execute = async () => {
    const operations = {
      [TERRA_OPERATIONS.DEPLOY]: this.abstractDeploy,
      [TERRA_OPERATIONS.QUERY]: this.abstractQuery,
      [TERRA_OPERATIONS.EXECUTE]: this.abstractExecute,
      help: this.abstractHelp,
    }

    const address = this.args[0]
    return operations[this.opts.action](this.params, address)
  }
}
