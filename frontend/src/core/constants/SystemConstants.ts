/**
 * System Constants
 * 
 * Core system constants for permissions, logging, and object categories.
 * Consolidated from legacy api.ts and system.json.
 */

// ============================================================================
// PERMISSIONS
// ============================================================================

export const PERMISSION_READ = 'Read';
export const PERMISSION_CREATE = 'Create';
export const PERMISSION_EDIT = 'Edit';
export const PERMISSION_DELETE = 'Delete';
export const PERMISSION_VIEW_ALL = 'ViewAll';
export const PERMISSION_MODIFY_ALL = 'ModifyAll';

export const PERMISSION_TYPES = [
    PERMISSION_READ,
    PERMISSION_CREATE,
    PERMISSION_EDIT,
    PERMISSION_DELETE,
    PERMISSION_VIEW_ALL,
    PERMISSION_MODIFY_ALL,
] as const;

export type PermissionType = typeof PERMISSION_TYPES[number];

// ============================================================================
// LOG LEVELS
// ============================================================================

export const LOG_LEVEL = {
    DEBUG: 'DEBUG',
    INFO: 'INFO',
    WARNING: 'WARNING',
    ERROR: 'ERROR',
} as const;

export type LogLevel = typeof LOG_LEVEL[keyof typeof LOG_LEVEL];

// ============================================================================
// DEFAULTS
// ============================================================================

export const PAGINATION_DEFAULTS = {
    LIMIT: 25,
    MAX_LIMIT: 1000,
    ORDER_DIR: 'ASC' as 'ASC' | 'DESC',
};

// ============================================================================
// CATEGORIES
// ============================================================================

export const APP_CATEGORIES = {
    STANDARD: 'Standard',
    CUSTOM: 'Custom',
    SYSTEM: 'System',
} as const;
