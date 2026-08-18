// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetApiTimeoutFallsBackToDefault(t *testing.T) {
	spec := Specification{}
	require.Equal(t, DefaultApiTimeout, spec.GetApiTimeout())
}

func TestGetApiTimeoutUsesConfiguredValue(t *testing.T) {
	spec := Specification{ApiTimeout: 3 * time.Second}
	require.Equal(t, 3*time.Second, spec.GetApiTimeout())
}

func TestCreateApiClientSetsHttpClientTimeout(t *testing.T) {
	spec := Specification{ApiTimeout: 7 * time.Second}
	client := spec.createApiClient()
	require.Equal(t, 7*time.Second, client.Cfg.HTTPClient.Timeout)
}
