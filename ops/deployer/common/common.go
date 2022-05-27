package common

import (
	"time"
)

type OCRinit struct {
	LinkToken                 string `json:"linkToken"`
	MinAnswer                 string `json:"minAnswer"`
	MaxAnswer                 string `json:"maxAnswer"`
	BillingAccessController   string `json:"billingAccessController"`
	RequesterAccessController string `json:"requesterAccessController"`
	Decimals                  int    `json:"decimals"`
	Description               string `json:"description"`
}

const BeginProposal = "begin_proposal"

type ProposeConfigDetails struct {
	ID            string   `json:"proposalId"`
	Payees        []string `json:"payees"`
	Signers       []string `json:"signers"`
	Transmitters  []string `json:"transmitters"`
	F             uint8    `json:"f"`
	OnchainConfig []byte   `json:"onchainConfig"`
}

type ProposeOffchainConfigDetails struct {
	ID                    string                `json:"proposalId"`
	OffchainConfigVersion uint64                `json:"offchainConfigVersion"`
	OffchainConfig        OffchainConfigDetails `json:"offchainConfig"`
}

type ReportingPluginConfig struct {
	AlphaReportInfinite bool          `json:"alphaReportInfinite"`
	AlphaReportPpb      uint64        `json:"alphaReportPpb"`
	AlphaAcceptInfinite bool          `json:"alphaAcceptInfinite"`
	AlphaAcceptPpb      uint64        `json:"alphaAcceptPpb"`
	DeltaCNanoseconds   time.Duration `json:"deltaCNanoseconds"`
}

type OffchainConfigDetails struct {
	DeltaProgressNanoseconds                           time.Duration         `json:"deltaProgressNanoseconds"`
	DeltaResendNanoseconds                             time.Duration         `json:"deltaResendNanoseconds"`
	DeltaRoundNanoseconds                              time.Duration         `json:"deltaRoundNanoseconds"`
	DeltaGraceNanoseconds                              time.Duration         `json:"deltaGraceNanoseconds"`
	DeltaStageNanoseconds                              time.Duration         `json:"deltaStageNanoseconds"`
	RMax                                               uint64                `json:"rMax"`
	S                                                  []int                 `json:"s"`
	OffchainPublicKeys                                 []string              `json:"offchainPublicKeys"`
	PeerIDs                                            []string              `json:"peerIds"`
	ReportingPluginConfig                              ReportingPluginConfig `json:"reportingPluginConfig"`
	MaxDurationQueryNanoseconds                        time.Duration         `json:"maxDurationQueryNanoseconds"`
	MaxDurationObservationNanoseconds                  time.Duration         `json:"maxDurationObservationNanoseconds"`
	MaxDurationReportNanoseconds                       time.Duration         `json:"maxDurationReportNanoseconds"`
	MaxDurationShouldAcceptFinalizedReportNanoseconds  time.Duration         `json:"maxDurationShouldAcceptFinalizedReportNanoseconds"`
	MaxDurationShouldTransmitAcceptedReportNanoseconds time.Duration         `json:"maxDurationShouldTransmitAcceptedReportNanoseconds"`
	ConfigPublicKeys                                   []string              `json:"configPublicKeys"`
}

type AcceptProposalDetails struct {
	ID     string `json:"proposalId"`
	Digest string `json:"digest"`
	Secret string `json:"secret"`
}

type ClearProposalDetails struct {
	ID     string `json:"proposalId"`
}

type FinalizeProposalDetails struct {
	ID     string `json:"proposalId"`
}

type Balance struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

type LINKinit struct {
	Name            string      `json:"name"`
	Symbol          string      `json:"symbol"`
	Decimals        int         `json:"decimals"`
	InitialBalances []Balance   `json:"initial_balances"`
	Mint            interface{} `json:"mint"`
	Marketing       interface{} `json:"marketing"`
}

type Send struct {
	Send SendDetails `json:"send"`
}

type SendDetails struct {
	Contract string `json:"contract"`
	Amount   string `json:"amount"`
	Msg      string `json:"msg"`
}
