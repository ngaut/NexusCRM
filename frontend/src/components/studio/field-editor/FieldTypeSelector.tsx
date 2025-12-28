import React from 'react';
import { Type, Hash, Calendar, List, Link2, ToggleLeft, Mail, AlignLeft, GitFork } from 'lucide-react';
import * as Icons from 'lucide-react';
import type { FieldType } from '../../../types';

export interface FieldTypeOption {
    type: FieldType;
    label: string;
    icon: React.ComponentType<{ size?: number | string; className?: string }>;
    description: string;
    color: string;
}

export const FIELD_TYPES: FieldTypeOption[] = [
    { type: 'Text', label: 'Text', icon: Type, description: 'Single line text', color: 'slate' },
    { type: 'TextArea', label: 'Text Area', icon: AlignLeft, description: 'Multi-line text', color: 'slate' },
    { type: 'Number', label: 'Number', icon: Hash, description: 'Numeric values', color: 'blue' },
    { type: 'Currency', label: 'Currency', icon: Icons.DollarSign, description: 'Money values', color: 'emerald' },
    { type: 'Percent', label: 'Percent', icon: Icons.Percent, description: 'Percentage values', color: 'blue' },
    { type: 'Date', label: 'Date', icon: Calendar, description: 'Date only', color: 'amber' },
    { type: 'DateTime', label: 'Date/Time', icon: Icons.Clock, description: 'Date and time', color: 'amber' },
    { type: 'Boolean', label: 'Checkbox', icon: ToggleLeft, description: 'True/False toggle', color: 'purple' },
    { type: 'Picklist', label: 'Picklist', icon: List, description: 'Dropdown selection', color: 'indigo' },
    { type: 'Lookup', label: 'Lookup', icon: Link2, description: 'Link to another object', color: 'cyan' },
    { type: 'MasterDetail', label: 'Master-Detail', icon: GitFork, description: 'Parent-Child relationship (Cascade Delete)', color: 'red' },
    { type: 'Email', label: 'Email', icon: Mail, description: 'Email address', color: 'rose' },
    { type: 'Phone', label: 'Phone', icon: Icons.Phone, description: 'Phone number', color: 'teal' },
    { type: 'Url', label: 'URL', icon: Icons.Link, description: 'Web address', color: 'sky' },
];

interface FieldTypeSelectorProps {
    onSelect: (type: FieldType | 'MasterDetail') => void;
}

export const FieldTypeSelector: React.FC<FieldTypeSelectorProps> = ({ onSelect }) => {
    return (
        <div>
            <p className="text-sm text-slate-600 mb-4">Select the type of field to create:</p>
            <div className="grid grid-cols-3 gap-2">
                {FIELD_TYPES.map(ft => (
                    <button
                        key={ft.type}
                        onClick={() => onSelect(ft.type === 'MasterDetail' ? 'MasterDetail' : ft.type)}
                        className={`flex flex-col items-center gap-2 p-4 border rounded-xl hover:border-blue-300 hover:bg-blue-50/50 transition-all text-center`}
                    >
                        <div className={`w-10 h-10 rounded-lg bg-${ft.color}-100 flex items-center justify-center`}>
                            <ft.icon size={20} className={`text-${ft.color}-600`} />
                        </div>
                        <div className="text-sm font-medium text-slate-800">{ft.label}</div>
                    </button>
                ))}
            </div>
        </div>
    );
};
