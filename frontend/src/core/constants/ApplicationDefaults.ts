export const APP_IDENTITY = {
    VERSION: '3.0.0',
    VERSION_NAME: 'MetaPaaS',
    TITLE: 'Nexus',
    SUBTITLE: 'CRM',
    VENDOR: 'NexusCRM',
    COPYRIGHT: `© ${new Date().getFullYear()} NexusCRM. All rights reserved.`,
} as const;

/**
 * Centralized localStorage key names
 * Prevents magic strings and makes key management easier
 */
export const STORAGE_KEYS = {
    AUTH_TOKEN: 'auth_token',
    CURRENT_APP: 'nexuscrm_current_app',
    SIDEBAR_COLLAPSED: 'sidebarCollapsed',
    AI_CONTEXT_FILES: 'nexus_ai_active_files',
    AI_CONTEXT_TOKENS: 'nexus_ai_total_tokens',
    AI_MESSAGES: 'nexus_ai_messages',
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

export const BREAKPOINTS = {
    lg: 1200,
    md: 996,
    sm: 768,
    xs: 480,
    xxs: 0,
} as const;

export const GRID_COLS = {
    lg: 12,
    md: 10,
    sm: 6,
    xs: 4,
    xxs: 2,
} as const;

/**
 * Common time constants in milliseconds
 */
export const TIME_MS = {
    SECOND: 1000,
    MINUTE: 60 * 1000,
    HOUR: 60 * 60 * 1000,
    DAY: 24 * 60 * 60 * 1000,
} as const;

export const APP_LOCALE = 'en-US';

export const KEYS = {
    ENTER: 'Enter',
    ESCAPE: 'Escape',
    TAB: 'Tab',
    ARROW_UP: 'ArrowUp',
    ARROW_DOWN: 'ArrowDown',
    ARROW_LEFT: 'ArrowLeft',
    ARROW_RIGHT: 'ArrowRight',
    SPACE: ' ',
    BACKSPACE: 'Backspace',
} as const;

export const DOM_EVENTS = {
    MOUSE_DOWN: 'mousedown',
    MOUSE_UP: 'mouseup',
    MOUSE_MOVE: 'mousemove',
    KEY_DOWN: 'keydown',
    KEY_UP: 'keyup',
    RESIZE: 'resize',
    SCROLL: 'scroll',
    BLUR: 'blur',
    FOCUS: 'focus',
} as const;

/**
 * UI Timing Constants - Centralized timeout and debounce values
 * Prevents magic numbers and ensures consistent UX timing
 */
export const UI_TIMING = {
    // Debounce delays
    DEBOUNCE_FAST_MS: 300,           // Search, typing input
    DEBOUNCE_NORMAL_MS: 500,         // Form validation, generic debounce
    DEBOUNCE_SLOW_MS: 1000,          // Heavy operations

    // Animation/transition durations
    ANIMATION_FAST_MS: 200,          // Tooltips, hover effects
    ANIMATION_NORMAL_MS: 300,        // Modal transitions, toast fade
    ANIMATION_SLOW_MS: 500,          // Refresh spinners, loading states

    // Feedback durations
    COPY_FEEDBACK_MS: 2000,          // "Copied!" message duration
    SUCCESS_FEEDBACK_MS: 1500,       // Brief success indicator

    // Polling intervals
    POLLING_FAST_MS: 30000,          // Notifications (30 sec)
    POLLING_NORMAL_MS: 60000,        // Approvals (1 min)
    POLLING_SLOW_MS: 300000,         // Background tasks (5 min)

    // AI-specific
    AI_SAVE_DELAY_MS: 2000,          // Debounce AI context save
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
