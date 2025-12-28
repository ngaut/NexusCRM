export const APP_IDENTITY = {
    VERSION: '3.0.0',
    VERSION_NAME: 'MetaPaaS',
    TITLE: 'Nexus',
    SUBTITLE: 'CRM',
    VENDOR: 'NexusCRM',
    COPYRIGHT: `© ${new Date().getFullYear()} NexusCRM. All rights reserved.`,
} as const;

export const DATABASE_DEFAULTS = {
    ISOLATION_LEVEL: 'READ_COMMITTED' as const,
    QUERY_TIMEOUT_MS: 30000,
    POOL_SIZE: 10,
} as const;

export const PLATFORM_LIMITS = {
    MAX_LIST_VIEW_ROWS: 200,
    MAX_SEARCH_RESULTS: 50,
    MAX_BATCH_SIZE: 200,
    MAX_RELATED_LIST_ROWS: 5,
    MAX_HIERARCHY_DEPTH: 10,
    MAX_FLOWS_PER_OBJECT: 50,
    MAX_FORMULA_EXECUTION_TIME_MS: 1000,
    MAX_VALIDATIONS_PER_OBJECT: 20,
    MAX_IMPORT_FILE_SIZE_BYTES: 10 * 1024 * 1024,
} as const;

export const CACHE_SETTINGS = {
    MAX_FORMULA_CACHE_SIZE: 300,
    METADATA_CACHE_TTL_MS: 5 * 60 * 1000,
    PERMISSION_CACHE_TTL_MS: 2 * 60 * 1000,
    QUERY_CACHE_TTL_MS: 30 * 1000,
    ENABLE_QUERY_CACHE: false,
} as const;

export const UI_DEFAULTS = {
    DEFAULT_THEME_COLOR: 'blue' as const,
    CHART_COLORS: [
        '#3b82f6',
        '#10b981',
        '#f59e0b',
        '#ef4444',
        '#8b5cf6',
        '#6366f1',
        '#ec4899',
        '#14b8a6',
    ] as const,
    DEFAULT_PAGE_SIZE: 25,
    TOAST_DURATION_MS: 3000,
    DEFAULT_FORM_COLUMNS: 2,
    MAX_RECENT_ITEMS: 20,
} as const;

export const RECORD_DEFAULTS = {
    DEFAULT_SHARING_MODEL: 'Private' as const,
    DEFAULT_ENABLE_HIERARCHY_SHARING: false,
    DEFAULT_SEARCHABLE: true,
    DEFAULT_LIST_VIEW_TYPE: 'List' as const,
    DEFAULT_DELETE_RULE: 'Restrict' as const,
    DEFAULT_CUSTOM_OBJECT_ICON: 'Box',
    DEFAULT_CUSTOM_OBJECT_COLOR: 'slate',
} as const;

export const SYSTEM_BEHAVIOR = {
    DEFAULT_TRACK_HISTORY: false,
    ENABLE_RECYCLE_BIN: true,
    RECYCLE_BIN_RETENTION_DAYS: 15,
    ENABLE_AUDIT_LOGGING: true,
    LOG_RETENTION_DAYS: 90,
    AUTO_REPAIR_SCHEMA: true,
    AUTO_SEED_DEMO_DATA: true,
} as const;

export const SECURITY_DEFAULTS = {
    REQUIRE_AUTH: false,
    ENABLE_RLS: true,
    ENABLE_FLS: true,
    ENABLE_SHARING_RULES: true,
    SESSION_TIMEOUT_MS: 0,
} as const;

export const FEATURE_FLAGS = {
    ENABLE_AI_ASSISTANT: true,
    ENABLE_AI_IMAGE_GEN: true,
    ENABLE_FLOWS: true,
    ENABLE_FORMULAS: true,
    ENABLE_ROLLUPS: true,
    ENABLE_VALIDATIONS: true,
    ENABLE_TRANSFORMATIONS: true,
    ENABLE_KANBAN: true,
    ENABLE_IMPORT: true,
    ENABLE_SCHEMA_GRAPH: true,
} as const;

export const DEV_SETTINGS = {
    VERBOSE_LOGGING: false,
    LOG_SQL_QUERIES: false,
    ENABLE_PROFILING: false,
    DEBUG_MODE: false,
} as const;

export function getLimit(limitName: keyof typeof PLATFORM_LIMITS, defaultValue?: number): number {
    return defaultValue ?? PLATFORM_LIMITS[limitName];
}

export function isFeatureEnabled(featureName: keyof typeof FEATURE_FLAGS): boolean {
    return FEATURE_FLAGS[featureName];
}
