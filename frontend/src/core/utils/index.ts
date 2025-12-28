/**
 * Core utility module - re-exports all utilities from a single location.
 * 
 * Usage:
 *   import { formatCurrency, parseCSV, getIcon } from '../core/utils';
 */

// Formatting utilities
export {
    formatCurrency,
    formatDate,
    formatDateTime,
    formatPercent,
    formatBoolean,
    formatNumber,
    formatValue,

} from './formatting';

// Parsing utilities
export {
    parseStandardValue,
    parseCSV,
} from './parsing';

// Icon utilities
export {
    getIcon,
} from './icons';

// Color utilities
export {
    getColorClasses,
} from './colorClasses';

// Expression evaluation
export {
    evaluateExpression,
    substituteTemplate,
    evaluateFormula,
} from './expressionEvaluator';

// Error handling utilities
export {
    formatApiError,
    getOperationErrorMessage,
} from './errorHandling';
export type { ApiError } from './errorHandling';
export * from './recordUtils';
