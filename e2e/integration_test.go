// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"context"
	"github.com/steadybit/action-kit/go/action_kit_test/e2e"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestWithMinikube(t *testing.T) {
	extlogging.InitZeroLog()
	server := createMockDatadogServer()
	defer server.Close()
	split := strings.SplitAfter(server.URL, ":")
	port := split[len(split)-1]

	extFactory := e2e.HelmExtensionFactory{
		Name: "extension-datadog",
		Port: 8090,
		ExtraArgs: func(m *e2e.Minikube) []string {
			return []string{
				"--set", "logging.level=debug",
				"--set", "datadog.apiKey=123456-7890",
				"--set", "datadog.applicationKey=555-666-777",
				"--set", "datadog.siteParameter=datadoghq.eu",
				"--set", "datadog.siteUrl=https://app.datadoghq.eu",
				"--set", "testing.scheme=http",
				"--set", "testing.host=host.minikube.internal:" + port,
			}
		},
	}

	mOpts := e2e.DefaultMiniKubeOpts
	mOpts.Runtimes = []e2e.Runtime{e2e.RuntimeDocker}

	e2e.WithMinikube(t, mOpts, &extFactory, []e2e.WithMinikubeTestCase{
		{
			Name: "target discovery",
			Test: testDiscovery,
		},
	})
}

func testDiscovery(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	target, err := e2e.PollForTarget(ctx, e, "com.steadybit.extension_datadog.monitor", func(target discovery_kit_api.Target) bool {
		return e2e.HasAttribute(target, "datadog.monitor.id", "8080808")
	})

	require.NoError(t, err)
	assert.Equal(t, target.TargetType, "com.steadybit.extension_datadog.monitor")
	assert.True(t, e2e.HasAttribute(target, "datadog.monitor.name", "[DEV] Monitor Kubernetes Deployments Replica Pods"))
}
