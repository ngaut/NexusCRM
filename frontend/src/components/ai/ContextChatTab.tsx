import React, { useState } from 'react';
import {
    MessageSquare,
    History,
    ChevronUp,
    ChevronDown,
    Trash2,
    User,
    Bot,
    Wrench,
    Database
} from 'lucide-react';
import { ChatMessage } from '../../infrastructure/api/agent';
import { formatRelativeTime, formatFullTime } from './utils';

interface ContextChatTabProps {
    messages: ChatMessage[];
    hasSummary: boolean;
    onRemoveMessage: (index: number) => void;
}

// Estimate tokens for a message
function estimateTokens(msg: ChatMessage): number {
    let tokens = Math.ceil((msg.content || '').length / 4);
    if (msg.tool_calls) {
        tokens += msg.tool_calls.reduce((acc, tc) => acc + Math.ceil((tc.function.arguments || '').length / 4), 0);
    }
    return tokens;
}

// Get role icon and color - Dark Mode Adjusted
function getRoleStyle(role: string) {
    switch (role) {
        case 'user':
            return { icon: User, bg: 'bg-indigo-500/10', text: 'text-indigo-300', label: 'You' };
        case 'assistant':
            return { icon: Bot, bg: 'bg-emerald-500/10', text: 'text-emerald-300', label: 'AI' };
        case 'tool':
            return { icon: Wrench, bg: 'bg-amber-500/10', text: 'text-amber-300', label: 'Tool' };
        case 'system':
            return { icon: Database, bg: 'bg-slate-800/50', text: 'text-slate-400', label: 'System' };
        default:
            return { icon: MessageSquare, bg: 'bg-slate-800/50', text: 'text-slate-400', label: role };
    }
}

export const ContextChatTab: React.FC<ContextChatTabProps> = ({
    messages,
    hasSummary,
    onRemoveMessage
}) => {
    const [expandedMessages, setExpandedMessages] = useState<Set<number>>(new Set());
    const [summaryExpanded, setSummaryExpanded] = useState(false);

    // Extract summary from system message if present
    const systemMessage = messages.find(m => m.role === 'system');
    let summaryContent: string | null = null;
    let summaryTokensSaved: number | null = null;
    let compactedAt: string | null = null;

    if (systemMessage?.content) {
        // Match: --- CONVERSATION SUMMARY (Saved ~123 tokens | Compacted: Dec 28, 11:30 AM) ---
        const summaryMatch = systemMessage.content.match(/--- CONVERSATION SUMMARY \(Saved ~(\d+) tokens(?: \| Compacted: ([^)]+))?\) ---\n([\s\S]*?)\n--- END SUMMARY ---/);
        if (summaryMatch) {
            summaryTokensSaved = parseInt(summaryMatch[1], 10);
            compactedAt = summaryMatch[2] || null;
            summaryContent = summaryMatch[3].trim();
        }
    }

    // Filter out system messages for display
    const displayMessages = messages.filter(m => m.role !== 'system');

    const toggleMessageExpand = (index: number) => {
        const newExpanded = new Set(expandedMessages);
        if (newExpanded.has(index)) {
            newExpanded.delete(index);
        } else {
            newExpanded.add(index);
        }
        setExpandedMessages(newExpanded);
    };

    if (displayMessages.length === 0 && !summaryContent) {
        return (
            <div className="flex flex-col items-center justify-center h-full px-6 py-8 text-center">
                <MessageSquare size={32} className="text-slate-300 mb-2" />
                <p className="text-xs text-slate-400">No conversation yet</p>
            </div>
        );
    }

    return (
        <div className="p-2 space-y-1">
            {/* Summary Card - shown when context has been compacted */}
            {summaryContent && (
                <div className="bg-gradient-to-r from-emerald-50 to-teal-50 rounded-lg p-3 border border-emerald-200 mb-2">
                    <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center gap-2">
                            <div className="w-6 h-6 rounded-md bg-emerald-100 flex items-center justify-center">
                                <History size={12} className="text-emerald-600" />
                            </div>
                            <div>
                                <div className="flex items-center gap-1.5">
                                    <span className="text-[11px] font-semibold text-emerald-700">Previous Context</span>
                                    {compactedAt && (
                                        <span className="text-[9px] text-emerald-500">• {compactedAt}</span>
                                    )}
                                </div>
                                {summaryTokensSaved && (
                                    <span className="text-[9px] text-emerald-500">
                                        saved ~{summaryTokensSaved.toLocaleString()} tokens
                                    </span>
                                )}
                            </div>
                        </div>
                        <button
                            onClick={() => setSummaryExpanded(!summaryExpanded)}
                            className="p-1 hover:bg-emerald-100 rounded transition-colors"
                        >
                            {summaryExpanded ? (
                                <ChevronUp size={14} className="text-emerald-500" />
                            ) : (
                                <ChevronDown size={14} className="text-emerald-500" />
                            )}
                        </button>
                    </div>

                    <div className={`text-[10px] text-emerald-700 leading-relaxed ${summaryExpanded ? '' : 'line-clamp-2'}`}>
                        {summaryContent}
                    </div>

                    {!summaryExpanded && summaryContent.length > 120 && (
                        <button
                            onClick={() => setSummaryExpanded(true)}
                            className="text-[10px] text-emerald-600 hover:text-emerald-700 mt-1 font-medium"
                        >
                            Show full summary
                        </button>
                    )}
                </div>
            )}

            {displayMessages.length === 0 ? (
                <div className="text-center py-4">
                    <p className="text-[11px] text-slate-400">Recent messages cleared by compaction</p>
                </div>
            ) : (
                displayMessages.map((msg, idx) => {
                    const style = getRoleStyle(msg.role);
                    const IconComponent = style.icon;
                    const tokens = estimateTokens(msg);
                    const isExpanded = expandedMessages.has(idx);

                    // Format content for display, handling tool calls and results
                    let displayContent = msg.content || '';
                    if (msg.role === 'assistant' && msg.tool_calls && msg.tool_calls.length > 0) {
                        const toolNames = msg.tool_calls.map(tc => tc.function.name).join(', ');
                        const label = `[Calling: ${toolNames}]`;
                        displayContent = displayContent ? `${displayContent}\n${label}` : label;
                    } else if (msg.role === 'tool') {
                        displayContent = `[Result (${msg.name || 'Tool'}): ${displayContent}]`;
                    }

                    const preview = displayContent.slice(0, 60);
                    const hasMore = displayContent.length > 60;
                    const relTime = formatRelativeTime(msg.timestamp);
                    const fullTime = formatFullTime(msg.timestamp);

                    const originalIndex = messages.indexOf(msg);

                    return (
                        <div
                            key={idx}
                            className="group bg-slate-50 hover:bg-slate-100 rounded-lg p-2 transition-colors"
                        >
                            <div className="flex items-start gap-2">
                                <div className={`w-6 h-6 rounded-md ${style.bg} flex items-center justify-center flex-shrink-0`}>
                                    <IconComponent size={12} className={style.text} />
                                </div>
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center justify-between mb-0.5">
                                        <span className={`text-[10px] font-medium ${style.text}`}>{style.label}</span>
                                        <span className="text-[9px] text-slate-400" title={fullTime}>
                                            {relTime ? `${relTime} • ` : ''}~{tokens} tok
                                        </span>
                                    </div>
                                    <div className="text-[11px] text-slate-600 leading-relaxed">
                                        {isExpanded ? displayContent : preview}
                                        {!isExpanded && hasMore && '...'}
                                    </div>
                                    {hasMore && (
                                        <button
                                            onClick={() => toggleMessageExpand(idx)}
                                            className="text-[10px] text-indigo-500 hover:text-indigo-600 mt-1 flex items-center gap-0.5"
                                        >
                                            {isExpanded ? (
                                                <>Show less <ChevronUp size={10} /></>
                                            ) : (
                                                <>Show more <ChevronDown size={10} /></>
                                            )}
                                        </button>
                                    )}
                                </div>
                                <button
                                    onClick={() => onRemoveMessage(originalIndex)}
                                    className="p-1 rounded text-slate-300 hover:text-red-500 hover:bg-red-50 transition-all opacity-0 group-hover:opacity-100"
                                    title="Remove from context"
                                >
                                    <Trash2 size={12} />
                                </button>
                            </div>
                        </div>
                    );
                })
            )}

            {displayMessages.length > 0 && (
                <button
                    onClick={() => {
                        for (let i = messages.length - 1; i >= 0; i--) {
                            if (messages[i].role !== 'system') {
                                onRemoveMessage(i);
                            }
                        }
                    }}
                    className="w-full mt-2 text-[10px] text-slate-400 hover:text-red-500 transition-colors py-1"
                >
                    Clear all conversation
                </button>
            )}
        </div>
    );
};
