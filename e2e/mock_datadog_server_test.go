package e2e

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
)

func createMockDatadogServer() *httptest.Server {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		panic(fmt.Sprintf("httptest: failed to listen: %v", err))
	}
	server := httptest.Server{
		Listener: listener,
		Config: &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info().Str("path", r.URL.Path).Str("method", r.Method).Str("query", r.URL.RawQuery).Msg("Request received")
			if r.URL.Path == "/api/v1/validate" {
				w.WriteHeader(http.StatusOK)
				w.Write(apiV1Validate())
			} else if r.URL.Path == "/api/v1/monitor" {
				page, _ := strconv.Atoi(r.URL.Query().Get("page"))
				w.WriteHeader(http.StatusOK)
				w.Write(apiV1Monitor(page))
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		})},
	}
	server.Start()
	log.Info().Str("url", server.URL).Msg("Started Mock-Server")
	return &server
}

func apiV1Validate() []byte {
	return []byte(`{"valid": true}`)
}

func apiV1Monitor(page int) []byte {
	if page == 0 {
		return []byte(
			`[
			{
				"id": 8080808,
				"org_id": 9090909090,
				"type": "query alert",
				"name": "[DEV] Monitor Kubernetes Deployments Replica Pods",
				"message": "More than one Deployments Replica's pods are down",
				"tags": [
					"integration:kubernetes",
					"env:dev"
				],
				"query": "min(last_5m):avg:kubernetes_state.deployment.replicas_desired{kube_cluster_name:dev-demo} by {kube_namespace,kube_deployment} - avg:kubernetes_state.deployment.replicas_available{kube_cluster_name:dev-demo} by {kube_namespace,kube_deployment} >= 2",
				"options": {
					"thresholds": {
						"critical": 2.0,
						"warning": 1.0
					},
					"notify_audit": true,
					"require_full_window": false,
					"notify_no_data": true,
					"renotify_interval": 0,
					"timeout_h": 0,
					"include_tags": true,
					"escalation_message": "",
					"new_group_delay": 60,
					"silenced": {}
				},
				"multi": true,
				"created_at": 1666859636000,
				"created": "2022-10-27T08:33:56.272148+00:00",
				"modified": "2023-06-20T10:15:46.744422+00:00",
				"deleted": null,
				"restricted_roles": null,
				"priority": null,
				"overall_state_modified": "2023-06-21T05:42:33+00:00",
				"overall_state": "OK",
				"creator": {
					"name": "Theo Test",
					"handle": "theo@test.com",
					"email": "theo@test.com",
					"id": 1010101010
				},
				"matching_downtimes": []
			}
		]`)
	} else {
		return []byte(`[]`)
	}
}
