import React, { useState } from 'react';
import * as Icons from 'lucide-react';
import { StickyNote, Clock, History, X, ChevronUp, ChevronDown } from 'lucide-react';
import type { UtilityItem } from '../types';

interface UtilityBarProps {
    items: UtilityItem[];
}

interface PanelState {
    isOpen: boolean;
    isMinimized: boolean;
}

export const UtilityBar: React.FC<UtilityBarProps> = ({ items }) => {
    const [panelStates, setPanelStates] = useState<Record<string, PanelState>>({});

    if (!items || items.length === 0) return null;

    const togglePanel = (id: string) => {
        setPanelStates(prev => ({
            ...prev,
            [id]: {
                isOpen: !prev[id]?.isOpen,
                isMinimized: false,
            }
        }));
    };

    const minimizePanel = (id: string) => {
        setPanelStates(prev => ({
            ...prev,
            [id]: {
                ...prev[id],
                isMinimized: !prev[id]?.isMinimized,
            }
        }));
    };

    const closePanel = (id: string) => {
        setPanelStates(prev => ({
            ...prev,
            [id]: { isOpen: false, isMinimized: false }
        }));
    };

    const getIconComponent = (iconName: string) => {
        const IconComponent = Icons[iconName as keyof typeof Icons] as React.ComponentType<any>;
        return IconComponent || Icons.Box;
    };

    const getUtilityIcon = (type: string) => {
        switch (type) {
            case 'notes': return StickyNote;
            case 'recent': return Clock;
            case 'history': return History;
            default: return Icons.Box;
        }
    };

    const renderPanelContent = (item: UtilityItem) => {
        switch (item.type) {
            case 'notes':
                return (
                    <div className="p-4 space-y-3">
                        <textarea
                            placeholder="Take notes here..."
                            className="w-full h-48 p-3 border rounded-lg resize-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
                        />
                        <p className="text-xs text-slate-500">Notes are saved automatically</p>
                    </div>
                );
            case 'recent':
                return (
                    <div className="p-4">
                        <p className="text-sm text-slate-500 text-center py-8">
                            Your recently viewed items will appear here
                        </p>
                    </div>
                );
            case 'history':
                return (
                    <div className="p-4">
                        <p className="text-sm text-slate-500 text-center py-8">
                            Record history will appear here
                        </p>
                    </div>
                );
            default:
                return (
                    <div className="p-4">
                        <p className="text-sm text-slate-500 text-center py-8">
                            Custom utility content
                        </p>
                    </div>
                );
        }
    };

    return (
        <>
            {/* Utility Panels */}
            {items.map(item => {
                const state = panelStates[item.id];
                if (!state?.isOpen) return null;

                const IconComponent = item.icon ? getIconComponent(item.icon) : getUtilityIcon(item.type);

                return (
                    <div
                        key={item.id}
                        className="fixed bottom-12 right-4 z-40 bg-white rounded-t-xl shadow-2xl border border-slate-200 overflow-hidden"
                        style={{
                            width: item.panel_width || 340,
                            maxHeight: state.isMinimized ? 48 : (item.panel_height || 400),
                        }}
                    >
                        {/* Panel Header */}
                        <div className="flex items-center justify-between px-4 py-3 bg-slate-50 border-b">
                            <div className="flex items-center gap-2">
                                <IconComponent size={16} className="text-slate-600" />
                                <span className="font-medium text-sm text-slate-800">{item.label}</span>
                            </div>
                            <div className="flex items-center gap-1">
                                <button
                                    onClick={() => minimizePanel(item.id)}
                                    className="p-1 hover:bg-slate-200 rounded"
                                >
                                    {state.isMinimized ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
                                </button>
                                <button
                                    onClick={() => closePanel(item.id)}
                                    className="p-1 hover:bg-slate-200 rounded"
                                >
                                    <X size={16} />
                                </button>
                            </div>
                        </div>

                        {/* Panel Content */}
                        {!state.isMinimized && (
                            <div className="overflow-y-auto" style={{ maxHeight: (item.panel_height || 400) - 48 }}>
                                {renderPanelContent(item)}
                            </div>
                        )}
                    </div>
                );
            })}

            {/* Utility Bar */}
            <div className="fixed bottom-0 left-0 right-0 h-12 bg-slate-900 border-t border-slate-700 flex items-center px-4 z-50">
                <div className="flex items-center gap-1">
                    {items.map(item => {
                        const state = panelStates[item.id];
                        const IconComponent = item.icon ? getIconComponent(item.icon) : getUtilityIcon(item.type);

                        return (
                            <button
                                key={item.id}
                                onClick={() => togglePanel(item.id)}
                                className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${state?.isOpen
                                    ? 'bg-blue-600 text-white'
                                    : 'text-slate-300 hover:bg-slate-800 hover:text-white'
                                    }`}
                            >
                                <IconComponent size={16} />
                                <span className="hidden md:inline">{item.label}</span>
                            </button>
                        );
                    })}
                </div>
            </div>
        </>
    );
};
