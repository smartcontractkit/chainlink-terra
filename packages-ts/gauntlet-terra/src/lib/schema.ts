import Ajv from 'ajv'
import JTD from 'ajv/dist/jtd'
import { JSONSchemaType } from 'ajv'
import { logger } from '@chainlink/gauntlet-core/dist/utils'

const ajv = new Ajv().addFormat('uint8', (value: any) => !isNaN(value))

ajv.addFormat('uint64', {
  type: 'number',
  validate: (x) => !isNaN(x),
})

ajv.addFormat('uint32', {
  type: 'number',
  validate: (x) => !isNaN(x),
})

export enum TERRA_OPERATIONS {
  DEPLOY = 'instantiate',
  EXECUTE = 'execute',
  QUERY = 'query',
}

export type TerraABI = {
  [TERRA_OPERATIONS.DEPLOY]: JSONSchemaType<any>
  [TERRA_OPERATIONS.EXECUTE]: JSONSchemaType<any>
  [TERRA_OPERATIONS.QUERY]: JSONSchemaType<any>
}

export const isValidFunction = (abi: TerraABI, functionName: string): boolean => {
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

export const isQueryFunction = (abi: TerraABI, functionName: string) => {
  const functionList = abi.query.oneOf || abi.query.anyOf
  return functionList.find((queryAbi: any) => {
    if (queryAbi.enum) return queryAbi.enum.includes(functionName)
    if (queryAbi.required) return queryAbi.required.includes(functionName)
    return false
  })
}

export default ajv

const jtd = new JTD()
export { jtd }
