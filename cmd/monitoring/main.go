package main

import (
	"context"

	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"github.com/smartcontractkit/chainlink-terra/pkg/monitoring"
	pkgClient "github.com/smartcontractkit/chainlink-terra/pkg/terra/client"
	"github.com/smartcontractkit/chainlink/core/logger"
)

func main() {
	ctx := context.Background()

	coreLog := logger.NewLogger()
	log := logWrapper{coreLog}

	terraConfig, err := monitoring.ParseTerraConfig()
	if err != nil {
		log.Fatalw("failed to parse terra specific configuration", "error", err)
		return
	}

	client, err := pkgClient.NewClientWithGRPCTransport(
		terraConfig.ChainID,
		terraConfig.GRPCAddr,
		terraConfig.GRPCAPIKey,
		coreLog,
	)
	if err != nil {
		log.Fatalw("failed to create a terra client", "error", err)
		return
	}

	envelopeSourceFactory := monitoring.NewEnvelopeSourceFactory(
		client,
		log.With("component", "source-envelope"),
	)
	txResultsFactory := monitoring.NewTxResultsSourceFactory(
		log.With("component", "source-txresults"),
	)

	entrypoint, err := relayMonitoring.NewEntrypoint(
		ctx,
		log,
		terraConfig,
		envelopeSourceFactory,
		txResultsFactory,
		monitoring.TerraFeedParser,
	)
	if err != nil {
		log.Fatalw("failed to build entrypoint", "error", err)
		return
	}

	proxySourceFactory := monitoring.NewProxySourceFactory(
		client,
		log.With("component", "source-proxy"),
	)
	if entrypoint.Config.Feature.TestOnlyFakeReaders {
		proxySourceFactory = monitoring.NewFakeProxySourceFactory(log.With("component", "fake-proxy-source"))
	}
	entrypoint.SourceFactories = append(entrypoint.SourceFactories, proxySourceFactory)

	prometheusExporterFactory := monitoring.NewPrometheusExporterFactory(
		log.With("component", "terra-prometheus-exporter"),
		monitoring.NewMetrics(log.With("component", "terra-metrics")),
	)
	entrypoint.ExporterFactories = append(entrypoint.ExporterFactories, prometheusExporterFactory)

	entrypoint.Run()
	log.Info("monitor stopped")
}

// adapt core/logger.Logger to monitoring logger.

type logWrapper struct {
	logger.Logger
}

func (l logWrapper) With(values ...interface{}) relayMonitoring.Logger {
	return logWrapper{l.Logger.With(values...)}
}
