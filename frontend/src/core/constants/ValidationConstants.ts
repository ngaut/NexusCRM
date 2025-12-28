/**
 * Validation messages and patterns for UI forms
 */

export const VALIDATION_MESSAGES = {
    PICKLIST_NO_OPTIONS: 'Picklist fields require at least one option',
    LOOKUP_NO_REFERENCE: 'Lookup fields require a referenced object',
    SAVE_FAILED: 'Failed to save field. Please check console for details.',
} as const;

export const VALIDATION_PATTERNS = {
    API_NAME: /^[a-zA-Z][a-zA-Z0-9]*$/,
    API_NAME_TITLE: 'Must start with a letter, contain only letters and numbers',
} as const;

export const EXCLUDED_FORMULA_RETURN_TYPES = ['Formula', 'Picklist', 'Lookup'] as const;
