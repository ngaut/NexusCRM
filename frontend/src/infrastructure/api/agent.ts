import { apiClient } from './client';
import { API_ENDPOINTS } from './endpoints';
import { API_CONFIG } from '../../core/constants/EnvironmentConfig';

export interface ToolCall {
    id: string;
    type: 'function';
    function: {
        name: string;
        arguments: string;
    };
}

export interface ChatMessage {
    role: string;
    content: string;
    name?: string;
    tool_calls?: ToolCall[];
    tool_call_id?: string;
    timestamp?: string; // ISO timestamp string from backend
}

export interface ChatRequest {
    messages: ChatMessage[];
    model?: string;
}

export interface ChatResponse {
    content: string;
    history: ChatMessage[];
}

// Streaming types
export type StreamEventType = 'thinking' | 'tool_call' | 'tool_result' | 'content' | 'done' | 'error' | 'auto_compact';

export interface StreamEvent {
    type: StreamEventType;
    content?: string;
    tool_name?: string;
    tool_call_id?: string;
    tool_args?: string;
    tool_result?: string;
    is_error?: boolean;
    history?: ChatMessage[];
    tokens_before?: number; // For auto_compact events
    tokens_after?: number;  // For auto_compact events
}

export interface ContextItem {
    path: string;
    token_size: number;
    content?: string;
}

export interface ContextState {
    items: ContextItem[];
    total_tokens: number;
    system_prompt?: string;
}

export interface CompactRequest {
    messages: ChatMessage[];
    keep?: string;
}

export interface CompactResponse {
    messages: ChatMessage[];
    tokens_before: number;
    tokens_after: number;
    warning?: string;
}

export const agentApi = {
    chat: async (request: ChatRequest): Promise<ChatResponse> => {
        return await apiClient.post<ChatResponse>(API_ENDPOINTS.AGENT.CHAT, request);
    },

    getContext: async (includeContent: boolean = false): Promise<ContextState> => {
        const query = includeContent ? '?include_content=true' : '';
        return await apiClient.get<ContextState>(API_ENDPOINTS.AGENT.CONTEXT(query));
    },

    compact: async (request: CompactRequest): Promise<CompactResponse> => {
        return await apiClient.post<CompactResponse>(API_ENDPOINTS.AGENT.COMPACT, request);
    },

    // Streaming chat using EventSource pattern
    chatStream: (
        request: ChatRequest,
        onEvent: (event: StreamEvent) => void,
        onError: (error: Error) => void,
        onComplete: () => void
    ): AbortController => {
        const controller = new AbortController();
        const token = apiClient.getToken();

        if (!token) {
            onError(new Error('Not authenticated. Please log in.'));
            return controller;
        }

        fetch(`${API_CONFIG.BACKEND_URL}${API_ENDPOINTS.AGENT.CHAT_STREAM}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
            },
            body: JSON.stringify(request),
            signal: controller.signal,
        })
            .then((response) => {
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}`);
                }
                const reader = response.body?.getReader();
                if (!reader) {
                    throw new Error('No reader available');
                }

                const decoder = new TextDecoder();
                let buffer = '';

                const processChunk = (chunk: string) => {
                    buffer += chunk;
                    const lines = buffer.split('\n');
                    buffer = lines.pop() || '';

                    for (const line of lines) {
                        if (line.startsWith('data:')) {
                            const data = line.slice(5).trim();
                            if (data) {
                                try {
                                    const event: StreamEvent = JSON.parse(data);
                                    onEvent(event);
                                } catch {
                                    // Skip malformed JSON
                                }
                            }
                        }
                    }
                };

                const readStream = async () => {
                    let done = false;
                    while (!done) {
                        const { done: streamDone, value } = await reader.read();
                        done = streamDone;
                        if (value) {
                            processChunk(decoder.decode(value, { stream: true }));
                        }
                    }
                    onComplete();
                };

                readStream().catch(onError);
            })
            .catch(onError);

        return controller;
    },

    // Conversation persistence
    getConversation: async (id?: string): Promise<ConversationResponse> => {
        const query = id ? `?id=${id}` : '';
        return await apiClient.get<ConversationResponse>(`${API_ENDPOINTS.AGENT.CONVERSATION}${query}`);
    },

    saveConversation: async (messages: ChatMessage[], conversationId?: string, title?: string): Promise<{ id: string; status: string }> => {
        return await apiClient.post<{ id: string; status: string }>(API_ENDPOINTS.AGENT.CONVERSATION, {
            messages,
            conversation_id: conversationId,
            title,
        });
    },

    clearConversation: async (): Promise<void> => {
        await apiClient.delete(API_ENDPOINTS.AGENT.CONVERSATION);
    },

    // Multiple conversations
    listConversations: async (): Promise<{ conversations: ConversationSummary[] }> => {
        return await apiClient.get<{ conversations: ConversationSummary[] }>(API_ENDPOINTS.AGENT.CONVERSATIONS);
    },

    deleteConversation: async (id: string): Promise<void> => {
        await apiClient.delete(`${API_ENDPOINTS.AGENT.CONVERSATIONS}/${id}`);
    },
};

export interface ConversationSummary {
    id: string;
    title: string;
    is_active: boolean;
    created_date: string;
    last_modified_date: string;
}

export interface ConversationResponse {
    conversation: { id: string; title: string } | null;
    messages: ChatMessage[];
}

