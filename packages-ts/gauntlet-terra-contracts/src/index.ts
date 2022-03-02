import { executeCLI } from '@chainlink/gauntlet-core'
import { multisigWrapCommand, commands as CWPlusCommands } from '@chainlink/gauntlet-terra-cw-plus'
import { existsSync } from 'fs'
import path from 'path'
import { io } from '@chainlink/gauntlet-core/dist/utils'
import { CONTRACT_LIST } from './lib/contracts'
import { abstract } from './commands'
import { defaultFlags } from './lib/args'
import Terra from './commands'

const commands = {
  custom: [...Terra, ...Terra.map(multisigWrapCommand), ...CWPlusCommands],
  loadDefaultFlags: () => defaultFlags,
  abstract: {
    findPolymorphic: () => undefined,
    makeCommand: abstract.makeAbstractCommand,
  },
}

;(async () => {
  try {
    const networkPossiblePaths = ['./packages-ts/gauntlet-terra-contracts/networks']
    const networkPath = networkPossiblePaths.filter((networkPath) =>
      existsSync(path.join(process.cwd(), networkPath)),
    )[0]
    const result = await executeCLI(commands, networkPath)
    if (result) {
      io.saveJSON(result, process.env['REPORT_NAME'] ? process.env['REPORT_NAME'] : 'report')
    }
  } catch (e) {
    console.log(e)
    console.log('Terra Command execution error', e.message)
    process.exitCode = 1
  }
})()
