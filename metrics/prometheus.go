package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

const (
	Namespace = "galaxia"

	UnauthenticatedRequestsCountMetric = "unauthenticated_requests_count"
	UserMessagesSentCountMetric        = "user_messages_sent_count"
	BotMessagesSentCountMetric         = "bot_messages_sent_count"
	RequestDurationBucketMetric        = "request_duration"

	CallbacksProcessedCountMetric   = "callbacks_processed_count"
	CmdExecutedCountMetric          = "cmd_executed_count"
	StageReachedCountMetric         = "stage_reached_count"
	StageActionProcessedCountMetric = "stage_action_processed_count"

	CallbackHandlerRefLabel = "callback_handler_ref"
	StageRefLabel           = "stage_ref"
	ActionRefLabel          = "action_ref"
	CmdRefLabel             = "cmd_ref"

	DefaultListen = ":9000"
)

type PrometheusExporter struct {
	Listen string

	reg *prometheus.Registry

	counters    map[string]prometheus.Counter
	counterVecs map[string]*prometheus.CounterVec
	hists       map[string]*prometheus.HistogramVec
}

func NewPrometheusExporter() *PrometheusExporter {
	p := &PrometheusExporter{
		Listen:      DefaultListen,
		reg:         prometheus.NewRegistry(),
		counters:    make(map[string]prometheus.Counter),
		counterVecs: make(map[string]*prometheus.CounterVec),
		hists:       make(map[string]*prometheus.HistogramVec),
	}
	unauthenticatedRequestsCount := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      UnauthenticatedRequestsCountMetric,
		Help:      "Number of unauthenticated requests",
	})
	p.reg.MustRegister(unauthenticatedRequestsCount)
	p.counters[UnauthenticatedRequestsCountMetric] = unauthenticatedRequestsCount

	userMessagesSentCount := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      UserMessagesSentCountMetric,
		Help:      "Number of user messages sent",
	})
	p.reg.MustRegister(userMessagesSentCount)
	p.counters[UserMessagesSentCountMetric] = userMessagesSentCount

	botMessagesSentCount := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      BotMessagesSentCountMetric,
		Help:      "Number of bot/messages sent",
	})
	p.reg.MustRegister(botMessagesSentCount)
	p.counters[BotMessagesSentCountMetric] = botMessagesSentCount

	requestDurationBucket := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: Namespace,
			Name:      RequestDurationBucketMetric,
			Help:      "Bucketed histogram of request latencies",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{ActionRefLabel},
	)
	p.reg.MustRegister(requestDurationBucket)
	p.hists[RequestDurationBucketMetric] = requestDurationBucket

	callbacksProcessedCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      CallbacksProcessedCountMetric,
			Help:      "Number of callbacks processed",
		},
		[]string{CallbackHandlerRefLabel},
	)
	p.reg.MustRegister(callbacksProcessedCount)
	p.counterVecs[CallbacksProcessedCountMetric] = callbacksProcessedCount

	cmdExecutedCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      CmdExecutedCountMetric,
			Help:      "Number of commands executed",
		},
		[]string{CmdRefLabel},
	)
	p.reg.MustRegister(cmdExecutedCount)
	p.counterVecs[CmdExecutedCountMetric] = cmdExecutedCount

	stageReachedCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      StageReachedCountMetric,
			Help:      "Number stages reached",
		},
		[]string{StageRefLabel},
	)
	p.reg.MustRegister(stageReachedCount)
	p.counterVecs[StageReachedCountMetric] = stageReachedCount

	stageActionProcessed := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      StageActionProcessedCountMetric,
			Help:      "Number stage action processed",
		},
		[]string{StageRefLabel, ActionRefLabel},
	)
	p.reg.MustRegister(stageActionProcessed)
	p.counterVecs[StageActionProcessedCountMetric] = stageActionProcessed
	return p
}

func (p *PrometheusExporter) Increase(metric string) {
	counter, ok := p.counters[metric]
	if !ok {
		log.Println("metric not found in counters", metric)
		return
	}
	counter.Inc()
}

func (p *PrometheusExporter) IncreaseWithLabels(metric string, labels map[string]string) {
	counterVec, ok := p.counterVecs[metric]
	if !ok {
		log.Println("metric not found in counterVecs", metric)
		return
	}
	counterVec.With(labels).Inc()
}

func (p *PrometheusExporter) ObserveWithLabels(metric string, d time.Duration, labels map[string]string) {
	hist, ok := p.hists[metric]
	if !ok {
		log.Println("metric not found in histVecs", metric)
		return
	}
	hist.With(labels).Observe(d.Seconds())
}

func (p *PrometheusExporter) Serve(ctx context.Context) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(p.reg, promhttp.HandlerOpts{}))
	srv := &http.Server{
		Addr:    p.Listen,
		Handler: mux,
	}
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutDownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v; forcing close", err)
		_ = srv.Close()
	}
}
