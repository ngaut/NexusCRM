import React, { useState, useEffect, useRef } from 'react';
import { Search, X, Loader2, ChevronDown } from 'lucide-react';
import { dataAPI } from '../infrastructure/api/data';
import type { SObject } from '../types';
import { useDebounce } from '../core/hooks/useDebounce';
import { getRecordDisplayName } from '../core/utils/recordUtils';
import { COMMON_FIELDS } from '../core/constants';

interface SearchableLookupProps {
    objectApiName: string | string[];
    value: string | null | undefined;
    onChange: (value: string | null, record?: SObject) => void;
    disabled?: boolean;
    placeholder?: string;
    error?: boolean;
    objectType?: string; // For Polymorphic Lookups: The specific type of the current value
}

export const SearchableLookup: React.FC<SearchableLookupProps> = ({
    objectApiName,
    value,
    onChange,
    disabled = false,
    placeholder = 'Search...',
    error = false,
    objectType
}) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [results, setResults] = useState<SObject[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [isOpen, setIsOpen] = useState(false);
    const [selectedRecord, setSelectedRecord] = useState<SObject | null>(null);
    const [isInitialLoad, setIsInitialLoad] = useState(true);
    const [highlightedIndex, setHighlightedIndex] = useState(-1);

    const containerRef = useRef<HTMLDivElement>(null);
    const inputRef = useRef<HTMLInputElement>(null);
    const debouncedSearchTerm = useDebounce(searchTerm, 300);

    // Initial load handling
    useEffect(() => {
        const loadInitialValue = async () => {
            if (value && !selectedRecord) {
                try {
                    setIsLoading(true);
                    // Determine which object to query
                    // If objectType is provided (best case), use it.
                    // If not, and objectApiName is string, use it.
                    // If array and no objectType, we might fail to resolve name unless we brute force.
                    let targetObject = objectType;
                    if (!targetObject && typeof objectApiName === 'string') {
                        targetObject = objectApiName;
                    }

                    if (targetObject) {
                        const record = await dataAPI.getRecord(targetObject, value);
                        setSelectedRecord(record);
                        setSearchTerm(getRecordDisplayName(record));
                    } else if (Array.isArray(objectApiName)) {
                        // Polymorphic catch-22: We have ID but don't know Type.
                        // Try to find it in one of the allowed objects?
                        // For now, if no type is known, just show ID or try first?
                        // Optimistic: Try first one? Or just show value.
                        // Better: Parallel Query?
                        // Let's try parallel query for robustness if list is small (< 5)
                        const promises = objectApiName.map(obj =>
                            dataAPI.getRecord(obj, value).then(r => ({ obj, r })).catch(() => null)
                        );
                        const results = await Promise.all(promises);
                        const match = results.find(res => res !== null);
                        if (match) {
                            setSelectedRecord(match.r);
                            setSearchTerm(getRecordDisplayName(match.r));
                        } else {
                            setSearchTerm(value);
                        }
                    } else {
                        setSearchTerm(value);
                    }
                } catch {
                    setSearchTerm(value); // Fallback to ID
                } finally {
                    setIsLoading(false);
                    setIsInitialLoad(false);
                }
            } else if (!value) {
                setSelectedRecord(null);
                setSearchTerm('');
                setIsInitialLoad(false);
            }
        };

        if (isInitialLoad || (value && value !== (selectedRecord?.[COMMON_FIELDS.ID] as string))) {
            loadInitialValue();
        }
    }, [value, objectApiName, objectType, isInitialLoad, selectedRecord]);

    // Search effect
    useEffect(() => {
        const search = async () => {
            if (!debouncedSearchTerm) {
                setResults([]);
                return;
            }

            // Don't search if the search term matches the selected record
            if (selectedRecord && (
                debouncedSearchTerm === selectedRecord[COMMON_FIELDS.NAME]
            )) {
                return;
            }

            try {
                setIsLoading(true);
                let records: SObject[] = [];

                if (Array.isArray(objectApiName)) {
                    // Polymorphic Search: Search all allowed objects
                    const promises = objectApiName.map(obj =>
                        dataAPI.searchSingleObject(obj, debouncedSearchTerm)
                            .then(recs => recs.map(r => ({ ...r, _object_type: obj }))) // Tag with type
                            .catch(() => [])
                    );
                    const results = await Promise.all(promises);
                    records = results.flat();
                } else {
                    records = await dataAPI.searchSingleObject(objectApiName, debouncedSearchTerm);
                }

                setResults(records);
                setHighlightedIndex(-1); // Reset highlight on new results
                setIsOpen(true);
            } catch {
                setResults([]);
            } finally {
                setIsLoading(false);
            }
        };

        if (debouncedSearchTerm && isOpen) {
            search();
        }
    }, [debouncedSearchTerm, objectApiName, isOpen, selectedRecord]);

    // Click outside to close
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
                setIsOpen(false);
                setHighlightedIndex(-1);
                // Reset search term to selected record name if closed without selection
                if (selectedRecord) {
                    setSearchTerm(getRecordDisplayName(selectedRecord));
                } else if (!value) {
                    setSearchTerm('');
                }
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, [selectedRecord, value]);

    const handleSelect = (record: SObject) => {
        setSelectedRecord(record);
        setSearchTerm(getRecordDisplayName(record));
        onChange(record[COMMON_FIELDS.ID] as string, record);
        setIsOpen(false);
        setHighlightedIndex(-1);
        setResults([]);
    };

    const handleClear = (e: React.MouseEvent) => {
        e.stopPropagation();
        setSelectedRecord(null);
        setSearchTerm('');
        onChange(null);
        setResults([]);
        setHighlightedIndex(-1);
        inputRef.current?.focus();
    };

    const handleFocus = () => {
        if (!disabled) {
            setIsOpen(true);
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (!isOpen || results.length === 0) {
            // If dropdown not open but we have a search term, open it
            if (e.key === 'ArrowDown' && searchTerm) {
                setIsOpen(true);
            }
            return;
        }

        switch (e.key) {
            case 'ArrowDown':
                e.preventDefault();
                setHighlightedIndex(prev =>
                    prev < results.length - 1 ? prev + 1 : 0
                );
                break;
            case 'ArrowUp':
                e.preventDefault();
                setHighlightedIndex(prev =>
                    prev > 0 ? prev - 1 : results.length - 1
                );
                break;
            case 'Enter':
                e.preventDefault();
                if (highlightedIndex >= 0 && highlightedIndex < results.length) {
                    handleSelect(results[highlightedIndex]);
                }
                break;
            case 'Escape':
                e.preventDefault();
                setIsOpen(false);
                setHighlightedIndex(-1);
                break;
            case 'Tab':
                // Select highlighted item on Tab if one is highlighted
                if (highlightedIndex >= 0 && highlightedIndex < results.length) {
                    handleSelect(results[highlightedIndex]);
                }
                setIsOpen(false);
                break;
        }
    };

    return (
        <div className="relative" ref={containerRef}>
            <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <Search className="h-4 w-4 text-slate-400" />
                </div>
                <input
                    ref={inputRef}
                    type="text"
                    value={searchTerm}
                    onChange={(e) => {
                        setSearchTerm(e.target.value);
                        setIsOpen(true);
                        setHighlightedIndex(-1);
                    }}
                    onFocus={handleFocus}
                    onKeyDown={handleKeyDown}
                    disabled={disabled}
                    className={`block w-full pl-10 pr-10 py-2 border ${error ? 'border-red-300 focus:border-red-500 focus:ring-red-500' : 'border-slate-300 focus:border-blue-500 focus:ring-blue-500'} rounded-lg text-sm bg-white disabled:bg-slate-50 disabled:text-slate-500 transition-colors shadow-sm outline-none`}
                    placeholder={placeholder}
                    autoComplete="off"
                    role="combobox"
                    aria-expanded={isOpen}
                    aria-haspopup="listbox"
                    aria-activedescendant={highlightedIndex >= 0 ? `lookup-option-${highlightedIndex}` : undefined}
                />
                <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
                    {isLoading ? (
                        <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />
                    ) : value && !disabled ? (
                        <button
                            type="button"
                            onClick={handleClear}
                            className="text-slate-400 hover:text-slate-600 rounded-full p-0.5 hover:bg-slate-100 transition-colors"
                        >
                            <X className="h-4 w-4" />
                        </button>
                    ) : (
                        <ChevronDown className="h-4 w-4 text-slate-400 pointer-events-none opacity-50" />
                    )}
                </div>
            </div>

            {isOpen && !disabled && (
                <div
                    className="absolute z-50 w-full mt-1 bg-white rounded-lg shadow-lg border border-slate-200 max-h-60 overflow-auto"
                    role="listbox"
                >
                    {results.length > 0 ? (
                        <ul className="py-1">
                            {results.map((record, index) => (
                                <li
                                    key={record[COMMON_FIELDS.ID] as string}
                                    id={`lookup-option-${index}`}
                                    role="option"
                                    aria-selected={highlightedIndex === index}
                                >
                                    <button
                                        type="button"
                                        onClick={() => handleSelect(record)}
                                        onMouseEnter={() => setHighlightedIndex(index)}
                                        className={`w-full px-4 py-2 text-left text-sm flex items-center gap-3 transition-colors group ${highlightedIndex === index
                                            ? 'bg-blue-50 text-blue-900'
                                            : 'hover:bg-slate-50'
                                            }`}
                                    >
                                        <div className={`w-8 h-8 rounded-full flex items-center justify-center font-medium transition-colors shrink-0 ${highlightedIndex === index
                                            ? 'bg-blue-200 text-blue-700'
                                            : 'bg-blue-100 text-blue-600 group-hover:bg-blue-200'
                                            }`}>
                                            {getRecordDisplayName(record).charAt(0).toUpperCase()}
                                        </div>
                                        <div className="min-w-0">
                                            <div className="font-medium text-slate-900 truncate">
                                                {getRecordDisplayName(record)}
                                            </div>
                                            <div className="text-xs text-slate-500 font-mono truncate">
                                                {(record as SObject & { _object_type?: string })._object_type || objectApiName} • {record[COMMON_FIELDS.ID]}
                                            </div>
                                        </div>
                                    </button>
                                </li>
                            ))}
                        </ul>
                    ) : searchTerm && !isLoading ? (
                        <div className="px-4 py-3 text-sm text-slate-500 text-center italic">
                            No results found for "{searchTerm}"
                        </div>
                    ) : (
                        <div className="px-4 py-3 text-sm text-slate-400 text-center italic">
                            Start typing to search...
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};


