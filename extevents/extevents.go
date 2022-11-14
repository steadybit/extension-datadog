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
	"os"
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
	// extract the request body to EventRequestBody
	tags := convertSteadybitEventToDataDogEventTags(w, r, body)
	if tags == nil {
		return
	}

	datadogEventBody := datadogV1.EventCreateRequest{
		Title:          "Experiment started",
		Text:           "An experiment has been started by the Steadybit platform",
		Tags:           tags,
		SourceTypeName: extutil.Ptr("Steadybit"),
	}

	result := SendDatadogEvent(r.Context(), &config.Config, datadogEventBody)

	exthttp.WriteBody(w, result)
}

func onExperimentCompleted(w http.ResponseWriter, r *http.Request, body []byte) {
	// extract the request body to EventRequestBody
	tags := convertSteadybitEventToDataDogEventTags(w, r, body)
	if tags == nil {
		return
	}

	datadogEventBody := datadogV1.EventCreateRequest{
		Title:          "Experiment ended",
		Text:           "An experiment has ended by the Steadybit platform",
		Tags:           tags,
		SourceTypeName: extutil.Ptr("Steadybit"),
	}

	result := SendDatadogEvent(r.Context(), &config.Config, datadogEventBody)

	exthttp.WriteBody(w, result)
}

func convertSteadybitEventToDataDogEventTags(w http.ResponseWriter, r *http.Request, body []byte) []string {
	var event event_kit_api.EventRequestBody
	err := json.Unmarshal(body, &event)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to decode event request body", err))
		return nil
	}

	log.Debug().Msgf("Req %s body: %s", r, event)

	tags := []string{
		"source:Steadybit",
		"environment:" + event.Environment.Name,
		"event_name:" + event.EventName,
		"event_time:" + event.EventTime.String(),
		"event_id:" + event.Id.String(),
		"team:" + event.Team.Name,
		"tenant:" + event.Tenant.Name,
		"execution_id:" + fmt.Sprintf("%f", event.ExperimentExecution.ExecutionId),
		"experiment_key:" + event.ExperimentExecution.ExperimentKey,
		"experiment_name:" + event.ExperimentExecution.Name,
		string("state:" + event.ExperimentExecution.State),
		//"Principal:" + event.Principal, //TODO: add principal to event
	}

	if len(event.ExperimentExecution.Hypothesis) > 0 {
		tags = append(tags, "hypothesis:"+event.ExperimentExecution.Hypothesis)
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

func SendDatadogEvent(ctx context.Context, api SendEventApi, datadogEventBody datadogV1.EventCreateRequest) datadogV1.EventCreateResponse {
	result, r, err := api.SendEvent(ctx, datadogEventBody)

	if err != nil {
		log.Err(err).Msgf("Failed to send Datadog event. Full response %v",
			r)
		return result
	}

	if r.StatusCode != 202 && r.StatusCode != 200 {
		log.Error().Msgf("Datadog API responded with unexpected status code %d while sending Event. Full response: %v",
			r.StatusCode,
			r)
		return result
	}

	responseContent, _ := json.MarshalIndent(result, "", "  ")
	fmt.Fprintf(os.Stdout, "Response from `EventsApi.CreateEvent`:\n%s\n", responseContent)

	return result
}
