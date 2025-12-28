/**
 * Parsing utilities for CSV, data coercion, etc.
 * Extracted from utils.ts
 */

import { FieldType } from '../../types';

// --- Type Coercion ---

/**
 * Centralized Type Coercion logic.
 * Converts raw string/mixed input (from CSV, Forms, Flow Actions) into database-ready typed values.
 */
export const parseStandardValue = (value: unknown, type: FieldType): unknown => {
    if (value === undefined || value === null) return null;
    if (typeof value === 'string' && value.trim() === '') return null;

    switch (type) {
        case 'Number':
        case 'Currency':
        case 'Percent':
        case 'RollupSummary':
            const strVal = String(value).replace(/[$,%]/g, ''); // Strip formatting
            const num = parseFloat(strVal);
            return isNaN(num) ? null : num;

        case 'Boolean':
            const lowerVal = String(value).toLowerCase();
            return ['true', '1', 'yes', 'y', 'on'].includes(lowerVal) ? 1 : 0; // Standardize to SQL 1/0

        case 'Date':
            const d = new Date(value as string | number | Date);
            return isNaN(d.getTime()) ? null : d.toISOString().split('T')[0];

        case 'DateTime':
            const dt = new Date(value as string | number | Date);
            return isNaN(dt.getTime()) ? null : dt.toISOString().slice(0, 19).replace('T', ' ');

        default:
            // Text, TextArea, Picklist, JSON, etc.
            return String(value);
    }
};

// --- CSV Parsing ---

/**
 * Simple CSV parser for import functionality.
 * Returns an array of objects with header names as keys.
 */
export const parseCSV = (csvText: string): Record<string, string>[] => {
    const lines = csvText.trim().split('\n');
    if (lines.length < 2) return [];

    const headers = lines[0].split(',').map(h => h.trim().replace(/^"|"$/g, ''));
    const results = [];

    for (let i = 1; i < lines.length; i++) {
        const currentLine = lines[i].split(',').map(val => val.trim().replace(/^"|"$/g, ''));

        if (currentLine.length === headers.length) {
            const obj: Record<string, string> = {};
            headers.forEach((header, index) => {
                if (header) obj[header] = currentLine[index];
            });
            results.push(obj);
        }
    }
    return results;
};
