// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extevents

import (
	"context"
	"encoding/json"
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
	exthttp.RegisterHttpHandler("/events/experiment-started", onExperimentStarted)
	exthttp.RegisterHttpHandler("/events/experiment-completed", onExperimentCompleted)
	exthttp.RegisterHttpHandler("/events/experiment-step-started", onExperimentStepStarted)
	exthttp.RegisterHttpHandler("/events/experiment-step-completed", onExperimentStepCompleted)
}

type SendEventApi interface {
	SendEvent(ctx context.Context, datadogEventBody datadogV1.EventCreateRequest) (datadogV1.EventCreateResponse, *http.Response, error)
}

func onExperimentStarted(w http.ResponseWriter, r *http.Request, body []byte) {
	event, err := parseBodyToEventRequestBody(body)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to decode event request body", err))
		return
	}
	tags := convertSteadybitEventToDataDogEventTags(event)
	if tags == nil {
		return
	}

	datadogEventBody := datadogV1.EventCreateRequest{
		Title: fmt.Sprintf("Experiment '%s' started", event.ExperimentExecution.ExperimentKey),
		Text: fmt.Sprintf("%%%%%% \nExperiment `%s` - `%s` (execution `%.0f`) started.\n\n_The experiment is executed through [Steadybit](https://steadybit.com/?utm_campaign=extension-datadog&utm_source=extension-datadog-event)._\n %%%%%%",
			event.ExperimentExecution.ExperimentKey,
			event.ExperimentExecution.Name,
			event.ExperimentExecution.ExecutionId),
		Tags:           tags,
		SourceTypeName: extutil.Ptr("Steadybit"),
		AggregationKey: extutil.Ptr(fmt.Sprintf("steadybit-execution-%.0f", event.ExperimentExecution.ExecutionId)),
	}

	SendDatadogEvent(r.Context(), &config.Config, datadogEventBody)

	exthttp.WriteBody(w, "{}")
}

func onExperimentCompleted(w http.ResponseWriter, r *http.Request, body []byte) {
	event, err := parseBodyToEventRequestBody(body)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to decode event request body", err))
		return
	}

	lastStartedStepsMux.Lock()
	delete(lastStartedSteps, event.ExperimentExecution.ExecutionId)
	lastStartedStepsMux.Unlock()
	lastCompletedStepsMux.Lock()
	delete(lastCompletedSteps, event.ExperimentExecution.ExecutionId)
	lastCompletedStepsMux.Unlock()

	tags := convertSteadybitEventToDataDogEventTags(event)
	if tags == nil {
		return
	}

	duration := event.ExperimentExecution.EndedTime.Sub(event.ExperimentExecution.PreparedTime)
	datadogEventBody := datadogV1.EventCreateRequest{
		Title: fmt.Sprintf("Experiment '%s' ended", event.ExperimentExecution.ExperimentKey),
		Text: fmt.Sprintf("%%%%%% \nExperiment `%s` - `%s` (execution `%.0f`) ended with state `%s` after %.2f seconds.\n\n_The experiment is executed through [Steadybit](https://steadybit.com/?utm_campaign=extension-datadog&utm_source=extension-datadog-event)._\n %%%%%%",
			event.ExperimentExecution.ExperimentKey,
			event.ExperimentExecution.Name,
			event.ExperimentExecution.ExecutionId,
			event.ExperimentExecution.State,
			duration.Seconds()),
		Tags:           tags,
		SourceTypeName: extutil.Ptr("Steadybit"),
		AggregationKey: extutil.Ptr(fmt.Sprintf("steadybit-execution-%.0f", event.ExperimentExecution.ExecutionId)),
	}

	SendDatadogEvent(r.Context(), &config.Config, datadogEventBody)

	exthttp.WriteBody(w, "{}")
}

var lastStartedSteps = make(map[float32]time.Time)
var lastStartedStepsMux = &sync.RWMutex{}
var lastCompletedSteps = make(map[float32]time.Time)
var lastCompletedStepsMux = &sync.RWMutex{}

func onExperimentStepStarted(w http.ResponseWriter, r *http.Request, body []byte) {
	event, err := parseBodyToEventRequestBody(body)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to decode event request body", err))
		return
	}
	lastStartedStepsMux.RLock()
	lastStartedStep, ok := lastStartedSteps[event.ExperimentExecution.ExecutionId]
	lastStartedStepsMux.RUnlock()
	if !ok {
		lastStartedStep = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	for _, step := range event.ExperimentExecution.Steps {
		if step.ActionKind == nil || *step.ActionKind != event_kit_api.Attack {
			continue
		}
		if step.StartedTime != nil && step.StartedTime.After(lastStartedStep) {
			lastStartedStepsMux.Lock()
			lastStartedSteps[event.ExperimentExecution.ExecutionId] = *step.StartedTime
			lastStartedStepsMux.Unlock()
			executionAndStepTags := convertSteadybitEventToDataDogEventTags(event)
			if executionAndStepTags == nil {
				return
			}
			executionAndStepTags = append(executionAndStepTags, getStepTags(step)...)
			for _, target := range *step.TargetExecutions {
				targetTags := getTargetTags(target)
				allTags := make([]string, 0, len(executionAndStepTags)+len(targetTags))
				allTags = append(allTags, executionAndStepTags...)
				allTags = append(allTags, targetTags...)
				actionName := *step.ActionId
				if step.ActionName != nil {
					actionName = *step.ActionName
				}
				if step.CustomLabel != nil {
					actionName = *step.CustomLabel
				}
				datadogEventBody := datadogV1.EventCreateRequest{
					Title: fmt.Sprintf("Experiment '%s' - Attack started", event.ExperimentExecution.ExperimentKey),
					Text: fmt.Sprintf("%%%%%% \nExperiment `%s` - `%s` (execution `%.0f`) - Attack `%s` started.\n\nTarget:%s\n\n_The experiment is executed through [Steadybit](https://steadybit.com/?utm_campaign=extension-datadog&utm_source=extension-datadog-event)._\n %%%%%%",
						event.ExperimentExecution.ExperimentKey,
						event.ExperimentExecution.Name,
						event.ExperimentExecution.ExecutionId,
						actionName,
						getTargetName(target),
					),
					Tags:           allTags,
					SourceTypeName: extutil.Ptr("Steadybit"),
					AggregationKey: extutil.Ptr(fmt.Sprintf("steadybit-execution-%.0f", event.ExperimentExecution.ExecutionId)),
				}
				SendDatadogEvent(r.Context(), &config.Config, datadogEventBody)
			}
		}
	}
	exthttp.WriteBody(w, "{}")
}

func onExperimentStepCompleted(w http.ResponseWriter, r *http.Request, body []byte) {
	event, err := parseBodyToEventRequestBody(body)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to decode event request body", err))
		return
	}
	lastCompletedStepsMux.RLock()
	lastCompletedStep, ok := lastCompletedSteps[event.ExperimentExecution.ExecutionId]
	lastCompletedStepsMux.RUnlock()
	if !ok {
		lastCompletedStep = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	for _, step := range event.ExperimentExecution.Steps {
		if step.ActionKind == nil || *step.ActionKind != event_kit_api.Attack {
			continue
		}
		if step.Type != event_kit_api.Wait && step.EndedTime != nil && step.EndedTime.After(lastCompletedStep) && step.StartedTime != nil {
			lastCompletedStepsMux.Lock()
			lastCompletedSteps[event.ExperimentExecution.ExecutionId] = *step.EndedTime
			lastCompletedStepsMux.Unlock()
			executionAndStepTags := convertSteadybitEventToDataDogEventTags(event)
			if executionAndStepTags == nil {
				return
			}
			executionAndStepTags = append(executionAndStepTags, getStepTags(step)...)
			duration := step.EndedTime.Sub(*step.StartedTime)

			for _, target := range *step.TargetExecutions {
				targetTags := getTargetTags(target)
				allTags := make([]string, 0, len(executionAndStepTags)+len(targetTags))
				allTags = append(allTags, executionAndStepTags...)
				allTags = append(allTags, targetTags...)
				actionName := *step.ActionId
				if step.ActionName != nil {
					actionName = *step.ActionName
				}
				if step.CustomLabel != nil {
					actionName = *step.CustomLabel
				}
				datadogEventBody := datadogV1.EventCreateRequest{
					Title: fmt.Sprintf("Experiment '%s' - Attack ended", event.ExperimentExecution.ExperimentKey),
					Text: fmt.Sprintf("%%%%%% \nExperiment `%s` - `%s` (execution `%.0f`) - Attack `%s` ended with state `%s` after %.2f seconds.\n\nTarget:%s\n\n_The experiment is executed through [Steadybit](https://steadybit.com/?utm_campaign=extension-datadog&utm_source=extension-datadog-event)._\n %%%%%%",
						event.ExperimentExecution.ExperimentKey,
						event.ExperimentExecution.Name,
						event.ExperimentExecution.ExecutionId,
						actionName,
						step.State,
						duration.Seconds(),
						getTargetName(target),
					),
					Tags:           allTags,
					SourceTypeName: extutil.Ptr("Steadybit"),
					AggregationKey: extutil.Ptr(fmt.Sprintf("steadybit-execution-%.0f", event.ExperimentExecution.ExecutionId)),
				}
				SendDatadogEvent(r.Context(), &config.Config, datadogEventBody)
			}

		}
	}
	exthttp.WriteBody(w, "{}")
}

func getTargetName(target event_kit_api.ExperimentStepExecutionTarget) string {
	if values, ok := target.TargetAttributes["steadybit.label"]; ok {
		return values[0]
	}
	return target.TargetName
}

func convertSteadybitEventToDataDogEventTags(event event_kit_api.EventRequestBody) []string {
	tags := []string{
		"source:Steadybit",
		"environment_name:" + event.Environment.Name,
		"event_name:" + event.EventName,
		"event_time:" + event.EventTime.String(),
		"event_id:" + event.Id.String(),
		"team_name:" + event.Team.Name,
		"team_key:" + event.Team.Key,
		"tenant_name:" + event.Tenant.Name,
		"tenant_key:" + event.Tenant.Key,
		"execution_id:" + fmt.Sprintf("%g", event.ExperimentExecution.ExecutionId),
		"experiment_key:" + event.ExperimentExecution.ExperimentKey,
		"experiment_name:" + event.ExperimentExecution.Name,
		string("execution_state:" + event.ExperimentExecution.State),
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

	if len(event.ExperimentExecution.Hypothesis) > 0 {
		tags = append(tags, "experiment_hypothesis:"+event.ExperimentExecution.Hypothesis)
	}

	if event.ExperimentExecution.StartedTime.IsZero() {
		tags = append(tags, "started_time:"+time.Now().Format(time.RFC3339))
	} else {
		tags = append(tags, "started_time:"+event.ExperimentExecution.StartedTime.Format(time.RFC3339))
	}

	if event.ExperimentExecution.EndedTime != nil {
		tags = append(tags, "ended_time:"+event.ExperimentExecution.EndedTime.Format(time.RFC3339))
	}

	return tags
}

func getStepTags(step event_kit_api.ExperimentStepExecution) []string {
	var tags []string
	tags = append(tags, "step_state:"+string(step.State))
	if step.StartedTime != nil {
		tags = append(tags, "step_started_time:"+step.StartedTime.Format(time.RFC3339))
	}
	if step.EndedTime != nil {
		tags = append(tags, "step_ended_time:"+step.EndedTime.Format(time.RFC3339))
	}
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

func getTargetTags(target event_kit_api.ExperimentStepExecutionTarget) []string {
	var tags []string
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
	}

	tags = append(tags, getHostnameTag(target)...)
	tags = append(tags, translateToDatadog(target, "container.id.stripped", "container_id")...)

	//AWS tags
	tags = append(tags, translateToDatadog(target, "aws.region", "aws_region")...)
	tags = append(tags, translateToDatadog(target, "aws.zone", "aws_zone")...)
	tags = append(tags, translateToDatadog(target, "aws.account", "aws_account")...)

	return tags
}

func getHostnameTag(target event_kit_api.ExperimentStepExecutionTarget) []string {
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

func translateToDatadog(target event_kit_api.ExperimentStepExecutionTarget, steadybitAttribute string, datadogTag string) []string {
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

func SendDatadogEvent(ctx context.Context, api SendEventApi, datadogEventBody datadogV1.EventCreateRequest) {
	_, r, err := api.SendEvent(ctx, datadogEventBody)

	if err != nil {
		log.Err(err).Msgf("Failed to send Datadog event. Full response %v",
			r)
	} else if r.StatusCode != 202 && r.StatusCode != 200 {
		log.Error().Msgf("Datadog API responded with unexpected status code %d while sending Event. Full response: %v",
			r.StatusCode,
			r)
	}
}
