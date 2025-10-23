// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extmonitor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_commons"
	"github.com/steadybit/discovery-kit/go/discovery_kit_sdk"
	"github.com/steadybit/extension-datadog/config"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
)

type monitorDiscovery struct {
}

var (
	_ discovery_kit_sdk.TargetDescriber    = (*monitorDiscovery)(nil)
	_ discovery_kit_sdk.AttributeDescriber = (*monitorDiscovery)(nil)
)

func NewMonitorDiscovery() discovery_kit_sdk.TargetDiscovery {
	discovery := &monitorDiscovery{}
	return discovery_kit_sdk.NewCachedTargetDiscovery(discovery,
		discovery_kit_sdk.WithRefreshTargetsNow(),
		discovery_kit_sdk.WithRefreshTargetsInterval(context.Background(), 1*time.Minute),
	)
}
func (d *monitorDiscovery) Describe() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id: monitorTargetId,
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("1m"),
		},
	}
}

func (d *monitorDiscovery) DescribeTarget() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       monitorTargetId,
		Label:    discovery_kit_api.PluralLabel{One: "Datadog monitor", Other: "Datadog monitors"},
		Category: extutil.Ptr("monitoring"),
		Version:  extbuild.GetSemverVersionStringOrUnknown(),
		Icon:     extutil.Ptr(monitorIcon),
		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: "datadog.monitor.name"},
				{Attribute: "datadog.monitor.tags"},
			},
			OrderBy: []discovery_kit_api.OrderBy{
				{
					Attribute: "datadog.monitor.name",
					Direction: "ASC",
				},
			},
		},
	}
}

func (d *monitorDiscovery) DescribeAttributes() []discovery_kit_api.AttributeDescription {
	return []discovery_kit_api.AttributeDescription{
		{
			Attribute: "datadog.monitor.name",
			Label: discovery_kit_api.PluralLabel{
				One:   "Datadog monitor name",
				Other: "Datadog monitor names",
			},
		}, {
			Attribute: "datadog.monitor.id",
			Label: discovery_kit_api.PluralLabel{
				One:   "Datadog monitor ID",
				Other: "Datadog monitor IDs",
			},
		}, {
			Attribute: "datadog.monitor.tags",
			Label: discovery_kit_api.PluralLabel{
				One:   "Datadog monitor tags",
				Other: "Datadog monitor tags",
			},
		},
	}
}

func (d *monitorDiscovery) DiscoverTargets(ctx context.Context) ([]discovery_kit_api.Target, error) {
	return getAllMonitors(ctx, &config.Config, config.Config.SiteUrl), nil
}

type ListMonitorsApi interface {
	ListMonitors(ctx context.Context, params datadogV1.ListMonitorsOptionalParameters) ([]datadogV1.Monitor, *http.Response, error)
}

func getAllMonitors(ctx context.Context, api ListMonitorsApi, siteUrl string) []discovery_kit_api.Target {
	result := make([]discovery_kit_api.Target, 0, 500)

	parameters := datadogV1.NewListMonitorsOptionalParameters()
	parameters.PageSize = extutil.Ptr(int32(200))
	parameters.Page = extutil.Ptr(int64(0))

	start := time.Now()
	for {
		log.Debug().Int64("page", *parameters.Page).Msg("Fetch monitors from Datadog")
		monitors, r, err := api.ListMonitors(ctx, *parameters)
		if err != nil {
			log.Err(err).Msgf("Failed to retrieve monitors from Datadog for page %d and page size %d. Full response: %v",
				*parameters.Page,
				*parameters.PageSize,
				r)
			return result
		}

		if r.StatusCode != 200 {
			log.Error().Msgf("Datadog API responded with unexpected status code %d while retrieving monitors for page %d and page size %d. Full response: %v",
				r.StatusCode,
				*parameters.Page,
				*parameters.PageSize,
				r)
			return result
		}

		if len(monitors) == 0 {
			// end of list reached
			break
		}

		for _, monitor := range monitors {
			result = append(result, toTarget(monitor, siteUrl))
		}

		parameters.Page = extutil.Ptr(*parameters.Page + 1)
	}
	log.Debug().Msgf("Discovery took %s, returning %d monitors.", time.Since(start), len(result))
	return discovery_kit_commons.ApplyAttributeExcludes(result, config.Config.DiscoveryAttributesExcludesMonitor)
}

func toTarget(monitor datadogV1.Monitor, siteUrl string) discovery_kit_api.Target {
	id := strconv.FormatInt(*monitor.Id, 10)
	name := *monitor.Name

	attributes := make(map[string][]string)
	attributes["steadybit.url"] = []string{fmt.Sprintf("%s/monitors/%s", siteUrl, id)}
	attributes["datadog.monitor.name"] = []string{name}
	attributes["datadog.monitor.id"] = []string{id}
	attributes["datadog.monitor.tags"] = monitor.Tags
	if monitor.Multi != nil {
		attributes["datadog.monitor.multi-alert"] = []string{fmt.Sprintf("%t", *monitor.Multi)}
	} else {
		attributes["datadog.monitor.multi-alert"] = []string{"false"}
	}

	return discovery_kit_api.Target{
		Id:         id,
		Label:      name,
		TargetType: monitorTargetId,
		Attributes: attributes,
	}
}
