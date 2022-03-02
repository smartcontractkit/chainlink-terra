import { io, logger } from '@chainlink/gauntlet-core/dist/utils'
import { existsSync, readFileSync } from 'fs'
import path from 'path'
import fetch from 'node-fetch'
import { DEFAULT_RELEASE_VERSION, DEFAULT_CWPLUS_VERSION } from './constants'
import { Contract } from '@chainlink/gauntlet-terra'
import { TerraABI } from '@chainlink/gauntlet-terra/dist/lib/schema'

export enum CONTRACT_LIST {
  FLAGS = 'flags',
  DEVIATION_FLAGGING_VALIDATOR = 'deviation_flagging_validator',
  OCR_2 = 'ocr2',
  PROXY_OCR_2 = 'proxy_ocr2',
  ACCESS_CONTROLLER = 'access_controller',
  CW20_BASE = 'cw20_base',
  MULTISIG = 'cw3_flex_multisig',
  CW4_GROUP = 'cw4_group',
}

export const getContractCode = async (contractId: CONTRACT_LIST, version): Promise<string> => {
  if (version === 'local') {
    // Possible paths depending on how/where gauntlet is being executed
    const possibleContractPaths = [
      path.join(__dirname, '../../artifacts/bin'),
      path.join(process.cwd(), './artifacts/bin'),
      path.join(process.cwd(), './tests/e2e/common_artifacts'),
      path.join(process.cwd(), './packages-ts/gauntlet-terra-contracts/artifacts/bin'),
    ]

    const codes = possibleContractPaths
      .filter((contractPath) => existsSync(`${contractPath}/${contractId}.wasm`))
      .map((contractPath) => {
        const wasm = readFileSync(`${contractPath}/${contractId}.wasm`)
        return wasm.toString('base64')
      })
    return codes[0]
  } else {
    let url
    switch (contractId) {
      case CONTRACT_LIST.CW20_BASE:
      case CONTRACT_LIST.CW4_GROUP:
      case CONTRACT_LIST.MULTISIG:
        url = `https://github.com/CosmWasm/cw-plus/releases/download/${version}/${contractId}.wasm`
        break
      default:
        url = `https://github.com/smartcontractkit/chainlink-terra/releases/download/${version}/${contractId}.wasm`
    }
    logger.loading(`Fetching ${url}...`)
    const response = await fetch(url)
    const body = await response.arrayBuffer()
    return Buffer.from(body).toString('base64')
  }
}

const contractDirName = {
  [CONTRACT_LIST.FLAGS]: 'flags',
  [CONTRACT_LIST.DEVIATION_FLAGGING_VALIDATOR]: 'deviation-flagging-validator',
  [CONTRACT_LIST.OCR_2]: 'ocr2',
  [CONTRACT_LIST.PROXY_OCR_2]: 'proxy-ocr2',
  [CONTRACT_LIST.ACCESS_CONTROLLER]: 'access-controller',
  [CONTRACT_LIST.CW20_BASE]: 'cw20_base',
  [CONTRACT_LIST.CW4_GROUP]: 'cw4_group',
  [CONTRACT_LIST.MULTISIG]: 'cw3_flex_multisig',
}

const defaultContractVersions = {
  [CONTRACT_LIST.FLAGS]: DEFAULT_RELEASE_VERSION,
  [CONTRACT_LIST.DEVIATION_FLAGGING_VALIDATOR]: DEFAULT_RELEASE_VERSION,
  [CONTRACT_LIST.OCR_2]: DEFAULT_RELEASE_VERSION,
  [CONTRACT_LIST.ACCESS_CONTROLLER]: DEFAULT_RELEASE_VERSION,
  [CONTRACT_LIST.CW20_BASE]: DEFAULT_CWPLUS_VERSION,
  [CONTRACT_LIST.CW4_GROUP]: DEFAULT_CWPLUS_VERSION,
  [CONTRACT_LIST.MULTISIG]: DEFAULT_CWPLUS_VERSION,
}
export const getContractABI = (contractId: CONTRACT_LIST): TerraABI => {
  // Possible paths depending on how/where gauntlet is being executed
  const possibleContractPaths = [
    path.join(__dirname, './artifacts/contracts'),
    path.join(process.cwd(), './contracts'),
    path.join(process.cwd(), '../../contracts'),
    path.join(process.cwd(), './packages-ts/gauntlet-terra-contracts/artifacts/contracts'),
  ]

  const toDirName = (contractId: CONTRACT_LIST) => contractDirName[contractId]

  const abi = possibleContractPaths
    .filter((path) => existsSync(`${path}/${toDirName(contractId)}/schema`))
    .map((contractPath) => {
      const toPath = (type) => {
        if (contractId == CONTRACT_LIST.CW20_BASE && type == 'execute_msg') {
          return path.join(contractPath, `./${toDirName(contractId)}/schema/cw20_${type}`)
        } else {
          return path.join(contractPath, `./${toDirName(contractId)}/schema/${type}`)
        }
      }

      return {
        execute: io.readJSON(toPath('execute_msg')),
        query: io.readJSON(toPath('query_msg')),
        instantiate: io.readJSON(toPath('instantiate_msg')),
      }
    })
  if (abi.length === 0) {
    logger.error(`ABI not found for contract ${contractId}`)
  }

  return abi[0]
}

export const getContract = async (contractName: CONTRACT_LIST, version): Promise<Contract> => {
  // First convert contactName (a string restricted to the values of CONTRACT_LIST) to an actual member of
  //  the enum; typescript sees both as the same, but conversion must happen at runtime since javascript
  //  treats them as two completely different types
  const id: CONTRACT_LIST = CONTRACT_LIST[contractName]
  version = version ? version : defaultContractVersions[id]
  // Preload contracts
  return {
    id: id,
    abi: getContractABI(id),
    bytecode: await getContractCode(id, version),
  }
}
