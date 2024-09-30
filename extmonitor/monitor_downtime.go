// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extmonitor

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-datadog/config"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"net/http"
	"strconv"
	"time"
)

type MonitorDowntimeAction struct{}

// Make sure action implements all required interfaces
var (
	_ action_kit_sdk.Action[MonitorDowntimeState]         = (*MonitorDowntimeAction)(nil)
	_ action_kit_sdk.ActionWithStop[MonitorDowntimeState] = (*MonitorDowntimeAction)(nil)
)

type MonitorDowntimeState struct {
	MonitorId     int64
	End           time.Time
	Notify        bool
	DowntimeId    *string
	ExperimentUri *string
	ExecutionUri  *string
}

func NewMonitorDowntimeAction() action_kit_sdk.Action[MonitorDowntimeState] {
	return &MonitorDowntimeAction{}
}
func (m *MonitorDowntimeAction) NewEmptyState() MonitorDowntimeState {
	return MonitorDowntimeState{}
}

func (m *MonitorDowntimeAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.downtime", monitorTargetId),
		Label:       "Create Downtime",
		Description: "Start a Monitor Downtime for a given duration.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(monitorIcon),
		TargetSelection: extutil.Ptr(action_kit_api.TargetSelection{
			TargetType:          monitorTargetId,
			QuantityRestriction: extutil.Ptr(action_kit_api.All),
			SelectionTemplates: extutil.Ptr([]action_kit_api.TargetSelectionTemplate{
				{
					Label: "by monitor name",
					Query: "datadog.monitor.name=\"\"",
				},
			}),
		}),
		Technology:  extutil.Ptr("Datadog"),
		Kind:        action_kit_api.Other,
		TimeControl: action_kit_api.TimeControlExternal,
		Parameters: []action_kit_api.ActionParameter{
			{
				Name:         "duration",
				Label:        "Duration",
				Description:  extutil.Ptr(""),
				Type:         action_kit_api.Duration,
				DefaultValue: extutil.Ptr("30s"),
				Order:        extutil.Ptr(1),
				Required:     extutil.Ptr(true),
			},
			{
				Name:         "notify",
				Label:        "Notify after Downtime if unhealthy",
				Description:  extutil.Ptr("Should datadog notify after the Downtime if the monitor is in an unhealthy state?"),
				Type:         action_kit_api.Boolean,
				DefaultValue: extutil.Ptr("true"),
				Order:        extutil.Ptr(2),
				Required:     extutil.Ptr(true),
			},
		},
		Stop: extutil.Ptr(action_kit_api.MutatingEndpointReference{}),
	}
}

func (m *MonitorDowntimeAction) Prepare(_ context.Context, state *MonitorDowntimeState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	monitorId := request.Target.Attributes["datadog.monitor.id"]
	if len(monitorId) == 0 {
		return nil, extension_kit.ToError("Target is missing the 'datadog.monitor.id' tag.", nil)
	}

	parsedMonitorId, err := strconv.ParseInt(monitorId[0], 10, 64)
	if err != nil {
		return nil, extension_kit.ToError(fmt.Sprintf("Failed to parse monitor ID '%s' as int64.", monitorId[0]), nil)
	}

	duration := request.Config["duration"].(float64)
	end := time.Now().Add(time.Millisecond * time.Duration(duration))

	state.MonitorId = parsedMonitorId
	state.End = end
	state.Notify = request.Config["notify"].(bool)
	state.ExperimentUri = request.ExecutionContext.ExperimentUri
	state.ExecutionUri = request.ExecutionContext.ExecutionUri
	return nil, nil
}

func (m *MonitorDowntimeAction) Start(ctx context.Context, state *MonitorDowntimeState) (*action_kit_api.StartResult, error) {
	return MonitorDowntimeStart(ctx, state, &config.Config)
}

func (m *MonitorDowntimeAction) Stop(ctx context.Context, state *MonitorDowntimeState) (*action_kit_api.StopResult, error) {
	return MonitorDowntimeStop(ctx, state, &config.Config)
}

type MonitorDowntimeApi interface {
	CreateDowntime(ctx context.Context, downtimeBody datadogV2.DowntimeCreateRequest) (datadogV2.DowntimeResponse, *http.Response, error)
	CancelDowntime(ctx context.Context, downtimeId string) (*http.Response, error)
}

func MonitorDowntimeStart(ctx context.Context, state *MonitorDowntimeState, api MonitorDowntimeApi) (*action_kit_api.StartResult, error) {
	var notifyEndType []datadogV2.DowntimeNotifyEndStateActions
	if state.Notify {
		notifyEndType = []datadogV2.DowntimeNotifyEndStateActions{datadogV2.DOWNTIMENOTIFYENDSTATEACTIONS_CANCELED, datadogV2.DOWNTIMENOTIFYENDSTATEACTIONS_EXPIRED}
	}

	message := "Created by ![Steadybit](https://downloads.steadybit.com/logo-extension-datadog.jpg)"
	if state.ExperimentUri != nil {
		message = message + fmt.Sprintf("\n\n[Open Experiment](%s)", *state.ExperimentUri)
	}
	if state.ExecutionUri != nil {
		message = message + fmt.Sprintf("\n\n[Open Execution](%s)", *state.ExecutionUri)
	}

	downtimeRequest := datadogV2.DowntimeCreateRequest{
		Data: datadogV2.DowntimeCreateRequestData{
			Type: datadogV2.DOWNTIMERESOURCETYPE_DOWNTIME,
			Attributes: datadogV2.DowntimeCreateRequestAttributes{
				MonitorIdentifier: datadogV2.DowntimeMonitorIdentifier{
					DowntimeMonitorIdentifierId: &datadogV2.DowntimeMonitorIdentifierId{
						MonitorId: state.MonitorId,
					},
				},
				Message: *datadog.NewNullableString(extutil.Ptr(message)),
				Schedule: &datadogV2.DowntimeScheduleCreateRequest{
					DowntimeScheduleOneTimeCreateUpdateRequest: &datadogV2.DowntimeScheduleOneTimeCreateUpdateRequest{
						End: *datadog.NewNullableTime(extutil.Ptr(state.End)),
					},
				},
				MuteFirstRecoveryNotification: extutil.Ptr(true),
				NotifyEndTypes:                notifyEndType,
				Scope:                         "*",
			},
		},
	}

	downtime, resp, err := api.CreateDowntime(ctx, downtimeRequest)
	if err != nil {
		return nil, extension_kit.ToError(fmt.Sprintf("Failed to create Downtime for monitor %d. Full response: %v", state.MonitorId, resp), err)
	}

	state.DowntimeId = downtime.Data.Id

	return &action_kit_api.StartResult{
		Messages: &action_kit_api.Messages{
			action_kit_api.Message{Level: extutil.Ptr(action_kit_api.Info), Message: fmt.Sprintf("Downtime started. (monitor %d, downtime %s)", state.MonitorId, *state.DowntimeId)},
		},
	}, nil
}

func MonitorDowntimeStop(ctx context.Context, state *MonitorDowntimeState, api MonitorDowntimeApi) (*action_kit_api.StopResult, error) {
	if state.DowntimeId == nil {
		return nil, nil
	}

	resp, err := api.CancelDowntime(ctx, *state.DowntimeId)
	if err != nil {
		return nil, extension_kit.ToError(fmt.Sprintf("Failed to cancel Downtime (monitor %d, downtime %s). Full response: %v", state.MonitorId, *state.DowntimeId, resp), err)
	}

	return &action_kit_api.StopResult{
		Messages: &action_kit_api.Messages{
			action_kit_api.Message{Level: extutil.Ptr(action_kit_api.Info), Message: fmt.Sprintf("Downtime canceled. (monitor %d, downtime %s)", state.MonitorId, *state.DowntimeId)},
		},
	}, nil
}
