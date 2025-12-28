
import React, { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { User, Bot, Copy, Check } from 'lucide-react';
import { ChatMessage } from '../../infrastructure/api/agent';
import { formatRelativeTime, formatFullTime } from './utils';

interface MessageBubbleProps {
    msg: ChatMessage;
}

export function MessageBubble({ msg }: MessageBubbleProps) {
    const [copied, setCopied] = useState(false);
    const isUser = msg.role === 'user';

    const handleCopy = async () => {
        try {
            await navigator.clipboard.writeText(msg.content || '');
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.warn('Failed to copy to clipboard:', err);
        }
    };

    return (
        <div className={`group relative flex gap-3 ${isUser ? 'flex-row-reverse' : ''} animate-slide-up pb-4`}>
            <div className={`w-9 h-9 rounded-xl flex items-center justify-center flex-shrink-0 shadow-sm ${isUser
                ? 'bg-gradient-to-br from-indigo-500 to-purple-600 text-white'
                : 'bg-gradient-to-br from-emerald-400 to-teal-500 text-white'
                }`}>
                {isUser ? <User size={18} /> : <Bot size={18} />}
            </div>
            <div className="relative max-w-[85%]">
                <div className={`rounded-2xl px-4 py-3 text-sm shadow-sm ${isUser
                    ? 'bg-gradient-to-br from-indigo-600 to-purple-600 text-white rounded-tr-sm'
                    : 'bg-white border border-slate-100 text-slate-700 rounded-tl-sm'
                    }`}>
                    {isUser ? (
                        <div className="whitespace-pre-wrap leading-relaxed">{msg.content}</div>
                    ) : (
                        <div className="prose prose-sm prose-slate max-w-none prose-p:my-1.5 prose-headings:my-2 prose-ul:my-1.5 prose-ol:my-1.5 prose-li:my-0.5 prose-pre:my-2 prose-code:text-indigo-600 prose-code:bg-indigo-50 prose-code:px-1 prose-code:py-0.5 prose-code:rounded prose-code:text-xs">
                            <ReactMarkdown remarkPlugins={[remarkGfm]}>
                                {msg.content || ''}
                            </ReactMarkdown>
                        </div>
                    )}
                </div>
                {/* Copy button - only show for assistant messages */}
                {!isUser && msg.content && (
                    <button
                        onClick={handleCopy}
                        className="absolute -bottom-2 right-2 opacity-0 group-hover:opacity-100 p-1.5 bg-white border border-slate-200 rounded-lg shadow-sm hover:bg-slate-50 transition-all duration-150 text-slate-400 hover:text-slate-600"
                        title={copied ? 'Copied!' : 'Copy message'}
                    >
                        {copied ? (
                            <Check size={14} className="text-emerald-500" />
                        ) : (
                            <Copy size={14} />
                        )}
                    </button>
                )}
            </div>
            {/* Timestamp */}
            {msg.timestamp && (
                <div className={`absolute -bottom-5 ${isUser ? 'right-12' : 'left-12'} opacity-0 group-hover:opacity-100 transition-opacity text-[10px] text-slate-400 select-none`} title={formatFullTime(msg.timestamp)}>
                    {formatRelativeTime(msg.timestamp)}
                </div>
            )}
        </div>
    );
}
