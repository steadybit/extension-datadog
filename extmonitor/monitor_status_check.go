// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extmonitor

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-datadog/config"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type MonitorStatusCheckAction struct{}

// Make sure action implements all required interfaces
var (
	_ action_kit_sdk.Action[MonitorStatusCheckState]           = (*MonitorStatusCheckAction)(nil)
	_ action_kit_sdk.ActionWithStatus[MonitorStatusCheckState] = (*MonitorStatusCheckAction)(nil)
)

type MonitorStatusCheckState struct {
	MonitorId      int64
	End            time.Time
	ExpectedStatus string
}

func NewMonitorStatusCheckAction() action_kit_sdk.Action[MonitorStatusCheckState] {
	return &MonitorStatusCheckAction{}
}

func (m *MonitorStatusCheckAction) NewEmptyState() MonitorStatusCheckState {
	return MonitorStatusCheckState{}
}

func (m *MonitorStatusCheckAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.status_check", monitorTargetId),
		Label:       "Monitor Status",
		Description: "collects information about the monitor status and optionally verifies that the monitor has an expected status.",
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
		Category:    extutil.Ptr("monitoring"),
		Kind:        action_kit_api.Check,
		TimeControl: action_kit_api.Internal,
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
				Name:        "expectedStatus",
				Label:       "Expected Status",
				Description: extutil.Ptr(""),
				Type:        action_kit_api.String,
				Options: extutil.Ptr([]action_kit_api.ParameterOption{
					action_kit_api.ExplicitParameterOption{
						Label: "Ok",
						Value: string(datadogV1.MONITOROVERALLSTATES_OK),
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Alert",
						Value: string(datadogV1.MONITOROVERALLSTATES_ALERT),
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Warn",
						Value: string(datadogV1.MONITOROVERALLSTATES_WARN),
					},
					action_kit_api.ExplicitParameterOption{
						Label: "No Data",
						Value: string(datadogV1.MONITOROVERALLSTATES_NO_DATA),
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Unknown",
						Value: string(datadogV1.MONITOROVERALLSTATES_UNKNOWN),
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Skipped",
						Value: string(datadogV1.MONITOROVERALLSTATES_SKIPPED),
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Ignored",
						Value: string(datadogV1.MONITOROVERALLSTATES_IGNORED),
					},
				}),
				Required: extutil.Ptr(false),
				Order:    extutil.Ptr(2),
			},
		},
		Widgets: extutil.Ptr([]action_kit_api.Widget{
			action_kit_api.StateOverTimeWidget{
				Type:  action_kit_api.ComSteadybitWidgetStateOverTime,
				Title: "Datadog Monitor Status",
				Identity: action_kit_api.StateOverTimeWidgetIdentityConfig{
					From: "datadog.monitor.id",
				},
				Label: action_kit_api.StateOverTimeWidgetLabelConfig{
					From: "datadog.monitor.name",
				},
				State: action_kit_api.StateOverTimeWidgetStateConfig{
					From: "state",
				},
				Tooltip: action_kit_api.StateOverTimeWidgetTooltipConfig{
					From: "tooltip",
				},
				Url: extutil.Ptr(action_kit_api.StateOverTimeWidgetUrlConfig{
					From: extutil.Ptr("url"),
				}),
				Value: extutil.Ptr(action_kit_api.StateOverTimeWidgetValueConfig{
					Hide: extutil.Ptr(true),
				}),
			},
		}),
		Prepare: action_kit_api.MutatingEndpointReference{},
		Start:   action_kit_api.MutatingEndpointReference{},
		Status: extutil.Ptr(action_kit_api.MutatingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("5s"),
		}),
	}
}

func (m *MonitorStatusCheckAction) Prepare(_ context.Context, state *MonitorStatusCheckState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	monitorId := request.Target.Attributes["datadog.monitor.id"]
	if len(monitorId) == 0 {
		return nil, extutil.Ptr(extension_kit.ToError("Target is missing the 'datadog.monitor.id' tag.", nil))
	}

	duration := request.Config["duration"].(float64)
	end := time.Now().Add(time.Millisecond * time.Duration(duration))

	var expectedStatus string
	if request.Config["expectedStatus"] != nil {
		expectedStatus = fmt.Sprintf("%v", request.Config["expectedStatus"])
	}

	parsedMonitorId, err := strconv.ParseInt(monitorId[0], 10, 64)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError(fmt.Sprintf("Failed to parse monitor ID '%s' as int64.", monitorId[0]), nil))
	}

	state.MonitorId = parsedMonitorId
	state.End = end
	state.ExpectedStatus = expectedStatus

	return nil, nil
}

func (m *MonitorStatusCheckAction) Start(_ context.Context, _ *MonitorStatusCheckState) (*action_kit_api.StartResult, error) {
	return nil, nil
}

func (m *MonitorStatusCheckAction) Status(ctx context.Context, state *MonitorStatusCheckState) (*action_kit_api.StatusResult, error) {
	return MonitorStatusCheckStatus(ctx, state, &config.Config, config.Config.SiteUrl)
}

type GetMonitorApi interface {
	GetMonitor(ctx context.Context, monitorId int64, params datadogV1.GetMonitorOptionalParameters) (datadogV1.Monitor, *http.Response, error)
}

func MonitorStatusCheckStatus(ctx context.Context, state *MonitorStatusCheckState, api GetMonitorApi, siteUrl string) (*action_kit_api.StatusResult, error) {
	now := time.Now()
	monitor, resp, err := api.GetMonitor(ctx, state.MonitorId, *datadogV1.NewGetMonitorOptionalParameters())
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError(fmt.Sprintf("Failed to retrieve monitor %d from Datadog. Full response: %v", state.MonitorId, resp), err))
	}

	completed := now.After(state.End)
	var checkError *action_kit_api.ActionKitError
	if len(state.ExpectedStatus) > 0 && monitor.OverallState != nil && string(*monitor.OverallState) != state.ExpectedStatus {
		tags := strings.Join(monitor.Tags, ", ")
		if len(tags) == 0 {
			tags = "<none>"
		}
		checkError = extutil.Ptr(action_kit_api.ActionKitError{
			Title: fmt.Sprintf("Monitor '%s' (id %d, tags: %s) has status '%s' whereas '%s' is expected.",
				*monitor.Name,
				state.MonitorId,
				tags,
				*monitor.OverallState,
				state.ExpectedStatus),
			Status: extutil.Ptr(action_kit_api.Failed),
		})
	}

	metrics := []action_kit_api.Metric{
		*toMetric(&monitor, now, siteUrl),
	}

	return &action_kit_api.StatusResult{
		Completed: completed,
		Error:     checkError,
		Metrics:   extutil.Ptr(metrics),
	}, nil
}

func toMetric(monitor *datadogV1.Monitor, now time.Time, siteUrl string) *action_kit_api.Metric {
	var tooltip string
	var state string

	if monitor.OverallState == nil || *monitor.OverallState == datadogV1.MONITOROVERALLSTATES_UNKNOWN {
		state = "warn"
		tooltip = "Monitor status is: Unknown"
	} else {
		tooltip = fmt.Sprintf("Monitor status is: %s", *monitor.OverallState)
		switch *monitor.OverallState {
		case datadogV1.MONITOROVERALLSTATES_ALERT:
			state = "danger"
		case datadogV1.MONITOROVERALLSTATES_IGNORED:
			state = "warn"
		case datadogV1.MONITOROVERALLSTATES_NO_DATA:
			state = "info"
		case datadogV1.MONITOROVERALLSTATES_OK:
			state = "success"
		case datadogV1.MONITOROVERALLSTATES_SKIPPED:
			state = "info"
		case datadogV1.MONITOROVERALLSTATES_WARN:
			state = "warn"
		default:
			state = "danger"
		}
	}

	monitorId := strconv.FormatInt(*monitor.Id, 10)
	return extutil.Ptr(action_kit_api.Metric{
		Name: extutil.Ptr("datadog_monitor_status"),
		Metric: map[string]string{
			"datadog.monitor.id":   monitorId,
			"datadog.monitor.name": *monitor.Name,
			"state":                state,
			"tooltip":              tooltip,
			"url":                  fmt.Sprintf("%s/monitors/%s", siteUrl, monitorId),
		},
		Timestamp: now,
		Value:     0,
	})
}
