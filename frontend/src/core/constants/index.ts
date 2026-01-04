/**
 * Core Constants - Barrel Export
 * 
 * Central export point for all platform constants.
 * Import from here instead of individual files for cleaner code.
 * 
 * @example
 * ```typescript
 * import { PROFILE_IDS, SYSTEM_TABLE_NAMES, APP_IDENTITY } from '@/core/constants';
 * ```
 * 
 * @module core/constants
 */

// System Profiles
export {
    PROFILE_IDS,
    SYSTEM_PROFILES,
    isSuperUserProfile,
    isSystemProfile,
    type ProfileId,
    type ProfileMetadata,
} from './SystemProfiles';

// System Objects
export {
    SYSTEM_TABLE_NAMES,
    type ObjectCategory,
    isSystemTable,
    isCustomObject,
    getObjectCategory,
    type SystemTableName,
} from './SystemObjects';

// Application Defaults
export {
    APP_IDENTITY,
    DATABASE_DEFAULTS,
    PLATFORM_LIMITS,
    CACHE_SETTINGS,
    UI_DEFAULTS,
    UI_TIMING,
    BREAKPOINTS,
    GRID_COLS,
    TIME_MS,
    KEYS,
    DOM_EVENTS,
    APP_LOCALE,
    RECORD_DEFAULTS,
    SYSTEM_BEHAVIOR,
    SECURITY_DEFAULTS,
    FEATURE_FLAGS,
    DEV_SETTINGS,
    STORAGE_KEYS,
    getLimit,
    isFeatureEnabled,
} from './ApplicationDefaults';

// Environment Configuration
export {
    IS_PRODUCTION,
    IS_DEVELOPMENT,
    IS_TEST,
    API_CONFIG,
    EXTERNAL_SERVICES,
    LOGGING_CONFIG,
    FEATURE_FLAGS_ENV,
    validateEnvironment,
    printEnvironmentSummary,
    getConfigSummary,
} from './EnvironmentConfig';

// Field Constants
export {
    COMMON_FIELDS,
} from './CommonFields';

// Approval Constants
export {
    APPROVAL_STATUS,
    type ApprovalStatus,
} from './ApprovalConstants';

// System Constants
export {
    PERMISSION_TYPES,
    PERMISSION_READ,
    PERMISSION_CREATE,
    PERMISSION_EDIT,
    PERMISSION_DELETE,
    PERMISSION_VIEW_ALL,
    PERMISSION_MODIFY_ALL,
    LOG_LEVEL,
    PAGINATION_DEFAULTS,
    APP_CATEGORIES,
    type PermissionType,
    type LogLevel,
} from './SystemConstants';

// Automation Constants
export {
    FLOW_STATUS,
    TRIGGER_TYPE,
    ACTION_TYPE,
    EVENT_TYPE,
    isBeforeTrigger,
    isAfterTrigger,
    type FlowStatus,
    type TriggerType,
    type ActionType,
    type EventType,
} from './FlowConstants';

// Route Constants
export {
    ROUTES,
    buildRoute,
} from './Routes';
