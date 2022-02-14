import { FlowCommand } from '@chainlink/gauntlet-core'
import { CATEGORIES } from '../../../lib/constants'
import { waitExecute, TransactionResponse } from '@chainlink/gauntlet-terra'

import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import { makeAbstractCommand } from '../../abstract'
import DeployOCR2 from './deploy'
import SetBilling from './setBilling'
import SetConfig from './setConfig'
import SetPayees from './setPayees'
import Inspect from './inspection/inspect'

export default class OCR2InitializeFlow extends FlowCommand<TransactionResponse> {
  static id = 'ocr2:initialize:flow'
  static category = CATEGORIES.OCR
  static examples = ['yarn gauntlet ocr2:initialize:flow --network=local --id=[ID] --rdd=[PATH_TO_RDD]']

  constructor(flags, args) {
    super(flags, args, waitExecute, makeAbstractCommand)

    this.stepIds = {
      OCR_2: 1,
    }

    this.flow = [
      {
        name: 'Deploy OCR 2',
        command: DeployOCR2,
        id: this.stepIds.OCR_2,
      },
      {
        name: 'Change RDD',
        exec: this.showRddInstructions,
      },
      {
        name: 'Set Billing',
        command: SetBilling,
        args: [this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.OCR_2))],
      },
      {
        name: 'Set Config',
        command: SetConfig,
        args: [this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.OCR_2))],
      },
      {
        name: 'Set Payees',
        command: SetPayees,
        args: [this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.OCR_2))],
      },
      // Inspection here
      {
        name: 'Inspection',
        command: Inspect,
        args: [this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.OCR_2))],
      },
    ]
  }

  showRddInstructions = async () => {
    logger.info(
      `
        Change the RDD ID with the new contract address: 
          - Contract Address: ${this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.OCR_2))}
      `,
    )

    await prompt('Ready? Continue')
  }
}
