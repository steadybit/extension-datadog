package extmonitor

import (
	"context"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

type datadogDowntimeClientMock struct {
	mock.Mock
}

func (m *datadogDowntimeClientMock) CreateDowntime(ctx context.Context, downtimeBody datadogV2.DowntimeCreateRequest) (datadogV2.DowntimeResponse, *http.Response, error) {
	args := m.Called(ctx, downtimeBody)
	return args.Get(0).(datadogV2.DowntimeResponse), args.Get(1).(*http.Response), args.Error(2)
}

func (m *datadogDowntimeClientMock) CancelDowntime(ctx context.Context, downtimeId string) (*http.Response, error) {
	args := m.Called(ctx, downtimeId)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestMonitorDowntimePrepareExtractsState(t *testing.T) {
	// Given
	request := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"duration": 1000 * 60,
			"notify":   true,
		},
		Target: extutil.Ptr(action_kit_api.Target{
			Attributes: map[string][]string{
				"datadog.monitor.id": {"12349876"},
			},
		}),
		ExecutionContext: extutil.Ptr(action_kit_api.ExecutionContext{
			ExperimentUri: extutil.Ptr("<uri-to-experiment>"),
			ExecutionUri:  extutil.Ptr("<uri-to-execution>"),
		}),
	})
	action := MonitorDowntimeAction{}
	state := action.NewEmptyState()

	// When
	result, err := action.Prepare(context.TODO(), &state, request)

	// Then
	require.Nil(t, result)
	require.Nil(t, err)
	require.Equal(t, int64(12349876), state.MonitorId)
	require.Equal(t, *state.ExperimentUri, "<uri-to-experiment>")
	require.Equal(t, *state.ExecutionUri, "<uri-to-execution>")
	require.True(t, state.End.After(time.Now()))
	require.True(t, state.Notify)
}

func TestMonitorDowntimeStartSuccess(t *testing.T) {
	// Given
	mockedApi := new(datadogDowntimeClientMock)
	mockedApi.On("CreateDowntime", mock.Anything, mock.Anything, mock.Anything).Return(datadogV2.DowntimeResponse{
		Data: &datadogV2.DowntimeResponseData{
			Id: extutil.Ptr("4711"),
		},
	}, extutil.Ptr(http.Response{
		StatusCode: 200,
	}), nil).Once()

	action := MonitorDowntimeAction{}
	state := action.NewEmptyState()
	state.MonitorId = 1234
	state.End = time.Now().Add(time.Minute)
	state.Notify = true
	state.ExecutionUri = extutil.Ptr("<uri-to-execution>")
	state.ExperimentUri = extutil.Ptr("<uri-to-experiment>")

	// When
	result, err := MonitorDowntimeStart(context.Background(), &state, mockedApi)

	// Then
	require.Nil(t, err)
	require.Nil(t, result.State)
	require.Equal(t, "4711", *state.DowntimeId)
	require.Equal(t, "Downtime started. (monitor 1234, downtime 4711)", (*result.Messages)[0].Message)
}

func TestMonitorDowntimeStopSuccess(t *testing.T) {
	// Given
	mockedApi := new(datadogDowntimeClientMock)
	mockedApi.On("CancelDowntime", mock.Anything, mock.Anything, mock.Anything).Return(extutil.Ptr(http.Response{
		StatusCode: 200,
	}), nil).Once()

	action := MonitorDowntimeAction{}
	state := action.NewEmptyState()
	state.MonitorId = 1234
	state.End = time.Now().Add(time.Minute)
	state.Notify = true
	state.DowntimeId = extutil.Ptr("4711")

	// When
	result, err := MonitorDowntimeStop(context.Background(), &state, mockedApi)

	// Then
	require.Nil(t, err)
	require.Equal(t, "Downtime canceled. (monitor 1234, downtime 4711)", (*result.Messages)[0].Message)
}
