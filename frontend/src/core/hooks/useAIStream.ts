import { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import { agentApi, ChatMessage, StreamEvent, CompactRequest, ConversationSummary } from '../../infrastructure/api/agent';
import { ProcessStep } from '../../components/ai/types';
import { STORAGE_KEYS } from '../constants/ApplicationDefaults';

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
                } else {
                    // Fallback to localStorage for migration
                    const savedMessages = localStorage.getItem(STORAGE_KEYS.AI_MESSAGES);
                    if (savedMessages) {
                        const parsed = JSON.parse(savedMessages);
                        setMessages(parsed);
                        // Migrate to server
                        const result = await agentApi.saveConversation(parsed);
                        setCurrentConversationId(result.id);
                        // Refresh list
                        const newList = await agentApi.listConversations();
                        setConversations(newList.conversations || []);
                    }
                }
            } catch {
                // Fallback to localStorage if API fails
                const savedMessages = localStorage.getItem(STORAGE_KEYS.AI_MESSAGES);
                if (savedMessages) setMessages(JSON.parse(savedMessages));
            } finally {
                setIsInitialized(true);
            }
        };
        loadConversations();
    }, []);

    // Auto-save to server when messages change (debounced)
    useEffect(() => {
        if (!isInitialized || messages.length === 0) return;

        // Also save to localStorage as backup
        localStorage.setItem(STORAGE_KEYS.AI_MESSAGES, JSON.stringify(messages));

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

    // Note: processSteps are NOT persisted - they are transient streaming state
    // and already captured in message history as tool_calls/tool results

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
        const stepId = `step-${Date.now()}-${Math.random()}`;

        switch (event.type) {
            case 'thinking':
                const thinkContent = event.content || 'Thinking...';
                setAccumulatedThinking(prev => prev + thinkContent);
                setProcessSteps(prev => [
                    ...prev,
                    {
                        id: stepId,
                        type: 'thinking',
                        content: thinkContent,
                        timestamp: new Date()
                    }
                ]);
                break;

            case 'tool_call':
                setProcessSteps(prev => {
                    // Mark any active thinking steps as done
                    const steps = prev.map(s =>
                        s.type === 'thinking' && !s.isDone
                            ? { ...s, isDone: true }
                            : s
                    );

                    return [
                        ...steps,
                        {
                            id: stepId,
                            type: 'tool_call',
                            content: `Calling ${event.tool_name}`,
                            toolName: event.tool_name,
                            toolArgs: event.tool_args,
                            timestamp: new Date()
                        }
                    ];
                });

                if (event.tool_name && event.tool_args) {
                    setMessages(prev => {
                        const lastMsg = prev[prev.length - 1];
                        // If the last message is an assistant message and already has tool_calls, we append to it.
                        // We also need to ensure we preserve/update the reasoning_content on this message.
                        if (lastMsg && lastMsg.role === 'assistant' && lastMsg.tool_calls) {
                            return [...prev.slice(0, -1), {
                                ...lastMsg,
                                reasoning_content: accumulatedThinking, // Persist accumulated thinking
                                tool_calls: [...lastMsg.tool_calls, {
                                    id: event.tool_call_id || `call_${Date.now()}`,
                                    type: 'function',
                                    function: {
                                        name: event.tool_name!,
                                        arguments: event.tool_args!
                                    }
                                }]
                            }];
                        } else {
                            // New assistant message frame
                            return [...prev, {
                                role: 'assistant',
                                content: '',
                                reasoning_content: accumulatedThinking, // Persist accumulated thinking
                                tool_calls: [{
                                    id: event.tool_call_id || `call_${Date.now()}`,
                                    type: 'function',
                                    function: {
                                        name: event.tool_name!,
                                        arguments: event.tool_args!
                                    }
                                }]
                            }];
                        }
                    });
                }
                break;

            case 'tool_result':
                setProcessSteps(prev => {
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
                    setMessages(prev => {
                        return [...prev, {
                            role: 'tool',
                            name: event.tool_name,
                            content: event.tool_result!,
                            tool_call_id: event.tool_call_id || `call_unknown`
                        }];
                    });
                }
                break;

            case 'content':
                setProcessSteps(prev => prev.map(s =>
                    s.type === 'thinking' && !s.isDone
                        ? { ...s, isDone: true }
                        : s
                ));
                setStreamingContent(prev => prev + (event.content || ''));

                // Also update the message list if we are streaming the final response, 
                // ensuring reasoning is attached if it exists (though usually content stream is separate)
                // Note: We don't usually update 'messages' for streaming content until 'done', 
                // but if we wanted real-time persistence of reasoning + content, we'd do it here.
                // For now, streamingContent handles the visual part, but let's make sure 
                // the final 'done' event captures it, or we rely on 'streamingContent' visual.
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
                setMessages(prev => [
                    ...prev,
                    { role: 'assistant', content: `Error: ${event.content}` }
                ]);
                setStreamingContent('');
                setProcessSteps([]);
                break;

            case 'auto_compact':
                if (event.history && event.history.length > 0) {
                    setMessages(event.history);
                }
                setProcessSteps(prev => [
                    ...prev,
                    {
                        id: `auto-compact-${Date.now()}`,
                        type: 'thinking' as const,
                        content: event.content || `Context auto-compacted (${event.tokens_before?.toLocaleString()} → ${event.tokens_after?.toLocaleString()} tokens)`,
                        timestamp: new Date(),
                        isDone: true
                    }
                ]);
                break;
        }
    }, []);

    const sendMessage = async (input: string) => {
        if (!input.trim() || isLoading) return;

        const userMsg: ChatMessage = { role: 'user', content: input };
        const newHistory = [...messages, userMsg];

        setMessages(newHistory);
        setIsLoading(true);
        setProcessSteps([]);
        setStreamingContent('');
        setAccumulatedThinking('');
        setIsProcessExpanded(true);

        abortControllerRef.current = agentApi.chatStream(
            { messages: newHistory },
            handleStreamEvent,
            (error) => {
                console.warn('Stream error:', error);
                setIsLoading(false);
                setMessages(prev => [
                    ...prev,
                    { role: 'assistant', content: `Error: ${error.message}` }
                ]);
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
                setMessages(prev => [
                    ...prev,
                    { role: 'assistant', content: streamingContent + '\n\n*(Cancelled)*' }
                ]);
                setStreamingContent('');
            }
            setIsProcessExpanded(false);
        }
    };

    const clearChat = () => {
        setMessages([]);
        setProcessSteps([]);
        setStreamingContent('');
        setCurrentConversationId(null);
        localStorage.removeItem(STORAGE_KEYS.AI_MESSAGES);
        agentApi.clearConversation().catch(() => { });
    };

    // Create a new conversation
    const newChat = useCallback(async () => {
        // Clear current state
        setMessages([]);
        setProcessSteps([]);
        setStreamingContent('');
        setCurrentConversationId(null);
        localStorage.removeItem(STORAGE_KEYS.AI_MESSAGES);
    }, []);

    // Select and load a specific conversation
    const selectConversation = useCallback(async (id: string) => {
        if (id === currentConversationId) return;

        try {
            const response = await agentApi.getConversation(id);
            if (response.conversation) {
                setMessages(response.messages as ChatMessage[]);
                setCurrentConversationId(response.conversation.id);
                setProcessSteps([]);
                setStreamingContent('');
                localStorage.setItem(STORAGE_KEYS.AI_MESSAGES, JSON.stringify(response.messages));
            }
        } catch (error) {
            console.warn('Failed to load conversation:', error);
        }
    }, [currentConversationId]);

    // Delete a specific conversation
    const deleteConversation = useCallback(async (id: string) => {
        try {
            await agentApi.deleteConversation(id);
            // Update list
            setConversations(prev => prev.filter(c => c.id !== id));
            // If deleting current conversation, clear it
            if (id === currentConversationId) {
                setMessages([]);
                setCurrentConversationId(null);
                localStorage.removeItem(STORAGE_KEYS.AI_MESSAGES);
            }
        } catch (error) {
            console.warn('Failed to delete conversation:', error);
        }
    }, [currentConversationId]);

    // Exposed primarily for ContextPanel interactions
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
            setMessages(prev => [...prev, {
                role: 'assistant',
                content: `❌ Failed to compact context: ${err instanceof Error ? err.message : 'Unknown error'}`
            }]);
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
