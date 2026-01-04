/**
 * Application Routes
 * 
 * Centralized route path definitions to avoid magic strings.
 * Use these constants for navigation and route matching.
 */

export const ROUTES = {
    // Root routes
    HOME: '/',
    LOGIN: '/login',

    // Object routes
    OBJECT: {
        LIST: (objectApiName: string) => `/object/${objectApiName}`,
        DETAIL: (objectApiName: string, recordId: string) => `/object/${objectApiName}/${recordId}`,
        NEW: (objectApiName: string) => `/object/${objectApiName}/new`,
        EDIT: (objectApiName: string, recordId: string) => `/object/${objectApiName}/${recordId}?edit=true`,
    },

    // Setup routes
    SETUP: {
        ROOT: '/setup',
        OBJECTS: '/setup/objects',
        OBJECT_DETAIL: (objectApiName: string) => `/setup/objects/${objectApiName}`,
        OBJECT_LAYOUT: (objectApiName: string) => `/setup/objects/${objectApiName}/layout`,
        FLOWS: '/setup/flows',
        SHARING_RULES: '/setup/sharing-rules',
        GROUPS: '/setup/groups',
        USERS: '/setup/users',
        PROFILES: '/setup/profiles',
        PERMISSION_SETS: '/setup/permissionsets',
        APP_MANAGER: '/setup/appmanager',
        APPS: '/setup/apps',
        RECYCLE_BIN: '/setup/recyclebin',
    },

    // Dashboard
    DASHBOARD: '/dashboard',

    // Approval
    APPROVAL_QUEUE: '/approvals',
} as const;

/**
 * Build a route path with query parameters
 */
export function buildRoute(basePath: string, params?: Record<string, string>): string {
    if (!params || Object.keys(params).length === 0) {
        return basePath;
    }
    const queryString = new URLSearchParams(params).toString();
    return `${basePath}?${queryString}`;
}
