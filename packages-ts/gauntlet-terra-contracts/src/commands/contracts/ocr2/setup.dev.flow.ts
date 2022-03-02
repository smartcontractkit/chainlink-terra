import { FlowCommand } from '@chainlink/gauntlet-core'
import { CATEGORIES } from '../../../lib/constants'
import { waitExecute, TransactionResponse } from '@chainlink/gauntlet-terra'
import { MnemonicKey } from '@terra-money/terra.js'
import UploadContractCode from '../../tooling/upload'
import DeployLink from '../link/deploy'
import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import DeployOCR2 from './deploy'
import SetBilling from './setBilling'
import ProposeConfig from './proposeConfig'
import ProposeOffchainConfig from './proposeOffchainConfig'
import BeginProposal from './proposal/beginProposal'
import FinalizeProposal from './proposal/finalizeProposal'
import AcceptProposal from './proposal/acceptProposal'
import { abstract } from '../..'

export default class DeployOCR2Flow extends FlowCommand<TransactionResponse> {
  static id = 'ocr2:setup:flow'
  static category = CATEGORIES.OCR

  constructor(flags, args) {
    super(flags, args, waitExecute, abstract.makeAbstractCommand)

    const oraclesLength = this.flags.oracles || 16

    const oracles = new Array(oraclesLength).fill('').map((_, i) => ({
      offchainPublicKey: '5cd10bf991c8b0db7bee3ec371c7795a69297b6bccf7b4d738e0920b56131772',
      peerId: 'DxRwKpwNBuMzKf5YEG1vLpnRbWeKo1Z4tKHfFGt8vUkj',
      configPublicKey: '5cd10bf991c8b0db7bee3ec371c7795a69297b6bccf7b4d738e0920b56131772',
      signer: new Array(64).fill(i.toString(16)).join(''),
      payee: new MnemonicKey().publicKey?.address(),
      transmitter: new MnemonicKey().publicKey?.address(),
    }))

    this.stepIds = {
      BILLING_ACCESS_CONTROLLER: 1,
      REQUEST_ACCESS_CONTROLLER: 2,
      LINK: 3,
      OCR_2: 4,
      BEGIN_PROPOSAL: 5,
      FINALIZE_PROPOSAL: 6,
    }

    const billingInput = {
      recommendedGasPriceMicro: '1',
      observationPaymentGjuels: 1,
      transmissionPaymentGjuels: 1,
    }

    const offchainConfigInput = {
      f: 1,
      deltaProgressNanoseconds: 300000000,
      deltaResendNanoseconds: 300000000,
      deltaRoundNanoseconds: 30,
      deltaGraceNanoseconds: 30,
      deltaStageNanoseconds: 30,
      rMax: 30,
      s: oracles.map(() => 1),
      offchainPublicKeys: oracles.map((o) => o.offchainPublicKey),
      peerIds: oracles.map((o) => o.peerId),
      reportingPluginConfig: {
        alphaReportInfinite: false,
        alphaReportPpb: 0,
        alphaAcceptInfinite: false,
        alphaAcceptPpb: 30,
        deltaCNanoseconds: 30,
      },
      maxDurationQueryNanoseconds: 30,
      maxDurationObservationNanoseconds: 30,
      maxDurationReportNanoseconds: 30,
      maxDurationShouldAcceptFinalizedReportNanoseconds: 30,
      maxDurationShouldTransmitAcceptedReportNanoseconds: 30,
      configPublicKeys: oracles.map((o) => o.configPublicKey),
    }

    const transmitters = oracles.map((o) => o.transmitter)
    const configInput = {
      signers: oracles.map((o) => o.signer),
      transmitters,
      offchainConfig: offchainConfigInput,
      offchainConfigVersion: 2,
      onchainConfig: '',
    }

    const payeesInput = {
      payees: oracles.map((o) => o.payee),
      transmitters,
    }

    this.flow = [
      // {
      //   name: 'Upload Contracts',
      //   command: UploadContractCode,
      // },
      // {
      //   name: 'Deploy LINK',
      //   command: DeployLink,
      //   id: this.stepIds.LINK,
      // },
      // {
      //   name: 'Deploy Billing Access Controller',
      //   command: 'access_controller:deploy',
      //   id: this.stepIds.BILLING_ACCESS_CONTROLLER,
      // },
      // {
      //   name: 'Deploy Request Access Controller',
      //   command: 'access_controller:deploy',
      //   id: this.stepIds.REQUEST_ACCESS_CONTROLLER,
      // },
      // {
      //   name: 'Set environment',
      //   exec: this.setEnvironment,
      // },
      {
        name: 'Deploy OCR 2',
        command: DeployOCR2,
        flags: {
          input: {
            billingAccessController: process.env.BILLING_ACCESS_CONTROLLER,
            requesterAccessController: process.env.REQUESTER_ACCESS_CONTROLLER,
            linkToken: process.env.LINK,
            decimals: 18,
            description: 'TEST',
            maxAnswer: '100000000000',
            minAnswer: '1',
          },
        },
        id: this.stepIds.OCR_2,
      },
      {
        name: 'Set Billing',
        command: SetBilling,
        flags: {
          input: billingInput,
        },
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
          input: {
            f: offchainConfigInput.f,
            proposalId: this.getReportStepDataById(FlowCommand.ID.data(this.stepIds.BEGIN_PROPOSAL, 'proposalId')),
            signers: configInput.signers,
            transmitters: configInput.transmitters,
            payees: payeesInput.payees,
            onchainConfig: '',
          },
        },
        args: [this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.OCR_2))],
      },
      {
        name: 'Propose Offchain Config',
        command: ProposeOffchainConfig,
        flags: {
          input: {
            proposalId: this.getReportStepDataById(FlowCommand.ID.data(this.stepIds.BEGIN_PROPOSAL, 'proposalId')),
            offchainConfigVersion: 2,
            offchainConfig: offchainConfigInput,
          },
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
    ]
  }

  setEnvironment = async () => {
    const linkAddress = this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.LINK))
    const billingAC = this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.BILLING_ACCESS_CONTROLLER))
    const requesterAC = this.getReportStepDataById(FlowCommand.ID.contract(this.stepIds.REQUEST_ACCESS_CONTROLLER))
    logger.info(`
      Setting the following env variables. Include them into .env.${this.flags.network} for future runs
        LINK=${linkAddress}
        BILLING_ACCESS_CONTROLLER=${billingAC}
        REQUESTER_ACCESS_CONTROLLER=${requesterAC}
      `)
    process.env.LINK = linkAddress
    process.env.BILLING_ACCESS_CONTROLLER = billingAC
    process.env.REQUESTER_ACCESS_CONTROLLER = requesterAC
    await prompt('Continue?')
  }
}
