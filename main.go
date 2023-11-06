// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package main

import (
	"github.com/rs/zerolog"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/event-kit/go/event_kit_api"
	"github.com/steadybit/extension-datadog/config"
	"github.com/steadybit/extension-datadog/extevents"
	"github.com/steadybit/extension-datadog/extmonitor"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthealth"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-kit/extruntime"
	"github.com/steadybit/extension-kit/extutil"
	_ "net/http/pprof" //allow pprof
)

func main() {
	extlogging.InitZeroLog()
	extbuild.PrintBuildInformation()
	extruntime.LogRuntimeInformation(zerolog.DebugLevel)
	exthealth.SetReady(false)
	exthealth.StartProbes(8091)

	config.ParseConfiguration()
	config.ValidateConfiguration()

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(getExtensionList))
	extmonitor.RegisterMonitorDiscoveryHandlers()
	action_kit_sdk.RegisterAction(extmonitor.NewMonitorStatusCheckAction())
	action_kit_sdk.RegisterAction(extmonitor.NewMonitorDowntimeAction())
	extevents.RegisterEventListenerHandlers()

	action_kit_sdk.RegisterCoverageEndpoints()
	action_kit_sdk.InstallSignalHandler()

	exthealth.SetReady(true)
	exthttp.Listen(exthttp.ListenOpts{
		Port: 8090,
	})
}

// ExtensionListResponse exists to merge the possible root path responses supported by the
// various extension kits. In this case, the response for ActionKit, DiscoveryKit and EventKit.
type ExtensionListResponse struct {
	action_kit_api.ActionList       `json:",inline"`
	discovery_kit_api.DiscoveryList `json:",inline"`
	event_kit_api.EventListenerList `json:",inline"`
}

func getExtensionList() ExtensionListResponse {
	return ExtensionListResponse{
		ActionList: action_kit_sdk.GetActionList(),
		DiscoveryList: discovery_kit_api.DiscoveryList{
			Discoveries: []discovery_kit_api.DescribingEndpointReference{
				{
					Method: "GET",
					Path:   "/monitor/discovery",
				},
			},
			TargetTypes: []discovery_kit_api.DescribingEndpointReference{
				{
					Method: "GET",
					Path:   "/monitor/discovery/target-description",
				},
			},
			TargetAttributes: []discovery_kit_api.DescribingEndpointReference{
				{
					Method: "GET",
					Path:   "/monitor/discovery/attribute-descriptions",
				},
			},
			TargetEnrichmentRules: []discovery_kit_api.DescribingEndpointReference{},
		},
		EventListenerList: event_kit_api.EventListenerList{
			EventListeners: []event_kit_api.EventListener{
				{
					Method:   "POST",
					Path:     "/events/experiment-started",
					ListenTo: []string{"experiment.execution.created"},

					RestrictTo: extutil.Ptr(event_kit_api.Leader),
				},
				{
					Method:     "POST",
					Path:       "/events/experiment-completed",
					ListenTo:   []string{"experiment.execution.completed", "experiment.execution.failed", "experiment.execution.canceled", "experiment.execution.errored"},
					RestrictTo: extutil.Ptr(event_kit_api.Leader),
				},
				{
					Method:     "POST",
					Path:       "/events/experiment-step-started",
					ListenTo:   []string{"experiment.execution.step-started"},
					RestrictTo: extutil.Ptr(event_kit_api.Leader),
				},
				{
					Method:     "POST",
					Path:       "/events/experiment-step-completed",
					ListenTo:   []string{"experiment.execution.step-completed", "experiment.execution.step-canceled", "experiment.execution.step-errored", "experiment.execution.step-failed"},
					RestrictTo: extutil.Ptr(event_kit_api.Leader),
				},
				{
					Method:     "POST",
					Path:       "/events/experiment-target-started",
					ListenTo:   []string{"experiment.execution.target-started"},
					RestrictTo: extutil.Ptr(event_kit_api.Leader),
				},
				{
					Method:     "POST",
					Path:       "/events/experiment-target-completed",
					ListenTo:   []string{"experiment.execution.target-completed", "experiment.execution.target-canceled", "experiment.execution.target-errored", "experiment.execution.target-failed"},
					RestrictTo: extutil.Ptr(event_kit_api.Leader),
				},
			},
		},
	}
}
