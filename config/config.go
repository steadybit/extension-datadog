// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package config

import (
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

var (
	Config Specification
)

func ParseConfiguration() {
	err := envconfig.Process("steadybit_extension", &Config)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to parse configuration from environment.")
	}
}

func ValidateConfiguration() {
	resp, r, err := Config.ValidateCredentials(context.Background())

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
