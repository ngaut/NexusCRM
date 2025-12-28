/**
 * Metadata-driven error handling utilities
 * Provides user-friendly error messages based on error type and context
 */

export interface ApiError {
    message: string;
    code?: string;
    field?: string;
    details?: unknown;
}

export class AppError extends Error {
    code?: string;
    field?: string;
    details?: unknown;

    constructor(message: string, code?: string, field?: string, details?: unknown) {
        super(message);
        this.name = 'AppError';
        this.code = code;
        this.field = field;
        this.details = details;
    }
}

/**
 * Type guard for error objects with message property
 */
function isErrorWithMessage(error: unknown): error is { message: string } {
    return (
        typeof error === 'object' &&
        error !== null &&
        'message' in error &&
        typeof (error as Record<string, unknown>).message === 'string'
    );
}

interface BackendError {
    status?: number;
    data?: {
        message?: string;
        field?: string;
        details?: unknown;
    };
    response?: {
        status: number;
        data?: {
            message?: string;
            field?: string;
            details?: unknown;
        };
    };
}

/**
 * Type guard for backend API error response
 */
function isBackendError(error: unknown): error is BackendError {
    return typeof error === 'object' && error !== null;
}

/**
 * Convert backend API errors to user-friendly messages
 */
export function formatApiError(error: unknown): AppError {
    // Network errors
    if (isErrorWithMessage(error)) {
        if (error.message === 'Failed to fetch' || error.message === 'Network request failed') {
            return new AppError(
                'Unable to connect to the server. Please check your internet connection and try again.',
                'NETWORK_ERROR'
            );
        }
    }

    if (!isBackendError(error)) {
        return new AppError(
            'An detailed error occurred. Please try again.',
            'UNKNOWN_ERROR'
        );
    }

    // Handle APIError from our fetch-based client (error.status) or axios-style (error.response.status)
    const status = error.status ?? error.response?.status;
    const data = error.data ?? error.response?.data;

    if (status) {
        switch (status) {
            case 400:
                return new AppError(
                    data?.message || 'Invalid request. Please check your input and try again.',
                    'VALIDATION_ERROR',
                    data?.field
                );

            case 401:
                return new AppError(
                    'Your session has expired. Please log in again.',
                    'UNAUTHORIZED'
                );

            case 403:
                return new AppError(
                    data?.message || 'You don\'t have permission to perform this action.',
                    'FORBIDDEN'
                );

            case 404:
                return new AppError(
                    data?.message || 'The requested resource was not found.',
                    'NOT_FOUND'
                );

            case 409:
                return new AppError(
                    data?.message || 'A record with this name or identifier already exists. Please use a different value.',
                    'CONFLICT',
                    data?.field
                );

            case 422:
                return new AppError(
                    data?.message || 'Validation failed. Please check your input.',
                    'VALIDATION_ERROR',
                    data?.field,
                    data?.details
                );

            case 500:
                return new AppError(
                    'An unexpected error occurred on the server. Our team has been notified.',
                    'SERVER_ERROR'
                );

            case 503:
                return new AppError(
                    'The service is temporarily unavailable. Please try again in a few moments.',
                    'SERVICE_UNAVAILABLE'
                );

            default:
                return new AppError(
                    data?.message || `An error occurred (${status}). Please try again.`,
                    'UNKNOWN_ERROR'
                );
        }
    }

    // Generic errors
    const message = isErrorWithMessage(error) ? error.message : 'An unexpected error occurred. Please try again.';

    return new AppError(
        message,
        'UNKNOWN_ERROR'
    );
}

/**
 * Get user-friendly error message for specific operations
 */
export function getOperationErrorMessage(operation: string, objectName: string, error: AppError): string {
    const objectLabel = objectName.replace(/([A-Z])/g, ' $1').trim(); // "LeadConversion" -> "Lead Conversion"

    const templates: Record<string, string> = {
        create: `Failed to create ${objectLabel}. ${error.message}`,
        update: `Failed to update ${objectLabel}. ${error.message}`,
        delete: `Failed to delete ${objectLabel}. ${error.message}`,
        fetch: `Failed to load ${objectLabel}. ${error.message}`,
        search: `Failed to search ${objectLabel}. ${error.message}`,
    };

    return templates[operation] || error.message;
}

/**
 * Validation error formatter for forms
 */
export function formatValidationErrors(errors: Record<string, string>): string {
    const errorMessages = Object.entries(errors)
        .map(([field, message]) => `${field}: ${message}`)
        .join('\n');

    return `Please fix the following errors:\n${errorMessages}`;
}

/**
 * Get suggested action for error recovery
 */
export function getErrorRecoveryAction(error: AppError): {
    label: string;
    action: 'retry' | 'refresh' | 'login' | 'contact_support' | 'dismiss';
} {
    switch (error.code) {
        case 'NETWORK_ERROR':
            return { label: 'Retry', action: 'retry' };

        case 'UNAUTHORIZED':
            return { label: 'Log In', action: 'login' };

        case 'NOT_FOUND':
            return { label: 'Go Back', action: 'dismiss' };

        case 'VALIDATION_ERROR':
            return { label: 'Fix Errors', action: 'dismiss' };

        case 'SERVER_ERROR':
            return { label: 'Contact Support', action: 'contact_support' };

        case 'SERVICE_UNAVAILABLE':
            return { label: 'Try Again', action: 'retry' };

        default:
            return { label: 'Retry', action: 'retry' };
    }
}

/**
 * Log errors for debugging (in development)
 */
export function logError(context: string, error: unknown) {
    if (process.env.NODE_ENV === 'development') {
        console.group(`🔴 Error in ${context}`);
        console.error('Error:', error);

        if (typeof error === 'object' && error !== null) {
            const errObj = error as Record<string, unknown>;
            if (errObj.response) {
                console.error('Response:', errObj.response);
            }
            if (errObj.stack) {
                console.error('Stack:', errObj.stack);
            }
        }
        console.groupEnd();
    }
}
