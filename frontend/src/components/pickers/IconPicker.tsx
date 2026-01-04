import React, { useState } from 'react';
import { createPortal } from 'react-dom';
import * as Icons from 'lucide-react';
import { Search, X } from 'lucide-react';

interface IconPickerProps {
    value: string;
    onChange: (iconName: string) => void;
    onClose?: () => void;
}

const COMMON_ICONS = [
    'Home', 'Users', 'Settings', 'Database', 'Layout', 'FileText',
    'Briefcase', 'ShoppingCart', 'TrendingUp', 'Calendar', 'Mail', 'Phone',
    'MapPin', 'Building', 'Package', 'Truck', 'DollarSign', 'CreditCard',
    'BarChart', 'PieChart', 'Activity', 'Target', 'Award', 'Star'
];

export const IconPicker: React.FC<IconPickerProps> = ({ value, onChange, onClose }) => {
    const [search, setSearch] = useState('');

    // Get all Lucide icon names
    const allIcons = Object.keys(Icons).filter(
        key => key !== 'createLucideIcon' && typeof Icons[key as keyof typeof Icons] === 'function'
    );

    const filteredIcons = search
        ? allIcons.filter(name => name.toLowerCase().includes(search.toLowerCase()))
        : COMMON_ICONS;

    const renderIcon = (iconName: string) => {
        const IconComponent = Icons[iconName as keyof typeof Icons] as React.ComponentType<{ size?: number }>;
        if (!IconComponent) return null;
        return <IconComponent size={20} />;
    };

    return createPortal(
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[100]">
            <div className="bg-white rounded-lg shadow-xl w-full max-w-2xl mx-4">
                <div className="flex items-center justify-between p-4 border-b">
                    <h3 className="text-lg font-semibold">Select Icon</h3>
                    {onClose && (
                        <button onClick={onClose} className="text-slate-400 hover:text-slate-600">
                            <X size={20} />
                        </button>
                    )}
                </div>

                <div className="p-4 border-b">
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={18} />
                        <input
                            type="text"
                            placeholder="Search icons..."
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            className="w-full pl-10 pr-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                    </div>
                </div>

                <div className="p-4 max-h-96 overflow-y-auto">
                    <div className="grid grid-cols-6 gap-2">
                        {filteredIcons.slice(0, 60).map(iconName => (
                            <button
                                key={iconName}
                                onClick={() => {
                                    onChange(iconName);
                                    onClose?.();
                                }}
                                className={`
                  p-3 rounded-lg border-2 hover:border-blue-500 hover:bg-blue-50 transition-colors
                  flex items-center justify-center
                  ${value === iconName ? 'border-blue-500 bg-blue-50' : 'border-slate-200'}
                `}
                                title={iconName}
                            >
                                {renderIcon(iconName)}
                            </button>
                        ))}
                    </div>
                    {filteredIcons.length === 0 && (
                        <div className="text-center py-8 text-slate-500">
                            No icons found matching "{search}"
                        </div>
                    )}
                </div>

                <div className="p-4 border-t bg-slate-50 flex justify-between items-center">
                    <div className="text-sm text-slate-600">
                        Selected: <span className="font-mono font-semibold">{value || 'None'}</span>
                    </div>
                    <button
                        onClick={onClose}
                        className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                    >
                        Done
                    </button>
                </div>
            </div>
        </div>,
        document.body
    );
};
