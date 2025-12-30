/**
 * Core Constants - Barrel Export
 * 
 * Central export point for all platform constants.
 * Import from here instead of individual files for cleaner code.
 * 
 * @example
 * ```typescript
 * import { PROFILE_IDS, SYSTEM_TABLES, APP_IDENTITY } from '@/core/constants';
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
    SYSTEM_TABLES,
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
    RECORD_DEFAULTS,
    SYSTEM_BEHAVIOR,
    SECURITY_DEFAULTS,
    FEATURE_FLAGS,
    DEV_SETTINGS,
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
