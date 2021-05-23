/*
 *@Description
 *@author          lirui
 *@create          2021-05-23 10:54
 */
package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	addedNamespaces = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name:      "added_namespaces",
			Namespace: "tenant_controller",
			Help:      "Number of added namespaces",
			ConstLabels: map[string]string{
				"ssss": "ssss",
			},
		},
	)
)

func init() {
	metrics.Registry.MustRegister(addedNamespaces)
}
