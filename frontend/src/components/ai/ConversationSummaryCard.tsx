
import React, { useState } from 'react';
import { ChevronDown, History } from 'lucide-react';
import ReactMarkdown from 'react-markdown';

interface ConversationSummaryCardProps {
    summary: string;
    stats?: string;
    compactedAt?: string;
    timestamp?: Date | string;
}

export function ConversationSummaryCard({ summary, stats, compactedAt, timestamp }: ConversationSummaryCardProps) {
    const [isExpanded, setIsExpanded] = useState(false);

    return (
        <div className="mx-4 my-2 animate-fade-in group/card">
            <div
                className={`
          rounded-xl overflow-hidden border transition-all duration-200
          ${isExpanded
                        ? 'bg-slate-50 border-slate-200 shadow-sm'
                        : 'bg-white border-slate-200 hover:border-slate-300 hover:bg-slate-50/50'
                    }
        `}
            >
                <button
                    onClick={() => setIsExpanded(!isExpanded)}
                    className="w-full flex items-center justify-between p-3 group"
                >
                    <div className="flex items-center gap-3 overflow-hidden">
                        <div className={`
              w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 transition-colors
              ${isExpanded ? 'bg-slate-200 text-slate-700' : 'bg-slate-100 text-slate-500 group-hover:bg-slate-200 group-hover:text-slate-600'}
            `}>
                            <History size={16} />
                        </div>

                        <div className="text-left flex-1 min-w-0">
                            <div className="flex items-center gap-2">
                                <h4 className={`text-sm font-medium transition-colors ${isExpanded ? 'text-slate-800' : 'text-slate-700'}`}>
                                    Previous Context Summary
                                </h4>
                                {(stats || compactedAt) && (
                                    <div className="flex items-center gap-1.5">
                                        {compactedAt && (
                                            <span className="text-[10px] text-slate-400 font-medium whitespace-nowrap">
                                                • {compactedAt}
                                            </span>
                                        )}
                                        {stats && (
                                            <span className="text-[10px] font-medium px-1.5 py-0.5 rounded-full bg-emerald-100 text-emerald-700 border border-emerald-200/50 whitespace-nowrap">
                                                {stats}
                                            </span>
                                        )}
                                    </div>
                                )}
                            </div>

                            {!isExpanded && (
                                <p className="text-xs text-slate-500 truncate max-w-full opacity-80 group-hover:opacity-100 transition-opacity">
                                    {summary.split('\n')[0].replace(/^[#-]\s*/, '')}
                                </p>
                            )}
                        </div>
                    </div>

                    <div className={`
            text-slate-400 transition-transform duration-200 ml-2 flex-shrink-0
            ${isExpanded ? 'rotate-180 text-slate-600' : 'group-hover:text-slate-600'}
          `}>
                        <ChevronDown size={18} />
                    </div>
                </button>

                {isExpanded && (
                    <div className="px-4 pb-4 pt-1 border-t border-slate-200/50">
                        <div className="prose prose-sm prose-slate max-w-none prose-p:my-1 prose-headings:my-2 text-slate-600">
                            <ReactMarkdown>{summary}</ReactMarkdown>
                        </div>
                        <div className="mt-3 text-[10px] text-slate-400 font-medium uppercase tracking-wider text-right">
                            Archived {new Date(timestamp || Date.now()).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
