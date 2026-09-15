// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package config

import (
	"strings"

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
	// envconfig's `required:"true"` only checks that the variable is *set*: an empty
	// value satisfies it, so the extension would start with a blank configuration and
	// fail much later against the target system. Reject blank values here instead.
	if strings.TrimSpace(Config.SiteParameter) == "" {
		log.Fatal().Msg("STEADYBIT_EXTENSION_SITE_PARAMETER must not be empty.")
	}
	if strings.TrimSpace(Config.SiteUrl) == "" {
		log.Fatal().Msg("STEADYBIT_EXTENSION_SITE_URL must not be empty.")
	}
	if strings.TrimSpace(Config.ApiKey) == "" {
		log.Fatal().Msg("STEADYBIT_EXTENSION_API_KEY must not be empty.")
	}
	if strings.TrimSpace(Config.ApplicationKey) == "" {
		log.Fatal().Msg("STEADYBIT_EXTENSION_APPLICATION_KEY must not be empty.")
	}

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
