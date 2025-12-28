import React, { useState } from 'react';
import {
    Database,
    X,
    RefreshCw,
    MessageSquare,
    FolderOpen,
    Eye,
    Settings,
    Plus
} from 'lucide-react';
import { ChatMessage, agentApi, ContextItem } from '../../infrastructure/api/agent';
import { ContextInspector } from './ContextInspector';
import { ContextUsageStats } from './ContextUsageStats';
import { ContextSystemTab } from './ContextSystemTab';
import { ContextFileList } from './ContextFileList';
import { ContextChatTab } from './ContextChatTab';

interface ContextFile {
    path: string;
    tokenSize: number;
}

interface ContextPanelProps {
    files: ContextFile[];
    messages: ChatMessage[];
    toolsList: string[];
    systemPromptTokens: number;
    toolsTokens: number;
    totalTokens: number;
    conversationTokens: number;
    maxTokens: number;
    isOpen: boolean;
    hasSummary: boolean;
    tokensSaved: number;
    isLoading: boolean;
    onClose: () => void;
    onRemoveFile: (path: string) => void;
    onRemoveMessage: (index: number) => void;
    onCompact: () => void;
    onRefresh: () => void;
    onAddFiles?: () => void;
    width: number;
}

// System prompt preview
const SYSTEM_PROMPT_PREVIEW = `You are Nexus, an AI assistant for NexusCRM.

PRINCIPLES:
1. EXPLORE BEFORE ACTING - Like exploring a new codebase, first understand what's available, then dive into specifics.
2. TREE EXPLORATION - Start broad (list all), then narrow down (get details), then act (CRUD). Don't try to do everything at once.

You have access to a dynamic CRM system. Objects and fields are metadata-driven. Think step by step. If a tool fails, read the error and adapt.`;

type TabType = 'system' | 'conversation' | 'files';

export function ContextPanel({
    files,
    messages,
    toolsList,
    systemPromptTokens,
    toolsTokens,
    totalTokens,
    conversationTokens,
    maxTokens,
    isOpen,
    hasSummary,
    tokensSaved,
    isLoading,
    onClose,
    onRemoveFile,
    onRemoveMessage,
    onCompact,
    onRefresh,
    onAddFiles,
    width
}: ContextPanelProps) {
    const [activeTab, setActiveTab] = useState<TabType>('conversation');
    const [isInspectorOpen, setIsInspectorOpen] = useState(false);
    const [realSystemPrompt, setRealSystemPrompt] = useState<string | null>(null);
    const [fullFiles, setFullFiles] = useState<ContextItem[] | null>(null);

    // Fetch full context (including content) when inspector opens
    React.useEffect(() => {
        if (isInspectorOpen) {
            agentApi.getContext(true)
                .then(data => {
                    setRealSystemPrompt(data.system_prompt || null);
                    setFullFiles(data.items);
                })
                .catch(err => console.error("Failed to fetch full context inspection:", err));
        }
    }, [isInspectorOpen]);

    if (!isOpen) return null;

    // Filter out system messages for display count
    const displayMessagesCount = messages.filter(m => m.role !== 'system').length;

    return (
        <div
            className="border-r border-slate-200 bg-white flex flex-col h-full animate-slide-right transition-[width] duration-0"
            style={{ width: `${width}px`, minWidth: '320px' }}
        >
            {/* Header */}
            <div className="px-4 py-3 border-b border-slate-200 flex justify-between items-center">
                <div>
                    <h3 className="font-semibold text-sm text-slate-800 flex items-center gap-2">
                        <Database size={16} className="text-indigo-500" />
                        Active Context
                    </h3>
                    <p className="text-[10px] text-slate-400 mt-0.5">Everything the AI receives</p>
                </div>
                <div className="flex items-center gap-0.5">
                    {onAddFiles && (
                        <button
                            onClick={onAddFiles}
                            className="p-1.5 hover:bg-indigo-50 rounded-lg text-indigo-500 hover:text-indigo-600 transition-colors"
                            title="Add files to context"
                        >
                            <Plus size={16} />
                        </button>
                    )}
                    <button
                        onClick={() => setIsInspectorOpen(true)}
                        className="p-1.5 hover:bg-indigo-50 rounded-lg text-slate-400 hover:text-indigo-600 transition-colors"
                        title="Inspect full context prompt"
                    >
                        <Eye size={14} />
                    </button>
                    <button
                        onClick={onRefresh}
                        disabled={isLoading}
                        className="p-1.5 hover:bg-slate-100 rounded-lg text-slate-400 hover:text-slate-600 transition-colors disabled:opacity-50"
                        title="Refresh context"
                    >
                        <RefreshCw size={14} className={isLoading ? 'animate-spin' : ''} />
                    </button>
                    <button
                        onClick={onClose}
                        className="p-1.5 hover:bg-slate-100 rounded-lg text-slate-400 hover:text-slate-600 transition-colors"
                        title="Close panel"
                    >
                        <X size={14} />
                    </button>
                </div>
            </div>

            {/* Token Usage */}
            <ContextUsageStats
                systemPromptTokens={systemPromptTokens}
                toolsTokens={toolsTokens}
                totalTokens={totalTokens}
                conversationTokens={conversationTokens}
                maxTokens={maxTokens}
                hasSummary={hasSummary}
                onCompact={onCompact}
                isLoading={isLoading}
            />

            {/* Summary Indicator */}
            {
                hasSummary && (
                    <div className="px-3 py-2 bg-emerald-50 border-b border-emerald-100 flex items-center gap-2">
                        <Settings size={14} className="text-emerald-600 flex-shrink-0" />
                        <div className="flex-1 min-w-0">
                            <div className="text-[11px] font-medium text-emerald-800">Summarized</div>
                            {tokensSaved > 0 && (
                                <div className="text-[10px] text-emerald-600">Saved ~{tokensSaved.toLocaleString()} tokens</div>
                            )}
                        </div>
                    </div>
                )
            }

            {/* Tabs - 3 tabs */}
            <div className="flex border-b border-white/5">
                <button
                    onClick={() => setActiveTab('system')}
                    className={`flex-1 px-2 py-2 text-[11px] font-medium transition-colors ${activeTab === 'system'
                        ? 'text-indigo-400 border-b-2 border-indigo-500 bg-indigo-500/10'
                        : 'text-slate-500 hover:text-slate-300 hover:bg-white/5'
                        }`}
                >
                    <Settings size={11} className="inline mr-1" />
                    System
                </button>
                <button
                    onClick={() => setActiveTab('conversation')}
                    className={`flex-1 px-2 py-2 text-[11px] font-medium transition-colors ${activeTab === 'conversation'
                        ? 'text-indigo-400 border-b-2 border-indigo-500 bg-indigo-500/10'
                        : 'text-slate-500 hover:text-slate-300 hover:bg-white/5'
                        }`}
                >
                    <MessageSquare size={11} className="inline mr-1" />
                    Chat ({displayMessagesCount})
                </button>
                <button
                    onClick={() => setActiveTab('files')}
                    className={`flex-1 px-2 py-2 text-[11px] font-medium transition-colors ${activeTab === 'files'
                        ? 'text-indigo-400 border-b-2 border-indigo-500 bg-indigo-500/10'
                        : 'text-slate-500 hover:text-slate-300 hover:bg-white/5'
                        }`}
                >
                    <FolderOpen size={11} className="inline mr-1" />
                    Files ({files.length})
                </button>
            </div>

            {/* Content Area */}
            <div className="flex-1 overflow-y-auto">
                {activeTab === 'system' ? (
                    <ContextSystemTab
                        systemPromptTokens={systemPromptTokens}
                        toolsTokens={toolsTokens}
                        toolsList={toolsList}
                        systemPromptPreview={SYSTEM_PROMPT_PREVIEW}
                    />
                ) : activeTab === 'conversation' ? (
                    <ContextChatTab
                        messages={messages}
                        hasSummary={hasSummary}
                        onRemoveMessage={onRemoveMessage}
                    />
                ) : (
                    <ContextFileList
                        files={files}
                        onAddFiles={onAddFiles}
                        onRemoveFile={onRemoveFile}
                    />
                )}
            </div>
            <ContextInspector
                isOpen={isInspectorOpen}
                onClose={() => setIsInspectorOpen(false)}
                systemPrompt={realSystemPrompt || SYSTEM_PROMPT_PREVIEW}
                files={fullFiles ? fullFiles.map(f => ({ path: f.path, tokenSize: f.token_size })) : files}
                messages={messages}
            />
        </div>
    );
}
