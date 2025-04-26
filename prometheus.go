package main

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type usageCollector struct {
	metric       *prometheus.Desc
	process_name string
	command      string
}

func newUsageCollector(process_name string, command string) *usageCollector {
	return &usageCollector{
		metric:       prometheus.NewDesc(process_name, command, nil, nil),
		process_name: process_name,
		command:      command,
	}
}
func (collector *usageCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.metric
}
func (collector *usageCollector) Collect(ch chan<- prometheus.Metric) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	query := prometheus.MustNewConstMetric(
		collector.metric,
		prometheus.GaugeValue,
		ExecCommand(collector.command))
	//query = prometheus.NewMetricWithTimestamp(time.Now(), query)
	ch <- query
}
