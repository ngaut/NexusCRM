import React, { useEffect, useRef } from 'react';
import { Loader2, ChevronDown, ChevronRight } from 'lucide-react';
import { ChatMessage } from '../../infrastructure/api/agent';
import { ProcessStep, ToolBlock } from './types';
import { ToolExecutionSummary } from './ToolExecutionSummary';
import { ProcessStepCard } from './ProcessStepCard';
import { MessageBubble } from './MessageBubble';
import { ConversationSummaryCard } from './ConversationSummaryCard';
import { StreamingContent } from './StreamingContent';
import { EmptyState } from './EmptyState';
import { formatToolName, buildDisplayItems } from './utils';

interface MessageListProps {
    messages: ChatMessage[];
    isLoading: boolean;
    streamingContent: string;
    processSteps: ProcessStep[];
    isProcessExpanded: boolean;
    setIsProcessExpanded: (expanded: boolean) => void;
    onPromptSelect: (prompt: string) => void;
}

export const MessageList: React.FC<MessageListProps> = ({
    messages,
    isLoading,
    streamingContent,
    processSteps,
    isProcessExpanded,
    setIsProcessExpanded,
    onPromptSelect,
}) => {
    const messagesEndRef = useRef<HTMLDivElement>(null);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    };

    useEffect(() => {
        scrollToBottom();
    }, [messages, isLoading, processSteps, streamingContent]);

    return (
        <div className="flex-1 overflow-y-auto p-4 pb-32 space-y-4 bg-gradient-to-b from-slate-50/50 to-white">
            {/* Empty State */}
            {messages.length === 0 && !isLoading && !streamingContent && (
                <EmptyState onPromptSelect={onPromptSelect} />
            )}

            {buildDisplayItems(messages).map((item, idx) => {
                // Skip ToolBlock during active streaming - processSteps shows live progress
                // This prevents duplicate display of "X operations completed" alongside "Running X operations..."
                if ('type' in item && item.type === 'tool_block' && isLoading) {
                    return null;
                }

                // Render Tool Block (only when not loading)
                if ('type' in item && item.type === 'tool_block') {
                    return <ToolExecutionSummary key={item.id} block={item as ToolBlock} formatToolName={formatToolName} />;
                }

                // Render Summary Block
                if ('type' in item && item.type === 'summary_block') {
                    return <ConversationSummaryCard
                        key={item.id}
                        summary={item.summary}
                        stats={item.stats}
                        compactedAt={item.compactedAt}
                        timestamp={item.timestamp}
                    />;
                }

                // Render Chat Message
                const msg = item as ChatMessage;
                return <MessageBubble key={`msg-${idx}`} msg={msg} />;
            })}

            {/* Process Steps - Collapsible Summary (only during active loading) */}
            {isLoading && processSteps.length > 0 && (
                <div className="ml-12">
                    <button
                        onClick={() => setIsProcessExpanded(!isProcessExpanded)}
                        className="inline-flex items-center gap-2 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 rounded-full text-xs text-slate-600 font-medium transition-colors"
                    >
                        <Loader2 size={12} className="animate-spin" />
                        {isProcessExpanded ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
                        <span>
                            {`Running ${processSteps.length} operation${processSteps.length > 1 ? 's' : ''}...`}
                        </span>
                    </button>

                    {isProcessExpanded && (
                        <div className="mt-2 space-y-2 pl-2 border-l-2 border-slate-200">
                            {processSteps.map((step) => (
                                <ProcessStepCard key={step.id} step={step} formatToolName={formatToolName} />
                            ))}
                        </div>
                    )}
                </div>
            )}

            {/* Streaming Content */}
            <StreamingContent content={streamingContent} />

            <div ref={messagesEndRef} />
        </div>
    );
};
