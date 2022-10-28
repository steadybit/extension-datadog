// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/extconversion"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

type datadogGetMonitorClientMock struct {
	mock.Mock
}

func (m *datadogGetMonitorClientMock) GetMonitor(ctx context.Context, monitorId int64, params datadogV1.GetMonitorOptionalParameters) (datadogV1.Monitor, *http.Response, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(datadogV1.Monitor), args.Get(1).(*http.Response), args.Error(2)
}

func getStatusRequestBody(t *testing.T, state MonitorStatusCheckState) []byte {
	var encodedState action_kit_api.ActionState
	err := extconversion.Convert(state, &encodedState)
	require.NoError(t, err)
	request := action_kit_api.ActionStatusRequestBody{
		State: encodedState,
	}
	json, err := json.Marshal(request)
	require.NoError(t, err)
	return json
}

func TestPrepareExtractsState(t *testing.T) {
	// Given
	request := action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"duration":       1000 * 60,
			"expectedStatus": datadogV1.MONITOROVERALLSTATES_OK,
		},
		Target: extutil.Ptr(action_kit_api.Target{
			Attributes: map[string][]string{
				"datadog.monitor.id": {"12349876"},
			},
		}),
	}
	json, err := json.Marshal(request)
	require.NoError(t, err)

	// When
	state, extErr := PrepareMonitorStatusCheck(json)

	// Then
	require.Nil(t, extErr)
	require.Equal(t, int64(12349876), state.MonitorId)
	require.True(t, state.End.After(time.Now()))
	require.Equal(t, "OK", state.ExpectedStatus)
}

func TestPrepareSupportsMissingExpectedStatus(t *testing.T) {
	// Given
	request := action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"duration": 1000 * 60,
		},
		Target: extutil.Ptr(action_kit_api.Target{
			Attributes: map[string][]string{
				"datadog.monitor.id": {"12349876"},
			},
		}),
	}
	json, err := json.Marshal(request)
	require.NoError(t, err)

	// When
	state, extErr := PrepareMonitorStatusCheck(json)

	// Then
	require.Nil(t, extErr)
	require.Equal(t, int64(12349876), state.MonitorId)
	require.True(t, state.End.After(time.Now()))
	require.Empty(t, state.ExpectedStatus)
}

func TestPrepareReportsMonitorIdProblems(t *testing.T) {
	// Given
	request := action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"duration": 1000 * 60,
		},
		Target: extutil.Ptr(action_kit_api.Target{
			Attributes: map[string][]string{
				"datadog.monitor.id": {"NOT AN INT"},
			},
		}),
	}
	json, err := json.Marshal(request)
	require.NoError(t, err)

	// When
	state, extErr := PrepareMonitorStatusCheck(json)

	// Then
	require.Nil(t, state)
	require.Equal(t, "Failed to parse monitor ID 'NOT AN INT' as int64.", extErr.Title)
}

func TestStatusReportsIssuesOnMissingMonitor(t *testing.T) {
	// Given
	json := getStatusRequestBody(t, MonitorStatusCheckState{
		MonitorId:      1234,
		End:            time.Now().Add(time.Minute),
		ExpectedStatus: string(datadogV1.MONITOROVERALLSTATES_OK),
	})
	mockedApi := new(datadogGetMonitorClientMock)
	mockedApi.On("GetMonitor", mock.Anything, mock.Anything, mock.Anything).Return(datadogV1.Monitor{}, extutil.Ptr(http.Response{
		StatusCode: 200,
	}), fmt.Errorf("intentional error"))

	// When
	result := MonitorStatusCheckStatus(context.Background(), json, mockedApi)

	// Then
	require.Nil(t, result.State)
	require.False(t, result.Completed)
	require.Nil(t, result.Metrics)
	require.Contains(t, result.Error.Title, "Failed to retrieve monitor 1234 from Datadog")
	require.Contains(t, *result.Error.Status, action_kit_api.Errored)
}

func TestExpectationMismatch(t *testing.T) {
	// Given
	json := getStatusRequestBody(t, MonitorStatusCheckState{
		MonitorId:      1234,
		End:            time.Now().Add(time.Minute * -1),
		ExpectedStatus: string(datadogV1.MONITOROVERALLSTATES_OK),
	})
	mockedApi := new(datadogGetMonitorClientMock)
	mockedApi.On("GetMonitor", mock.Anything, mock.Anything, mock.Anything).Return(datadogV1.Monitor{
		Name:         extutil.Ptr("gateway pods ready"),
		Id:           extutil.Ptr(int64(1234)),
		OverallState: extutil.Ptr(datadogV1.MONITOROVERALLSTATES_WARN),
	}, extutil.Ptr(http.Response{
		StatusCode: 200,
	}), nil)

	// When
	result := MonitorStatusCheckStatus(context.Background(), json, mockedApi)

	// Then
	require.Nil(t, result.State)
	require.True(t, result.Completed)
	require.Equal(t, "Monitor 'gateway pods ready' (id 1234, tags: <none>) has status 'Warn' whereas 'OK' is expected.", result.Error.Title)
	require.Contains(t, *result.Error.Status, action_kit_api.Failed)
	metric := (*result.Metrics)[0]
	require.Equal(t, "datadog_monitor_status", *metric.Name)
	require.Equal(t, "1234", metric.Metric["datadog.monitor.id"])
	require.Equal(t, "gateway pods ready", metric.Metric["datadog.monitor.name"])
	require.NotNil(t, metric.Timestamp)
	require.Equal(t, float64(0), metric.Value)
}

func TestExpectationSuccess(t *testing.T) {
	// Given
	json := getStatusRequestBody(t, MonitorStatusCheckState{
		MonitorId:      1234,
		End:            time.Now().Add(time.Minute * -1),
		ExpectedStatus: string(datadogV1.MONITOROVERALLSTATES_WARN),
	})
	mockedApi := new(datadogGetMonitorClientMock)
	mockedApi.On("GetMonitor", mock.Anything, mock.Anything, mock.Anything).Return(datadogV1.Monitor{
		Name:         extutil.Ptr("gateway pods ready"),
		Id:           extutil.Ptr(int64(1234)),
		OverallState: extutil.Ptr(datadogV1.MONITOROVERALLSTATES_WARN),
	}, extutil.Ptr(http.Response{
		StatusCode: 200,
	}), nil)

	// When
	result := MonitorStatusCheckStatus(context.Background(), json, mockedApi)

	// Then
	require.Nil(t, result.State)
	require.True(t, result.Completed)
	require.Nil(t, result.Error)
	require.Len(t, *result.Metrics, 1)
}
