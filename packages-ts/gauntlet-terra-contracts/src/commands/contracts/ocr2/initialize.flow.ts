import { FlowCommand } from '@chainlink/gauntlet-core'
import { CATEGORIES } from '../../../lib/constants'
import { waitExecute, TransactionResponse } from '@chainlink/gauntlet-terra'
import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import DeployOCR2 from './deploy'
import SetBilling from './setBilling'
import ProposeConfig from './proposeConfig'
import ProposeOffchainConfig from './proposeOffchainConfig'
import BeginProposal from './proposal/beginProposal'
import AcceptProposal from './proposal/acceptProposal'
import FinalizeProposal from './proposal/finalizeProposal'
import Inspect from './inspection/inspect'
import { abstract } from '../..'

export default class OCR2InitializeFlow extends FlowCommand<TransactionResponse> {
  static id = 'ocr2:initialize:flow'
  static category = CATEGORIES.OCR
  static examples = ['yarn gauntlet ocr2:initialize:flow --network=local --id=[ID] --rdd=[PATH_TO_RDD]']

  constructor(flags, args) {
    super(flags, args, waitExecute, abstract.makeAbstractCommand)

    this.stepIds = {
      OCR_2: 1,
      BEGIN_PROPOSAL: 2,
      FINALIZE_PROPOSAL: 3,
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
        id: this.stepIds.BEGIN_PROPOSAL,
        name: 'Begin Proposal',
        command: BeginProposal,
        args: [this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.OCR_2))],
      },
      {
        name: 'Propose Config',
        command: ProposeConfig,
        flags: {
          proposalId: this.getReportStepDataById(FlowCommand.ID.data(this.stepIds.BEGIN_PROPOSAL, 'proposalId')),
        },
        args: [this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.OCR_2))],
      },
      {
        name: 'Propose Offchain Config',
        command: ProposeOffchainConfig,
        flags: {
          proposalId: this.getReportStepDataById(FlowCommand.ID.data(this.stepIds.BEGIN_PROPOSAL, 'proposalId')),
        },
        args: [this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.OCR_2))],
      },
      {
        id: this.stepIds.FINALIZE_PROPOSAL,
        name: 'Finalize Proposal',
        command: FinalizeProposal,
        flags: {
          proposalId: this.getReportStepDataById(FlowCommand.ID.data(this.stepIds.BEGIN_PROPOSAL, 'proposalId')),
        },
        args: [this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.OCR_2))],
      },
      {
        name: 'Accept Proposal',
        command: AcceptProposal,
        flags: {
          proposalId: this.getReportStepDataById(FlowCommand.ID.data(this.stepIds.BEGIN_PROPOSAL, 'proposalId')),
          digest: this.getReportStepDataById(FlowCommand.ID.data(this.stepIds.FINALIZE_PROPOSAL, 'digest')),
        },
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
