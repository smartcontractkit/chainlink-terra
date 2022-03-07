// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Metrics is an autogenerated mock type for the Metrics type
type Metrics struct {
	mock.Mock
}

// Cleanup provides a mock function with given fields: proxyContractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName
func (_m *Metrics) Cleanup(proxyContractAddress string, feedID string, chainID string, contractStatus string, contractType string, feedName string, feedPath string, networkID string, networkName string) {
	_m.Called(proxyContractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName)
}

// SetProxyAnswers provides a mock function with given fields: answer, proxyContractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName
func (_m *Metrics) SetProxyAnswers(answer float64, proxyContractAddress string, feedID string, chainID string, contractStatus string, contractType string, feedName string, feedPath string, networkID string, networkName string) {
	_m.Called(answer, proxyContractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName)
}

// SetProxyAnswersRaw provides a mock function with given fields: answer, proxyContractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName
func (_m *Metrics) SetProxyAnswersRaw(answer float64, proxyContractAddress string, feedID string, chainID string, contractStatus string, contractType string, feedName string, feedPath string, networkID string, networkName string) {
	_m.Called(answer, proxyContractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName)
}
