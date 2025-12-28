package domain

import (
	"fmt"
)

// FlowState represents the current state of a flow instance
type FlowState string

const (
	// FlowStateRunning indicates the flow is actively executing
	FlowStateRunning FlowState = "Running"
	// FlowStatePaused indicates the flow is waiting for approval
	FlowStatePaused FlowState = "Paused"
	// FlowStateCompleted indicates the flow finished successfully
	FlowStateCompleted FlowState = "Completed"
	// FlowStateFailed indicates the flow encountered an error
	FlowStateFailed FlowState = "Failed"
)

// FlowTransition represents an action that can change flow state
type FlowTransition string

const (
	// TransitionStart initiates a new flow
	TransitionStart FlowTransition = "Start"
	// TransitionPause pauses flow at an approval step
	TransitionPause FlowTransition = "Pause"
	// TransitionResume resumes flow after approval
	TransitionResume FlowTransition = "Resume"
	// TransitionComplete marks flow as completed
	TransitionComplete FlowTransition = "Complete"
	// TransitionFail marks flow as failed
	TransitionFail FlowTransition = "Fail"
)

// FlowStateMachine enforces valid state transitions for flow instances.
// Invalid transitions return an error (fail-fast approach).
type FlowStateMachine struct {
	// transitions maps (current state, transition) -> next state
	transitions map[stateTransitionKey]FlowState
}

type stateTransitionKey struct {
	state      FlowState
	transition FlowTransition
}

// NewFlowStateMachine creates a new state machine with the flow lifecycle rules.
// State diagram:
//
//	          Start
//	            │
//	            ▼
//	       [Running] ◄──Resume──┐
//	         │    \             │
//	      Pause   Complete      │
//	         │      \           │
//	         ▼       ▼          │
//	     [Paused]  [Completed]  │
//	         │                  │
//	         └──────────────────┘
//
//	Any state can transition to [Failed] via Fail
func NewFlowStateMachine() *FlowStateMachine {
	sm := &FlowStateMachine{
		transitions: make(map[stateTransitionKey]FlowState),
	}

	// Define valid transitions
	sm.addTransition(FlowStateRunning, TransitionPause, FlowStatePaused)
	sm.addTransition(FlowStateRunning, TransitionComplete, FlowStateCompleted)
	sm.addTransition(FlowStateRunning, TransitionFail, FlowStateFailed)
	sm.addTransition(FlowStatePaused, TransitionResume, FlowStateRunning)
	sm.addTransition(FlowStatePaused, TransitionFail, FlowStateFailed)
	// Note: Start transition is handled separately (creates new instance)

	return sm
}

func (sm *FlowStateMachine) addTransition(from FlowState, via FlowTransition, to FlowState) {
	key := stateTransitionKey{state: from, transition: via}
	sm.transitions[key] = to
}

// Transition attempts to transition from the current state using the given action.
// Returns the new state or an error if the transition is invalid.
func (sm *FlowStateMachine) Transition(current FlowState, action FlowTransition) (FlowState, error) {
	key := stateTransitionKey{state: current, transition: action}
	next, ok := sm.transitions[key]
	if !ok {
		return current, fmt.Errorf("invalid state transition: cannot %s from %s", action, current)
	}
	return next, nil
}

// CanTransition checks if a transition is valid without performing it.
func (sm *FlowStateMachine) CanTransition(current FlowState, action FlowTransition) bool {
	key := stateTransitionKey{state: current, transition: action}
	_, ok := sm.transitions[key]
	return ok
}

// ValidTransitions returns all valid transitions from the given state.
func (sm *FlowStateMachine) ValidTransitions(state FlowState) []FlowTransition {
	var result []FlowTransition
	for key := range sm.transitions {
		if key.state == state {
			result = append(result, key.transition)
		}
	}
	return result
}

// IsTerminal returns true if the state is a terminal state (no further transitions).
func (sm *FlowStateMachine) IsTerminal(state FlowState) bool {
	return state == FlowStateCompleted || state == FlowStateFailed
}
