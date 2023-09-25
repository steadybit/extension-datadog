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
	SiteParameter  string `json:"siteParameter" split_words:"true" required:"true"`
	SiteUrl        string `json:"siteUrl" split_words:"true" required:"true"`
	ApiKey         string `json:"apiKey" split_words:"true" required:"true"`
	ApplicationKey string `json:"applicationKey" split_words:"true" required:"true"`
	// Only used for testing:
	TestingScheme                     *string `json:"testingScheme" split_words:"true" required:"false"`
	TestingHost                       *string `json:"testingHost" split_words:"true" required:"false"`
	DiscoveryAttributesExcludesMonitor []string `json:"discoveryAttributesExcludesMonitor" split_words:"true" required:"false"`
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

func (s *Specification) createApiClient() *datadog.APIClient {
	configuration := datadog.NewConfiguration()
	if s.TestingHost != nil && s.TestingScheme != nil {
		configuration.Scheme = *s.TestingScheme
		configuration.Host = *s.TestingHost
	}
	return datadog.NewAPIClient(configuration)
}

func (s *Specification) ValidateCredentials(ctx context.Context) (datadogV1.AuthenticationValidationResponse, *http.Response, error) {
	api := datadogV1.NewAuthenticationApi(s.createApiClient())
	return api.Validate(s.WrapContextWithDatadogContextValues(ctx))
}

func (s *Specification) ListMonitors(ctx context.Context, params datadogV1.ListMonitorsOptionalParameters) ([]datadogV1.Monitor, *http.Response, error) {
	api := datadogV1.NewMonitorsApi(s.createApiClient())
	return api.ListMonitors(s.WrapContextWithDatadogContextValues(ctx), params)
}

func (s *Specification) GetMonitor(ctx context.Context, monitorId int64, params datadogV1.GetMonitorOptionalParameters) (datadogV1.Monitor, *http.Response, error) {
	api := datadogV1.NewMonitorsApi(s.createApiClient())
	return api.GetMonitor(s.WrapContextWithDatadogContextValues(ctx), monitorId, params)
}

func (s *Specification) CreateDowntime(ctx context.Context, downtimeBody datadogV1.Downtime) (datadogV1.Downtime, *http.Response, error) {
	api := datadogV1.NewDowntimesApi(s.createApiClient())
	return api.CreateDowntime(s.WrapContextWithDatadogContextValues(ctx), downtimeBody)
}

func (s *Specification) CancelDowntime(ctx context.Context, downtimeId int64) (*http.Response, error) {
	api := datadogV1.NewDowntimesApi(s.createApiClient())
	return api.CancelDowntime(s.WrapContextWithDatadogContextValues(ctx), downtimeId)
}

func (s *Specification) SendEvent(ctx context.Context, datadogEventBody datadogV1.EventCreateRequest) (datadogV1.EventCreateResponse, *http.Response, error) {
	api := datadogV1.NewEventsApi(s.createApiClient())
	return api.CreateEvent(s.WrapContextWithDatadogContextValues(ctx), datadogEventBody)
}
