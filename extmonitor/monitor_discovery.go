// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extmonitor

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-datadog/config"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"net/http"
	"strconv"
)

func RegisterMonitorDiscoveryHandlers() {
	exthttp.RegisterHttpHandler("/monitor/discovery", exthttp.GetterAsHandler(getMonitorDiscoveryDescription))
	exthttp.RegisterHttpHandler("/monitor/discovery/target-description", exthttp.GetterAsHandler(getMonitorTargetDescription))
	exthttp.RegisterHttpHandler("/monitor/discovery/attribute-descriptions", exthttp.GetterAsHandler(getMonitorAttributeDescriptions))
	exthttp.RegisterHttpHandler("/monitor/discovery/discovered-targets", getMonitorDiscoveryResults)
}

func getMonitorDiscoveryDescription() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id:         monitorTargetId,
		RestrictTo: extutil.Ptr(discovery_kit_api.LEADER),
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			Method:       "GET",
			Path:         "/monitor/discovery/discovered-targets",
			CallInterval: extutil.Ptr("30s"),
		},
	}
}

func getMonitorTargetDescription() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       monitorTargetId,
		Label:    discovery_kit_api.PluralLabel{One: "Datadog monitor", Other: "Datadog monitors"},
		Category: extutil.Ptr("monitoring"),
		Version:  "1.0.0",
		Icon:     extutil.Ptr(monitorIcon),
		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: "steadybit.label"},
				{Attribute: "datadog.monitor.tag"},
			},
			OrderBy: []discovery_kit_api.OrderBy{
				{
					Attribute: "steadybit.label",
					Direction: "ASC",
				},
			},
		},
	}
}

func getMonitorAttributeDescriptions() discovery_kit_api.AttributeDescriptions {
	return discovery_kit_api.AttributeDescriptions{
		Attributes: []discovery_kit_api.AttributeDescription{
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
		},
	}
}

func getMonitorDiscoveryResults(w http.ResponseWriter, r *http.Request, _ []byte) {
	targets := GetAllMonitors(r.Context(), &config.Config, config.Config.SiteUrl)
	exthttp.WriteBody(w, discovery_kit_api.DiscoveredTargets{Targets: targets})
}

type ListMonitorsApi interface {
	ListMonitors(ctx context.Context, params datadogV1.ListMonitorsOptionalParameters) ([]datadogV1.Monitor, *http.Response, error)
}

func GetAllMonitors(ctx context.Context, api ListMonitorsApi, siteUrl string) []discovery_kit_api.Target {
	result := make([]discovery_kit_api.Target, 0, 500)

	parameters := datadogV1.NewListMonitorsOptionalParameters()
	parameters.PageSize = extutil.Ptr(int32(200))
	parameters.Page = extutil.Ptr(int64(0))

	for {
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

	return result
}

func toTarget(monitor datadogV1.Monitor, siteUrl string) discovery_kit_api.Target {
	id := strconv.FormatInt(*monitor.Id, 10)
	name := *monitor.Name

	attributes := make(map[string][]string)
	attributes["steadybit.label"] = []string{name}
	attributes["steadybit.url"] = []string{fmt.Sprintf("%s/monitors/%s", siteUrl, id)}
	attributes["datadog.monitor.name"] = []string{name}
	attributes["datadog.monitor.id"] = []string{id}
	attributes["datadog.monitor.tags"] = monitor.Tags

	return discovery_kit_api.Target{
		Id:         id,
		Label:      name,
		TargetType: monitorTargetId,
		Attributes: attributes,
	}
}
