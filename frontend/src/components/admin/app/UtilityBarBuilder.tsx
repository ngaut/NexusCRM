import React from 'react';
import * as Icons from 'lucide-react';
import type { UtilityItem } from '../../../types';

interface UtilityBarBuilderProps {
    utilityItems: UtilityItem[];
    onChange: (items: UtilityItem[]) => void;
}

export const UtilityBarBuilder: React.FC<UtilityBarBuilderProps> = ({ utilityItems, onChange }) => {
    return (
        <div>
            <div className="flex items-center gap-2 mb-3">
                <Icons.PanelBottom size={16} className="text-slate-600" />
                <label className="text-sm font-medium text-slate-700">Utility Bar</label>
            </div>
            <p className="text-xs text-slate-500 mb-3">Add productivity tools to the footer bar</p>

            <div className="space-y-2">
                {[
                    { id: 'notes', type: 'notes' as const, label: 'Notes', icon: 'StickyNote', description: 'Quick notes panel' },
                    { id: 'recent', type: 'recent' as const, label: 'Recent Items', icon: 'Clock', description: 'Recently viewed records' },
                    { id: 'history', type: 'history' as const, label: 'History', icon: 'History', description: 'Record change history' },
                ].map(util => {
                    const isEnabled = utilityItems.some(item => item.type === util.type);
                    const UtilIcon = Icons[util.icon as keyof typeof Icons] as React.ComponentType<{ size?: number; className?: string }>;
                    return (
                        <label
                            key={util.id}
                            className="flex items-center gap-3 p-3 border rounded-lg cursor-pointer hover:bg-slate-50"
                        >
                            <input
                                type="checkbox"
                                checked={isEnabled}
                                onChange={(e) => {
                                    if (e.target.checked) {
                                        onChange([...utilityItems, {
                                            id: `util-${util.type}-${Date.now()}`,
                                            type: util.type,
                                            label: util.label,
                                            icon: util.icon,
                                        }]);
                                    } else {
                                        onChange(utilityItems.filter(item => item.type !== util.type));
                                    }
                                }}
                                className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                            />
                            <div className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center">
                                <UtilIcon size={16} className="text-slate-600" />
                            </div>
                            <div className="flex-1">
                                <div className="font-medium text-slate-800 text-sm">{util.label}</div>
                                <div className="text-xs text-slate-500">{util.description}</div>
                            </div>
                        </label>
                    );
                })}
            </div>
        </div>
    );
};
