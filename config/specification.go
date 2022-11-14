// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package config

import (
	"context"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"net/http"
)

type Specification struct {
	// see https://docs.datadoghq.com/getting_started/site/#access-the-datadog-site
	SiteParameter  string `json:"site" split_words:"true" required:"true"`
	SiteUrl        string `json:"site" split_words:"true" required:"true"`
	ApiKey         string `json:"apiKey" split_words:"true" required:"true"`
	ApplicationKey string `json:"applicationKey" split_words:"true" required:"true"`
}

func (s *Specification) WrapContextWithDatadogContextValues(ctx context.Context) context.Context {
	ctx = context.WithValue(
		ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: Config.ApiKey,
			},
			"appKeyAuth": {
				Key: Config.ApplicationKey,
			},
		},
	)

	ctx = context.WithValue(
		ctx,
		datadog.ContextServerVariables,
		map[string]string{
			"site": Config.SiteParameter,
		},
	)

	return ctx
}

func (s *Specification) ValidateCredentials(ctx context.Context) (datadogV1.AuthenticationValidationResponse, *http.Response, error) {
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewAuthenticationApi(apiClient)
	return api.Validate(s.WrapContextWithDatadogContextValues(ctx))
}

func (s *Specification) ListMonitors(ctx context.Context, params datadogV1.ListMonitorsOptionalParameters) ([]datadogV1.Monitor, *http.Response, error) {
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewMonitorsApi(apiClient)
	return api.ListMonitors(s.WrapContextWithDatadogContextValues(ctx), params)
}

func (s *Specification) GetMonitor(ctx context.Context, monitorId int64, params datadogV1.GetMonitorOptionalParameters) (datadogV1.Monitor, *http.Response, error) {
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewMonitorsApi(apiClient)
	return api.GetMonitor(s.WrapContextWithDatadogContextValues(ctx), monitorId, params)
}

func (s *Specification) SendEvent(ctx context.Context, datadogEventBody datadogV1.EventCreateRequest) (datadogV1.EventCreateResponse, *http.Response, error) {
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewEventsApi(apiClient)
	return api.CreateEvent(s.WrapContextWithDatadogContextValues(ctx), datadogEventBody)
}
