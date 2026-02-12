// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package main

import (
	_ "github.com/KimMachineGun/automemlimit" // By default, it sets `GOMEMLIMIT` to 90% of cgroup's memory limit.
	"github.com/rs/zerolog"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_sdk"
	"github.com/steadybit/event-kit/go/event_kit_api"
	"github.com/steadybit/extension-datadog/config"
	"github.com/steadybit/extension-datadog/extevents"
	"github.com/steadybit/extension-datadog/extmonitor"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthealth"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-kit/extruntime"
	"github.com/steadybit/extension-kit/extsignals"
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
	discovery_kit_sdk.Register(extmonitor.NewMonitorDiscovery())
	action_kit_sdk.RegisterAction(extmonitor.NewMonitorStatusCheckAction())
	action_kit_sdk.RegisterAction(extmonitor.NewMonitorDowntimeAction())
	extevents.RegisterEventListenerHandlers()

	action_kit_sdk.RegisterCoverageEndpoints()
	extsignals.ActivateSignalHandlers()

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
		ActionList:    action_kit_sdk.GetActionList(),
		DiscoveryList: discovery_kit_sdk.GetDiscoveryList(),
		EventListenerList: event_kit_api.EventListenerList{
			EventListeners: []event_kit_api.EventListener{
				{
					Method:   "POST",
					Path:     "/events/experiment-started",
					ListenTo: []string{"experiment.execution.created"},
				},
				{
					Method:   "POST",
					Path:     "/events/experiment-completed",
					ListenTo: []string{"experiment.execution.completed", "experiment.execution.failed", "experiment.execution.canceled", "experiment.execution.errored"},
				},
				{
					Method:   "POST",
					Path:     "/events/experiment-step-started",
					ListenTo: []string{"experiment.execution.step-started"},
				},
				{
					Method:   "POST",
					Path:     "/events/experiment-target-started",
					ListenTo: []string{"experiment.execution.target-started"},
				},
				{
					Method:   "POST",
					Path:     "/events/experiment-target-completed",
					ListenTo: []string{"experiment.execution.target-completed", "experiment.execution.target-canceled", "experiment.execution.target-errored", "experiment.execution.target-failed"},
				},
			},
		},
	}
}
