import React, { useState } from 'react';
import { Layers } from 'lucide-react';
import * as Icons from 'lucide-react';
import { IconPicker } from '../../pickers/IconPicker';
import { ColorPicker } from '../../pickers/ColorPicker';
import { getColorClasses } from '../../../core/utils/colorClasses';
import type { AppConfig } from '../../../types';

interface AppIconPickerProps {
    icon: string;
    color: string;
    onChange: (updates: Partial<AppConfig>) => void;
}

export const AppIconPicker: React.FC<AppIconPickerProps> = ({ icon, color, onChange }) => {
    const [showIconPicker, setShowIconPicker] = useState(false);

    // Resolve Icon Component
    const IconComponent = Icons[icon as keyof typeof Icons] as React.ComponentType<{ size?: number; className?: string }> || Layers;

    return (
        <div className="grid grid-cols-2 gap-4">
            {/* Icon */}
            <div>
                <label className="block text-sm font-medium text-slate-700 mb-2">
                    Icon
                </label>
                <button
                    type="button"
                    onClick={() => setShowIconPicker(true)}
                    className="flex items-center gap-3 px-4 py-3 border rounded-lg hover:bg-slate-50 w-full"
                >
                    <div className={`w-10 h-10 ${getColorClasses(color).bg} rounded-lg flex items-center justify-center ${getColorClasses(color).text}`}>
                        <IconComponent size={20} />
                    </div>
                    <span className="font-mono text-sm text-slate-600">{icon}</span>
                </button>
            </div>

            {/* Color */}
            <ColorPicker
                value={color}
                onChange={(newColor) => onChange({ color: newColor })}
                label="Theme Color"
            />

            {/* Modal */}
            {showIconPicker && (
                <IconPicker
                    value={icon}
                    onChange={(newIcon) => onChange({ icon: newIcon })}
                    onClose={() => setShowIconPicker(false)}
                />
            )}
        </div>
    );
};
