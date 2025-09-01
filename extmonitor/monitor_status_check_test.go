// Copyright 2025 steadybit GmbH. All rights reserved.

package extmonitor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type datadogGetMonitorClientMock struct {
	mock.Mock
}

func (m *datadogGetMonitorClientMock) GetMonitor(ctx context.Context, monitorId int64, params datadogV1.GetMonitorOptionalParameters) (datadogV1.Monitor, *http.Response, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(datadogV1.Monitor), args.Get(1).(*http.Response), args.Error(2)
}

func TestPrepareExtractsState(t *testing.T) {
	// Given
	request := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"duration":           1000 * 60,
			"expectedStatus":     datadogV1.MONITOROVERALLSTATES_OK,
			"expectedStatusList": []string{string(datadogV1.MONITOROVERALLSTATES_OK), string(datadogV1.MONITOROVERALLSTATES_NO_DATA)},
			"statusCheckMode":    statusCheckModeAtLeastOnce,
		},
		Target: extutil.Ptr(action_kit_api.Target{
			Attributes: map[string][]string{
				"datadog.monitor.id": {"12349876"},
			},
		}),
	})
	attack := MonitorStatusCheckAction{}
	state := attack.NewEmptyState()

	// When
	result, err := attack.Prepare(context.TODO(), &state, request)

	// Then
	require.Nil(t, result)
	require.Nil(t, err)
	require.Equal(t, int64(12349876), state.MonitorId)
	require.NotNil(t, state.Start)
	require.True(t, state.End.After(time.Now()))
	require.Equal(t, []string{"OK", "No Data"}, state.ExpectedStatus)
	require.Equal(t, statusCheckModeAtLeastOnce, state.StatusCheckMode)
}

func TestPrepareExtractsStateWithoutStatusCheck(t *testing.T) {
	// Given
	request := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"duration": 1000 * 60,
		},
		Target: extutil.Ptr(action_kit_api.Target{
			Attributes: map[string][]string{
				"datadog.monitor.id": {"12349876"},
			},
		}),
	})
	attack := MonitorStatusCheckAction{}
	state := attack.NewEmptyState()

	// When
	result, err := attack.Prepare(context.TODO(), &state, request)

	// Then
	require.Nil(t, result)
	require.Nil(t, err)
	require.Equal(t, int64(12349876), state.MonitorId)
	require.NotNil(t, state.Start)
	require.True(t, state.End.After(time.Now()))
	require.Equal(t, []string{}, state.ExpectedStatus)
	require.Equal(t, state.StatusCheckMode, "")
}

func TestPrepareExtractsStateDeprecatedExpextedStatus(t *testing.T) {
	// Given
	request := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"duration":        1000 * 60,
			"expectedStatus":  datadogV1.MONITOROVERALLSTATES_OK,
			"statusCheckMode": statusCheckModeAtLeastOnce,
		},
		Target: extutil.Ptr(action_kit_api.Target{
			Attributes: map[string][]string{
				"datadog.monitor.id": {"12349876"},
			},
		}),
	})
	attack := MonitorStatusCheckAction{}
	state := attack.NewEmptyState()

	// When
	result, err := attack.Prepare(context.TODO(), &state, request)

	// Then
	require.Nil(t, result)
	require.Nil(t, err)
	require.Equal(t, int64(12349876), state.MonitorId)
	require.NotNil(t, state.Start)
	require.True(t, state.End.After(time.Now()))
	require.Equal(t, []string{"OK"}, state.ExpectedStatus)
	require.Equal(t, statusCheckModeAtLeastOnce, state.StatusCheckMode)
}

func TestPrepareSupportsMissingExpectedStatus(t *testing.T) {
	// Given
	request := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"duration": 1000 * 60,
		},
		Target: extutil.Ptr(action_kit_api.Target{
			Attributes: map[string][]string{
				"datadog.monitor.id": {"12349876"},
			},
		}),
	})
	action := MonitorStatusCheckAction{}
	state := action.NewEmptyState()

	// When
	result, err := action.Prepare(context.TODO(), &state, request)

	// Then
	require.Nil(t, result)
	require.Nil(t, err)
	require.Equal(t, int64(12349876), state.MonitorId)
	require.True(t, state.End.After(time.Now()))
	require.Empty(t, state.ExpectedStatus)
}

func TestPrepareReportsMonitorIdProblems(t *testing.T) {
	// Given
	request := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"duration": 1000 * 60,
		},
		Target: extutil.Ptr(action_kit_api.Target{
			Attributes: map[string][]string{
				"datadog.monitor.id": {"NOT AN INT"},
			},
		}),
	})
	action := MonitorStatusCheckAction{}
	state := action.NewEmptyState()

	// When
	result, err := action.Prepare(context.TODO(), &state, request)

	// Then
	require.Nil(t, result)
	require.Equal(t, "Failed to parse monitor ID 'NOT AN INT' as int64.", err.Error())
}

func TestStatusReportsIssuesOnMissingMonitor(t *testing.T) {
	// Given
	mockedApi := new(datadogGetMonitorClientMock)
	mockedApi.On("GetMonitor", mock.Anything, mock.Anything, mock.Anything).Return(datadogV1.Monitor{}, extutil.Ptr(http.Response{
		StatusCode: 200,
	}), fmt.Errorf("intentional error"))

	action := MonitorStatusCheckAction{}
	state := action.NewEmptyState()
	state.MonitorId = 1234
	state.End = time.Now().Add(time.Minute)
	state.ExpectedStatus = []string{string(datadogV1.MONITOROVERALLSTATES_OK)}

	// When
	result, err := monitorStatusCheckStatus(context.Background(), &state, mockedApi, "http://example.com")

	// Then
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Failed to retrieve monitor 1234 from Datadog")
}

func TestAllTheTimeSuccess(t *testing.T) {
	// Given
	mockedApi := new(datadogGetMonitorClientMock)
	mockedApi.On("GetMonitor", mock.Anything, mock.Anything, mock.Anything).Return(datadogV1.Monitor{
		Name:         extutil.Ptr("gateway pods ready"),
		Id:           extutil.Ptr(int64(1234)),
		OverallState: extutil.Ptr(datadogV1.MONITOROVERALLSTATES_WARN),
	}, extutil.Ptr(http.Response{
		StatusCode: 200,
	}), nil)

	action := MonitorStatusCheckAction{}
	state := action.NewEmptyState()
	state.MonitorId = 1234
	state.End = time.Now().Add(time.Minute * -1)
	state.ExpectedStatus = []string{string(datadogV1.MONITOROVERALLSTATES_WARN)}
	state.StatusCheckMode = statusCheckModeAllTheTime

	// When
	result, err := monitorStatusCheckStatus(context.Background(), &state, mockedApi, "https://example.com")

	// Then
	require.Nil(t, result.State)
	require.Nil(t, err)
	require.True(t, result.Completed)
	require.Nil(t, result.Error)
	require.Len(t, *result.Metrics, 1)
}

func TestAllTheTimeExpectationMismatch(t *testing.T) {
	// Given
	mockedApi := new(datadogGetMonitorClientMock)
	mockedApi.On("GetMonitor", mock.Anything, mock.Anything, mock.Anything).Return(datadogV1.Monitor{
		Name:         extutil.Ptr("gateway pods ready"),
		Id:           extutil.Ptr(int64(1234)),
		OverallState: extutil.Ptr(datadogV1.MONITOROVERALLSTATES_WARN),
	}, extutil.Ptr(http.Response{
		StatusCode: 200,
	}), nil)

	action := MonitorStatusCheckAction{}
	state := action.NewEmptyState()
	state.MonitorId = 1234
	state.End = time.Now().Add(time.Minute * 1) // time not yet up - early exit
	state.ExpectedStatus = []string{string(datadogV1.MONITOROVERALLSTATES_OK)}
	state.StatusCheckMode = statusCheckModeAllTheTime

	// When
	result, err := monitorStatusCheckStatus(context.Background(), &state, mockedApi, "http://example.com")

	// Then
	require.Nil(t, err)
	require.Nil(t, result.State)
	require.False(t, result.Completed)
	require.Equal(t, "Monitor 'gateway pods ready' (id 1234, tags: <none>) has status 'Warn' whereas '[OK]' is expected.", result.Error.Title)
	require.Contains(t, *result.Error.Status, action_kit_api.Failed)
	metric := (*result.Metrics)[0]
	require.Equal(t, "datadog_monitor_status", *metric.Name)
	require.Equal(t, "1234", metric.Metric["datadog.monitor.id"])
	require.Equal(t, "gateway pods ready", metric.Metric["datadog.monitor.name"])
	require.NotNil(t, metric.Timestamp)
	require.Equal(t, float64(0), metric.Value)
}

func TestAtLeastOnceSuccess(t *testing.T) {
	// ----------------------------------------
	// First Call: Status is not ok - StatusCheckSuccess in State is still false - no exit (End not yet reached)
	// ----------------------------------------
	// Given
	mockedApi := new(datadogGetMonitorClientMock)
	mockedApi.On("GetMonitor", mock.Anything, mock.Anything, mock.Anything).Return(datadogV1.Monitor{
		Name:         extutil.Ptr("gateway pods ready"),
		Id:           extutil.Ptr(int64(1234)),
		OverallState: extutil.Ptr(datadogV1.MONITOROVERALLSTATES_WARN),
	}, extutil.Ptr(http.Response{
		StatusCode: 200,
	}), nil).Once()

	action := MonitorStatusCheckAction{}
	state := action.NewEmptyState()
	state.MonitorId = 1234
	state.End = time.Now().Add(time.Minute * 1) //time not yet up - no early exit if status is ok at least once
	state.ExpectedStatus = []string{string(datadogV1.MONITOROVERALLSTATES_OK)}
	state.StatusCheckMode = statusCheckModeAtLeastOnce

	// When
	result, err := monitorStatusCheckStatus(context.Background(), &state, mockedApi, "http://example.com")

	// Then
	require.Nil(t, err)
	require.Nil(t, result.State)
	require.False(t, result.Completed)
	require.False(t, state.StatusCheckSuccess)
	require.Nil(t, result.Error)
	metric := (*result.Metrics)[0]
	require.Equal(t, "datadog_monitor_status", *metric.Name)
	require.Equal(t, "1234", metric.Metric["datadog.monitor.id"])
	require.Equal(t, "gateway pods ready", metric.Metric["datadog.monitor.name"])
	require.NotNil(t, metric.Timestamp)
	require.Equal(t, float64(0), metric.Value)

	// ----------------------------------------
	// Second Call: Status is ok - StatusCheckSuccess in State is true - no exit (End not yet reached)
	// ----------------------------------------
	// Given
	mockedApi.On("GetMonitor", mock.Anything, mock.Anything, mock.Anything).Return(datadogV1.Monitor{
		Name:         extutil.Ptr("gateway pods ready"),
		Id:           extutil.Ptr(int64(1234)),
		OverallState: extutil.Ptr(datadogV1.MONITOROVERALLSTATES_OK),
	}, extutil.Ptr(http.Response{
		StatusCode: 200,
	}), nil).Once()

	// When
	result, err = monitorStatusCheckStatus(context.Background(), &state, mockedApi, "http://example.com")

	// Then
	require.Nil(t, err)
	require.Nil(t, result.State)
	require.False(t, result.Completed)
	require.True(t, state.StatusCheckSuccess)
	require.Nil(t, result.Error)
	require.Len(t, *result.Metrics, 1)

	// ----------------------------------------
	// Thirds Call: Status is not ok - but StatusCheckSuccess in State was true (call 2) - successfully exit (End reached)
	// ----------------------------------------
	//Given
	mockedApi.On("GetMonitor", mock.Anything, mock.Anything, mock.Anything).Return(datadogV1.Monitor{
		Name:         extutil.Ptr("gateway pods ready"),
		Id:           extutil.Ptr(int64(1234)),
		OverallState: extutil.Ptr(datadogV1.MONITOROVERALLSTATES_WARN),
	}, extutil.Ptr(http.Response{
		StatusCode: 200,
	}), nil).Once()
	state.End = time.Now().Add(time.Minute * -1) //Simulate that the time has passed

	// When
	result, err = monitorStatusCheckStatus(context.Background(), &state, mockedApi, "http://example.com")

	// Then
	require.Nil(t, err)
	require.Nil(t, result.State)
	require.True(t, result.Completed)
	require.Nil(t, result.Error)
	require.Len(t, *result.Metrics, 1)
}

func TestAtLeastOnceExpectationMismatch(t *testing.T) {
	// Given
	mockedApi := new(datadogGetMonitorClientMock)
	mockedApi.On("GetMonitor", mock.Anything, mock.Anything, mock.Anything).Return(datadogV1.Monitor{
		Name:         extutil.Ptr("gateway pods ready"),
		Id:           extutil.Ptr(int64(1234)),
		OverallState: extutil.Ptr(datadogV1.MONITOROVERALLSTATES_WARN),
	}, extutil.Ptr(http.Response{
		StatusCode: 200,
	}), nil).Once()

	action := MonitorStatusCheckAction{}
	state := action.NewEmptyState()
	state.MonitorId = 1234
	state.End = time.Now().Add(time.Minute * -1) //Simulate that the time has passed
	state.ExpectedStatus = []string{string(datadogV1.MONITOROVERALLSTATES_OK)}
	state.StatusCheckMode = statusCheckModeAtLeastOnce

	// When
	result, err := monitorStatusCheckStatus(context.Background(), &state, mockedApi, "http://example.com")

	// Then
	require.Nil(t, err)
	require.Nil(t, result.State)
	require.True(t, result.Completed)
	require.Equal(t, "Monitor 'gateway pods ready' (id 1234, tags: <none>) didn't have status '[OK]' at least once.", result.Error.Title)
	require.Contains(t, *result.Error.Status, action_kit_api.Failed)
	metric := (*result.Metrics)[0]
	require.Equal(t, "datadog_monitor_status", *metric.Name)
	require.Equal(t, "1234", metric.Metric["datadog.monitor.id"])
	require.Equal(t, "gateway pods ready", metric.Metric["datadog.monitor.name"])
	require.NotNil(t, metric.Timestamp)
	require.Equal(t, float64(0), metric.Value)
}

func Test_Should_Retry_On_Error(t *testing.T) {
	// Given
	mockedApi := new(datadogGetMonitorClientMock)
	mockedApi.On("GetMonitor", mock.Anything, mock.Anything, mock.Anything).
		Return(datadogV1.Monitor{}, extutil.Ptr(http.Response{StatusCode: 500}), errors.New("500 Internal Server Error")).Once().
		Return(datadogV1.Monitor{}, extutil.Ptr(http.Response{StatusCode: 500}), errors.New("500 Internal Server Error")).Once().
		Return(datadogV1.Monitor{
			Name:         extutil.Ptr("gateway pods ready"),
			Id:           extutil.Ptr(int64(1234)),
			OverallState: extutil.Ptr(datadogV1.MONITOROVERALLSTATES_WARN),
		},
			extutil.Ptr(http.Response{StatusCode: 200}),
			nil).Once()

	action := MonitorStatusCheckAction{}
	state := action.NewEmptyState()
	state.MonitorId = 1234
	state.End = time.Now().Add(time.Minute * -1) //Simulate that the time has passed
	state.ExpectedStatus = []string{string(datadogV1.MONITOROVERALLSTATES_OK)}
	state.StatusCheckMode = statusCheckModeAtLeastOnce

	// When
	result, err := monitorStatusCheckStatus(context.Background(), &state, mockedApi, "http://example.com")

	// Then
	require.Nil(t, err)
	require.Nil(t, result.State)
	require.True(t, result.Completed)
}

func TestCreateMetric(t *testing.T) {
	// Given
	now := time.Now()
	siteUrl := "https://app.datadoghq.eu"
	monitor := extutil.Ptr(datadogV1.Monitor{
		Id:           extutil.Ptr(int64(42)),
		Name:         extutil.Ptr("gateway readiness"),
		OverallState: extutil.Ptr(datadogV1.MONITOROVERALLSTATES_ALERT),
	})

	// When
	start := time.Date(2023, 7, 4, 13, 0, 0, 0, time.UTC)
	end := time.Date(2023, 7, 4, 13, 10, 0, 0, time.UTC)
	metric := toMetric(monitor, now, start, end, siteUrl)

	// Then
	require.Equal(t, "datadog_monitor_status", *metric.Name)
	require.Equal(t, float64(0), metric.Value)
	require.Equal(t, now, metric.Timestamp)
	require.Equal(t, "42", metric.Metric["datadog.monitor.id"])
	require.Equal(t, "gateway readiness", metric.Metric["datadog.monitor.name"])
	require.Equal(t, "danger", metric.Metric["state"])
	require.Equal(t, "Monitor status is: Alert", metric.Metric["tooltip"])
	require.Equal(t, "https://app.datadoghq.eu/monitors/42?from_ts=1688475600000&to_ts=1688476200000", metric.Metric["url"])
}

func TestCreateMetricForUnknownState(t *testing.T) {
	// Given
	now := time.Now()
	siteUrl := "https://app.datadoghq.eu"
	monitor := extutil.Ptr(datadogV1.Monitor{
		Id:           extutil.Ptr(int64(42)),
		Name:         extutil.Ptr("gateway readiness"),
		OverallState: nil,
	})

	// When
	start := time.Date(2023, 7, 4, 13, 0, 0, 0, time.UTC)
	end := time.Date(2023, 7, 4, 13, 10, 0, 0, time.UTC)
	metric := toMetric(monitor, now, start, end, siteUrl)

	// Then
	require.Equal(t, "datadog_monitor_status", *metric.Name)
	require.Equal(t, float64(0), metric.Value)
	require.Equal(t, now, metric.Timestamp)
	require.Equal(t, "42", metric.Metric["datadog.monitor.id"])
	require.Equal(t, "gateway readiness", metric.Metric["datadog.monitor.name"])
	require.Equal(t, "warn", metric.Metric["state"])
	require.Equal(t, "Monitor status is: Unknown", metric.Metric["tooltip"])
	require.Equal(t, "https://app.datadoghq.eu/monitors/42?from_ts=1688475600000&to_ts=1688476200000", metric.Metric["url"])
}
