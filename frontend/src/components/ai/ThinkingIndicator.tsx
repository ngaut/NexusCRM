import React from 'react';
import { Loader2 } from 'lucide-react';

interface ThinkingIndicatorProps {
    message?: string;
    step?: number;
    maxSteps?: number;
}

export function ThinkingIndicator({ message = 'Thinking...', step, maxSteps }: ThinkingIndicatorProps) {
    return (
        <div className="flex items-center gap-2.5 px-4 py-2.5 bg-gradient-to-r from-slate-50 to-slate-100 rounded-xl border border-slate-200/60">
            <Loader2 size={16} className="text-indigo-500 animate-spin flex-shrink-0" />
            <span className="text-sm text-slate-600 font-medium">{message}</span>
            {step && maxSteps && (
                <span className="text-xs text-slate-400 ml-auto">
                    {step}/{maxSteps}
                </span>
            )}
            <div className="flex gap-1 ml-auto">
                <div className="w-1.5 h-1.5 bg-indigo-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                <div className="w-1.5 h-1.5 bg-indigo-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                <div className="w-1.5 h-1.5 bg-indigo-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
            </div>
        </div>
    );
}
