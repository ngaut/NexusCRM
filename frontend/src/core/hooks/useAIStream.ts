import { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import { agentApi, ChatMessage, StreamEvent, CompactRequest, ConversationSummary } from '../../infrastructure/api/agent';
import { ProcessStep } from '../../components/ai/types';
import { STORAGE_KEYS } from '../constants/ApplicationDefaults';
import { idGenerator, createMessage, createToolCall, createProcessStep } from '../ai';

export function useAIStream() {
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [processSteps, setProcessSteps] = useState<ProcessStep[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [streamingContent, setStreamingContent] = useState<string>('');
    const [accumulatedThinking, setAccumulatedThinking] = useState<string>('');
    const [isProcessExpanded, setIsProcessExpanded] = useState(false);
    const [isInitialized, setIsInitialized] = useState(false);

    // Multi-conversation state
    const [currentConversationId, setCurrentConversationId] = useState<string | null>(null);
    const [conversations, setConversations] = useState<ConversationSummary[]>([]);
    const [isSidebarOpen, setIsSidebarOpen] = useState(false);

    const abortControllerRef = useRef<AbortController | null>(null);
    const saveTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    // Load conversations list and active conversation on mount
    useEffect(() => {
        const loadConversations = async () => {
            try {
                // Load conversation list
                const listResponse = await agentApi.listConversations();
                setConversations(listResponse.conversations || []);

                // Load active conversation
                const response = await agentApi.getConversation();
                if (response.conversation) {
                    setCurrentConversationId(response.conversation.id);
                    setMessages(response.messages as ChatMessage[]);
                }
            } finally {
                setIsInitialized(true);
            }
        };
        loadConversations();
    }, []);

    // Auto-save to server when messages change (debounced)
    useEffect(() => {
        if (!isInitialized || messages.length === 0) return;

        // Debounce server save to avoid too many requests
        if (saveTimeoutRef.current) clearTimeout(saveTimeoutRef.current);
        saveTimeoutRef.current = setTimeout(async () => {
            try {
                const result = await agentApi.saveConversation(messages, currentConversationId || undefined);
                if (!currentConversationId && result.id) {
                    setCurrentConversationId(result.id);
                }
                // Refresh conversation list
                const listResponse = await agentApi.listConversations();
                setConversations(listResponse.conversations || []);
            } catch { }
        }, 2000);

        return () => {
            if (saveTimeoutRef.current) clearTimeout(saveTimeoutRef.current);
        };
    }, [messages, isInitialized, currentConversationId]);

    // Token Calculation
    const conversationTokens = useMemo(() => {
        return messages.reduce((total, msg) => {
            let tokens = Math.ceil((msg.content || '').length / 4);
            if (msg.tool_calls) {
                tokens += msg.tool_calls.reduce((acc, tc) => acc + Math.ceil((tc.function.arguments || '').length / 4), 0);
            }
            return total + tokens;
        }, 0);
    }, [messages]);

    const summaryInfo = useMemo(() => {
        for (const msg of messages) {
            if (msg.role === 'system' && msg.content?.includes('--- CONVERSATION SUMMARY')) {
                const match = msg.content.match(/Saved ~(\d+) tokens/);
                return {
                    hasSummary: true,
                    tokensSaved: match ? parseInt(match[1], 10) : 0
                };
            }
        }
        return { hasSummary: false, tokensSaved: 0 };
    }, [messages]);

    const handleStreamEvent = useCallback((event: StreamEvent) => {
        switch (event.type) {
            case 'thinking': {
                const thinkContent = event.content || 'Thinking...';
                setAccumulatedThinking(prev => prev + thinkContent);
                setProcessSteps(prev => [...prev, createProcessStep.thinking(thinkContent)]);
                break;
            }

            case 'tool_call': {
                setProcessSteps(prev => {
                    // Mark any active thinking steps as done
                    const steps = prev.map(s =>
                        s.type === 'thinking' && !s.isDone ? { ...s, isDone: true } : s
                    );
                    return [...steps, createProcessStep.toolCall(event.tool_name || '', event.tool_args)];
                });

                if (event.tool_name && event.tool_args) {
                    setMessages(prev => {
                        const lastMsg = prev[prev.length - 1];
                        const newToolCall = createToolCall.call(event.tool_name!, event.tool_args!, event.tool_call_id);

                        // Append to existing assistant message with tool_calls
                        if (lastMsg && lastMsg.role === 'assistant' && lastMsg.tool_calls) {
                            return [...prev.slice(0, -1), {
                                ...lastMsg,
                                reasoning_content: accumulatedThinking,
                                tool_calls: [...lastMsg.tool_calls, newToolCall]
                            }];
                        }

                        // Create new assistant message with tool calls
                        return [...prev, createMessage.assistantWithToolCalls([newToolCall], accumulatedThinking)];
                    });
                }
                break;
            }

            case 'tool_result': {
                setProcessSteps(prev => {
                    // Find last matching tool_call (backwards search)
                    let lastMatchingIndex = -1;
                    for (let i = prev.length - 1; i >= 0; i--) {
                        if (prev[i].toolName === event.tool_name && prev[i].type === 'tool_call') {
                            lastMatchingIndex = i;
                            break;
                        }
                    }
                    if (lastMatchingIndex === -1) return prev;

                    return prev.map((s, idx) =>
                        idx === lastMatchingIndex
                            ? {
                                ...s,
                                type: 'tool_result' as const,
                                toolResult: event.tool_result,
                                isError: event.is_error,
                                isDone: true,
                                content: event.is_error
                                    ? `${event.tool_name} failed`
                                    : `${event.tool_name} completed`
                            }
                            : s
                    );
                });

                if (event.tool_name && event.tool_result) {
                    setMessages(prev => [
                        ...prev,
                        createMessage.tool(event.tool_name!, event.tool_result!, event.tool_call_id || 'call_unknown')
                    ]);
                }
                break;
            }

            case 'content':
                setProcessSteps(prev => prev.map(s =>
                    s.type === 'thinking' && !s.isDone ? { ...s, isDone: true } : s
                ));
                setStreamingContent(prev => prev + (event.content || ''));
                break;

            case 'done':
                setIsLoading(false);
                if (event.history) setMessages(event.history);
                setStreamingContent('');
                setProcessSteps([]);
                setIsProcessExpanded(false);
                break;

            case 'error':
                setIsLoading(false);
                setMessages(prev => [...prev, createMessage.error(event.content || 'Unknown error')]);
                setStreamingContent('');
                setProcessSteps([]);
                break;

            case 'auto_compact':
                if (event.history && event.history.length > 0) {
                    setMessages(event.history);
                }
                const compactMsg = event.content ||
                    `Context auto-compacted (${event.tokens_before?.toLocaleString()} → ${event.tokens_after?.toLocaleString()} tokens)`;
                setProcessSteps(prev => [...prev, createProcessStep.autoCompact(compactMsg)]);
                break;
        }
    }, [accumulatedThinking]);

    const sendMessage = async (input: string) => {
        if (!input.trim() || isLoading) return;

        const userMsg = createMessage.user(input);
        const newHistory = [...messages, userMsg];

        setMessages(newHistory);
        setIsLoading(true);
        setProcessSteps([]);
        idGenerator.reset(); // Reset ID counters for new turn
        setStreamingContent('');
        setAccumulatedThinking('');
        setIsProcessExpanded(true);

        abortControllerRef.current = agentApi.chatStream(
            { messages: newHistory },
            handleStreamEvent,
            (error) => {
                console.warn('Stream error:', error);
                setIsLoading(false);
                setMessages(prev => [...prev, createMessage.error(error.message)]);
            },
            () => {
                setIsLoading(false);
            }
        );
    };

    const cancelStream = () => {
        if (abortControllerRef.current) {
            abortControllerRef.current.abort();
            setIsLoading(false);
            if (streamingContent) {
                setMessages(prev => [...prev, createMessage.cancelled(streamingContent)]);
                setStreamingContent('');
            }
            setAccumulatedThinking('');
            setIsProcessExpanded(false);
        }
    };

    const clearChat = useCallback(async () => {
        // Just clear state
        setMessages([]);
        setProcessSteps([]);
        setStreamingContent('');
        setCurrentConversationId(null);
    }, []);

    const newChat = useCallback(async () => {
        setMessages([]);
        setProcessSteps([]);
        setStreamingContent('');
        setCurrentConversationId(null);
    }, []);

    const selectConversation = useCallback(async (id: string) => {
        if (id === currentConversationId) return;

        try {
            const response = await agentApi.getConversation(id);
            if (response.conversation) {
                setMessages(response.messages as ChatMessage[]);
                setCurrentConversationId(response.conversation.id);
                setProcessSteps([]);
                setStreamingContent('');
            }
        } catch (error) {
            console.warn('Failed to load conversation:', error);
        }
    }, [currentConversationId]);

    const deleteConversation = useCallback(async (id: string) => {
        try {
            await agentApi.deleteConversation(id);
            setConversations(prev => prev.filter(c => c.id !== id));
            if (id === currentConversationId) {
                setMessages([]);
                setCurrentConversationId(null);
                setCurrentConversationId(null);
            }
        } catch (error) {
            console.warn('Failed to delete conversation:', error);
        }
    }, [currentConversationId]);

    const compactMessages = async (keepInstruction?: string) => {
        if (messages.length === 0) return;
        setIsLoading(true);
        const compactRequest: CompactRequest = {
            messages: messages,
            keep: keepInstruction,
        };
        try {
            const response = await agentApi.compact(compactRequest);
            setMessages(response.messages);
        } catch (err) {
            console.warn('Failed to compact messages:', err);
            setMessages(prev => [...prev, createMessage.compactError(
                err instanceof Error ? err.message : 'Unknown error'
            )]);
        } finally {
            setIsLoading(false);
        }
    };

    return {
        messages,
        setMessages,
        processSteps,
        setProcessSteps,
        isLoading,
        setIsLoading,
        streamingContent,
        setStreamingContent,
        isProcessExpanded,
        setIsProcessExpanded,
        sendMessage,
        cancelStream,
        clearChat,
        compactMessages,
        conversationTokens,
        summaryInfo,
        // Multi-conversation
        currentConversationId,
        conversations,
        isSidebarOpen,
        setIsSidebarOpen,
        newChat,
        selectConversation,
        deleteConversation,
    };
}
