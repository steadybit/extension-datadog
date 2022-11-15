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
						State:                event_kit_api.Created,
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
				"execution_id:42.000000",
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
						State:                event_kit_api.Created,
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
				"execution_id:42.000000",
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
						State:                event_kit_api.Created,
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
				"execution_id:42.000000",
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
