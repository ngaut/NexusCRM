import React, { useState } from 'react';
import { Cpu, ChevronDown, ChevronUp, Check, X, Loader2 } from 'lucide-react';

interface ToolCallCardProps {
    toolName: string;
    toolArgs?: string;
    toolResult?: string;
    isError?: boolean;
    status: 'pending' | 'running' | 'complete' | 'error';
}

export function ToolCallCard({ toolName, toolArgs, toolResult, isError, status }: ToolCallCardProps) {
    const [isExpanded, setIsExpanded] = useState(false);

    const formatToolName = (name: string) => {
        // Convert create_Account -> Create Account
        return name.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
    };

    const statusIcon = () => {
        switch (status) {
            case 'pending':
                return <div className="w-4 h-4 rounded-full bg-slate-200" />;
            case 'running':
                return <Loader2 size={16} className="text-indigo-500 animate-spin" />;
            case 'complete':
                return <Check size={16} className="text-emerald-500" />;
            case 'error':
                return <X size={16} className="text-red-500" />;
        }
    };

    const statusBorder = () => {
        switch (status) {
            case 'running':
                return 'border-indigo-200 bg-indigo-50/50';
            case 'complete':
                return 'border-emerald-200 bg-emerald-50/50';
            case 'error':
                return 'border-red-200 bg-red-50/50';
            default:
                return 'border-slate-200 bg-slate-50/50';
        }
    };

    return (
        <div className={`rounded-lg border ${statusBorder()} overflow-hidden transition-all duration-200`}>
            {/* Header */}
            <button
                onClick={() => setIsExpanded(!isExpanded)}
                className="w-full flex items-center gap-2 px-3 py-2 hover:bg-white/50 transition-colors"
            >
                {statusIcon()}
                <Cpu size={14} className="text-slate-500" />
                <span className="flex-1 text-left text-sm font-medium text-slate-700">
                    {formatToolName(toolName)}
                </span>
                {(toolArgs || toolResult) && (
                    isExpanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />
                )}
            </button>

            {/* Expandable Content */}
            {isExpanded && (toolArgs || toolResult) && (
                <div className="px-3 pb-3 space-y-2 border-t border-slate-200/50 animate-fade-in">
                    {toolArgs && (
                        <div className="mt-2">
                            <div className="text-xs font-medium text-slate-500 mb-1">Arguments</div>
                            <pre className="text-xs bg-slate-100 p-2 rounded overflow-x-auto font-mono text-slate-600">
                                {(() => {
                                    try {
                                        return JSON.stringify(JSON.parse(toolArgs), null, 2);
                                    } catch {
                                        return toolArgs;
                                    }
                                })()}
                            </pre>
                        </div>
                    )}
                    {toolResult && (
                        <div>
                            <div className="text-xs font-medium text-slate-500 mb-1">Result</div>
                            <pre className={`text-xs p-2 rounded overflow-x-auto font-mono ${isError ? 'bg-red-100 text-red-700' : 'bg-emerald-50 text-emerald-700'
                                }`}>
                                {toolResult}
                            </pre>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
