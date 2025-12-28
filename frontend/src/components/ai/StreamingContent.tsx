
import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Bot } from 'lucide-react';

interface StreamingContentProps {
    content: string;
}

export function StreamingContent({ content }: StreamingContentProps) {
    if (!content) return null;

    return (
        <div className="flex gap-3">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-emerald-400 to-teal-500 text-white flex items-center justify-center flex-shrink-0 shadow-sm">
                <Bot size={18} />
            </div>
            <div className="max-w-[85%] rounded-2xl rounded-tl-sm px-4 py-3 text-sm bg-white border border-slate-100 text-slate-700 shadow-sm">
                <div className="prose">
                    <ReactMarkdown remarkPlugins={[remarkGfm]}>
                        {content}
                    </ReactMarkdown>
                    <span className="inline-block w-2 h-4 bg-indigo-500 animate-pulse ml-0.5 rounded-sm" />
                </div>
            </div>
        </div>
    );
}
