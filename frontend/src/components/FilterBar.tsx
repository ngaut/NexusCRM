import React, { useState, useEffect } from 'react';
import { User, Calendar, X } from 'lucide-react';
import { dataAPI } from '../infrastructure/api/data';
import { SObject } from '../types';

export interface GlobalFilters {
    ownerId?: string;
    startDate?: string;
    endDate?: string;
}

interface FilterBarProps {
    filters: GlobalFilters;
    onFilterChange: (filters: GlobalFilters) => void;
}

export const FilterBar: React.FC<FilterBarProps> = ({ filters, onFilterChange }) => {
    const [users, setUsers] = useState<SObject[]>([]);
    const [loadingUsers, setLoadingUsers] = useState(false);
    const [usersFetched, setUsersFetched] = useState(false);

    useEffect(() => {
        // Fetch users for the owner dropdown
        // Assuming 'User' is a queryable object and has 'name' field
        if (usersFetched) return; // Only fetch once

        const fetchUsers = async () => {
            setLoadingUsers(true);
            try {
                // If User object doesn't exist, this might fail, so we catch error
                const records = await dataAPI.query({
                    objectApiName: 'user',
                    limit: 50 // Limit to 50 users for now
                });
                setUsers(records);
            } catch (err) {
                // Silently fail - User object might not be created yet
                // Don't spam console with warnings
            } finally {
                setLoadingUsers(false);
                setUsersFetched(true);
            }
        };
        fetchUsers();
    }, [usersFetched]);

    const handleChange = (key: keyof GlobalFilters, value: string | undefined) => {
        const newFilters = { ...filters, [key]: value };
        // Clean undefined/empty string
        if (!value) delete newFilters[key];
        onFilterChange(newFilters);
    };

    const clearFilters = () => {
        onFilterChange({});
    };

    const hasFilters = Object.keys(filters).length > 0;

    return (
        <div className="bg-white border border-slate-200 rounded-lg p-3 shadow-sm flex flex-wrap items-center gap-4">
            <div className="flex items-center text-slate-500 text-sm font-medium">
                <span className="mr-2">Global Filters:</span>
            </div>

            {/* Owner Filter */}
            <div className="flex items-center gap-2">
                <User className="w-4 h-4 text-slate-400" />
                <select
                    value={filters.ownerId || ''}
                    onChange={(e) => handleChange('ownerId', e.target.value)}
                    className="text-sm border-slate-300 rounded-md focus:ring-blue-500 focus:border-blue-500 min-w-[150px]"
                >
                    <option value="">All Owners</option>
                    {users.map(u => (
                        <option key={u.id} value={u.id}>{String(u.name || u.username || 'Unknown User')}</option>
                    ))}
                </select>
            </div>

            <div className="h-6 w-px bg-slate-200 mx-2 hidden sm:block"></div>

            {/* Date Range Filter */}
            <div className="flex items-center gap-2">
                <Calendar className="w-4 h-4 text-slate-400" />
                <input
                    type="date"
                    value={filters.startDate || ''}
                    onChange={(e) => handleChange('startDate', e.target.value)}
                    className="text-sm border-slate-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                    placeholder="Start Date"
                />
                <span className="text-slate-400">-</span>
                <input
                    type="date"
                    value={filters.endDate || ''}
                    onChange={(e) => handleChange('endDate', e.target.value)}
                    className="text-sm border-slate-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                    placeholder="End Date"
                />
            </div>

            {hasFilters && (
                <button
                    onClick={clearFilters}
                    className="ml-auto text-sm text-slate-500 hover:text-red-600 flex items-center gap-1"
                >
                    <X className="w-3 h-3" />
                    Clear
                </button>
            )}
        </div>
    );
};
