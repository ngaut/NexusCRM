
import React, { useState } from 'react';
import { ChevronDown, ChevronRight, Loader2, Cpu, AlertCircle, Check } from 'lucide-react';
import { ProcessStep } from './types';
import { formatFullTime } from './utils';

interface ProcessStepCardProps {
    step: ProcessStep;
    formatToolName: (name: string) => string;
}

export function ProcessStepCard({ step, formatToolName }: ProcessStepCardProps) {
    const [isExpanded, setIsExpanded] = useState(false);

    const getIcon = () => {
        if (step.type === 'thinking') {
            return <Loader2 size={14} className="text-indigo-500 animate-spin" />;
        }
        if (step.type === 'tool_call') {
            if (step.isDone) {
                return <Cpu size={14} className="text-amber-600" />;
            }
            return <Loader2 size={14} className="text-amber-500 animate-spin" />;
        }
        if (step.isError) {
            return <AlertCircle size={14} className="text-red-500" />;
        }
        return <Check size={14} className="text-emerald-500" />;
    };

    const getBgColor = () => {
        if (step.type === 'thinking') return 'bg-indigo-50 border-indigo-100';
        if (step.type === 'tool_call') return 'bg-amber-50 border-amber-100';
        if (step.isError) return 'bg-red-50 border-red-100';
        return 'bg-emerald-50 border-emerald-100';
    };

    const hasDetails = step.toolArgs || step.toolResult;

    return (
        <div className={`rounded-lg border ${getBgColor()} overflow-hidden text-xs`}>
            <button
                onClick={() => hasDetails && setIsExpanded(!isExpanded)}
                className={`w-full flex items-center gap-2 px-3 py-2 ${hasDetails ? 'cursor-pointer hover:bg-white/50' : 'cursor-default'}`}
                disabled={!hasDetails}
            >
                {getIcon()}
                <Cpu size={12} className="text-slate-400" />
                <span className="flex-1 text-left font-medium text-slate-700">
                    {step.toolName ? formatToolName(step.toolName) : step.content}
                </span>
                <span className="text-[10px] text-slate-400 font-mono mr-1" title={formatFullTime(step.timestamp)}>
                    {step.timestamp && new Date(step.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false })}
                </span>
                {hasDetails && (
                    isExpanded ? <ChevronDown size={12} /> : <ChevronRight size={12} />
                )}
            </button>

            {isExpanded && hasDetails && (
                <div className="px-3 pb-3 space-y-2 border-t border-slate-200/50 bg-white/50">
                    {step.toolArgs && (
                        <div className="mt-2">
                            <div className="text-[10px] uppercase tracking-wider font-semibold text-slate-400 mb-1">Arguments</div>
                            <pre className="text-[11px] bg-slate-100 p-2 rounded overflow-x-auto font-mono text-slate-600 max-h-32 overflow-y-auto">
                                {(() => {
                                    try {
                                        return JSON.stringify(JSON.parse(step.toolArgs), null, 2);
                                    } catch {
                                        return step.toolArgs;
                                    }
                                })()}
                            </pre>
                        </div>
                    )}
                    {step.toolResult && (
                        <div>
                            <div className="text-[10px] uppercase tracking-wider font-semibold text-slate-400 mb-1">Result</div>
                            <pre className={`text-[11px] p-2 rounded overflow-x-auto font-mono max-h-40 overflow-y-auto ${step.isError ? 'bg-red-100 text-red-700' : 'bg-emerald-50 text-emerald-700'
                                }`}>
                                {step.toolResult}
                            </pre>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
