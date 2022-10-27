// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extmonitor

import (
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
)

func RegisterMonitorDiscoveryHandlers() {
	exthttp.RegisterHttpHandler("/monitor/discovery", exthttp.GetterAsHandler(getMonitorDiscoveryDescription))
	exthttp.RegisterHttpHandler("/monitor/discovery/target-description", exthttp.GetterAsHandler(getMonitorTargetDescription))
	exthttp.RegisterHttpHandler("/monitor/discovery/attribute-descriptions", exthttp.GetterAsHandler(getMonitorAttributeDescriptions))
	//exthttp.RegisterHttpHandler("/monitor/discovery/discovered-targets", getRdsInstanceDiscoveryResults)
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
		// TODO
		Version: "1.0.0-SNAPSHOT",
		Icon:    extutil.Ptr(monitorIcon),
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
				Attribute: "datadog.monitor.tag",
				Label: discovery_kit_api.PluralLabel{
					One:   "Datadog monitor tag",
					Other: "Datadog monitor tags",
				},
			},
		},
	}
}

//func getRdsInstanceDiscoveryResults(w http.ResponseWriter, r *http.Request, _ []byte) {
//	client := rds.NewFromConfig(utils.AwsConfig)
//	targets, err := GetAllRdsInstances(r.Context(), client)
//	if err != nil {
//		exthttp.WriteError(w, extension_kit.ToError("Failed to collect RDS instance information", err))
//	} else {
//		exthttp.WriteBody(w, discovery_kit_api.DiscoveredTargets{Targets: targets})
//	}
//}
//
//type RdsDescribeInstancesApi interface {
//	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
//}
//
//func GetAllRdsInstances(ctx context.Context, rdsApi RdsDescribeInstancesApi) ([]discovery_kit_api.Target, error) {
//	result := make([]discovery_kit_api.Target, 0, 20)
//
//	var marker *string = nil
//	for {
//		output, err := rdsApi.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
//			Marker: marker,
//		})
//		if err != nil {
//			return result, err
//		}
//
//		for _, dbInstance := range output.DBInstances {
//			result = append(result, toTarget(dbInstance))
//		}
//
//		if output.Marker == nil {
//			break
//		} else {
//			marker = output.Marker
//		}
//	}
//
//	return result, nil
//}
//
//func toTarget(dbInstance types.DBInstance) discovery_kit_api.Target {
//	arn := aws.ToString(dbInstance.DBInstanceArn)
//	label := aws.ToString(dbInstance.DBInstanceIdentifier)
//
//	attributes := make(map[string][]string)
//	attributes["steadybit.label"] = []string{label}
//	attributes["aws.account"] = []string{utils.AwsAccountNumber}
//	attributes["aws.arn"] = []string{arn}
//	attributes["aws.zone"] = []string{aws.ToString(dbInstance.AvailabilityZone)}
//	attributes["aws.rds.engine"] = []string{aws.ToString(dbInstance.Engine)}
//	attributes["aws.rds.instance.id"] = []string{label}
//	attributes["aws.rds.instance.status"] = []string{aws.ToString(dbInstance.DBInstanceStatus)}
//
//	if dbInstance.DBClusterIdentifier != nil {
//		attributes["aws.rds.cluster"] = []string{aws.ToString(dbInstance.DBClusterIdentifier)}
//	}
//
//	return discovery_kit_api.Target{
//		Id:         arn,
//		Label:      label,
//		TargetType: rdsTargetId,
//		Attributes: attributes,
//	}
//}
