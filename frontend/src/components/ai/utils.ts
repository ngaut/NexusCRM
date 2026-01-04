
import { ChatMessage } from '../../infrastructure/api/agent';
import { TIME_MS, APP_LOCALE } from '../../core/constants';
import { DisplayItem, ToolBlock, ProcessStep, ThinkingBlock } from './types';

export const formatToolName = (name: string) => {
    return name.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
};

export const shouldRenderContent = (content?: string) => {
    if (!content) return false;
    const trimmed = content.trim();
    if (!trimmed) return false;
    if (trimmed === '{}') return false;
    if (trimmed === 'Thinking...') return false;
    return true;
};

/**
 * Format a date as a relative time string (e.g., "just now", "2m", "1h")
 */
export const formatRelativeTime = (date: Date | string | undefined): string => {
    if (!date) return '';
    const now = Date.now();
    const timestamp = typeof date === 'string' ? new Date(date).getTime() : date.getTime();
    if (isNaN(timestamp)) return '';

    const diffMs = now - timestamp;

    if (diffMs < TIME_MS.MINUTE) return 'just now';
    if (diffMs < TIME_MS.HOUR) return `${Math.floor(diffMs / TIME_MS.MINUTE)}m`;
    if (diffMs < TIME_MS.DAY) return `${Math.floor(diffMs / TIME_MS.HOUR)}h`;
    return new Date(timestamp).toLocaleDateString(APP_LOCALE, { month: 'short', day: 'numeric' });
};

export const formatDate = (timestamp: number): string => {
    return new Date(timestamp).toLocaleDateString(APP_LOCALE, { month: 'short', day: 'numeric' });
};

/**
 * Format a date as a full timestamp for tooltip display
 */
export const formatFullTime = (date: Date | string | undefined): string => {
    if (!date) return '';
    const timestamp = typeof date === 'string' ? new Date(date) : date;
    if (isNaN(timestamp.getTime())) return '';
    return timestamp.toLocaleString(APP_LOCALE, {
        month: 'short',
        day: 'numeric',
        hour: 'numeric',
        minute: '2-digit',
        hour12: true
    });
};

export const formatDateTime = (timestamp: number): string => {
    return new Date(timestamp).toLocaleString(APP_LOCALE, {
        month: 'short',
        day: 'numeric',
        hour: 'numeric',
        minute: 'numeric'
    });
};

export const buildDisplayItems = (messages: ChatMessage[]): DisplayItem[] => {
    const displayItems: DisplayItem[] = [];
    const systemMessages = new Set<number>();
    const consumedToolResults = new Set<number>();

    // First pass: Identify system messages and look for summaries
    messages.forEach((msg, idx) => {
        if (msg.role === 'system') {
            systemMessages.add(idx);

            // Check for conversation summary signature
            if (msg.content && msg.content.includes('--- CONVERSATION SUMMARY')) {
                // Match header with optional stats, compacted time, body, and footer
                // Header: --- CONVERSATION SUMMARY (Saved ~123 tokens | Compacted: Dec 28, 12:17 PM) ---
                const summaryMatch = msg.content.match(/--- CONVERSATION SUMMARY \(Saved ~(\d+) tokens(?: \| Compacted: ([^)]+))?\) ---\n([\s\S]*?)\n--- END SUMMARY ---/);

                if (summaryMatch && summaryMatch[3]) {
                    // It's a summary! Add it to display items immediately
                    displayItems.push({
                        type: 'summary_block',
                        id: `summary-${idx}`,
                        summary: summaryMatch[3].trim(),
                        stats: `Saved ~${summaryMatch[1]} tokens`,
                        compactedAt: summaryMatch[2],
                        timestamp: new Date()
                    });
                }
            }
        }
    });

    // Second pass: Group tool calls
    for (let i = 0; i < messages.length; i++) {
        if (systemMessages.has(i) || consumedToolResults.has(i)) continue;

        const msg = messages[i];

        // 1. Render Thinking Block if reasoning exists
        if (msg.role === 'assistant' && msg.reasoning_content) {
            displayItems.push({
                type: 'thinking_block',
                id: `think-${i}`,
                content: msg.reasoning_content,
                timestamp: msg.timestamp ? new Date(msg.timestamp) : new Date()
            });
        }

        // Handle Assistant with Tool Calls
        if (msg.role === 'assistant' && msg.tool_calls && msg.tool_calls.length > 0) {
            const toolBlock: ToolBlock = {
                type: 'tool_block',
                id: `block-${i}`,
                timestamp: new Date(),
                tools: []
            };

            // Consume consecutive tool interactions
            let currentIdx = i;
            while (currentIdx < messages.length) {
                const currentMsg = messages[currentIdx];
                const isAssistantToolCall = currentMsg.role === 'assistant' && currentMsg.tool_calls && currentMsg.tool_calls.length > 0;
                if (!isAssistantToolCall) break;

                currentMsg.tool_calls!.forEach((tc, tcIdx) => {
                    const callStep: ProcessStep = {
                        id: `hist-tc-${currentIdx}-${tcIdx}`,
                        type: 'tool_call',
                        content: '',
                        toolName: tc.function.name,
                        toolArgs: tc.function.arguments,
                        isDone: true,
                        timestamp: new Date()
                    };

                    let resultStep: ProcessStep | undefined;
                    const resultIdx = messages.findIndex((m, mIdx) =>
                        mIdx > currentIdx &&
                        m.role === 'tool' &&
                        m.tool_call_id === tc.id
                    );

                    if (resultIdx !== -1) {
                        consumedToolResults.add(resultIdx);
                        const resMsg = messages[resultIdx];
                        resultStep = {
                            id: `hist-${resultIdx}`,
                            type: 'tool_result',
                            content: '',
                            toolName: 'Tool Result',
                            toolResult: resMsg.content,
                            timestamp: new Date()
                        };
                    }
                    toolBlock.tools.push({ call: callStep, result: resultStep });
                });

                // Mark as consumed if not the initial one
                if (currentIdx !== i) {
                    consumedToolResults.add(currentIdx);
                }

                // Look ahead for next assistant message
                let nextIdx = currentIdx + 1;
                while (nextIdx < messages.length && (consumedToolResults.has(nextIdx) || messages[nextIdx].role === 'system')) {
                    nextIdx++;
                }
                if (nextIdx < messages.length) {
                    const nextMsg = messages[nextIdx];
                    if (nextMsg.role === 'assistant' && nextMsg.tool_calls && nextMsg.tool_calls.length > 0) {
                        currentIdx = nextIdx;
                        continue;
                    }
                }
                break;
            }

            displayItems.push(toolBlock);

            // If there's content alongside tool calls, add it as a separate message
            if (shouldRenderContent(msg.content)) {
                displayItems.push({ ...msg, tool_calls: undefined });
            }
            continue;
        }

        // Regular messages (User, Assistant without tools, or Orphan Tool Results)
        if (shouldRenderContent(msg.content)) {
            displayItems.push(msg);
        }
    }

    return displayItems;
};
