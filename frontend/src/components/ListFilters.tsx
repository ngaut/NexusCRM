import React, { useState, useEffect } from 'react';
import { Filter, X } from 'lucide-react';
import type { FieldMetadata } from '../types';
import { FormulaEditor } from './formula/FormulaEditor';

interface ListFiltersProps {
    objectApiName: string;
    fields: FieldMetadata[];
    activeFilterExpr?: string;
    onFiltersChange: (filterExpr: string) => void;
}

export const ListFilters: React.FC<ListFiltersProps> = ({ objectApiName, fields, activeFilterExpr, onFiltersChange }) => {
    const [isOpen, setIsOpen] = useState(false);
    const [localExpr, setLocalExpr] = useState(activeFilterExpr || '');

    useEffect(() => {
        setLocalExpr(activeFilterExpr || '');
    }, [activeFilterExpr, isOpen]);

    const handleApply = () => {
        onFiltersChange(localExpr);
        setIsOpen(false);
    };

    const handleClear = () => {
        setLocalExpr('');
        onFiltersChange('');
        setIsOpen(false);
    };

    const activeFilterCount = activeFilterExpr ? 1 : 0;

    return (
        <div className="mb-4 relative">
            <div className="flex items-center justify-between">
                <button
                    onClick={() => setIsOpen(!isOpen)}
                    className={`flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors ${activeFilterCount > 0
                        ? 'bg-blue-600 text-white hover:bg-blue-700'
                        : 'bg-white text-slate-700 border border-slate-300 hover:bg-slate-50'
                        }`}
                >
                    <Filter size={16} />
                    Filters
                    {activeFilterCount > 0 && (
                        <span className="px-2 py-0.5 bg-white/20 rounded-full text-xs font-semibold">
                            {activeFilterCount}
                        </span>
                    )}
                </button>

                {activeFilterCount > 0 && (
                    <button
                        onClick={handleClear}
                        className="px-3 py-2 text-sm text-slate-600 hover:text-slate-900"
                    >
                        Clear all
                    </button>
                )}
            </div>

            {isOpen && (
                <div className="absolute top-12 left-0 z-20 w-[600px] max-w-[90vw] bg-white border border-slate-200 rounded-lg shadow-xl p-4">
                    <div className="flex justify-between items-center mb-4">
                        <h3 className="font-semibold text-slate-800">Filter Records</h3>
                        <button onClick={() => setIsOpen(false)} className="text-slate-400 hover:text-slate-600">
                            <X size={16} />
                        </button>
                    </div>

                    <div className="mb-4">
                        <FormulaEditor
                            objectApiName={objectApiName}
                            fields={fields}
                            value={localExpr}
                            onChange={setLocalExpr}
                            label="Filter Logic"
                        />
                    </div>

                    <div className="flex justify-end gap-2 border-t pt-3">
                        <button
                            onClick={() => setIsOpen(false)}
                            className="px-3 py-1.5 text-slate-600 hover:text-slate-800 text-sm font-medium"
                        >
                            Cancel
                        </button>
                        <button
                            onClick={handleApply}
                            className="px-3 py-1.5 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700"
                        >
                            Apply Filters
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};
