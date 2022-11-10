// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extevents

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/event-kit/go/event_kit_api"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/exthttp"
	"net/http"
	"os"
)

func RegisterEventListenerHandlers() {
	exthttp.RegisterHttpHandler("/events/experiment-started", onExperimentStarted)
	exthttp.RegisterHttpHandler("/events/experiment-completed", onExperimentCompleted)
}

func onExperimentStarted(w http.ResponseWriter, r *http.Request, body []byte) {
	// extract the request body to EventRequestBody
	var event event_kit_api.EventRequestBody
	err := json.Unmarshal(body, &event)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to decode event request body", err))
		return
	}

	log.Debug().Msgf("Req %s body: %s", r, event)

	tags := []string{
		"Environment:" + event.Environment.Name,
		"EventName:" + event.EventName,
		"EventTime:" + event.EventTime.String(),
		"EventId:" + event.Id.String(),
		"Team:" + event.Team.Name,
		"Tenant:" + event.Tenant.Name,
		//"EndedTime:" + event.ExperimentExecution.EndedTime,
		"ExecutionId:" + fmt.Sprintf("%f", event.ExperimentExecution.ExecutionId),
		"ExperimentKey:" + event.ExperimentExecution.ExperimentKey,
		"Hypothesis:" + event.ExperimentExecution.Hypothesis,
		"ExperimentName:" + event.ExperimentExecution.Name,
		"StartedTime:" + event.ExperimentExecution.StartedTime.String(),
		string("State:" + event.ExperimentExecution.State),
	}

	for key, value := range event.ExperimentExecution.Variables {
		tags = append(tags, key+":"+value)
	}

	datadogEventBody := datadogV1.EventCreateRequest{
		Title: "Experiment started",
		Text:  "An experiment has been started by the Steadybit platform",
		Tags:  tags,
	}
	ctx := datadog.NewDefaultContext(context.Background())
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewEventsApi(apiClient)
	resp, _, err := api.CreateEvent(ctx, datadogEventBody)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EventsApi.CreateEvent`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	responseContent, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Fprintf(os.Stdout, "Response from `EventsApi.CreateEvent`:\n%s\n", responseContent)

	exthttp.WriteBody(w, "ok")
}

func onExperimentCompleted(w http.ResponseWriter, r *http.Request, _ []byte) {
	exthttp.WriteBody(w, "ok")
}
