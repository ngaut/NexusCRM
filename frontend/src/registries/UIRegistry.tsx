


import React from 'react';
import { useNavigate } from 'react-router-dom';
import { FieldMetadata, SObject } from '../types';
import { ExternalLink, Check, X, DollarSign, Calendar, Hash, Percent, Clock } from 'lucide-react';
import * as Icons from 'lucide-react';
import { formatCurrency, formatDate, formatDateTime } from '../core/utils/formatting';
import { SearchableLookup } from '../components/SearchableLookup';
import { UI_DEFAULTS } from '../core/constants';
import { dataAPI } from '../infrastructure/api/data';
import { COMMON_FIELDS } from '../core/constants/CommonFields';

// --- Prop Type Definitions ---

export interface FieldRendererProps {
    field: FieldMetadata;
    value: unknown;
    onNavigate?: (obj: string, id: string) => void;
    record?: SObject;
    variant?: 'table' | 'detail';
}

export interface FieldInputProps {
    field: FieldMetadata;
    value: unknown;
    onChange: (value: unknown) => void;
    disabled?: boolean;
    placeholder?: string;
    required?: boolean;
    options?: string[];
    onKeyDown?: (e: React.KeyboardEvent) => void;
    autoFocus?: boolean;
}


// --- Default Implementations ---

const BooleanRenderer: React.FC<FieldRendererProps> = ({ value }) => (
    value ? (
        <div className="flex items-center gap-1 text-emerald-600 font-medium">
            <Check size={16} /> <span className="text-xs">Yes</span>
        </div>
    ) : (
        <div className="flex items-center gap-1 text-slate-400">
            <X size={16} /> <span className="text-xs">No</span>
        </div>
    )
);

const PicklistRenderer: React.FC<FieldRendererProps> = ({ value, field }) => {
    // Fail Visible: Warn if we have a picklist but no options
    const hasOptions = (field.options && field.options.length > 0) || (field.type === 'Picklist' && field.options && field.options.length > 0);
    // Note: Some legacy fields might rely on checks. If strictly picklist, it must have options.
    // For now, if value matches an option, good. If no options defined at all, warning.

    if (field.type === 'Picklist' && (!field.options || field.options.length === 0)) {
        return (
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-amber-100 text-amber-800 border border-amber-200" title="Metadata Error: Picklist has no options defined">
                ⚠️ Broken Picklist
            </span>
        );
    }

    if (value === null || value === undefined || value === '') return null;
    const valStr = String(value);
    let hash = 0;
    for (let i = 0; i < valStr.length; i++) hash = valStr.charCodeAt(i) + ((hash << 5) - hash);
    const hue = Math.abs(hash % 360);
    return (
        <span
            className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border transition-all hover:scale-105"
            style={{ backgroundColor: `hsl(${hue}, 70%, 96%)`, color: `hsl(${hue}, 70%, 35%)`, borderColor: `hsl(${hue}, 70%, 85%)` }}
        >
            {valStr}
        </span>
    );
};

const UrlRenderer: React.FC<FieldRendererProps> = ({ value, variant, field }) => {
    if (value === null || value === undefined || value === '') return null;
    const strVal = String(value);
    if (strVal.startsWith('data:image')) {
        return (
            <div className="relative group w-fit">
                <img
                    src={strVal}
                    alt={field.label}
                    className={`object-cover border border-slate-200 rounded-lg shadow-sm ${variant === 'table' ? 'w-10 h-10' : 'w-32 h-32'}`}
                />
            </div>
        );
    }
    return (
        <a href={strVal} target="_blank" rel="noreferrer" className="text-blue-600 hover:underline inline-flex items-center gap-1 truncate max-w-[200px]" onClick={(e) => e.stopPropagation()}>
            {strVal} <ExternalLink size={10} />
        </a>
    );
};

const LookupRenderer: React.FC<FieldRendererProps> = ({ field, value, onNavigate, record }) => {
    // Always try to use the resolved name if available
    const displayLabel = (record && record[`${field.api_name}_Name`])
        ? String(record[`${field.api_name}_Name`])
        : String(value);

    // If we have navigation capability and reference_to, make it clickable
    if (onNavigate && field.reference_to) {
        // Determine target object
        let targetObject: string | undefined;
        if (Array.isArray(field.reference_to)) {
            // Polymorphic: Get type from record
            if (record && record[`${field.api_name}_type`]) {
                targetObject = String(record[`${field.api_name}_type`]);
            }
        } else {
            targetObject = field.reference_to;
        }

        // Type Hint for polymorphic
        const typeHint = (Array.isArray(field.reference_to) && targetObject)
            ? <span className="text-xs text-gray-400 mr-1">({targetObject})</span>
            : null;

        if (targetObject) {
            return (
                <div className="flex items-center">
                    {typeHint}
                    <button
                        onClick={(e) => { e.stopPropagation(); onNavigate(targetObject!, String(value)); }}
                        className="text-blue-600 hover:text-blue-800 font-medium hover:underline text-left truncate"
                    >
                        {displayLabel}
                    </button>
                </div>
            );
        }
    }
    // Non-clickable display - still use the resolved name
    return <span className="text-slate-700">{displayLabel}</span>;
};



// --- Registry Implementation ---

import { RegistryBase } from '@shared/utils';

// Concrete implementation for internal use
class SimpleRegistry<T> extends RegistryBase<T> { }

class UIRegistryClass {
    private renderers: RegistryBase<React.FC<FieldRendererProps>>;
    private inputs: RegistryBase<React.FC<FieldInputProps>>;


    constructor() {
        this.renderers = new SimpleRegistry<React.FC<FieldRendererProps>>({
            validator: (key, value) => typeof value === 'function' || typeof value === 'object' ? true : `Renderer "${key}" must be a component`,
            enableEvents: true
        });

        this.inputs = new SimpleRegistry<React.FC<FieldInputProps>>({
            validator: (key, value) => typeof value === 'function' || typeof value === 'object' ? true : `Input "${key}" must be a component`,
            enableEvents: true
        });



        this.registerDefaults();
    }

    private registerDefaults() {
        // Renderers
        this.registerFieldRenderer('Boolean', BooleanRenderer);
        this.registerFieldRenderer('Picklist', PicklistRenderer);
        this.registerFieldRenderer('Url', UrlRenderer);
        this.registerFieldRenderer('Lookup', LookupRenderer);
        this.registerFieldRenderer('Currency', ({ value }) => <span>{formatCurrency(Number(value))}</span>);
        this.registerFieldRenderer('Date', ({ value }) => <span>{formatDate(String(value))}</span>);
        this.registerFieldRenderer('DateTime', ({ value }) => <span>{formatDateTime(String(value))}</span>);
        this.registerFieldRenderer('Number', ({ value }) => <span>{Number(value).toLocaleString()}</span>);
        this.registerFieldRenderer('Percent', ({ value }) => <span>{Number(value)}%</span>);

        // Inputs
        this.registerFieldInput('Picklist', ({ field, value, onChange, disabled, required, options, onKeyDown, autoFocus }) => (
            <select className="w-full border border-slate-300 rounded-md px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white disabled:bg-slate-100 disabled:text-slate-500" required={required} onChange={(e) => onChange(e.target.value)} value={(value as string) || ''} disabled={disabled} onKeyDown={onKeyDown} autoFocus={autoFocus}>
                <option value="">-- Select {field.label} --</option>
                {(options || field.options || []).map(opt => <option key={opt} value={opt}>{opt}</option>)}
            </select>
        ));
        this.registerFieldInput('Lookup', ({ field, value, onChange, disabled, placeholder }) => (
            <SearchableLookup objectApiName={field.reference_to || ''} value={(value as string) || ''} onChange={(v, r) => onChange(v)} disabled={disabled} placeholder={placeholder || `Search...`} />
        ));
        this.registerFieldInput('TextArea', ({ value, onChange, disabled, required, placeholder, onKeyDown, autoFocus }) => (
            <textarea className="w-full border border-slate-300 rounded-md px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none disabled:bg-slate-100 disabled:text-slate-500" required={required} onChange={(e) => onChange(e.target.value)} value={(value as string) || ''} rows={3} placeholder={placeholder} disabled={disabled} onKeyDown={onKeyDown} autoFocus={autoFocus} />
        ));
        this.registerFieldInput('Boolean', ({ field, value, onChange, disabled, onKeyDown, autoFocus }) => (
            <div className="flex items-center gap-2 pt-2">
                <input type="checkbox" id={`field-${field.api_name}`} className="rounded border-slate-300 text-blue-600 focus:ring-blue-500 h-4 w-4 disabled:opacity-50" checked={!!value} onChange={(e) => onChange(e.target.checked)} disabled={disabled} onKeyDown={onKeyDown} autoFocus={autoFocus} />
                <label htmlFor={`field-${field.api_name}`} className={`text-sm font-medium cursor-pointer select-none ${disabled ? 'text-slate-400' : 'text-slate-700'}`}>{field.label}</label>
            </div>
        ));

        // Widgets

    }

    registerFieldRenderer(type: string, component: React.FC<FieldRendererProps>) {
        this.renderers.register(type, component);
    }

    getFieldRenderer(type: string): React.FC<FieldRendererProps> {
        return this.renderers.get(type) || (({ value }) => <span className="text-slate-700">{String(value ?? '')}</span>);
    }

    registerFieldInput(type: string, component: React.FC<FieldInputProps>) {
        this.inputs.register(type, component);
    }

    getFieldInput(type: string): React.FC<FieldInputProps> {
        return this.inputs.get(type) || ((props) => {
            let inputType = 'text';
            if (props.field.type === 'Number' || props.field.type === 'Currency' || props.field.type === 'Percent') inputType = 'number';
            if (props.field.type === 'Date') inputType = 'date';
            if (props.field.type === 'DateTime') inputType = 'datetime-local';

            // Read-only fields check
            if (props.field.api_name === COMMON_FIELDS.CREATED_DATE || props.field.api_name === COMMON_FIELDS.LAST_MODIFIED_DATE) {
                return <span className="text-slate-700">{String(props.value ?? '')}</span>;
            }

            return (
                <input
                    type={inputType}
                    className="w-full border border-slate-300 rounded-md px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none disabled:bg-slate-100 disabled:text-slate-500"
                    required={props.required}
                    onChange={(e) => props.onChange(e.target.value)}
                    value={(props.value as string) || ''}
                    placeholder={props.placeholder}
                    disabled={props.disabled}
                    step={props.field.type === 'Currency' || props.field.type === 'Percent' ? '0.01' : undefined}
                    onKeyDown={props.onKeyDown}
                    autoFocus={props.autoFocus}
                />
            );
        });
    }


}

export const UIRegistry = new UIRegistryClass();