// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package config

import (
	"context"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

type Specification struct {
	// see https://docs.datadoghq.com/getting_started/site/#access-the-datadog-site
	Site           string `json:"site" split_words:"true" required:"true"`
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
			"site": Config.Site,
		},
	)

	return ctx
}
