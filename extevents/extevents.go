// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extevents

import (
	"github.com/steadybit/extension-kit/exthttp"
	"net/http"
)

func RegisterEventListenerHandlers() {
	exthttp.RegisterHttpHandler("/events/experiment-started", onExperimentStarted)
	exthttp.RegisterHttpHandler("/events/experiment-completed", onExperimentCompleted)
}

func onExperimentStarted(w http.ResponseWriter, r *http.Request, _ []byte) {
	exthttp.WriteBody(w, "ok")
}

func onExperimentCompleted(w http.ResponseWriter, r *http.Request, _ []byte) {
	exthttp.WriteBody(w, "ok")
}
