import React, { useState } from 'react';
import { Lock, Cpu, ChevronUp, ChevronDown, Wrench } from 'lucide-react';

interface ContextSystemTabProps {
    systemPromptTokens: number;
    toolsTokens: number;
    toolsList: string[];
    systemPromptPreview: string;
}

export const ContextSystemTab: React.FC<ContextSystemTabProps> = ({
    systemPromptTokens,
    toolsTokens,
    toolsList,
    systemPromptPreview
}) => {
    const [toolsExpanded, setToolsExpanded] = useState(false);

    return (
        <div className="p-3 space-y-3">
            {/* System Prompt Section */}
            <div className="bg-slate-50 rounded-lg p-3">
                <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-1.5">
                        <Lock size={12} className="text-slate-400" />
                        <span className="text-[11px] font-medium text-slate-600">System Prompt</span>
                    </div>
                    <span className="text-[10px] text-slate-400">~{systemPromptTokens} tokens</span>
                </div>
                <div className="text-[10px] text-slate-500 leading-relaxed whitespace-pre-wrap max-h-32 overflow-y-auto bg-white rounded p-2 border border-slate-100">
                    {systemPromptPreview}
                </div>
                <div className="mt-1.5 text-[9px] text-slate-400 flex items-center gap-1">
                    <Lock size={9} />
                    Read-only (fixed overhead)
                </div>
            </div>

            {/* Tools Section */}
            <div className="bg-purple-50 rounded-lg p-3">
                <button
                    onClick={() => setToolsExpanded(!toolsExpanded)}
                    className="w-full flex items-center justify-between"
                >
                    <div className="flex items-center gap-1.5">
                        <Cpu size={12} className="text-purple-500" />
                        <span className="text-[11px] font-medium text-purple-700">
                            Tools ({toolsList.length})
                        </span>
                    </div>
                    <div className="flex items-center gap-2">
                        <span className="text-[10px] text-purple-400">~{toolsTokens} tok</span>
                        {toolsExpanded ? (
                            <ChevronUp size={12} className="text-purple-400" />
                        ) : (
                            <ChevronDown size={12} className="text-purple-400" />
                        )}
                    </div>
                </button>

                {toolsExpanded && (
                    <div className="mt-2 space-y-1 max-h-48 overflow-y-auto">
                        {toolsList.map((tool, idx) => (
                            <div
                                key={idx}
                                className="flex items-center gap-2 px-2 py-1.5 bg-white rounded text-[10px] text-slate-600"
                            >
                                <Wrench size={10} className="text-purple-400" />
                                {tool}
                            </div>
                        ))}
                    </div>
                )}

                <div className="mt-1.5 text-[9px] text-purple-400 flex items-center gap-1">
                    <Lock size={9} />
                    Available tools (fixed overhead)
                </div>
            </div>
        </div>
    );
};
