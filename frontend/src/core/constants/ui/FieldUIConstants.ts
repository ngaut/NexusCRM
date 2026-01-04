import { FieldType } from '../SchemaDefinitions';
import { Type, Hash, Calendar, List, Link2, ToggleLeft, Mail, AlignLeft, GitFork } from 'lucide-react';
import * as Icons from 'lucide-react';

export interface FieldTypeOption {
    type: FieldType | 'MasterDetail';
    label: string;
    icon: React.ComponentType<{ size?: number | string; className?: string }>;
    description: string;
    color: string;
}

export const FIELD_TYPE_OPTIONS: FieldTypeOption[] = [
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
    { type: 'AutoNumber', label: 'Auto Number', icon: Hash, description: 'System-generated sequence', color: 'slate' },
];
