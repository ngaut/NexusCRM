import React, { useMemo, useState } from 'react';
import { X, Copy, Check, FileText, Database, MessageSquare } from 'lucide-react';
import { ChatMessage } from '../../infrastructure/api/agent';
import { Z_LAYERS } from '../../core/constants/zIndex';

interface ContextFile {
    path: string;
    tokenSize: number;
    content?: string; // Optional if we can fetch it later
}

interface ContextInspectorProps {
    isOpen: boolean;
    onClose: () => void;
    systemPrompt: string; // The base prompt
    files: ContextFile[];
    messages: ChatMessage[];
}

export function ContextInspector({
    isOpen,
    onClose,
    systemPrompt,
    files,
    messages
}: ContextInspectorProps) {
    const [copied, setCopied] = useState(false);

    // Reconstruct the full prompt simulation
    const fullPrompt = useMemo(() => {
        const parts = [];

        // 1. System Prompt
        parts.push(`[SYSTEM PROMPT]
${systemPrompt}`);

        // 2. Injected Files
        if (files.length > 0) {
            parts.push(`\n[INJECTED CONTEXT FILES]
(These files are injected directly into the system context)`);
            files.forEach(file => {
                parts.push(`
--- FILE: ${file.path} ---
${file.content ? file.content : `[... Content of ${file.path} (~${file.tokenSize} tokens) ...]`}
--- END FILE ---`);
            });
        }

        // 3. Chat History
        if (messages.length > 0) {
            parts.push(`\n[CONVERSATION HISTORY]`);
            messages.forEach((msg, idx) => {
                // Formatting based on compactor logic
                let content = msg.content || '';
                if (msg.role === 'assistant' && msg.tool_calls?.length) {
                    msg.tool_calls.forEach(tc => {
                        content += `\n(Tool Call: ${tc.function.name})`;
                    });
                }
                if (msg.role === 'tool') {
                    // Show full tool result
                    content = `(Tool Result: ${msg.name})\n${content}`;
                }
                parts.push(`[${msg.role}]: ${content}`);
            });
        }

        return parts.join('\n');
    }, [systemPrompt, files, messages]);

    const handleCopy = () => {
        navigator.clipboard.writeText(fullPrompt);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-fade-in" style={{ zIndex: Z_LAYERS.MAX }}>
            <div className="bg-white rounded-xl shadow-2xl w-full max-w-4xl h-[85vh] flex flex-col border border-slate-200" onClick={(e) => e.stopPropagation()}>
                {/* Header */}
                <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200">
                    <div>
                        <h2 className="text-lg font-semibold text-slate-800 flex items-center gap-2">
                            <Database size={18} className="text-indigo-600" />
                            Context Inspector
                        </h2>
                        <p className="text-xs text-slate-500 mt-1">
                            Visualization of the actual prompt sent to the LLM (System + Files + Chat).
                        </p>
                    </div>
                    <div className="flex items-center gap-2">
                        <button
                            onClick={handleCopy}
                            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-600 hover:text-indigo-600 hover:bg-slate-100 rounded-lg transition-colors"
                        >
                            {copied ? <Check size={14} className="text-emerald-500" /> : <Copy size={14} />}
                            {copied ? 'Copied' : 'Copy Full Text'}
                        </button>
                        <button
                            onClick={onClose}
                            className="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-lg transition-colors"
                        >
                            <X size={20} />
                        </button>
                    </div>
                </div>

                {/* Content */}
                <div className="flex-1 overflow-auto p-0 bg-slate-950">
                    <div className="p-6 font-mono text-xs leading-relaxed text-slate-300">
                        {/* Render with syntax highlighting-ish logic */}
                        {fullPrompt.split('\n').map((line, i) => {
                            if (line.startsWith('[SYSTEM') || line.startsWith('[INJECTED') || line.startsWith('[CONVERSATION')) {
                                return <div key={i} className="text-emerald-400 font-bold mt-4 mb-2 border-b border-white/10 pb-1">{line}</div>;
                            }
                            if (line.startsWith('--- FILE:')) {
                                return <div key={i} className="text-cyan-400 font-semibold mt-2">{line}</div>;
                            }
                            if (line.startsWith('--- END FILE')) {
                                return <div key={i} className="text-cyan-400 font-semibold mb-2">{line}</div>;
                            }
                            if (line.startsWith('[user]:')) {
                                return <div key={i} className="text-indigo-300 mt-2 font-bold">{line}</div>;
                            }
                            if (line.startsWith('[assistant]:')) {
                                return <div key={i} className="text-purple-300 mt-1 font-bold">{line}</div>;
                            }
                            if (line.startsWith('[tool]:') || line.includes('(Tool Result')) {
                                return <div key={i} className="text-amber-300 mt-1">{line}</div>;
                            }
                            if (line.includes('[... Content')) {
                                return <div key={i} className="text-slate-500 italic pl-4">{line}</div>;
                            }
                            return <div key={i} className="pl-0">{line}</div>;
                        })}
                    </div>
                </div>

                {/* Footer */}
                <div className="px-6 py-3 border-t border-slate-200 bg-slate-50 flex justify-between items-center text-xs text-slate-500">
                    <div className="flex gap-4">
                        <span className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-emerald-500" /> System</span>
                        <span className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-cyan-500" /> Files</span>
                        <span className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-indigo-500" /> Chat</span>
                    </div>
                    <div>
                        Total Context: ~{files.reduce((acc, f) => acc + f.tokenSize, 0) + messages.reduce((acc, m) => acc + (m.content ? m.content.length / 4 : 0), 0) + 200} tokens
                    </div>
                </div>
            </div>
        </div>
    );
}
