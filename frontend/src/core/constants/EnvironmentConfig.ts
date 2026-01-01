/**
 * Environment Configuration
 *
 * Single source of truth for all environment-dependent values.
 * All configuration must come from environment variables in production.
 *
 * @module core/constants/EnvironmentConfig
 */

/**
 * Check if running in production environment
 */
export const IS_PRODUCTION = process.env.NODE_ENV === 'production';

/**
 * Check if running in development environment
 */
export const IS_DEVELOPMENT = process.env.NODE_ENV === 'development';

/**
 * Check if running in test environment
 */
export const IS_TEST = process.env.NODE_ENV === 'test';

/**
 * Backend API Configuration
 */
export const API_CONFIG = {
    /**
     * Backend API Base URL
     * Defaults to localhost in development, requires explicit config in production
     */
    BACKEND_URL: process.env.REACT_APP_BACKEND_URL ||
        (IS_PRODUCTION
            ? (() => { throw new Error('REACT_APP_BACKEND_URL must be set in production'); })()
            : ''), // Use relative path in dev to leverage Vite proxy

    /**
     * Frontend Application URL (used for CORS configuration)
     */
    FRONTEND_URL: process.env.REACT_APP_FRONTEND_URL ||
        (IS_PRODUCTION
            ? (() => { throw new Error('REACT_APP_FRONTEND_URL must be set in production'); })()
            : 'http://localhost:3000'),

    /**
     * API Request Timeout (milliseconds)
     */
    REQUEST_TIMEOUT_MS: parseInt(process.env.REACT_APP_REQUEST_TIMEOUT_MS || '30000', 10),

    /**
     * Enable request retry on failure
     */
    ENABLE_RETRY: process.env.REACT_APP_ENABLE_RETRY !== 'false',

    /**
     * Maximum number of retry attempts
     */
    MAX_RETRIES: parseInt(process.env.REACT_APP_MAX_RETRIES || '2', 10),
} as const;

/**
 * UI Behavior Configuration
 */
export const UI_CONFIG = {
    /**
     * Enable auto-fill of API Name field from Label
     * When true: typing in Label auto-generates API Name (can cause duplication with browser autofill)
     * When false: API Name must be entered manually
     * Default: false (disabled to prevent browser autofill conflicts)
     */
    ENABLE_AUTO_FILL_API_NAME: process.env.REACT_APP_ENABLE_AUTO_FILL_API_NAME !== 'false',
} as const;

/**
 * NOTE: Database, Authentication, and Server configurations
 * are handled by the backend Go server.
 * Frontend should only communicate via the backend API.
 */

/**
 * AI / External Services Configuration
 */
export const EXTERNAL_SERVICES = {
    /**
     * Google Gemini API Key
     * Optional - AI features disabled if not set
     */
    GEMINI_API_KEY: process.env.REACT_APP_API_KEY || process.env.API_KEY || '',

    /**
     * Enable AI features
     */
    AI_ENABLED: Boolean(process.env.REACT_APP_API_KEY || process.env.API_KEY),

    /**
     * AI Request Timeout (milliseconds)
     */
    AI_TIMEOUT_MS: parseInt(process.env.AI_TIMEOUT_MS || '30000', 10),
} as const;

/**
 * Logging Configuration
 */
export const LOGGING_CONFIG = {
    /**
     * Log Level
     * Options: 'error', 'warn', 'info', 'debug', 'trace'
     */
    LOG_LEVEL: process.env.LOG_LEVEL || (IS_PRODUCTION ? 'info' : 'debug'),

    /**
     * Enable SQL query logging
     */
    LOG_SQL: process.env.LOG_SQL === 'true' || (!IS_PRODUCTION && process.env.LOG_SQL !== 'false'),

    /**
     * Enable performance profiling
     */
    ENABLE_PROFILING: process.env.ENABLE_PROFILING === 'true',

    /**
     * Enable verbose logging
     */
    VERBOSE: process.env.VERBOSE === 'true' || IS_DEVELOPMENT,
} as const;


/**
 * Feature Flags from Environment
 * These override the defaults in ApplicationDefaults.ts
 */
export const FEATURE_FLAGS_ENV = {
    /**
     * Override individual features via environment variables
     * Example: FEATURE_AI_ASSISTANT=false
     */
    AI_ASSISTANT: process.env.FEATURE_AI_ASSISTANT === undefined ? undefined : process.env.FEATURE_AI_ASSISTANT === 'true',
    FLOWS: process.env.FEATURE_FLOWS === undefined ? undefined : process.env.FEATURE_FLOWS === 'true',
    FORMULAS: process.env.FEATURE_FORMULAS === undefined ? undefined : process.env.FEATURE_FORMULAS === 'true',
    ROLLUPS: process.env.FEATURE_ROLLUPS === undefined ? undefined : process.env.FEATURE_ROLLUPS === 'true',
    VALIDATIONS: process.env.FEATURE_VALIDATIONS === undefined ? undefined : process.env.FEATURE_VALIDATIONS === 'true',
    TRANSFORMATIONS: process.env.FEATURE_TRANSFORMATIONS === undefined ? undefined : process.env.FEATURE_TRANSFORMATIONS === 'true',
    KANBAN: process.env.FEATURE_KANBAN === undefined ? undefined : process.env.FEATURE_KANBAN === 'true',
    IMPORT: process.env.FEATURE_IMPORT === undefined ? undefined : process.env.FEATURE_IMPORT === 'true',
    SCHEMA_GRAPH: process.env.FEATURE_SCHEMA_GRAPH === undefined ? undefined : process.env.FEATURE_SCHEMA_GRAPH === 'true',
} as const;

/**
 * Validate all required environment variables
 * Call this at application startup
 */
export function validateEnvironment(): { valid: boolean; errors: string[] } {
    const errors: string[] = [];

    // Production-only validations
    if (IS_PRODUCTION) {
        if (!process.env.REACT_APP_BACKEND_URL) {
            errors.push('REACT_APP_BACKEND_URL is required in production');
        }
        if (!process.env.REACT_APP_FRONTEND_URL) {
            errors.push('REACT_APP_FRONTEND_URL is required in production');
        }
    }

    // Validate timeouts
    if (API_CONFIG.REQUEST_TIMEOUT_MS < 1000 || API_CONFIG.REQUEST_TIMEOUT_MS > 120000) {
        errors.push('REQUEST_TIMEOUT_MS must be between 1000 and 120000');
    }

    return {
        valid: errors.length === 0,
        errors
    };
}

/**
 * Print environment configuration summary
 * Useful for debugging and startup diagnostics
 */
export function printEnvironmentSummary(): void {
    // Validate and show any errors
    const validation = validateEnvironment();
    if (!validation.valid) {
        console.error('⚠️  CONFIGURATION ERRORS:');
        validation.errors.forEach(err => console.error(`   - ${err}`));
    }
}

/**
 * Get a safe configuration summary for logging (no secrets)
 */
export function getConfigSummary(): Record<string, any> {
    return {
        environment: process.env.NODE_ENV || 'development',
        isProduction: IS_PRODUCTION,
        isDevelopment: IS_DEVELOPMENT,
        api: {
            backendUrl: API_CONFIG.BACKEND_URL,
            frontendUrl: API_CONFIG.FRONTEND_URL,
            timeoutMs: API_CONFIG.REQUEST_TIMEOUT_MS,
            retryEnabled: API_CONFIG.ENABLE_RETRY,
        },
        features: {
            aiEnabled: EXTERNAL_SERVICES.AI_ENABLED,
        },
        logging: {
            level: LOGGING_CONFIG.LOG_LEVEL,
            profiling: LOGGING_CONFIG.ENABLE_PROFILING,
            verbose: LOGGING_CONFIG.VERBOSE,
        },
    };
}
