// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package main

import (
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
	"github.com/steadybit/extension-kit/extutil"
)

func main() {
	extlogging.InitZeroLog()
	extbuild.PrintBuildInformation()
	exthealth.SetReady(false)
	exthealth.StartProbes(8091)

	config.ParseConfiguration()
	config.ValidateConfiguration()

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(getExtensionList))
	extmonitor.RegisterMonitorDiscoveryHandlers()
	action_kit_sdk.RegisterAction(extmonitor.NewMonitorStatusCheckAction())
	extevents.RegisterEventListenerHandlers()

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
			},
		},
	}
}
