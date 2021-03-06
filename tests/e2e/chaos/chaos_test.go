package chaos

import (
	"time"

	"github.com/smartcontractkit/chainlink-terra/tests/e2e/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/smartcontractkit/chainlink-terra/tests/e2e/smoke/common"
	"github.com/smartcontractkit/chainlink-testing-framework/actions"
)

var _ = Describe("Terra chaos suite", func() {
	var state = common.NewOCRv2State(1, 5)
	BeforeEach(func() {
		By("Deploying OCRv2 cluster", func() {
			state.DeployCluster(5, "2s", true, utils.ContractsDir)
			state.LabelChaosGroups()
			state.SetAllAdapterResponsesToTheSameValue(2)
		})
	})
	It("Can tolerate chaos experiments", func() {
		By("Stable and working", func() {
			state.ValidateAllRounds(time.Now(), common.NewRoundCheckTimeout, 10, false)
		})
		By("Can work with faulty nodes offline", func() {
			state.CanWorkWithFaultyNodesOffline()
		})
		By("Can't work with two parts network split, restored after", func() {
			state.RestoredAfterNetworkSplit()
		})
		By("Can recover from yellow group loss connection to validator", func() {
			state.CanWorkYellowGroupNoValidatorConnection()
		})
		By("Can recover after all nodes lost connection to validator", func() {
			state.CanRecoverAllNodesValidatorConnectionLoss()
		})
		By("Can work after all nodes restarted", func() {
			state.CanWorkAfterAllOraclesIPChange()
		})
		By("Can work when bootstrap migrated", func() {
			state.CanMigrateBootstrap()
		})
	})
	AfterEach(func() {
		By("Tearing down the environment", func() {
			err := actions.TeardownSuite(state.Env, nil, "logs", nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
