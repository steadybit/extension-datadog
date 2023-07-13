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
	"github.com/stretchr/testify/mock"
	"net/http"
	"reflect"
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

func TestSendDatadogEvent(t *testing.T) {
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
	SendDatadogEvent(context.Background(), mockedApi, datadogEventBody)

	// Then
	mockedApi.AssertNumberOfCalls(t, "SendEvent", 1)
}

func Test_convertSteadybitEventToDataDogEventTags(t *testing.T) {
	type args struct {
		w     http.ResponseWriter
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
			name: "Successfully convert event to datadog tags",
			args: args{
				event: event_kit_api.EventRequestBody{
					Environment: extutil.Ptr(event_kit_api.Environment{
						Id:   "test",
						Name: "gateway",
					}),
					EventName: "experiment.started",
					EventTime: eventTime,
					ExperimentExecution: extutil.Ptr(event_kit_api.ExperimentExecution{
						EndedTime:            extutil.Ptr(endedTime),
						ExecutionId:          42,
						ExperimentKey:        "ExperimentKey",
						FailureReason:        extutil.Ptr("FailureReason"),
						FailureReasonDetails: extutil.Ptr("FailureReasonDetails"),
						Hypothesis:           "Hypothesis",
						Name:                 "Name",
						PreparedTime:         eventTime,
						StartedTime:          startedTime,
						State:                event_kit_api.ExperimentExecutionStateCreated,
					}),
					Id: uuid.MustParse("ccf6a26e-588f-446e-8eaa-d16b086e150e"),
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
				"team_name:gateway",
				"team_key:test",
				"tenant_name:name",
				"tenant_key:key",
				"execution_id:42",
				"experiment_key:ExperimentKey",
				"experiment_name:Name",
				"execution_state:created",
				"principal_type:user",
				"principal_username:Pan",
				"principal_name:Peter",
				"experiment_hypothesis:Hypothesis",
				"started_time:" + startedTime.Format(time.RFC3339),
				"ended_time:" + endedTime.Format(time.RFC3339)},
		},
		{
			name: "Successfully convert event to datadog tags without Principal",
			args: args{
				event: event_kit_api.EventRequestBody{
					Environment: extutil.Ptr(event_kit_api.Environment{
						Id:   "test",
						Name: "gateway",
					}),
					EventName: "experiment.started",
					EventTime: eventTime,
					ExperimentExecution: extutil.Ptr(event_kit_api.ExperimentExecution{
						EndedTime:            extutil.Ptr(endedTime),
						ExecutionId:          42,
						ExperimentKey:        "ExperimentKey",
						FailureReason:        extutil.Ptr("FailureReason"),
						FailureReasonDetails: extutil.Ptr("FailureReasonDetails"),
						Hypothesis:           "Hypothesis",
						Name:                 "Name",
						PreparedTime:         eventTime,
						StartedTime:          startedTime,
						State:                event_kit_api.ExperimentExecutionStateCreated,
					}),
					Id: uuid.MustParse("ccf6a26e-588f-446e-8eaa-d16b086e150e"),
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
				"team_name:gateway",
				"team_key:test",
				"tenant_name:name",
				"tenant_key:key",
				"execution_id:42",
				"experiment_key:ExperimentKey",
				"experiment_name:Name",
				"execution_state:created",
				"principal_type:access_token",
				"principal_name:MyFancyToken",
				"experiment_hypothesis:Hypothesis",
				"started_time:" + startedTime.Format(time.RFC3339),
				"ended_time:" + endedTime.Format(time.RFC3339)},
		},
		{
			name: "Successfully convert event to datadog tags without hypothesis",
			args: args{
				event: event_kit_api.EventRequestBody{
					Environment: extutil.Ptr(event_kit_api.Environment{
						Id:   "test",
						Name: "gateway",
					}),
					EventName: "experiment.started",
					EventTime: eventTime,
					ExperimentExecution: extutil.Ptr(event_kit_api.ExperimentExecution{
						EndedTime:            extutil.Ptr(endedTime),
						ExecutionId:          42,
						ExperimentKey:        "ExperimentKey",
						FailureReason:        extutil.Ptr("FailureReason"),
						FailureReasonDetails: extutil.Ptr("FailureReasonDetails"),
						Hypothesis:           "",
						Name:                 "Name",
						PreparedTime:         eventTime,
						StartedTime:          startedTime,
						State:                event_kit_api.ExperimentExecutionStateCreated,
					}),
					Id:        uuid.MustParse("ccf6a26e-588f-446e-8eaa-d16b086e150e"),
					Principal: nil,
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
				"team_name:gateway",
				"team_key:test",
				"tenant_name:name",
				"tenant_key:key",
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
			if got := convertSteadybitEventToDataDogEventTags(tt.args.event); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertSteadybitEventToDataDogEventTags() = %v, want %v", got, tt.want)
			}
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
					ActionId:    extutil.Ptr("com.github.steadybit.action.example"),
					ActionName:  extutil.Ptr("example-action"),
					ActionKind:  extutil.Ptr(event_kit_api.Attack),
					CustomLabel: extutil.Ptr("My very own label"),
					State:       event_kit_api.ExperimentStepExecutionStateFailed,
					EndedTime:   extutil.Ptr(endedTime),
					StartedTime: extutil.Ptr(startedTime),
				},
			},
			want: []string{
				"step_state:failed",
				"step_started_time:2021-01-01T00:01:00Z",
				"step_ended_time:2021-01-01T00:02:00Z",
				"step_action_id:com.github.steadybit.action.example",
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
					ActionId:   extutil.Ptr("com.github.steadybit.action.example"),
					ActionKind: extutil.Ptr(event_kit_api.Attack),
					State:      event_kit_api.ExperimentStepExecutionStateCompleted,
				},
			},
			want: []string{
				"step_state:completed",
				"step_action_id:com.github.steadybit.action.example",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStepTags(tt.args.stepExecution); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getStepTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTargetTags(t *testing.T) {
	type args struct {
		w      http.ResponseWriter
		target event_kit_api.ExperimentStepExecutionTarget
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Successfully get tag for container targets",
			args: args{
				target: event_kit_api.ExperimentStepExecutionTarget{
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
					TargetName: "Container",
					TargetType: "container",
				},
			},
			want: []string{
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
				target: event_kit_api.ExperimentStepExecutionTarget{
					AgentHostname: "Agent-1",
					TargetAttributes: map[string][]string{
						"k8s.cluster-name":                                {"dev-cluster"},
						"k8s.pod.label.tags.datadoghq.com/service":        {"service-1"},
						"k8s.deployment.label.tags.datadoghq.com/service": {"service-1"},
					},
					TargetName: "Container",
					TargetType: "container",
				},
			},
			want: []string{
				"kube_cluster_name:dev-cluster",
				"cluster_name:dev-cluster",
				"service:service-1",
			},
		},
		{
			name: "Should add cluster name to hostname and deduplicate host names",
			args: args{
				target: event_kit_api.ExperimentStepExecutionTarget{
					AgentHostname: "Agent-1",
					TargetAttributes: map[string][]string{
						"k8s.cluster-name":     {"dev-cluster"},
						"container.host":       {"host-123"},
						"host.hostname":        {"host-123"},
						"application.hostname": {"host-123"},
					},
					TargetName: "Container",
					TargetType: "container",
				},
			},
			want: []string{
				"kube_cluster_name:dev-cluster",
				"cluster_name:dev-cluster",
				"host:host-123-dev-cluster",
			},
		},
		{
			name: "Ignore multiple values",
			args: args{
				target: event_kit_api.ExperimentStepExecutionTarget{
					AgentHostname: "Agent-1",
					TargetAttributes: map[string][]string{
						"host.hostname": {"Host-1"},
						"k8s.namespace": {"namespace-1", "namespace-2", "namespace-3"},
					},
					TargetName: "Host",
					TargetType: "host",
				},
			},
			want: []string{
				"host:Host-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTargetTags(tt.args.target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTargetTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
