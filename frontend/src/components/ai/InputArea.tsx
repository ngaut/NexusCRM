import React, { useState, useRef, useEffect } from 'react';
import { Send, Plus, FileText, Trash2, HelpCircle, Loader2, Minimize2 } from 'lucide-react';

interface InputAreaProps {
    input: string;
    isLoading: boolean;
    setInput: (value: string) => void;
    handleSubmit: (e: React.FormEvent) => void;
    handleCancel: () => void;
}

interface Command {
    id: string;
    label: string;
    icon: React.ElementType;
    desc: string;
}

const COMMANDS: Command[] = [
    { id: 'add', label: 'Add files', icon: FileText, desc: 'Add files to context' },
    { id: 'remove', label: 'Remove files', icon: Trash2, desc: 'Remove files from context' },
    { id: 'list', label: 'List context', icon: FileText, desc: 'View active files' },
    { id: 'compact', label: 'Compact context', icon: Minimize2, desc: 'Summarize conversation' },
    { id: 'clear', label: 'Clear chat', icon: Trash2, desc: 'Clear history' },
    { id: 'help', label: 'Help', icon: HelpCircle, desc: 'View commands' },
];

export function InputArea({ input, isLoading, setInput, handleSubmit, handleCancel }: InputAreaProps) {
    const [showSlashMenu, setShowSlashMenu] = useState(false);
    const [showActionMenu, setShowActionMenu] = useState(false);
    const [slashFilter, setSlashFilter] = useState('');
    const [selectedIndex, setSelectedIndex] = useState(0);
    const textareaRef = useRef<HTMLTextAreaElement>(null);
    const slashMenuRef = useRef<HTMLDivElement>(null);
    const actionMenuRef = useRef<HTMLDivElement>(null);

    // Auto-resize textarea
    useEffect(() => {
        if (textareaRef.current) {
            textareaRef.current.style.height = 'auto';
            textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 120)}px`;
        }
    }, [input]);

    // Handle Slash Command Detection
    useEffect(() => {
        const lastWord = input.split(' ').pop();
        if (lastWord && lastWord.startsWith('/')) {
            setShowSlashMenu(true);
            setSlashFilter(lastWord.slice(1).toLowerCase());
            setSelectedIndex(0);
        } else {
            setShowSlashMenu(false);
        }
    }, [input]);

    // Close menus on click outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (slashMenuRef.current && !slashMenuRef.current.contains(event.target as Node)) {
                setShowSlashMenu(false);
            }
            if (actionMenuRef.current && !actionMenuRef.current.contains(event.target as Node)) {
                setShowActionMenu(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const filteredCommands = COMMANDS.filter(cmd =>
        cmd.id.includes(slashFilter) || cmd.label.toLowerCase().includes(slashFilter)
    );

    const executeCommand = (cmdId: string) => {
        // Replace the slash command in input or set new input
        const words = input.split(' ');
        words.pop(); // Remove the partial command
        const prefix = words.join(' ');

        const newText = prefix ? `${prefix} /${cmdId} ` : `/${cmdId} `;
        setInput(newText);
        setShowSlashMenu(false);
        setShowActionMenu(false);
        textareaRef.current?.focus();
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.nativeEvent.isComposing) return;

        if (showSlashMenu) {
            if (e.key === 'ArrowDown') {
                e.preventDefault();
                setSelectedIndex(prev => (prev + 1) % filteredCommands.length);
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                setSelectedIndex(prev => (prev - 1 + filteredCommands.length) % filteredCommands.length);
            } else if (e.key === 'Enter' || e.key === 'Tab') {
                e.preventDefault();
                if (filteredCommands[selectedIndex]) {
                    executeCommand(filteredCommands[selectedIndex].id);
                }
            } else if (e.key === 'Escape') {
                setShowSlashMenu(false);
            }
        } else if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSubmit(e);
        }
    };

    return (
        <div className="absolute bottom-6 left-4 right-4 z-50">
            {/* Slash Command Popover */}
            {showSlashMenu && filteredCommands.length > 0 && (
                <div
                    ref={slashMenuRef}
                    className="absolute bottom-full left-0 mb-2 w-64 bg-white rounded-xl shadow-xl border border-slate-200 overflow-hidden animate-in fade-in slide-in-from-bottom-2"
                >
                    <div className="p-1">
                        {filteredCommands.map((cmd, idx) => (
                            <button
                                key={cmd.id}
                                onClick={() => executeCommand(cmd.id)}
                                className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-left transition-colors ${idx === selectedIndex ? 'bg-indigo-50 text-indigo-700' : 'hover:bg-slate-50 text-slate-700'
                                    }`}
                            >
                                <div className={`p-1.5 rounded-md ${idx === selectedIndex ? 'bg-indigo-100' : 'bg-slate-100'
                                    }`}>
                                    <cmd.icon size={14} />
                                </div>
                                <div className="flex-1">
                                    <div className="font-medium">/{cmd.id}</div>
                                    <div className="text-xs text-slate-400">{cmd.desc}</div>
                                </div>
                            </button>
                        ))}
                    </div>
                </div>
            )}

            {/* Action Menu Popover */}
            {showActionMenu && (
                <div
                    ref={actionMenuRef}
                    className="absolute bottom-full left-0 mb-2 w-56 bg-white rounded-xl shadow-xl border border-slate-200 overflow-hidden animate-in fade-in slide-in-from-bottom-2"
                >
                    <div className="p-1">
                        <div className="px-3 py-2 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                            Actions
                        </div>
                        {COMMANDS.map((cmd) => (
                            <button
                                key={cmd.id}
                                onClick={() => executeCommand(cmd.id)}
                                className="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-left hover:bg-slate-50 text-slate-700 transition-colors"
                            >
                                <cmd.icon size={16} className="text-slate-400" />
                                <span>{cmd.label}</span>
                            </button>
                        ))}
                    </div>
                </div>
            )}

            {/* Main Input Container */}
            <div className="relative flex items-end gap-2 p-2 bg-white/95 backdrop-blur-xl border border-slate-200/80 shadow-xl shadow-slate-900/5 rounded-2xl ring-1 ring-slate-900/5 transition-all focus-within:ring-2 focus-within:ring-indigo-500/30 focus-within:border-indigo-300 focus-within:shadow-indigo-500/10">
                {/* Action Button */}
                <button
                    type="button"
                    onClick={() => setShowActionMenu(!showActionMenu)}
                    className={`p-2.5 rounded-xl transition-all ${showActionMenu ? 'bg-indigo-100 text-indigo-600 rotate-45' : 'bg-slate-100 text-slate-500 hover:bg-slate-200 hover:text-slate-700'
                        }`}
                    title="Quick actions & commands"
                >
                    <Plus size={20} />
                </button>

                {/* Textarea */}
                <textarea
                    ref={textareaRef}
                    value={input}
                    onChange={e => setInput(e.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder="Ask a question or type / for commands..."
                    rows={1}
                    disabled={isLoading}
                    className="flex-1 max-h-[120px] py-3 px-2 bg-transparent border-0 focus:ring-0 focus:outline-none resize-none text-slate-800 placeholder:text-slate-400 text-[15px] leading-relaxed"
                />

                {/* Send/Loader Button */}
                {isLoading ? (
                    <div className="p-2.5 bg-slate-100 rounded-full animate-pulse">
                        <Loader2 size={20} className="text-slate-400 animate-spin" />
                    </div>
                ) : (
                    <button
                        onClick={handleSubmit}
                        disabled={!input.trim()}
                        className="p-2.5 bg-gradient-to-tr from-indigo-600 to-purple-600 text-white rounded-full shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 disabled:opacity-50 disabled:shadow-none transition-all hover:scale-105 active:scale-95"
                    >
                        <Send size={18} className="translate-x-0.5 translate-y-0.5" />
                    </button>
                )}
            </div>

            {/* Keyboard Hint / Cancel Button */}
            <div className="absolute -bottom-6 left-0 w-full flex items-center justify-center">
                {isLoading ? (
                    <button
                        onClick={handleCancel}
                        className="text-xs font-medium text-slate-400 hover:text-red-500 transition-colors flex items-center gap-1.5 px-2 py-0.5 rounded-full hover:bg-red-50"
                    >
                        <span className="w-1.5 h-1.5 rounded-full bg-red-400 animate-pulse" />
                        Stop generating
                    </button>
                ) : (
                    <span className="text-[10px] text-slate-400">
                        <kbd className="px-1.5 py-0.5 bg-slate-100 rounded text-slate-500 font-mono">↵</kbd> to send
                    </span>
                )}
            </div>
        </div>
    );
}
