import React, { useState, useEffect } from 'react';
import { Code, Layout, Plus, X, AlertTriangle, CheckCircle2 } from 'lucide-react';
import { FieldMetadata } from '../../types';
import { UIRegistry } from '../../registries/UIRegistry';
import { MetadataRegistry } from '../../registries/MetadataRegistry';

interface FormulaEditorProps {
    objectApiName?: string;
    fields?: FieldMetadata[];
    value: string;
    onChange: (value: string) => void;
    label?: string;
    helpText?: string;
}

interface CriterionRow {
    id: string;
    field: string;
    op: string;
    val: string | number | boolean;
}

// Simple parser for extracting basic conditions (e.g. `state == "CA" && amount > 1000`)
const parseFormulaToRows = (formula: string, fields: FieldMetadata[]): CriterionRow[] | null => {
    if (!formula) return [];

    // Split by top-level ANDs (simplistic: assumes no nested ANDs or quoted ANDs)
    // Real formula parsing is hard, this is a best-effort for MVP
    const parts = formula.split(' && ').map(p => p.trim());
    const rows: CriterionRow[] = [];

    for (const part of parts) {
        // Match standard operators: ==, !=, >, <, >=, <=, contains() logic is harder
        // Regex: (identifier) (operator) (value)
        // Value might be string "..." or number
        const match = part.match(/^([a-zA-Z0-9_]+)\s*(==|!=|>=|<=|>|<)\s*(.+)$/);

        if (match) {
            let [_, field, op, valStr] = match;

            // Clean up value quotes
            let val: string | number | boolean = valStr;
            if (valStr.startsWith('"') && valStr.endsWith('"')) {
                val = valStr.slice(1, -1);
            } else if (!isNaN(Number(valStr))) {
                val = Number(valStr);
            } else if (valStr === 'true') val = true;
            else if (valStr === 'false') val = false;

            // Validate field exists
            if (fields.length > 0 && !fields.find(f => f.api_name === field)) {
                return null; // Field not found, treat as complex formula
            }

            rows.push({
                id: Math.random().toString(36).substr(2, 9),
                field,
                op,
                val
            });
        } else {
            // Check for contains(field, "value")
            const containsMatch = part.match(/^contains\(([a-zA-Z0-9_]+),\s*"(.+)"\)$/i);
            if (containsMatch) {
                rows.push({
                    id: Math.random().toString(36).substr(2, 9),
                    field: containsMatch[1],
                    op: 'contains',
                    val: containsMatch[2]
                });
            } else {
                return null; // Can't parse, fallback to text mode
            }
        }
    }
    return rows;
};

// Convert rows back to formula string
const rowsToFormula = (rows: CriterionRow[], fields: FieldMetadata[]): string => {
    return rows
        .filter(r => r.field && r.val !== '' && r.val !== null && r.val !== undefined)
        .map(r => {
            const fieldDef = fields.find(f => f.api_name === r.field);
            const isString = fieldDef?.type === 'Text' || fieldDef?.type === 'Picklist' || typeof r.val === 'string';

            let valStr = r.val;
            if (isString) valStr = `"${r.val}"`;

            if (r.op === 'contains') {
                return `contains(${r.field}, ${valStr})`;
            }
            return `${r.field} ${r.op} ${valStr}`;
        })
        .join(' && ');
};

export const FormulaEditor: React.FC<FormulaEditorProps> = ({
    objectApiName,
    fields = [],
    value,
    onChange,
    label = "Criteria",
    helpText
}) => {
    const [mode, setMode] = useState<'visual' | 'code'>('visual');
    const [rows, setRows] = useState<CriterionRow[]>([]);
    const [parseError, setParseError] = useState<string | null>(null);

    // Initial load: try to parse formula
    useEffect(() => {
        if (!value) {
            setRows([{ id: '1', field: fields[0]?.api_name || '', op: '==', val: '' }]);
            return;
        }

        // Only try parsing if we haven't modified rows recently? 
        // No, we want to stay in sync. But careful of loops.
        // For MVP: simple one-way sync on mount or mode switch?
        // Let's rely on mode switching logic mostly.
    }, []);

    useEffect(() => {
        const parsed = parseFormulaToRows(value, fields);
        if (parsed) {
            setRows(parsed.length ? parsed : [{ id: '1', field: fields[0]?.api_name || '', op: '==', val: '' }]);
            setParseError(null);
        } else {
            if (mode === 'visual' && value) {
                setMode('code'); // Auto-switch if unparseable
                setParseError("Complex formula detected. Switched to code editor.");
            }
        }
    }, [value, mode, fields]);

    const handleRowChange = (id: string, updates: Partial<CriterionRow>) => {
        const newRows = rows.map(r => r.id === id ? { ...r, ...updates } : r);
        setRows(newRows);
        const newFormula = rowsToFormula(newRows, fields);
        onChange(newFormula);
    };

    const addRow = () => {
        const newRows = [...rows, { id: Math.random().toString(36), field: fields[0]?.api_name || '', op: '==', val: '' }];
        setRows(newRows);
        onChange(rowsToFormula(newRows, fields));
    };

    const removeRow = (id: string) => {
        const newRows = rows.filter(r => r.id !== id);
        setRows(newRows);
        onChange(rowsToFormula(newRows, fields));
    };

    // Helper to get input component
    const renderInput = (row: CriterionRow) => {
        const fieldDef = fields.find(f => f.api_name === row.field);
        if (!fieldDef) return <input type="text" value={String(row.val)} onChange={e => handleRowChange(row.id, { val: e.target.value })} className="border rounded px-2 py-1" />;

        const InputComponent = UIRegistry.getFieldInput(fieldDef.type);
        return (
            <div className="flex-1 min-w-[150px]">
                <InputComponent
                    field={fieldDef}
                    value={row.val}
                    onChange={(val: string | number | boolean) => handleRowChange(row.id, { val })}
                />
            </div>
        );
    };

    return (
        <div className="space-y-3">
            <div className="flex justify-between items-center">
                <label className="block text-sm font-medium text-slate-700">
                    {label}
                </label>
                <div className="flex bg-slate-100 rounded-lg p-1">
                    <button
                        onClick={() => setMode('visual')}
                        className={`px-3 py-1 text-xs font-medium rounded-md transition-all ${mode === 'visual' ? 'bg-white shadow text-blue-600' : 'text-slate-500 hover:text-slate-700'}`}
                        disabled={!!parseError && parseError.includes("Complex")}
                    >
                        <Layout size={14} className="inline mr-1" />
                        Builder
                    </button>
                    <button
                        onClick={() => setMode('code')}
                        className={`px-3 py-1 text-xs font-medium rounded-md transition-all ${mode === 'code' ? 'bg-white shadow text-blue-600' : 'text-slate-500 hover:text-slate-700'}`}
                    >
                        <Code size={14} className="inline mr-1" />
                        Code
                    </button>
                </div>
            </div>

            {mode === 'code' ? (
                <div>
                    <textarea
                        value={value}
                        onChange={(e) => onChange(e.target.value)}
                        className="w-full h-32 px-3 py-2 font-mono text-sm border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 bg-slate-900 text-slate-50"
                        placeholder="state == 'CA' && amount > 1000"
                    />
                    {parseError && (
                        <div className="mt-2 flex items-center text-amber-600 text-xs gap-1">
                            <AlertTriangle size={12} />
                            {parseError}
                        </div>
                    )}
                    <p className="mt-1 text-xs text-slate-500">
                        Supports: <code>==</code>, <code>!=</code>, <code>&gt;</code>, <code>&&</code>, <code>||</code>, <code>contains()</code>, etc.
                    </p>
                </div>
            ) : (
                <div className="bg-slate-50 p-4 rounded-lg border border-slate-200 space-y-3">
                    {rows.map(row => (
                        <div key={row.id} className="flex items-center gap-2">
                            <select
                                value={row.field}
                                onChange={(e) => handleRowChange(row.id, { field: e.target.value, val: '' })}
                                className="w-1/3 px-2 py-1.5 border border-slate-300 rounded text-sm"
                            >
                                {fields.map(f => (
                                    <option key={f.api_name} value={f.api_name}>{f.label}</option>
                                ))}
                            </select>
                            <select
                                value={row.op}
                                onChange={(e) => handleRowChange(row.id, { op: e.target.value })}
                                className="w-1/4 px-2 py-1.5 border border-slate-300 rounded text-sm"
                            >
                                <option value="==">equals</option>
                                <option value="!=">not equal</option>
                                <option value=">">gt</option>
                                <option value="<">lt</option>
                                <option value="contains">contains</option>
                            </select>
                            {renderInput(row)}
                            <button onClick={() => removeRow(row.id)} className="text-slate-400 hover:text-red-500">
                                <X size={16} />
                            </button>
                        </div>
                    ))}
                    <button onClick={addRow} className="text-sm text-blue-600 hover:text-blue-800 font-medium flex items-center gap-1">
                        <Plus size={14} /> Add Condition
                    </button>
                </div>
            )}

            {helpText && <p className="text-xs text-slate-500">{helpText}</p>}
        </div>
    );
};
