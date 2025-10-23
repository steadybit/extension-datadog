// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extmonitor

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/steadybit/extension-datadog/config"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type datadogClientMock struct {
	mock.Mock
}

func (m *datadogClientMock) ListMonitors(ctx context.Context, params datadogV1.ListMonitorsOptionalParameters) ([]datadogV1.Monitor, *http.Response, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]datadogV1.Monitor), args.Get(1).(*http.Response), args.Error(2)
}

func getPageMatcher(page int) interface{} {
	return mock.MatchedBy(func(params datadogV1.ListMonitorsOptionalParameters) bool {
		return *params.Page == int64(page)
	})
}

func TestIterateThroughMonitorsResponses(t *testing.T) {
	// Given
	mockedApi := new(datadogClientMock)
	page1 := []datadogV1.Monitor{
		{
			Id:    extutil.Ptr(int64(42)),
			Name:  extutil.Ptr("Test-42"),
			Tags:  []string{"tagA", "tagB"},
			Multi: extutil.Ptr(false),
		},
	}
	page2 := []datadogV1.Monitor{
		{
			Id:    extutil.Ptr(int64(69)),
			Name:  extutil.Ptr("Test-69"),
			Tags:  []string{"tagB", "tagC"},
			Multi: extutil.Ptr(true),
		},
	}
	page3 := []datadogV1.Monitor{}
	okResponse := http.Response{
		StatusCode: 200,
	}
	mockedApi.On("ListMonitors", mock.Anything, getPageMatcher(0)).Return(page1, &okResponse, nil)
	mockedApi.On("ListMonitors", mock.Anything, getPageMatcher(1)).Return(page2, &okResponse, nil)
	mockedApi.On("ListMonitors", mock.Anything, getPageMatcher(2)).Return(page3, &okResponse, nil)

	// When
	monitors := getAllMonitors(context.Background(), mockedApi, "https://app.datadoghq.eu")

	// Then
	require.Len(t, monitors, 2)
	require.Equal(t, "42", monitors[0].Id)
	require.Equal(t, "Test-42", monitors[0].Label)
	require.Equal(t, []string{"tagA", "tagB"}, monitors[0].Attributes["datadog.monitor.tags"])
	require.Equal(t, []string{"false"}, monitors[0].Attributes["datadog.monitor.multi-alert"])
	require.Equal(t, []string{"https://app.datadoghq.eu/monitors/42"}, monitors[0].Attributes["steadybit.url"])
	require.Equal(t, "69", monitors[1].Id)
	require.Equal(t, "Test-69", monitors[1].Label)
	require.Equal(t, []string{"tagB", "tagC"}, monitors[1].Attributes["datadog.monitor.tags"])
	require.Equal(t, []string{"true"}, monitors[1].Attributes["datadog.monitor.multi-alert"])
	mockedApi.AssertNumberOfCalls(t, "ListMonitors", 3)
}

func TestErrorResponseReturnsIntermediateResult(t *testing.T) {
	// Given
	mockedApi := new(datadogClientMock)
	page1 := []datadogV1.Monitor{
		{
			Id:    extutil.Ptr(int64(42)),
			Name:  extutil.Ptr("Test-42"),
			Tags:  []string{"tagA", "tagB"},
			Multi: extutil.Ptr(true),
		},
	}
	okResponse := http.Response{
		StatusCode: 200,
	}
	mockedApi.On("ListMonitors", mock.Anything, getPageMatcher(0)).Return(page1, &okResponse, nil)
	mockedApi.On("ListMonitors", mock.Anything, getPageMatcher(1)).Return([]datadogV1.Monitor{}, &okResponse, fmt.Errorf("Intentional Test error"))

	// When
	monitors := getAllMonitors(context.Background(), mockedApi, "https://app.datadoghq.eu")

	// Then
	require.Len(t, monitors, 1)
	require.Equal(t, "42", monitors[0].Id)
	require.Equal(t, []string{"tagA", "tagB"}, monitors[0].Attributes["datadog.monitor.tags"])
	mockedApi.AssertNumberOfCalls(t, "ListMonitors", 2)
}

func TestExlcudeAttributes(t *testing.T) {
	// Given
	config.Config.DiscoveryAttributesExcludesMonitor = []string{"datadog.monitor.tags"}
	mockedApi := new(datadogClientMock)
	page1 := []datadogV1.Monitor{
		{
			Id:    extutil.Ptr(int64(42)),
			Name:  extutil.Ptr("Test-42"),
			Tags:  []string{"tagA", "tagB"},
			Multi: extutil.Ptr(true),
		},
	}
	page2 := []datadogV1.Monitor{}
	okResponse := http.Response{
		StatusCode: 200,
	}
	mockedApi.On("ListMonitors", mock.Anything, getPageMatcher(0)).Return(page1, &okResponse, nil)
	mockedApi.On("ListMonitors", mock.Anything, getPageMatcher(1)).Return(page2, &okResponse, nil)

	// When
	monitors := getAllMonitors(context.Background(), mockedApi, "https://app.datadoghq.eu")

	// Then
	require.Len(t, monitors, 1)
	require.Equal(t, "42", monitors[0].Id)
	require.NotContains(t, monitors[0].Attributes, "datadog.monitor.tags")
	mockedApi.AssertNumberOfCalls(t, "ListMonitors", 2)
}
