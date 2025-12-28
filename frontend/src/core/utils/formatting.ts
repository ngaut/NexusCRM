/**
 * Formatting utilities for displaying data in the UI.
 * Consolidated from utils.ts and utils/formatting.ts
 */

import { FieldType } from '../../types';

// --- Currency & Number Formatting ---

export const formatCurrency = (value: number): string => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value);
};

export const formatPercent = (value: number): string => {
    return `${value}%`;
};

export const formatNumber = (value: number): string => {
    return Number(value).toLocaleString();
};

// --- Date Formatting ---

export const formatDate = (value: string | number | Date): string => {
    if (!value) return '';
    return new Date(value).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
};

export const formatDateTime = (value: string | number | Date): string => {
    if (!value) return '';
    return new Date(value).toLocaleString(undefined, { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
};

// --- Boolean Formatting ---

export const formatBoolean = (value: unknown): string => {
    return value ? 'Yes' : 'No';
};

// --- Unified Value Formatter ---

/**
 * Unified formatter for all data types.
 * Use this when displaying field values in the UI.
 */
export const formatValue = (value: unknown, type: string): string => {
    if (value === undefined || value === null || value === '') return '';

    switch (type) {
        case 'Currency': return formatCurrency(Number(value));
        case 'Date': return formatDate(String(value));
        case 'DateTime': return formatDateTime(String(value));
        case 'Boolean': return formatBoolean(value);
        case 'Percent': return formatPercent(Number(value));
        case 'Number': return formatNumber(Number(value));
        case 'Lookup': return String(value);
        case 'JSON': return typeof value === 'object' ? JSON.stringify(value) : String(value);
        default: return String(value);
    }
};


