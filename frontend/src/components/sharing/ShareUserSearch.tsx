import React, { useState } from 'react';
import { Search, User, Users, X, UserPlus } from 'lucide-react';

export interface UserOption {
    id: string;
    name: string;
    email?: string;
    type: 'user' | 'group';
}

interface ShareUserSearchProps {
    onSearch: (query: string) => Promise<UserOption[]>;
    onSelect: (option: UserOption) => void;
    searching: boolean;
    searchResults: UserOption[];
    selectedEntity: UserOption | null;
    onClearSelection: () => void;
    accessLevel: 'Read' | 'Edit';
    onAccessLevelChange: (level: 'Read' | 'Edit') => void;
}

export const ShareUserSearch: React.FC<ShareUserSearchProps> = ({
    onSearch,
    onSelect,
    searching,
    searchResults,
    selectedEntity,
    onClearSelection,
    accessLevel,
    onAccessLevelChange,
}) => {
    const [query, setQuery] = useState('');

    const handleSearchChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setQuery(value);
        if (value.length >= 2) {
            onSearch(value); // Trigger parent search which updates searchResults prop
        }
    };

    return (
        <>
            {/* Search Input */}
            <div className="relative mb-4">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                    type="text"
                    placeholder="Search users or groups..."
                    value={query}
                    onChange={handleSearchChange}
                    className="w-full pl-10 pr-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                />
                {searching && (
                    <div className="absolute right-3 top-1/2 -translate-y-1/2">
                        <div className="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
                    </div>
                )}
            </div>

            {/* Search Results */}
            {searchResults.length > 0 && !selectedEntity && (
                <div className="mb-4 border border-gray-200 rounded-xl overflow-hidden">
                    {searchResults.slice(0, 5).map((result) => (
                        <button
                            key={`${result.type}-${result.id}`}
                            onClick={() => {
                                onSelect(result);
                                setQuery(''); // Clear query on select
                            }}
                            className="w-full flex items-center gap-3 px-4 py-3 hover:bg-gray-50 transition-colors text-left border-b last:border-b-0"
                        >
                            {result.type === 'user' ? (
                                <div className="p-2 bg-blue-100 rounded-full">
                                    <User className="w-4 h-4 text-blue-600" />
                                </div>
                            ) : (
                                <div className="p-2 bg-purple-100 rounded-full">
                                    <Users className="w-4 h-4 text-purple-600" />
                                </div>
                            )}
                            <div>
                                <div className="font-medium text-gray-900">{result.name}</div>
                                <div className="text-sm text-gray-500">
                                    {result.type === 'user' ? result.email : 'Group'}
                                </div>
                            </div>
                        </button>
                    ))}
                </div>
            )}

            {/* Selected Entity + Access Level */}
            {selectedEntity && (
                <div className="mb-4 p-4 bg-blue-50 rounded-xl border border-blue-200">
                    <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-3">
                            {selectedEntity.type === 'user' ? (
                                <div className="p-2 bg-blue-100 rounded-full">
                                    <User className="w-4 h-4 text-blue-600" />
                                </div>
                            ) : (
                                <div className="p-2 bg-purple-100 rounded-full">
                                    <Users className="w-4 h-4 text-purple-600" />
                                </div>
                            )}
                            <div>
                                <div className="font-medium text-gray-900">{selectedEntity.name}</div>
                                <div className="text-sm text-gray-500 capitalize">{selectedEntity.type}</div>
                            </div>
                        </div>
                        <button
                            onClick={onClearSelection}
                            className="p-1 hover:bg-blue-200 rounded transition-colors"
                        >
                            <X className="w-4 h-4 text-blue-600" />
                        </button>
                    </div>

                    {/* Access Level Selection */}
                    <div className="flex gap-2">
                        <button
                            onClick={() => onAccessLevelChange('Read')}
                            className={`flex-1 flex items-center justify-center gap-2 px-4 py-2 rounded-lg font-medium transition-all ${accessLevel === 'Read'
                                ? 'bg-blue-600 text-white'
                                : 'bg-white text-gray-700 border border-gray-200 hover:border-blue-300'
                                }`}
                        >
                            Read Only
                        </button>
                        <button
                            onClick={() => onAccessLevelChange('Edit')}
                            className={`flex-1 flex items-center justify-center gap-2 px-4 py-2 rounded-lg font-medium transition-all ${accessLevel === 'Edit'
                                ? 'bg-blue-600 text-white'
                                : 'bg-white text-gray-700 border border-gray-200 hover:border-blue-300'
                                }`}
                        >
                            Read & Edit
                        </button>
                    </div>
                </div>
            )}
        </>
    );
};
