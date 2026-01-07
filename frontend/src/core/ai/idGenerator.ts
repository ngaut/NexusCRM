/**
 * Centralized ID generator for AI module.
 * Provides guaranteed unique IDs for process steps and tool calls within a session.
 */
class IdGeneratorService {
    private stepCounter = 0;
    private toolCallCounter = 0;

    /** Generate next unique step ID (e.g., "step-1", "step-2") */
    nextStepId(): string {
        return `step-${++this.stepCounter}`;
    }

    /** Generate next unique tool call ID (e.g., "call-1", "call-2") */
    nextToolCallId(): string {
        return `call-${++this.toolCallCounter}`;
    }

    /** Reset all counters (call when starting a new message/turn) */
    reset(): void {
        this.stepCounter = 0;
        this.toolCallCounter = 0;
    }

    /** Get current counter values for debugging */
    getState(): { stepCounter: number; toolCallCounter: number } {
        return {
            stepCounter: this.stepCounter,
            toolCallCounter: this.toolCallCounter
        };
    }
}

// Singleton instance for the application
export const idGenerator = new IdGeneratorService();
