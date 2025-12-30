import React, { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { dataAPI } from '../infrastructure/api/data';
import type { SearchResult, SObject } from '../types';
import { Search, Loader2, X } from 'lucide-react';
import { getRecordDisplayName } from '../core/utils/recordUtils';

export const GlobalSearch: React.FC = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [results, setResults] = useState<SearchResult[]>([]);
    const [loading, setLoading] = useState(false);
    const [isOpen, setIsOpen] = useState(false);
    const searchRef = useRef<HTMLDivElement>(null);
    const inputRef = useRef<HTMLInputElement>(null);
    const navigate = useNavigate();

    // Close dropdown on outside click
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (searchRef.current && !searchRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        if (isOpen) {
            document.addEventListener('mousedown', handleClickOutside);
        }

        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [isOpen]);

    // Debounced search
    useEffect(() => {
        if (!searchTerm || searchTerm.length < 2) {
            setResults([]);
            setIsOpen(false);
            return;
        }

        const timer = setTimeout(() => {
            performSearch(searchTerm);
        }, 300); // 300ms debounce

        return () => clearTimeout(timer);
    }, [searchTerm]);

    // Track latest search term for race conditions
    const currentSearchTermRef = useRef(searchTerm);
    useEffect(() => {
        currentSearchTermRef.current = searchTerm;
    }, [searchTerm]);

    const performSearch = async (term: string) => {
        setLoading(true);
        setIsOpen(true);

        try {
            const searchResults = await dataAPI.search(term);

            // Race condition check using Ref
            if (currentSearchTermRef.current !== term) {
                // If the current input differs from what we searched for, discard results.
                // This covers the case where user cleared input (current='', term='abc')
                return;
            }

            setResults(searchResults);
        } catch {
            setResults([]);
        } finally {
            // Only turn off loading if we processed the results (or if we want to ensure loading spinner stops)
            // Ideally check ref here too to avoid flicker, but loading=false is usually safe.
            setLoading(false);
        }
    };

    const handleResultClick = (objectApiName: string, recordId: string) => {
        navigate(`/object/${objectApiName}/${recordId}`);
        setSearchTerm('');
        setResults([]);
        setIsOpen(false);
        inputRef.current?.blur();
    };

    const handleClear = () => {
        setSearchTerm('');
        setResults([]);
        setIsOpen(false);
        inputRef.current?.focus();
    };

    const totalResults = results.reduce((sum, group) => sum + group.matches.length, 0);


    // Removed local getRecordDisplayName in favor of utility


    const getRecordSecondaryInfo = (record: SObject) => {
        // Collect secondary info to display (e.g. Email, Phone, or other common identifiers)
        const parts = [];

        // Common high-priority secondary fields
        if (record.email) parts.push(String(record.email));
        if (record.phone) parts.push(String(record.phone));
        if (record.company) parts.push(String(record.company));
        if (record.title) parts.push(String(record.title));

        return parts.join(' • ');
    };

    return (
        <div className="relative w-full max-w-md" ref={searchRef}>
            {/* Search Input */}
            <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    {loading ? (
                        <Loader2 size={18} className="text-slate-400 animate-spin" />
                    ) : (
                        <Search size={18} className="text-slate-400" />
                    )}
                </div>
                <input
                    ref={inputRef}
                    type="text"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    onFocus={() => searchTerm.length >= 2 && setIsOpen(true)}
                    placeholder="Search records..."
                    className="block w-full pl-10 pr-10 py-2 border border-slate-300 rounded-lg text-sm placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
                {searchTerm && (
                    <button
                        onClick={handleClear}
                        className="absolute inset-y-0 right-0 pr-3 flex items-center hover:bg-slate-100 rounded-r-lg transition-colors"
                    >
                        <X size={16} className="text-slate-400" />
                    </button>
                )}
            </div>

            {/* Results Dropdown */}
            {isOpen && (
                <div className="absolute top-full left-0 right-0 mt-2 bg-white border border-slate-200 rounded-lg shadow-lg max-h-96 overflow-y-auto z-50">
                    {loading && (
                        <div className="px-4 py-8 text-center text-slate-500 text-sm">
                            Searching...
                        </div>
                    )}

                    {!loading && results.length === 0 && searchTerm.length >= 2 && (
                        <div className="px-4 py-8 text-center text-slate-500 text-sm">
                            No results found for "{searchTerm}"
                        </div>
                    )}

                    {!loading && results.length > 0 && (
                        <>
                            {/* Header */}
                            <div className="px-4 py-2 border-b border-slate-200 bg-slate-50">
                                <p className="text-xs font-semibold text-slate-600 uppercase tracking-wide">
                                    {totalResults} result{totalResults !== 1 ? 's' : ''} found
                                </p>
                            </div>

                            {/* Results by Object */}
                            {results.map((group) => (
                                <div key={group.object_api_name} className="border-b border-slate-100 last:border-0">
                                    <div className="px-4 py-2 bg-slate-50">
                                        <div className="flex items-center gap-2">
                                            <span className="text-xl">{group.icon}</span>
                                            <h4 className="text-sm font-semibold text-slate-700">
                                                {group.object_label}
                                            </h4>
                                            <span className="text-xs text-slate-500">
                                                ({group.matches.length})
                                            </span>
                                        </div>
                                    </div>
                                    <div className="divide-y divide-slate-100">
                                        {group.matches.map((record) => (
                                            <button
                                                key={record.id}
                                                onClick={() => handleResultClick(group.object_api_name, record.id!)}
                                                className="w-full px-4 py-3 text-left hover:bg-blue-50 transition-colors flex items-start gap-3"
                                            >
                                                <div className="flex-1 min-w-0">
                                                    <p className="text-sm font-medium text-slate-900 truncate">
                                                        {String(getRecordDisplayName(record))}
                                                    </p>
                                                    <p className="text-xs text-slate-500 mt-0.5 truncate">
                                                        {getRecordSecondaryInfo(record)}
                                                    </p>
                                                </div>
                                            </button>
                                        ))}
                                    </div>
                                </div>
                            ))}
                        </>
                    )}
                </div>
            )}
        </div>
    );
};
