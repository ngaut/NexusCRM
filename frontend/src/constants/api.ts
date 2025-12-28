/**
 * API and Configuration Constants
 * 
 * Imports from shared JSON - single source of truth matching protobuf enums.
 * @see shared/constants/system.json
 * @see shared/proto/constants.proto (Permission, LogLevel enums)
 */

import systemConstants from '@shared/constants/system.json';

// Extract from shared JSON  
const { permissions, logLevels, defaults, recordStates } = systemConstants;

// Permission operations - derived from shared JSON
export const PERMISSION_READ = permissions.READ.value;
export const PERMISSION_CREATE = permissions.CREATE.value;
export const PERMISSION_EDIT = permissions.EDIT.value;
export const PERMISSION_DELETE = permissions.DELETE.value;
export const PERMISSION_VIEW_ALL = permissions.VIEW_ALL.value;
export const PERMISSION_MODIFY_ALL = permissions.MODIFY_ALL.value;

// All permission types
export const PERMISSION_TYPES = [
    PERMISSION_READ,
    PERMISSION_CREATE,
    PERMISSION_EDIT,
    PERMISSION_DELETE,
    PERMISSION_VIEW_ALL,
    PERMISSION_MODIFY_ALL,
] as const;

export type PermissionType = typeof PERMISSION_TYPES[number];

// Log levels
export const LOG_LEVEL_DEBUG = logLevels.DEBUG;
export const LOG_LEVEL_INFO = logLevels.INFO;
export const LOG_LEVEL_WARNING = logLevels.WARNING;
export const LOG_LEVEL_ERROR = logLevels.ERROR;

// Sort directions
export const SORT_ASC = 'ASC';
export const SORT_DESC = 'DESC';

// Pagination defaults
export const DEFAULT_LIMIT = defaults.pageSize;
export const DEFAULT_MAX_LIMIT = defaults.maxPageSize;
export const DEFAULT_ORDER_DIR = defaults.orderDir;

// Record states  
export const IS_DELETED_TRUE = recordStates.DELETED;
export const IS_DELETED_FALSE = recordStates.ACTIVE;

// Object categories
export const CATEGORY_STANDARD = 'Standard';
export const CATEGORY_CUSTOM = 'Custom';
export const CATEGORY_SYSTEM = 'System';
