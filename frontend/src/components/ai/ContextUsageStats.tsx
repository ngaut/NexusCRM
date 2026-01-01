import React from 'react';
import { AlertTriangle, Minimize2 } from 'lucide-react';

interface ContextUsageStatsProps {
    systemPromptTokens: number;
    toolsTokens: number;
    totalTokens: number;
    conversationTokens: number;
    maxTokens: number;
    hasSummary: boolean;
    onCompact: () => void;
    isLoading: boolean;
}

export const ContextUsageStats: React.FC<ContextUsageStatsProps> = ({
    systemPromptTokens,
    toolsTokens,
    totalTokens,
    conversationTokens,
    maxTokens,
    hasSummary,
    onCompact,
    isLoading
}) => {
    // Calculate total including system overhead
    const systemOverhead = systemPromptTokens + toolsTokens;
    const allTokens = systemOverhead + totalTokens + conversationTokens;
    const usagePercent = Math.min((allTokens / maxTokens) * 100, 100);

    // Individual percentages for segmented bar
    const systemPercent = (systemPromptTokens / maxTokens) * 100;
    const toolsPercent = (toolsTokens / maxTokens) * 100;
    const filesPercent = (totalTokens / maxTokens) * 100;
    const chatPercent = (conversationTokens / maxTokens) * 100;

    const showWarning = usagePercent >= 70;
    const showCompactButton = conversationTokens > 500;

    return (
        <div className="p-3 border-b border-white/5 bg-black/20">
            <div className="flex justify-between items-baseline mb-1.5">
                <span className="text-[11px] font-medium text-slate-400">Token Usage</span>
                <span className={`text-[11px] font-mono ${usagePercent > 70 ? 'text-amber-400' : 'text-slate-500'}`}>
                    {allTokens.toLocaleString()} / {(maxTokens / 1000).toFixed(0)}K
                </span>
            </div>

            {/* Segmented Progress Bar - 4 segments */}
            <div className="w-full h-2 bg-slate-800 rounded-full overflow-hidden flex">
                <div
                    className="h-full bg-slate-500 transition-all duration-500"
                    style={{ width: `${systemPercent}%` }}
                    title={`System Prompt: ~${systemPromptTokens.toLocaleString()} tokens`}
                />
                <div
                    className="h-full bg-purple-400 transition-all duration-500"
                    style={{ width: `${toolsPercent}%` }}
                    title={`Tools: ~${toolsTokens.toLocaleString()} tokens`}
                />
                <div
                    className="h-full bg-emerald-500 transition-all duration-500"
                    style={{ width: `${filesPercent}%` }}
                    title={`Files: ~${totalTokens.toLocaleString()} tokens`}
                />
                <div
                    className="h-full bg-indigo-400 transition-all duration-500"
                    style={{ width: `${chatPercent}%` }}
                    title={`Chat: ~${conversationTokens.toLocaleString()} tokens`}
                />
            </div>

            {/* Legend - 2 rows */}
            <div className="mt-2 grid grid-cols-2 gap-x-2 gap-y-1 text-[10px]">
                <div className="flex items-center gap-1">
                    <div className="w-2 h-2 rounded-full bg-slate-400" />
                    <span className="text-slate-500">System {systemPromptTokens}</span>
                </div>
                <div className="flex items-center gap-1">
                    <div className="w-2 h-2 rounded-full bg-purple-400" />
                    <span className="text-slate-500">Tools {toolsTokens}</span>
                </div>
                <div className="flex items-center gap-1">
                    <div className="w-2 h-2 rounded-full bg-emerald-500" />
                    <span className="text-slate-500">Files {totalTokens}</span>
                </div>
                <div className="flex items-center gap-1">
                    <div className="w-2 h-2 rounded-full bg-indigo-400" />
                    <span className="text-slate-500">Chat {conversationTokens}</span>
                </div>
            </div>

            {/* Warning & Compact */}
            {showWarning && (
                <div className="mt-2 p-2 bg-amber-900/20 border border-amber-900/40 rounded-lg flex items-start gap-2">
                    <AlertTriangle size={12} className="text-amber-500 mt-0.5 flex-shrink-0" />
                    <div className="text-[10px] text-amber-500">Context getting full</div>
                </div>
            )}
            {showCompactButton && !hasSummary && (
                <button
                    onClick={onCompact}
                    disabled={isLoading}
                    className="mt-2 w-full flex items-center justify-center gap-1.5 px-2 py-1.5 bg-indigo-50 hover:bg-indigo-100 border border-indigo-200 text-indigo-600 text-[11px] font-medium rounded-lg transition-colors disabled:opacity-50"
                >
                    <Minimize2 size={12} />
                    Compact Conversation
                </button>
            )}
        </div>
    );
};
