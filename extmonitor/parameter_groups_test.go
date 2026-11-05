// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package extmonitor

import (
	"testing"

	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/stretchr/testify/assert"
)

type describer interface {
	Describe() action_kit_api.ActionDescription
}

func TestParameterGroups(t *testing.T) {
	tests := []struct {
		name          string
		action        describer
		grouped       []string
		withFailEarly bool
	}{
		{name: "monitor status check", action: NewMonitorStatusCheckAction(), grouped: nil, withFailEarly: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParameterGroups(t, tt.action.Describe().Parameters, tt.grouped, tt.withFailEarly)
		})
	}
}

func TestMultiAlertGroup(t *testing.T) {
	desc := NewMonitorStatusCheckAction().Describe()
	found := false
	for _, p := range desc.Parameters {
		if p.Name == "multiAlertFilter" {
			found = true
			assert.Equal(t, groupMultiAlert, *p.Group)
			// Platforms without groups keep showing it in "Additional Settings".
			assert.True(t, *p.Advanced)
		}
	}
	assert.True(t, found, "multiAlertFilter parameter")
	assert.Equal(t, []string{groupMultiAlert}, []string{(*desc.ParameterGroups)[0].Name})
	assert.Len(t, *desc.ParameterGroups, 1)
}

func assertParameterGroups(t *testing.T, params []action_kit_api.ActionParameter, grouped []string, withFailEarly bool) {
	t.Helper()

	var groupedNames []string
	foundFailEarly := false
	for _, p := range params {
		if p.Advanced != nil && *p.Advanced {
			assert.NotNil(t, p.Group, "advanced parameter %s has no group", p.Name)
		}
		if p.Name == "failEarly" {
			foundFailEarly = true
			assert.False(t, p.Advanced != nil && *p.Advanced, "failEarly must not be advanced")
			assert.Nil(t, p.Group, "failEarly must not be grouped")
		}
		if p.Group != nil && *p.Group == action_kit_api.ParameterGroupAdditionalSettings {
			groupedNames = append(groupedNames, p.Name)
		}
	}
	assert.Equal(t, withFailEarly, foundFailEarly, "failEarly presence")
	assert.ElementsMatch(t, grouped, groupedNames)

	// The platform sorts by order (missing = 0); without any order the list order applies.
	usesOrder := false
	for _, p := range params {
		if p.Order != nil && *p.Order != 0 {
			usesOrder = true
		}
	}
	sortKey := func(i int) int {
		if !usesOrder {
			return i
		}
		if params[i].Order == nil {
			return 0
		}
		return *params[i].Order
	}
	for i, g := range params {
		if g.Group == nil || *g.Group != action_kit_api.ParameterGroupAdditionalSettings {
			continue
		}
		for j, o := range params {
			if o.Group != nil && *o.Group == action_kit_api.ParameterGroupAdditionalSettings {
				continue
			}
			assert.Greater(t, sortKey(i), sortKey(j), "grouped parameter %s must sort after %s", g.Name, o.Name)
		}
	}
}
