import React from 'react';
import { Sparkles, Search, Database, HelpCircle, MessageSquare, Zap } from 'lucide-react';

interface QuickAction {
    icon: React.ElementType;
    title: string;
    description: string;
    prompt: string;
    gradient: string;
}

const QUICK_ACTIONS: QuickAction[] = [
    {
        icon: Search,
        title: 'Query Your Data',
        description: 'Search and filter any records',
        prompt: 'Show me all recent records',
        gradient: 'from-blue-500 to-cyan-500'
    },
    {
        icon: Database,
        title: 'Explore Objects',
        description: 'Discover available data schemas',
        prompt: 'What objects are available in the system?',
        gradient: 'from-purple-500 to-pink-500'
    },
    {
        icon: Zap,
        title: 'Quick Analysis',
        description: 'Get insights from your data',
        prompt: 'Summarize key statistics from my data',
        gradient: 'from-amber-500 to-orange-500'
    },
    {
        icon: HelpCircle,
        title: 'Get Help',
        description: 'Learn what I can do',
        prompt: '/help',
        gradient: 'from-emerald-500 to-teal-500'
    }
];

const SUGGESTED_PROMPTS = [
    'How many accounts do I have?',
    'Show me all contacts created this week',
    'What are the open deals?',
    'List all Jira projects'
];

interface EmptyStateProps {
    onPromptSelect: (prompt: string) => void;
}

export function EmptyState({ onPromptSelect }: EmptyStateProps) {
    return (
        <div className="flex flex-col items-center justify-center h-full px-6 py-8 animate-fade-in">
            {/* Hero Section */}
            <div className="text-center mb-8">
                <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500 shadow-lg shadow-purple-500/30 mb-4">
                    <Sparkles size={32} className="text-white" />
                </div>
                <h2 className="text-xl font-semibold text-slate-800 mb-2">
                    How can I help you today?
                </h2>
                <p className="text-sm text-slate-500 max-w-xs mx-auto">
                    Ask questions about your data, query records, or get insights from your CRM.
                </p>
            </div>

            {/* Quick Action Cards */}
            <div className="w-full max-w-md grid grid-cols-2 gap-3 mb-8">
                {QUICK_ACTIONS.map((action) => (
                    <button
                        key={action.title}
                        onClick={() => onPromptSelect(action.prompt)}
                        className="group flex flex-col items-start p-4 bg-white rounded-xl border border-slate-200 shadow-sm hover:shadow-md hover:border-slate-300 transition-all duration-200 text-left"
                    >
                        <div className={`p-2 rounded-lg bg-gradient-to-br ${action.gradient} mb-3 group-hover:scale-110 transition-transform duration-200`}>
                            <action.icon size={18} className="text-white" />
                        </div>
                        <h3 className="text-sm font-medium text-slate-800 mb-0.5">
                            {action.title}
                        </h3>
                        <p className="text-xs text-slate-500 leading-relaxed">
                            {action.description}
                        </p>
                    </button>
                ))}
            </div>

            {/* Suggested Prompts */}
            <div className="w-full max-w-md">
                <div className="flex items-center gap-2 mb-3">
                    <MessageSquare size={14} className="text-slate-400" />
                    <span className="text-xs font-medium text-slate-500 uppercase tracking-wider">
                        Try asking
                    </span>
                </div>
                <div className="flex flex-wrap gap-2">
                    {SUGGESTED_PROMPTS.map((prompt) => (
                        <button
                            key={prompt}
                            onClick={() => onPromptSelect(prompt)}
                            className="px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-600 text-xs rounded-full transition-colors duration-150"
                        >
                            {prompt}
                        </button>
                    ))}
                </div>
            </div>
        </div>
    );
}
