// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

/*
 * Copyright 2022 steadybit GmbH. All rights reserved.
 */

package extevents

import (
	"context"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/google/uuid"
	"github.com/steadybit/event-kit/go/event_kit_api"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

type datadogClientMock struct {
	mock.Mock
}

func (m *datadogClientMock) SendEvent(ctx context.Context, datadogEventBody datadogV1.EventCreateRequest) (datadogV1.EventCreateResponse, *http.Response, error) {
	args := m.Called(ctx, datadogEventBody)
	return args.Get(0).(datadogV1.EventCreateResponse), args.Get(1).(*http.Response), args.Error(2)
}

func Test_sendDatadogEvent(t *testing.T) {
	// Given
	mockedApi := new(datadogClientMock)
	okResponse := http.Response{
		StatusCode: 202,
	}
	mockedApi.On("SendEvent", mock.Anything, mock.AnythingOfType("datadogV1.EventCreateRequest")).Return(datadogV1.EventCreateResponse{}, &okResponse, nil)

	// When
	datadogEventBody := datadogV1.EventCreateRequest{
		Title:          "Experiment started",
		Text:           "An experiment has been started by the Steadybit platform",
		Tags:           []string{},
		SourceTypeName: extutil.Ptr("Steadybit"),
	}
	sendDatadogEvent(mockedApi, &datadogEventBody)

	// Then
	mockedApi.AssertNumberOfCalls(t, "SendEvent", 1)
}

func Test_getEventBaseTags(t *testing.T) {
	type args struct {
		event event_kit_api.EventRequestBody
	}

	eventTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Successfully get base tags",
			args: args{
				event: event_kit_api.EventRequestBody{
					Environment: extutil.Ptr(event_kit_api.Environment{
						Id:   "test",
						Name: "gateway",
					}),
					EventName: "experiment.started",
					EventTime: eventTime,
					Id:        uuid.MustParse("ccf6a26e-588f-446e-8eaa-d16b086e150e"),
					Principal: event_kit_api.UserPrincipal{
						Email:         extutil.Ptr("email"),
						Name:          "Peter",
						Username:      "Pan",
						PrincipalType: string(event_kit_api.User),
					},
					Team: extutil.Ptr(event_kit_api.Team{
						Id:   "test",
						Key:  "test",
						Name: "gateway",
					}),
					Tenant: event_kit_api.Tenant{
						Key:  "key",
						Name: "name",
					},
				},
			},
			want: []string{
				"source:Steadybit",
				"environment_name:gateway",
				"event_name:experiment.started",
				"event_time:2021-01-01 00:00:00 +0000 UTC",
				"event_id:ccf6a26e-588f-446e-8eaa-d16b086e150e",
				"tenant_name:name",
				"tenant_key:key",
				"team_name:gateway",
				"team_key:test",
				"principal_type:user",
				"principal_username:Pan",
				"principal_name:Peter"},
		},
		{
			name: "Successfully get base tags without Principal",
			args: args{
				event: event_kit_api.EventRequestBody{
					Environment: extutil.Ptr(event_kit_api.Environment{
						Id:   "test",
						Name: "gateway",
					}),
					EventName: "experiment.started",
					EventTime: eventTime,
					Id:        uuid.MustParse("ccf6a26e-588f-446e-8eaa-d16b086e150e"),
					Principal: event_kit_api.AccessTokenPrincipal{
						Name:          "MyFancyToken",
						PrincipalType: string(event_kit_api.AccessToken),
					},
					Team: extutil.Ptr(event_kit_api.Team{
						Id:   "test",
						Key:  "test",
						Name: "gateway",
					}),
					Tenant: event_kit_api.Tenant{
						Key:  "key",
						Name: "name",
					},
				},
			},
			want: []string{
				"source:Steadybit",
				"environment_name:gateway",
				"event_name:experiment.started",
				"event_time:2021-01-01 00:00:00 +0000 UTC",
				"event_id:ccf6a26e-588f-446e-8eaa-d16b086e150e",
				"tenant_name:name",
				"tenant_key:key",
				"team_name:gateway",
				"team_key:test",
				"principal_type:access_token",
				"principal_name:MyFancyToken",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEventBaseTags(tt.args.event)
			assert.Equalf(t, tt.want, got, "getEventBaseTags(%v)", tt.args.event)
		})
	}
}

func Test_getExecutionTags(t *testing.T) {
	type args struct {
		event event_kit_api.EventRequestBody
	}

	eventTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	endedTime := time.Date(2021, 1, 1, 0, 2, 0, 0, time.UTC)
	startedTime := time.Date(2021, 1, 1, 0, 1, 0, 0, time.UTC)
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Successfully get execution tags",
			args: args{
				event: event_kit_api.EventRequestBody{
					ExperimentExecution: extutil.Ptr(event_kit_api.ExperimentExecution{
						EndedTime:     extutil.Ptr(endedTime),
						ExecutionId:   42,
						ExperimentKey: "ExperimentKey",
						Reason:        extutil.Ptr("Reason"),
						ReasonDetails: extutil.Ptr("ReasonDetails"),
						Hypothesis:    "Hypothesis",
						Name:          "Name",
						PreparedTime:  eventTime,
						StartedTime:   startedTime,
						State:         event_kit_api.ExperimentExecutionStateCreated,
					}),
				},
			},
			want: []string{
				"execution_id:42",
				"experiment_key:ExperimentKey",
				"experiment_name:Name",
				"execution_state:created",
				"experiment_hypothesis:Hypothesis",
				"started_time:" + startedTime.Format(time.RFC3339),
				"ended_time:" + endedTime.Format(time.RFC3339)},
		},
		{
			name: "Successfully get execution tags without hypothesis",
			args: args{
				event: event_kit_api.EventRequestBody{
					ExperimentExecution: extutil.Ptr(event_kit_api.ExperimentExecution{
						EndedTime:     extutil.Ptr(endedTime),
						ExecutionId:   42,
						ExperimentKey: "ExperimentKey",
						Reason:        extutil.Ptr("Reason"),
						ReasonDetails: extutil.Ptr("ReasonDetails"),
						Hypothesis:    "",
						Name:          "Name",
						PreparedTime:  eventTime,
						StartedTime:   startedTime,
						State:         event_kit_api.ExperimentExecutionStateCreated,
					}),
				},
			},
			want: []string{
				"execution_id:42",
				"experiment_key:ExperimentKey",
				"experiment_name:Name",
				"execution_state:created",
				"started_time:" + startedTime.Format(time.RFC3339),
				"ended_time:" + endedTime.Format(time.RFC3339)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getExecutionTags(tt.args.event)
			assert.Equalf(t, tt.want, got, "getExecutionTags(%v)", tt.args.event)
		})
	}
}

func Test_getStepTags(t *testing.T) {
	type args struct {
		w             http.ResponseWriter
		stepExecution event_kit_api.ExperimentStepExecution
	}

	endedTime := time.Date(2021, 1, 1, 0, 2, 0, 0, time.UTC)
	startedTime := time.Date(2021, 1, 1, 0, 1, 0, 0, time.UTC)
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Successfully get tags for started attack",
			args: args{
				stepExecution: event_kit_api.ExperimentStepExecution{
					Id:          uuid.UUID{},
					Type:        event_kit_api.Action,
					ActionId:    extutil.Ptr("com.steadybit.action.example"),
					ActionName:  extutil.Ptr("example-action"),
					ActionKind:  extutil.Ptr(event_kit_api.Attack),
					CustomLabel: extutil.Ptr("My very own label"),
					State:       event_kit_api.ExperimentStepExecutionStateFailed,
					EndedTime:   extutil.Ptr(endedTime),
					StartedTime: extutil.Ptr(startedTime),
				},
			},
			want: []string{
				"step_action_id:com.steadybit.action.example",
				"step_action_name:example-action",
				"step_custom_label:My very own label",
			},
		},
		{
			name: "Successfully get tags for not yet started attack",
			args: args{
				stepExecution: event_kit_api.ExperimentStepExecution{
					Id:         uuid.UUID{},
					Type:       event_kit_api.Action,
					ActionId:   extutil.Ptr("com.steadybit.action.example"),
					ActionKind: extutil.Ptr(event_kit_api.Attack),
					State:      event_kit_api.ExperimentStepExecutionStateCompleted,
				},
			},
			want: []string{
				"step_action_id:com.steadybit.action.example",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStepTags(tt.args.stepExecution)
			assert.Equalf(t, tt.want, got, "getStepTags(%v)", tt.args.stepExecution)
		})
	}
}

func Test_getTargetTags(t *testing.T) {
	type args struct {
		w      http.ResponseWriter
		target event_kit_api.ExperimentStepTargetExecution
	}

	endedTime := time.Date(2021, 1, 1, 0, 2, 0, 0, time.UTC)
	startedTime := time.Date(2021, 1, 1, 0, 1, 0, 0, time.UTC)
	id := uuid.New()
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Successfully get tag for container targets",
			args: args{
				target: event_kit_api.ExperimentStepTargetExecution{
					ExecutionId:   42,
					ExperimentKey: "ExperimentKey",
					Id:            id,
					State:         "completed",
					AgentHostname: "Agent-1",
					TargetAttributes: map[string][]string{
						"k8s.container.name":                       {"example-c1"},
						"k8s.pod.label.tags.datadoghq.com/service": {"example-service"},
						"container.host":                           {"host-123"},
						"k8s.namespace":                            {"namespace"},
						"k8s.deployment":                           {"example"},
						"k8s.pod.name":                             {"example-4711-123"},
						"k8s.cluster-name":                         {"dev-cluster"},
						"aws.zone":                                 {"eu-central-1a"},
						"aws.region":                               {"eu-central-1"},
						"aws.account":                              {"123456789"},
					},
					TargetName:  "Container",
					TargetType:  "com.steadybit.extension_container.container",
					StartedTime: &startedTime,
					EndedTime:   &endedTime,
				},
			},
			want: []string{
				"execution_id:42",
				"experiment_key:ExperimentKey",
				"execution_state:completed",
				"started_time:2021-01-01T00:01:00Z",
				"ended_time:2021-01-01T00:02:00Z",
				"kube_cluster_name:dev-cluster",
				"kube_namespace:namespace",
				"kube_deployment:example",
				"namespace:namespace",
				"pod_name:example-4711-123",
				"deployment:example",
				"container_name:example-c1",
				"cluster_name:dev-cluster",
				"service:example-service",
				"host:host-123-dev-cluster",
				"aws_region:eu-central-1",
				"aws_zone:eu-central-1a",
				"aws_account:123456789",
			},
		},
		{
			name: "Successfully deduplicate service tags",
			args: args{
				target: event_kit_api.ExperimentStepTargetExecution{
					ExecutionId:   42,
					ExperimentKey: "ExperimentKey",
					Id:            id,
					State:         "completed",
					AgentHostname: "Agent-1",
					TargetAttributes: map[string][]string{
						"k8s.cluster-name":                                {"dev-cluster"},
						"k8s.pod.label.tags.datadoghq.com/service":        {"service-1"},
						"k8s.deployment.label.tags.datadoghq.com/service": {"service-1"},
					},
					TargetName:  "Container",
					TargetType:  "com.steadybit.extension_container.container",
					StartedTime: &startedTime,
					EndedTime:   &endedTime,
				},
			},
			want: []string{
				"execution_id:42",
				"experiment_key:ExperimentKey",
				"execution_state:completed",
				"started_time:2021-01-01T00:01:00Z",
				"ended_time:2021-01-01T00:02:00Z",
				"kube_cluster_name:dev-cluster",
				"cluster_name:dev-cluster",
				"service:service-1",
			},
		},
		{
			name: "Should add cluster name to hostname and deduplicate host names",
			args: args{
				target: event_kit_api.ExperimentStepTargetExecution{
					ExecutionId:   42,
					ExperimentKey: "ExperimentKey",
					Id:            id,
					State:         "completed",
					AgentHostname: "Agent-1",
					TargetAttributes: map[string][]string{
						"k8s.cluster-name":     {"dev-cluster"},
						"container.host":       {"host-123"},
						"host.hostname":        {"host-123"},
						"application.hostname": {"host-123"},
					},
					TargetName:  "Container",
					TargetType:  "com.steadybit.extension_container.container",
					StartedTime: &startedTime,
					EndedTime:   &endedTime,
				},
			},
			want: []string{
				"execution_id:42",
				"experiment_key:ExperimentKey",
				"execution_state:completed",
				"started_time:2021-01-01T00:01:00Z",
				"ended_time:2021-01-01T00:02:00Z",
				"kube_cluster_name:dev-cluster",
				"cluster_name:dev-cluster",
				"host:host-123-dev-cluster",
			},
		},
		{
			name: "Ignore multiple values",
			args: args{
				target: event_kit_api.ExperimentStepTargetExecution{
					ExecutionId:   42,
					ExperimentKey: "ExperimentKey",
					Id:            id,
					State:         "completed",
					AgentHostname: "Agent-1",
					TargetAttributes: map[string][]string{
						"host.hostname": {"Host-1"},
						"k8s.namespace": {"namespace-1", "namespace-2", "namespace-3"},
					},
					TargetName:  "Host",
					TargetType:  "com.steadybit.extension_host.host",
					StartedTime: &startedTime,
					EndedTime:   &endedTime,
				},
			},
			want: []string{
				"execution_id:42",
				"experiment_key:ExperimentKey",
				"execution_state:completed",
				"started_time:2021-01-01T00:01:00Z",
				"ended_time:2021-01-01T00:02:00Z",
				"host:Host-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTargetTags(tt.args.target)
			assert.Equalf(t, tt.want, got, "getTargetTags(%v)", tt.args.target)
		})
	}
}

func Test_onExperimentStepStarted(t *testing.T) {
	eventTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	startedTime := time.Date(2021, 1, 1, 0, 1, 0, 0, time.UTC)
	endedTime := time.Date(2021, 1, 1, 0, 7, 0, 0, time.UTC)
	stepId := uuid.MustParse("ccf6a26e-588f-446e-8eaa-d16b086e150e")

	type args struct {
		stepEvent   event_kit_api.EventRequestBody
		targetEvent event_kit_api.EventRequestBody
	}
	tests := []struct {
		name string
		args args
		want *datadogV1.EventCreateRequest
	}{
		{
			name: "should emit event for experiment target started",
			args: args{
				stepEvent: event_kit_api.EventRequestBody{
					Environment: extutil.Ptr(event_kit_api.Environment{
						Id:   "test",
						Name: "gateway",
					}),
					EventName: "experiment.step.started",
					EventTime: eventTime,
					Id:        stepId,
					ExperimentStepExecution: extutil.Ptr(event_kit_api.ExperimentStepExecution{
						ExecutionId:   42,
						ExperimentKey: "ExperimentKey",
						Id:            stepId,
						ActionId:      extutil.Ptr("some_action_id"),
						ActionName:    extutil.Ptr("started step"),
						CustomLabel:   extutil.Ptr("custom label"),
						ActionKind:    extutil.Ptr(event_kit_api.Attack),
					}),
					Tenant: event_kit_api.Tenant{
						Key:  "key",
						Name: "name",
					},
				},
				targetEvent: event_kit_api.EventRequestBody{
					Environment: extutil.Ptr(event_kit_api.Environment{
						Id:   "test",
						Name: "gateway",
					}),
					EventName: "experiment.step.target.started",
					EventTime: eventTime,
					Id:        stepId,
					ExperimentStepTargetExecution: extutil.Ptr(event_kit_api.ExperimentStepTargetExecution{
						ExecutionId:     42,
						ExperimentKey:   "ExperimentKey",
						StepExecutionId: stepId,
						State:           "completed",
						TargetType:      "type",
						TargetName:      "test",
						StartedTime:     &startedTime,
						EndedTime:       &endedTime,
					}),
					Tenant: event_kit_api.Tenant{
						Key:  "key",
						Name: "name",
					},
				},
			},
			want: &datadogV1.EventCreateRequest{
				AggregationKey: extutil.Ptr("steadybit-execution-42"),
				SourceTypeName: extutil.Ptr("Steadybit"),
				Tags: []string{
					"step_action_name:started step",
					"step_custom_label:custom label",
					"source:Steadybit",
					"environment_name:gateway",
					"event_name:experiment.step.target.started",
					"event_time:2021-01-01 00:00:00 +0000 UTC",
					"event_id:ccf6a26e-588f-446e-8eaa-d16b086e150e",
					"tenant_name:name",
					"tenant_key:key",
					"execution_id:42",
					"experiment_key:ExperimentKey",
					"execution_state:completed",
					"started_time:2021-01-01T00:01:00Z",
					"ended_time:2021-01-01T00:07:00Z",
				},
				Title:        "Experiment 'ExperimentKey' - Attack started",
				Text:         "%%% \nExperiment `ExperimentKey` (execution `42`) - Attack `custom label` started.\n\nTarget:test\n %%%",
				DateHappened: extutil.Ptr(eventTime.Unix()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := onExperimentStepStarted(tt.args.stepEvent)
			require.NoError(t, err)
			got, err := onExperimentTargetStarted(tt.args.targetEvent)
			require.NoError(t, err)
			assert.Equalf(t, tt.want.Tags, got.Tags, "onExperimentTargetStarted - Tags different")
			assert.Equalf(t, tt.want.Text, got.Text, "onExperimentTargetStarted - Text different")
			assert.Equalf(t, tt.want, got, "onExperimentTargetStarted - Something else ist different, take a very close look ;-)")
		})
	}
}
