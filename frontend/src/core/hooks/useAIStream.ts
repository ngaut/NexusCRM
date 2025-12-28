import { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import { agentApi, ChatMessage, StreamEvent, CompactRequest } from '../../infrastructure/api/agent';
import { ProcessStep } from '../../components/ai/types';

const STORAGE_KEY_MESSAGES = 'nexus_ai_messages';
const STORAGE_KEY_PROCESS = 'nexus_ai_process_steps';

export function useAIStream() {
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [processSteps, setProcessSteps] = useState<ProcessStep[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [streamingContent, setStreamingContent] = useState<string>('');
    const [isProcessExpanded, setIsProcessExpanded] = useState(false);

    const abortControllerRef = useRef<AbortController | null>(null);

    // Persistence
    useEffect(() => {
        try {
            const savedMessages = localStorage.getItem(STORAGE_KEY_MESSAGES);
            const savedProcess = localStorage.getItem(STORAGE_KEY_PROCESS);
            if (savedMessages) setMessages(JSON.parse(savedMessages));
            if (savedProcess) setProcessSteps(JSON.parse(savedProcess));
        } catch (e) {
            console.error('Failed to load persisted AI state:', e);
        }
    }, []);

    useEffect(() => {
        localStorage.setItem(STORAGE_KEY_MESSAGES, JSON.stringify(messages));
    }, [messages]);

    useEffect(() => {
        localStorage.setItem(STORAGE_KEY_PROCESS, JSON.stringify(processSteps));
    }, [processSteps]);

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
                setProcessSteps(prev => [
                    ...prev.filter(s => s.type !== 'thinking'),
                    {
                        id: stepId,
                        type: 'thinking',
                        content: event.content || 'Thinking...',
                        timestamp: new Date()
                    }
                ]);
                break;

            case 'tool_call':
                setProcessSteps(prev => [
                    ...prev.filter(s => s.type !== 'thinking'),
                    {
                        id: stepId,
                        type: 'tool_call',
                        content: `Calling ${event.tool_name}`,
                        toolName: event.tool_name,
                        toolArgs: event.tool_args,
                        timestamp: new Date()
                    }
                ]);

                if (event.tool_name && event.tool_args) {
                    setMessages(prev => {
                        const lastMsg = prev[prev.length - 1];
                        if (lastMsg && lastMsg.role === 'assistant' && lastMsg.tool_calls) {
                            return [...prev.slice(0, -1), {
                                ...lastMsg,
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
                            return [...prev, {
                                role: 'assistant',
                                content: '',
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
                setProcessSteps(prev => prev.filter(s => s.type !== 'thinking'));
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
        setIsProcessExpanded(true);

        abortControllerRef.current = agentApi.chatStream(
            { messages: newHistory },
            handleStreamEvent,
            (error) => {
                console.error('Stream error:', error);
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
        localStorage.removeItem(STORAGE_KEY_MESSAGES);
        localStorage.removeItem(STORAGE_KEY_PROCESS);
    };

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
            console.error(err);
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
        summaryInfo
    };
}
