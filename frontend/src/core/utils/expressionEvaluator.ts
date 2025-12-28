/**
 * Expression evaluator utilities for safe client-side expression evaluation.
 * Used for visibility conditions, templates, and dynamic values.
 */

/**
 * Safely evaluate a simple boolean expression against a record.
 * Supports basic field comparisons like: record.status === 'Open'
 * 
 * @param expression - The expression string to evaluate
 * @param record - The record object to evaluate against
 * @returns boolean result of the expression, false on error
 */
export const evaluateExpression = (expression: string, record: Record<string, unknown>): boolean => {
    if (!expression || !record) return true;

    try {
        // Create a safe function that only has access to the record
        // SECURITY WARNING: This uses 'new Function' which is similar to eval. 
        // Ensure that the expression comes from a trusted source (Admin configuration).
        const safeEval = new Function('record', `"use strict"; return (${expression});`);
        const result = safeEval(record);
        return Boolean(result);
    } catch (e) {
        console.warn('Expression evaluation error:', e);
        return false;
    }
};

/**
 * Substitute template placeholders in a string with record values.
 * Placeholders are in the format {fieldName}.
 * 
 * @param template - String with {field} placeholders
 * @param record - Record object with field values
 * @returns String with placeholders replaced by actual values
 */
export const substituteTemplate = (template: string, record: Record<string, unknown>): string => {
    if (!template || !record) return template || '';

    return template.replace(/\{(\w+)\}/g, (match, fieldName) => {
        const value = record[fieldName];
        if (value === undefined || value === null) return '';
        return String(value);
    });
};

/**
 * Evaluate a formula expression that may return any type.
 * Used for computed values in action configs.
 * 
 * @param formula - Expression starting with = (e.g., =record.amount * 1.1)
 * @param context - Context object with record and optional user
 * @returns The evaluated value or undefined on error
 */
export const evaluateFormula = (formula: string, context: { record: Record<string, unknown>; user?: Record<string, unknown> }): unknown => {
    if (!formula) return undefined;

    // Remove leading = if present
    const expr = formula.startsWith('=') ? formula.substring(1) : formula;

    try {
        // SECURITY WARNING: This uses 'new Function'. Ensure formula comes from trusted source.
        const fn = new Function('record', 'user', `"use strict"; return (${expr});`);
        return fn(context.record, context.user);
    } catch (e) {
        console.error('Formula evaluation error:', e);
        return undefined;
    }
};
