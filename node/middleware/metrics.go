package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/yuriykis/microblocknet/common/proto"
)

// MetricsMiddleware is a middleware that collects metrics about node's
// it wraps the NodeServer interface
type MetricsMiddleware struct {
	handshakeCount      prometheus.Counter
	handshakeLatency    prometheus.Histogram
	handshakeErrorCount prometheus.Counter

	newTransactionCount   prometheus.Counter
	newTransactionLatency prometheus.Histogram
	newTransactionError   prometheus.Counter

	newBlockCount   prometheus.Counter
	newBlockLatency prometheus.Histogram
	newBlockError   prometheus.Counter

	getBlocksCount   prometheus.Counter
	getBlocksLatency prometheus.Histogram
	getBlocksError   prometheus.Counter

	next NodeServer
}

func NewMetricsMiddleware(next NodeServer) NodeServer {
	handshakeCount := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("handshake_count_%s", next),
		Help: "Number of handshakes",
	})
	handshakeLatency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    fmt.Sprintf("handshake_latency_%s", next),
		Help:    "Latency of handshakes",
		Buckets: prometheus.LinearBuckets(0, 1, 10),
	},
	)
	handshakeErrorCount := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("handshake_error_%s", next),
		Help: "Number of handshake errors",
	},
	)

	newTransactionCount := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("new_transaction_count_%s", next),
		Help: "Number of new transactions",
	},
	)
	newTransactionLatency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    fmt.Sprintf("new_transaction_latency_%s", next),
		Help:    "Latency of new transactions",
		Buckets: prometheus.LinearBuckets(0, 1, 10),
	},
	)
	newTransactionError := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("new_transaction_error_%s", next),
		Help: "Number of new transaction errors",
	},
	)

	newBlockCount := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("new_block_count_%s", next),
		Help: "Number of new blocks",
	},
	)
	newBlockLatency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    fmt.Sprintf("new_block_latency_%s", next),
		Help:    "Latency of new blocks",
		Buckets: prometheus.LinearBuckets(0, 1, 10),
	},
	)
	newBlockError := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("new_block_error_%s", next),
		Help: "Number of new block errors",
	},
	)

	getBlocksCount := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("get_blocks_count_%s", next),
		Help: "Number of get blocks",
	},
	)
	getBlocksLatency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    fmt.Sprintf("get_blocks_latency_%s", next),
		Help:    "Latency of get blocks",
		Buckets: prometheus.LinearBuckets(0, 1, 10),
	},
	)
	getBlocksError := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("get_blocks_error_%s", next),
		Help: "Number of get blocks errors",
	},
	)

	prometheus.MustRegister(handshakeCount)
	prometheus.MustRegister(handshakeLatency)
	prometheus.MustRegister(handshakeErrorCount)

	prometheus.MustRegister(newTransactionCount)
	prometheus.MustRegister(newTransactionLatency)
	prometheus.MustRegister(newTransactionError)

	prometheus.MustRegister(newBlockCount)
	prometheus.MustRegister(newBlockLatency)
	prometheus.MustRegister(newBlockError)

	prometheus.MustRegister(getBlocksCount)
	prometheus.MustRegister(getBlocksLatency)
	prometheus.MustRegister(getBlocksError)

	return &MetricsMiddleware{
		handshakeCount:      handshakeCount,
		handshakeLatency:    handshakeLatency,
		handshakeErrorCount: handshakeErrorCount,

		newTransactionCount:   newTransactionCount,
		newTransactionLatency: newTransactionLatency,
		newTransactionError:   newTransactionError,

		newBlockCount:   newBlockCount,
		newBlockLatency: newBlockLatency,
		newBlockError:   newBlockError,

		getBlocksCount:   getBlocksCount,
		getBlocksLatency: getBlocksLatency,
		getBlocksError:   getBlocksError,

		next: next,
	}
}

func (m *MetricsMiddleware) Handshake(ctx context.Context, v *proto.Version) (_ *proto.Version, err error) {
	defer func(begin time.Time) {
		m.handshakeCount.Inc()
		m.handshakeLatency.Observe(time.Since(begin).Seconds())
		if err != nil {
			m.handshakeErrorCount.Inc()
		}
	}(time.Now())
	return m.next.Handshake(ctx, v)
}

func (m *MetricsMiddleware) NewTransaction(
	ctx context.Context,
	t *proto.Transaction,
) (_ *proto.Transaction, err error) {
	defer func(begin time.Time) {
		m.newTransactionCount.Inc()
		m.newTransactionLatency.Observe(time.Since(begin).Seconds())
		if err != nil {
			m.newTransactionError.Inc()
		}
	}(time.Now())
	return m.next.NewTransaction(ctx, t)
}

func (m *MetricsMiddleware) NewBlock(ctx context.Context, b *proto.Block) (_ *proto.Block, err error) {
	defer func(begin time.Time) {
		m.newBlockCount.Inc()
		m.newBlockLatency.Observe(time.Since(begin).Seconds())
		if err != nil {
			m.newBlockError.Inc()
		}
	}(time.Now())
	return m.next.NewBlock(ctx, b)
}

func (m *MetricsMiddleware) GetBlocks(ctx context.Context, v *proto.Version) (_ *proto.Blocks, err error) {
	defer func(begin time.Time) {
		m.getBlocksCount.Inc()
		m.getBlocksLatency.Observe(time.Since(begin).Seconds())
		if err != nil {
			m.getBlocksError.Inc()
		}
	}(time.Now())
	return m.next.GetBlocks(ctx, v)
}

func (m *MetricsMiddleware) String() string {
	return fmt.Sprintf("metrics(%s)", m.next)
}
