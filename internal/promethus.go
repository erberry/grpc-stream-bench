package internal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const ()

var PromeHistogramClientRecv *prometheus.HistogramVec //统计消息处理执行耗时
var PromeHistogramClientSend *prometheus.HistogramVec //统计消息处理执行耗时

var PromeHistogramServerRecv prometheus.Counter //统计消息处理执行耗时
var PromeHistogramServerSend prometheus.Counter //统计消息处理执行耗时
var PromeUsageServerSendChan prometheus.Gauge

func InitPrometheus() {
	//TODO 如果希望收集 GO 运行状态（goroutine数量、内存使用统计、gc统计），注释本行代码
	prometheus.Unregister(collectors.NewGoCollector())

	PromeHistogramClientSend = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "client",
			Name:      "client_send",
		},
		[]string{},
	)
	prometheus.MustRegister(PromeHistogramClientSend)

	PromeHistogramClientRecv = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "client",
			Name:      "client_recv",
		},
		[]string{},
	)
	prometheus.MustRegister(PromeHistogramClientRecv)

	PromeHistogramServerSend = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "server",
			Name:      "server_send",
		},
	)
	prometheus.MustRegister(PromeHistogramServerSend)

	PromeHistogramServerRecv = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "server",
			Name:      "server_recv",
		},
	)
	prometheus.MustRegister(PromeHistogramServerRecv)

	PromeUsageServerSendChan = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "server",
			Name:      "server_chan",
		},
	)
	prometheus.MustRegister(PromeUsageServerSendChan)
}
