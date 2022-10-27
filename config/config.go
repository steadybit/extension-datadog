// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package config

import (
	"context"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

var (
	Config Specification
)

func init() {
	err := envconfig.Process("steadybit_extension", &Config)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to parse configuration from environment.")
	}
	validateConfiguration()
}

func validateConfiguration() {
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewAuthenticationApi(apiClient)
	resp, r, err := api.Validate(Config.WrapContextWithDatadogContextValues(context.Background()))

	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to validate extension configuration using the Datadog API. Full HTTP response: %v", r)
	}

	if !resp.HasValid() {
		log.Fatal().Msgf("Datadog API did not respond with expected 'valid' field while validating configuration. Full HTTP response: %v", r)
	}

	if !resp.GetValid() {
		log.Fatal().Msgf("Datadog API reported that the given configuration is invalid. Please fix the environment configuration. Full HTTP response: %v", r)
	}
}
