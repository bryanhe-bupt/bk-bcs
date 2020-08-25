/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cloud

import (
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/metrics"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	reqCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_network",
			Subsystem: "ingress_controller",
			Name:      "cloud_request_total",
			Help:      "request total counter",
		},
		[]string{"rpc", "errcode"},
	)

	respTimeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "bcs_network",
			Subsystem: "ingress_controller",
			Name:      "cloud_response_time",
			Help:      "response time(ms) summary.",
		},
		[]string{"rpc"},
	)
)

const (
	// MetricAPISuccess api call is successful
	MetricAPISuccess = "success"
	// MetricAPIFailed api call is failed
	MetricAPIFailed = "failed"
)

var (
	// ErrLoadbalancerNotFound error that loadbalancer not found
	ErrLoadbalancerNotFound = fmt.Errorf("loabalancer not found")
	// ErrListenerNotFound error that listener not found
	ErrListenerNotFound = fmt.Errorf("listener not found")
)

func init() {
	metrics.Registry.MustRegister(reqCounter, respTimeSummary)
}

// StatRequest report metrics for rpc requests
func StatRequest(rpc string, errcode string, inTime, outTime time.Time) int64 {
	reqCounter.With(prometheus.Labels{
		"rpc":     rpc,
		"errcode": errcode,
	}).Inc()

	cost := toMSTimestamp(outTime) - toMSTimestamp(inTime)
	respTimeSummary.With(prometheus.Labels{"rpc": rpc}).Observe(float64(cost))

	return cost
}

// toMSTimestamp converts time.Time to millisecond timestamp.
func toMSTimestamp(t time.Time) int64 {
	return t.UnixNano() / 1e6
}

// LoadBalanceObject lb object
type LoadBalanceObject struct {
	LbID string   `json:"lbID"`
	Type string   `json:"type"`
	Name string   `json:"name"`
	IPs  []string `json:"ips"`
	VIPs []string `json:"vips,omitempty"`
}

// LoadBalance interface for clb loadbalancer
type LoadBalance interface {
	// DescribeLoadBalancer get loadbalancer object by id
	DescribeLoadBalancer(lbID string) (*LoadBalanceObject, error)

	// EnsureListener ensure listener to cloud, and get listener info
	EnsureListener(listener *networkextensionv1.Listener) (string, error)

	// DeleteListener delete listener by name
	DeleteListener(listener *networkextensionv1.Listener) error

	// EnsureSegmentListener ensure segment listener
	EnsureSegmentListener(listener *networkextensionv1.Listener) (string, error)

	// DeleteSegmentListener delete segment listener
	DeleteSegmentListener(listener *networkextensionv1.Listener) error
}

// Validater validate parameter for cloud loadbalancer
type Validater interface {
	// IsIngressValid check bcs ingress parameter
	IsIngressValid(ingress *networkextensionv1.Ingress) (isValid bool, msg string)
}
