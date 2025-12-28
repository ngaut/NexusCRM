import React from 'react';
import { Check } from 'lucide-react';

interface ColorPickerProps {
    value: string;
    onChange: (color: string) => void;
    label?: string;
}

const COLORS = [
    { name: 'Blue', value: 'blue', bg: 'bg-blue-500', ring: 'ring-blue-500' },
    { name: 'Indigo', value: 'indigo', bg: 'bg-indigo-500', ring: 'ring-indigo-500' },
    { name: 'Purple', value: 'purple', bg: 'bg-purple-500', ring: 'ring-purple-500' },
    { name: 'Pink', value: 'pink', bg: 'bg-pink-500', ring: 'ring-pink-500' },
    { name: 'Red', value: 'red', bg: 'bg-red-500', ring: 'ring-red-500' },
    { name: 'Orange', value: 'orange', bg: 'bg-orange-500', ring: 'ring-orange-500' },
    { name: 'Yellow', value: 'yellow', bg: 'bg-yellow-500', ring: 'ring-yellow-500' },
    { name: 'Green', value: 'green', bg: 'bg-green-500', ring: 'ring-green-500' },
    { name: 'Teal', value: 'teal', bg: 'bg-teal-500', ring: 'ring-teal-500' },
    { name: 'Cyan', value: 'cyan', bg: 'bg-cyan-500', ring: 'ring-cyan-500' },
    { name: 'Slate', value: 'slate', bg: 'bg-slate-500', ring: 'ring-slate-500' },
    { name: 'Gray', value: 'gray', bg: 'bg-gray-500', ring: 'ring-gray-500' },
];

export const ColorPicker: React.FC<ColorPickerProps> = ({ value, onChange, label }) => {
    return (
        <div>
            {label && (
                <label className="block text-sm font-medium text-slate-700 mb-2">
                    {label}
                </label>
            )}
            <div className="grid grid-cols-6 gap-2">
                {COLORS.map(color => (
                    <button
                        key={color.value}
                        type="button"
                        onClick={() => onChange(color.value)}
                        className={`
              ${color.bg} rounded-lg h-10 relative
              hover:scale-105 transition-transform
              ${value === color.value ? 'ring-4 ' + color.ring + ' ring-offset-2' : ''}
            `}
                        title={color.name}
                    >
                        {value === color.value && (
                            <div className="absolute inset-0 flex items-center justify-center">
                                <Check className="text-white" size={20} strokeWidth={3} />
                            </div>
                        )}
                    </button>
                ))}
            </div>
        </div>
    );
};
