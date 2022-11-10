// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/event-kit/go/event_kit_api"
	"github.com/steadybit/extension-datadog/config"
	"github.com/steadybit/extension-datadog/extevents"
	"github.com/steadybit/extension-datadog/extmonitor"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-kit/extutil"
	"net/http"
)

func main() {
	extlogging.InitZeroLog()
	config.ParseConfiguration()
	config.ValidateConfiguration()

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(getExtensionList))
	extmonitor.RegisterMonitorDiscoveryHandlers()
	extmonitor.RegisterMonitorStatusCheckHandlers()
	extevents.RegisterEventListenerHandlers()

	port := 8090
	log.Log().Msgf("Starting extension-datadog server on port %d. Get started via /", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to start extension-datadog server on port %d", port)
	}
}

type ExtensionListResponse struct {
	Actions          []action_kit_api.DescribingEndpointReference    `json:"actions"`
	Discoveries      []discovery_kit_api.DescribingEndpointReference `json:"discoveries"`
	TargetAttributes []discovery_kit_api.DescribingEndpointReference `json:"targetAttributes"`
	TargetTypes      []discovery_kit_api.DescribingEndpointReference `json:"targetTypes"`
	EventListeners   []event_kit_api.EventListener                   `json:"eventListeners"`
}

func getExtensionList() ExtensionListResponse {
	return ExtensionListResponse{
		Actions: []action_kit_api.DescribingEndpointReference{
			{
				"GET",
				"/monitor/action/status-check",
			},
		},
		Discoveries: []discovery_kit_api.DescribingEndpointReference{
			{
				"GET",
				"/monitor/discovery",
			},
		},
		TargetTypes: []discovery_kit_api.DescribingEndpointReference{
			{
				"GET",
				"/monitor/discovery/target-description",
			},
		},
		TargetAttributes: []discovery_kit_api.DescribingEndpointReference{
			{
				"GET",
				"/monitor/discovery/attribute-descriptions",
			},
		},
		EventListeners: []event_kit_api.EventListener{
			{
				Method:     "POST",
				Path:       "/events/experiment-started",
				ListenTo:   []string{"experiment.execution.created"},

				RestrictTo: extutil.Ptr(event_kit_api.Leader),
			},
			{
				Method:     "POST",
				Path:       "/events/experiment-completed",
				ListenTo:   []string{"experiment.execution.completed"},
				RestrictTo: extutil.Ptr(event_kit_api.Leader),
			},
		},
	}
}
