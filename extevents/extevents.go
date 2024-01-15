// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extevents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/event-kit/go/event_kit_api"
	"github.com/steadybit/extension-datadog/config"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"net/http"
	"sync"
	"time"
)

func RegisterEventListenerHandlers() {
	exthttp.RegisterHttpHandler("/events/experiment-started", handle(onExperimentStarted))
	exthttp.RegisterHttpHandler("/events/experiment-completed", handle(onExperimentCompleted))
	exthttp.RegisterHttpHandler("/events/experiment-step-started", handle(onExperimentStepStarted))
	exthttp.RegisterHttpHandler("/events/experiment-target-started", handle(onExperimentTargetStarted))
	exthttp.RegisterHttpHandler("/events/experiment-target-completed", handle(onExperimentTargetCompleted))
}

type SendEventApi interface {
	SendEvent(ctx context.Context, datadogEventBody datadogV1.EventCreateRequest) (datadogV1.EventCreateResponse, *http.Response, error)
}

var (
	stepExecutions = sync.Map{}
)

type StepInfo struct {
	ActionName string
	Tags       []string
}

type eventHandler func(event event_kit_api.EventRequestBody) (*datadogV1.EventCreateRequest, error)

func handle(handler eventHandler) func(w http.ResponseWriter, r *http.Request, body []byte) {
	return func(w http.ResponseWriter, r *http.Request, body []byte) {

		event, err := parseBodyToEventRequestBody(body)
		if err != nil {
			exthttp.WriteError(w, extension_kit.ToError("Failed to decode event request body", err))
			return
		}

		if request, err := handler(event); err == nil {
			if request != nil {
				sendDatadogEvent(r.Context(), &config.Config, request)
			}
		} else {
			exthttp.WriteError(w, extension_kit.ToError(err.Error(), err))
			return
		}

		exthttp.WriteBody(w, "{}")
	}
}

func onExperimentStarted(event event_kit_api.EventRequestBody) (*datadogV1.EventCreateRequest, error) {
	tags := getEventBaseTags(event)
	tags = append(tags, getExecutionTags(event)...)
	return &datadogV1.EventCreateRequest{
		Title: fmt.Sprintf("Experiment '%s' started", event.ExperimentExecution.ExperimentKey),
		Text: fmt.Sprintf("%%%%%% \nExperiment `%s` - `%s` (execution `%.0f`) started.\n %%%%%%",
			event.ExperimentExecution.ExperimentKey,
			event.ExperimentExecution.Name,
			event.ExperimentExecution.ExecutionId),
		Tags:           tags,
		SourceTypeName: extutil.Ptr("Steadybit"),
		AggregationKey: extutil.Ptr(fmt.Sprintf("steadybit-execution-%.0f", event.ExperimentExecution.ExecutionId)),
		DateHappened:   extutil.Ptr(event.EventTime.Unix()),
	}, nil
}

func onExperimentCompleted(event event_kit_api.EventRequestBody) (*datadogV1.EventCreateRequest, error) {
	stepExecutions.Range(func(key, value interface{}) bool {
		stepExecution := value.(event_kit_api.ExperimentStepExecution)
		if stepExecution.ExecutionId == event.ExperimentExecution.ExecutionId {
			log.Debug().Msgf("Delete step execution data for id %.0f", stepExecution.ExecutionId)
			stepExecutions.Delete(key)
		}
		return true
	})

	duration := event.ExperimentExecution.EndedTime.Sub(event.ExperimentExecution.PreparedTime)
	tags := getEventBaseTags(event)
	tags = append(tags, getExecutionTags(event)...)
	return &datadogV1.EventCreateRequest{
		Title: fmt.Sprintf("Experiment '%s' ended", event.ExperimentExecution.ExperimentKey),
		Text: fmt.Sprintf("%%%%%% \nExperiment `%s` - `%s` (execution `%.0f`) ended with state `%s` after %.2f seconds.\n %%%%%%",
			event.ExperimentExecution.ExperimentKey,
			event.ExperimentExecution.Name,
			event.ExperimentExecution.ExecutionId,
			event.ExperimentExecution.State,
			duration.Seconds()),
		Tags:           tags,
		SourceTypeName: extutil.Ptr("Steadybit"),
		AggregationKey: extutil.Ptr(fmt.Sprintf("steadybit-execution-%.0f", event.ExperimentExecution.ExecutionId)),
		DateHappened:   extutil.Ptr(event.EventTime.Unix()),
	}, nil
}

func onExperimentStepStarted(event event_kit_api.EventRequestBody) (*datadogV1.EventCreateRequest, error) {
	if event.ExperimentStepExecution == nil {
		return nil, errors.New("missing ExperimentStepExecution in event")
	}
	stepExecutions.Store(event.ExperimentStepExecution.Id, *event.ExperimentStepExecution)
	return nil, nil
}

func onExperimentTargetStarted(event event_kit_api.EventRequestBody) (*datadogV1.EventCreateRequest, error) {
	if event.ExperimentStepTargetExecution == nil {
		return nil, errors.New("missing ExperimentStepTargetExecution in event")
	}

	var v, ok = stepExecutions.Load(event.ExperimentStepTargetExecution.StepExecutionId)
	if !ok {
		log.Warn().Msgf("Could not find step infos for step execution id %s", event.ExperimentStepTargetExecution.StepExecutionId)
		return nil, nil
	}
	stepExecution := v.(event_kit_api.ExperimentStepExecution)

	if stepExecution.ActionKind != nil && *stepExecution.ActionKind == event_kit_api.Attack {
		tags := append(getStepTags(stepExecution), getEventBaseTags(event)...)
		tags = append(tags, getTargetTags(*event.ExperimentStepTargetExecution)...)

		return &datadogV1.EventCreateRequest{
			Title: fmt.Sprintf("Experiment '%s' - Attack started", event.ExperimentStepTargetExecution.ExperimentKey),
			Text: fmt.Sprintf("%%%%%% \nExperiment `%s` (execution `%.0f`) - Attack `%s` started.\n\nTarget:%s\n %%%%%%",
				event.ExperimentStepTargetExecution.ExperimentKey,
				event.ExperimentStepTargetExecution.ExecutionId,
				getActionName(stepExecution),
				getTargetName(*event.ExperimentStepTargetExecution)),
			Tags:           tags,
			SourceTypeName: extutil.Ptr("Steadybit"),
			AggregationKey: extutil.Ptr(fmt.Sprintf("steadybit-execution-%.0f", event.ExperimentStepTargetExecution.ExecutionId)),
			DateHappened:   extutil.Ptr(event.EventTime.Unix()),
		}, nil
	}

	return nil, nil
}

func getActionName(stepExecution event_kit_api.ExperimentStepExecution) string {
	actionName := *stepExecution.ActionId
	if stepExecution.ActionName != nil {
		actionName = *stepExecution.ActionName
	}
	if stepExecution.CustomLabel != nil {
		actionName = *stepExecution.CustomLabel
	}
	return actionName
}

func onExperimentTargetCompleted(event event_kit_api.EventRequestBody) (*datadogV1.EventCreateRequest, error) {
	if event.ExperimentStepTargetExecution == nil {
		return nil, errors.New("missing ExperimentStepTargetExecution in event")
	}

	var v, ok = stepExecutions.Load(event.ExperimentStepTargetExecution.StepExecutionId)
	if !ok {
		log.Warn().Msgf("Could not find step infos for step execution id %s", event.ExperimentStepTargetExecution.StepExecutionId)
		return nil, nil
	}
	stepExecution := v.(event_kit_api.ExperimentStepExecution)

	if stepExecution.ActionKind != nil && *stepExecution.ActionKind == event_kit_api.Attack {
		tags := append(getStepTags(stepExecution), getEventBaseTags(event)...)
		tags = append(tags, getTargetTags(*event.ExperimentStepTargetExecution)...)
		duration := event.ExperimentStepTargetExecution.EndedTime.Sub(*event.ExperimentStepTargetExecution.StartedTime)

		return &datadogV1.EventCreateRequest{
			Title: fmt.Sprintf("Experiment '%s' - Attack ended", event.ExperimentStepTargetExecution.ExperimentKey),
			Text: fmt.Sprintf("%%%%%% \nExperiment `%s` (execution `%.0f`) - Attack `%s` ended with state `%s` after %.2f seconds.\n\nTarget:%s\n %%%%%%",
				event.ExperimentStepTargetExecution.ExperimentKey,
				event.ExperimentStepTargetExecution.ExecutionId,
				getActionName(stepExecution),
				event.ExperimentStepTargetExecution.State,
				duration.Seconds(),
				getTargetName(*event.ExperimentStepTargetExecution)),
			Tags:           tags,
			SourceTypeName: extutil.Ptr("Steadybit"),
			AggregationKey: extutil.Ptr(fmt.Sprintf("steadybit-execution-%.0f", event.ExperimentStepTargetExecution.ExecutionId)),
			DateHappened:   extutil.Ptr(event.EventTime.Unix()),
		}, nil
	}
	return nil, nil
}

func getTargetName(target event_kit_api.ExperimentStepTargetExecution) string {
	if values, ok := target.TargetAttributes["steadybit.label"]; ok {
		return values[0]
	}
	return target.TargetName
}

func getEventBaseTags(event event_kit_api.EventRequestBody) []string {
	tags := []string{
		"source:Steadybit",
		"environment_name:" + event.Environment.Name,
		"event_name:" + event.EventName,
		"event_time:" + event.EventTime.String(),
		"event_id:" + event.Id.String(),
		"tenant_name:" + event.Tenant.Name,
		"tenant_key:" + event.Tenant.Key,
	}

	if event.Team != nil {
		tags = append(tags, "team_name:"+event.Team.Name, "team_key:"+event.Team.Key)
	}

	userPrincipal, isUserPrincipal := event.Principal.(event_kit_api.UserPrincipal)
	if isUserPrincipal {
		tags = append(tags, "principal_type:"+userPrincipal.PrincipalType)
		tags = append(tags, "principal_username:"+userPrincipal.Username)
		tags = append(tags, "principal_name:"+userPrincipal.Name)
	}

	accessTokenPrincipal, isAccessTokenPrincipal := event.Principal.(event_kit_api.AccessTokenPrincipal)
	if isAccessTokenPrincipal {
		tags = append(tags, "principal_type:"+accessTokenPrincipal.PrincipalType)
		tags = append(tags, "principal_name:"+accessTokenPrincipal.Name)
	}

	batchPrincipal, isBatchPrincipal := event.Principal.(event_kit_api.BatchPrincipal)
	if isBatchPrincipal {
		tags = append(tags, "principal_type:"+batchPrincipal.PrincipalType)
		tags = append(tags, "principal_username:"+batchPrincipal.Username)
	}

	return tags
}

func getExecutionTags(event event_kit_api.EventRequestBody) []string {
	if event.ExperimentExecution == nil {
		return []string{}
	}
	tags := []string{
		"execution_id:" + fmt.Sprintf("%g", event.ExperimentExecution.ExecutionId),
		"experiment_key:" + event.ExperimentExecution.ExperimentKey,
		"experiment_name:" + event.ExperimentExecution.Name,
		"execution_state:" + string(event.ExperimentExecution.State),
	}

	if len(event.ExperimentExecution.Hypothesis) > 0 {
		tags = append(tags, "experiment_hypothesis:"+event.ExperimentExecution.Hypothesis)
	}

	if event.ExperimentExecution.StartedTime.IsZero() {
		tags = append(tags, "started_time:"+time.Now().Format(time.RFC3339))
	} else {
		tags = append(tags, "started_time:"+event.ExperimentExecution.StartedTime.Format(time.RFC3339))
	}

	if event.ExperimentExecution.EndedTime != nil && !(*event.ExperimentExecution.EndedTime).IsZero() {
		tags = append(tags, "ended_time:"+event.ExperimentExecution.EndedTime.Format(time.RFC3339))
	}

	return tags
}

func getStepTags(step event_kit_api.ExperimentStepExecution) []string {
	var tags []string
	if step.Type == event_kit_api.Action {
		tags = append(tags, "step_action_id:"+*step.ActionId)
	}
	if step.ActionName != nil {
		tags = append(tags, "step_action_name:"+*step.ActionName)
	}
	if step.CustomLabel != nil {
		tags = append(tags, "step_custom_label:"+*step.CustomLabel)
	}
	return tags
}

func getTargetTags(target event_kit_api.ExperimentStepTargetExecution) []string {
	tags := []string{
		"execution_id:" + fmt.Sprintf("%g", target.ExecutionId),
		"experiment_key:" + target.ExperimentKey,
		"execution_state:" + string(target.State),
	}

	if target.StartedTime != nil {
		tags = append(tags, "started_time:"+target.StartedTime.Format(time.RFC3339))
	}

	if target.EndedTime != nil {
		tags = append(tags, "ended_time:"+target.EndedTime.Format(time.RFC3339))
	}

	if _, ok := target.TargetAttributes["k8s.cluster-name"]; ok {
		//"kube_"-tags
		tags = append(tags, translateToDatadog(target, "k8s.cluster-name", "kube_cluster_name")...)
		tags = append(tags, translateToDatadog(target, "k8s.namespace", "kube_namespace")...)
		tags = append(tags, translateToDatadog(target, "k8s.deployment", "kube_deployment")...)
		tags = append(tags, translateToDatadog(target, "k8s.namespace", "namespace")...)
		tags = append(tags, translateToDatadog(target, "k8s.pod.name", "pod_name")...)
		tags = append(tags, translateToDatadog(target, "k8s.deployment", "deployment")...)
		tags = append(tags, translateToDatadog(target, "k8s.container.name", "container_name")...)
		tags = append(tags, translateToDatadog(target, "k8s.cluster-name", "cluster_name")...)
		tags = append(tags, translateToDatadog(target, "k8s.pod.label.tags.datadoghq.com/service", "service")...)
		tags = append(tags, translateToDatadog(target, "k8s.deployment.label.tags.datadoghq.com/service", "service")...)
	}

	tags = append(tags, getHostnameTag(target)...)
	tags = append(tags, translateToDatadog(target, "container.id.stripped", "container_id")...)

	//AWS tags
	tags = append(tags, translateToDatadog(target, "aws.region", "aws_region")...)
	tags = append(tags, translateToDatadog(target, "aws.zone", "aws_zone")...)
	tags = append(tags, translateToDatadog(target, "aws.account", "aws_account")...)

	return removeDuplicates(tags)
}

func getHostnameTag(target event_kit_api.ExperimentStepTargetExecution) []string {
	var tags []string
	tags = append(tags, translateToDatadog(target, "container.host", "host")...)
	tags = append(tags, translateToDatadog(target, "host.hostname", "host")...)
	tags = append(tags, translateToDatadog(target, "application.hostname", "host")...)
	tags = removeDuplicates(tags)

	//Add cluster-name to host -> https://docs.datadoghq.com/containers/guide/kubernetes-cluster-name-detection/
	if values, ok := target.TargetAttributes["k8s.cluster-name"]; ok {
		if len(tags) == 1 && len(values) == 1 {
			tags[0] = tags[0] + "-" + values[0]
		}
	}
	return tags
}

func translateToDatadog(target event_kit_api.ExperimentStepTargetExecution, steadybitAttribute string, datadogTag string) []string {
	var tags []string
	if values, ok := target.TargetAttributes[steadybitAttribute]; ok {
		//We don't want to add one-to-many attributes to datadog. For example when attacking a host, we don't want to add all namespaces or pods which are running on that host.
		if (len(values)) == 1 {
			tags = append(tags, datadogTag+":"+values[0])
		}
	}
	return tags
}

func removeDuplicates(tags []string) []string {
	allKeys := make(map[string]bool)
	var result []string
	for _, tag := range tags {
		if _, value := allKeys[tag]; !value {
			allKeys[tag] = true
			result = append(result, tag)
		}
	}
	return result
}

func parseBodyToEventRequestBody(body []byte) (event_kit_api.EventRequestBody, error) {
	var event event_kit_api.EventRequestBody
	err := json.Unmarshal(body, &event)
	return event, err
}

func sendDatadogEvent(ctx context.Context, api SendEventApi, datadogEventBody *datadogV1.EventCreateRequest) {
	_, r, err := api.SendEvent(ctx, *datadogEventBody)

	if err != nil {
		log.Err(err).Msgf("Failed to send Datadog event. Full response %v",
			r)
	} else if r.StatusCode != 202 && r.StatusCode != 200 {
		log.Error().Msgf("Datadog API responded with unexpected status code %d while sending Event. Full response: %v",
			r.StatusCode,
			r)
	}
}
