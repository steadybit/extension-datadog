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
	"time"
)

func RegisterEventListenerHandlers() {
	exthttp.RegisterHttpHandler("/events/experiment-started", onExperimentStarted)
	exthttp.RegisterHttpHandler("/events/experiment-completed", onExperimentCompleted)
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
		Title: fmt.Sprintf("Experiment '%s' (execution ID %.0f) started", event.ExperimentExecution.Name, event.ExperimentExecution.ExecutionId),
		Text: fmt.Sprintf("%%%%%% \nThe chaos engineering experiment `%s` (execution %.0f) started.\n\nThe experiment is executed through [Steadybit](https://steadybit.com/?utm_campaign=extension-datadog&utm_source=extension-datadog-event).\n %%%%%%",
			event.ExperimentExecution.Name,
			event.ExperimentExecution.ExecutionId),
		Tags:           tags,
		SourceTypeName: extutil.Ptr("Steadybit"),
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
	tags := convertSteadybitEventToDataDogEventTags(event)
	if tags == nil {
		return
	}

	duration := event.ExperimentExecution.EndedTime.Sub(event.ExperimentExecution.PreparedTime)
	datadogEventBody := datadogV1.EventCreateRequest{
		Title: fmt.Sprintf("Experiment '%s' (execution ID %.0f) ended", event.ExperimentExecution.Name, event.ExperimentExecution.ExecutionId),
		Text: fmt.Sprintf("%%%%%% \nThe chaos engineering experiment `%s` (execution %.0f) ended with state `%s` after %.2f seconds.\n\nThe experiment was executed through [Steadybit](https://steadybit.com/?utm_campaign=extension-datadog&utm_source=extension-datadog-event).\n %%%%%%",
			event.ExperimentExecution.Name,
			event.ExperimentExecution.ExecutionId,
			event.ExperimentExecution.State,
			duration.Seconds()),
		Tags:           tags,
		SourceTypeName: extutil.Ptr("Steadybit"),
	}

	SendDatadogEvent(r.Context(), &config.Config, datadogEventBody)

	exthttp.WriteBody(w, "{}")
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
		"execution_id:" + fmt.Sprintf("%f", event.ExperimentExecution.ExecutionId),
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
