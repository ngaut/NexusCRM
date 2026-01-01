
import React, { useState } from 'react';
import { ChevronDown, ChevronRight, Check, AlertCircle } from 'lucide-react';
import { ToolBlock } from './types';
import { ProcessStepCard } from './ProcessStepCard';

interface ToolExecutionSummaryProps {
    block: ToolBlock;
    formatToolName: (name: string) => string;
}

export function ToolExecutionSummary({ block, formatToolName }: ToolExecutionSummaryProps) {
    const [isExpanded, setIsExpanded] = useState(false);
    const toolCount = block.tools.length;
    const isAllSuccess = block.tools.every(t => !t.result?.isError);

    return (
        <div className="ml-12 mb-4 animate-slide-up">
            <button
                onClick={() => setIsExpanded(!isExpanded)}
                className="flex items-center gap-2 px-3 py-2 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-lg text-xs font-medium text-slate-600 transition-colors w-full sm:w-auto"
            >
                {isAllSuccess ? (
                    <Check size={14} className="text-emerald-500" />
                ) : (
                    <AlertCircle size={14} className="text-amber-500" />
                )}
                <span>
                    {toolCount} operation{toolCount !== 1 ? 's' : ''} completed
                </span>
                {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
            </button>

            {isExpanded && (
                <div className="mt-2 space-y-2 pl-2 border-l-2 border-slate-200">
                    {block.tools.map((tool, idx) => (
                        <React.Fragment key={`${block.id}-${idx}`}>
                            <ProcessStepCard step={tool.call} formatToolName={formatToolName} />
                            {tool.result && (
                                <ProcessStepCard step={tool.result} formatToolName={formatToolName} />
                            )}
                        </React.Fragment>
                    ))}
                </div>
            )}
        </div>
    );
}
