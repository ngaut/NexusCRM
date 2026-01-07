import { ChatMessage, ToolCall } from '../../infrastructure/api/agent';
import { ProcessStep } from '../../components/ai/types';
import { idGenerator } from './idGenerator';

/**
 * Factory functions for creating AI messages and process steps.
 * Single source of truth for message structure ensures consistency.
 */

// ============================================================================
// Message Factory
// ============================================================================

function createTimestamp(): string {
    return new Date().toISOString();
}

export const createMessage = {
    /** Create a user message */
    user(content: string): ChatMessage {
        return {
            role: 'user',
            content,
            timestamp: createTimestamp()
        };
    },

    /** Create an assistant message with optional reasoning */
    assistant(content: string, reasoningContent?: string): ChatMessage {
        return {
            role: 'assistant',
            content,
            reasoning_content: reasoningContent,
            timestamp: createTimestamp()
        };
    },

    /** Create an assistant message with tool calls */
    assistantWithToolCalls(toolCalls: ToolCall[], reasoningContent?: string): ChatMessage {
        return {
            role: 'assistant',
            content: '',
            reasoning_content: reasoningContent,
            timestamp: createTimestamp(),
            tool_calls: toolCalls
        };
    },

    /** Create a tool result message */
    tool(name: string, content: string, toolCallId: string): ChatMessage {
        return {
            role: 'tool',
            name,
            content,
            tool_call_id: toolCallId,
            timestamp: createTimestamp()
        };
    },

    /** Create an error message */
    error(errorMessage: string): ChatMessage {
        return {
            role: 'assistant',
            content: `Error: ${errorMessage}`,
            timestamp: createTimestamp()
        };
    },

    /** Create a cancelled message */
    cancelled(partialContent: string): ChatMessage {
        return {
            role: 'assistant',
            content: `${partialContent}\n\n*(Cancelled)*`,
            timestamp: createTimestamp()
        };
    },

    /** Create a compact failure message */
    compactError(errorMessage: string): ChatMessage {
        return {
            role: 'assistant',
            content: `❌ Failed to compact context: ${errorMessage}`,
            timestamp: createTimestamp()
        };
    }
};

// ============================================================================
// Tool Call Factory
// ============================================================================

export const createToolCall = {
    /** Create a tool call object */
    call(name: string, args: string, existingId?: string): ToolCall {
        return {
            id: existingId || idGenerator.nextToolCallId(),
            type: 'function',
            function: {
                name,
                arguments: args
            }
        };
    }
};

// ============================================================================
// Process Step Factory
// ============================================================================

export const createProcessStep = {
    /** Create a thinking step */
    thinking(content: string): ProcessStep {
        return {
            id: idGenerator.nextStepId(),
            type: 'thinking',
            content,
            timestamp: new Date()
        };
    },

    /** Create a tool call step */
    toolCall(toolName: string, toolArgs?: string): ProcessStep {
        return {
            id: idGenerator.nextStepId(),
            type: 'tool_call',
            content: `Calling ${toolName}`,
            toolName,
            toolArgs,
            timestamp: new Date()
        };
    },

    /** Create an auto-compact notification step */
    autoCompact(content: string): ProcessStep {
        return {
            id: idGenerator.nextStepId(),
            type: 'thinking',
            content,
            timestamp: new Date(),
            isDone: true
        };
    }
};
