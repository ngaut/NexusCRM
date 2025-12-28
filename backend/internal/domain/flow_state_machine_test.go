package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlowStateMachine_ValidTransitions(t *testing.T) {
	sm := NewFlowStateMachine()

	tests := []struct {
		name        string
		from        FlowState
		action      FlowTransition
		expectedTo  FlowState
		shouldError bool
	}{
		// Valid transitions
		{"Running -> Paused via Pause", FlowStateRunning, TransitionPause, FlowStatePaused, false},
		{"Running -> Completed via Complete", FlowStateRunning, TransitionComplete, FlowStateCompleted, false},
		{"Running -> Failed via Fail", FlowStateRunning, TransitionFail, FlowStateFailed, false},
		{"Paused -> Running via Resume", FlowStatePaused, TransitionResume, FlowStateRunning, false},
		{"Paused -> Failed via Fail", FlowStatePaused, TransitionFail, FlowStateFailed, false},

		// Invalid transitions
		{"Paused -> Completed (invalid)", FlowStatePaused, TransitionComplete, FlowStatePaused, true},
		{"Completed -> Running (terminal)", FlowStateCompleted, TransitionResume, FlowStateCompleted, true},
		{"Failed -> Running (terminal)", FlowStateFailed, TransitionResume, FlowStateFailed, true},
		{"Running -> Running via Resume (invalid)", FlowStateRunning, TransitionResume, FlowStateRunning, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			newState, err := sm.Transition(tc.from, tc.action)

			if tc.shouldError {
				assert.Error(t, err)
				assert.Equal(t, tc.from, newState, "State should not change on invalid transition")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTo, newState)
			}
		})
	}
}

func TestFlowStateMachine_CanTransition(t *testing.T) {
	sm := NewFlowStateMachine()

	assert.True(t, sm.CanTransition(FlowStateRunning, TransitionPause))
	assert.True(t, sm.CanTransition(FlowStateRunning, TransitionComplete))
	assert.True(t, sm.CanTransition(FlowStatePaused, TransitionResume))
	assert.False(t, sm.CanTransition(FlowStateCompleted, TransitionResume))
	assert.False(t, sm.CanTransition(FlowStateFailed, TransitionPause))
}

func TestFlowStateMachine_ValidTransitionsFromState(t *testing.T) {
	sm := NewFlowStateMachine()

	runningTransitions := sm.ValidTransitions(FlowStateRunning)
	assert.Len(t, runningTransitions, 3) // Pause, Complete, Fail

	pausedTransitions := sm.ValidTransitions(FlowStatePaused)
	assert.Len(t, pausedTransitions, 2) // Resume, Fail

	completedTransitions := sm.ValidTransitions(FlowStateCompleted)
	assert.Len(t, completedTransitions, 0) // Terminal state
}

func TestFlowStateMachine_IsTerminal(t *testing.T) {
	sm := NewFlowStateMachine()

	assert.False(t, sm.IsTerminal(FlowStateRunning))
	assert.False(t, sm.IsTerminal(FlowStatePaused))
	assert.True(t, sm.IsTerminal(FlowStateCompleted))
	assert.True(t, sm.IsTerminal(FlowStateFailed))
}
