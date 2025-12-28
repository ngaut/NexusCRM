/**
 * Action Handler Types
 *
 * Type definitions for the metadata-driven action handler system.
 *
 * @module core/actions/ActionHandlerTypes
 */

import { SObject, UserSession } from '../../types';
// Mock types for backend services not available in frontend
// check if these files should be here?
export interface Transaction {
    commit: () => Promise<void>;
    rollback: () => Promise<void>;
}
export interface TiDBServiceManager {
    persistence: {
        insert: (table: string, data: Record<string, unknown>, user: UserSession | null, tx?: Transaction) => Promise<unknown>;
        update: (table: string, id: string, data: Record<string, unknown>, user: UserSession | null, tx?: Transaction) => Promise<void>;
        delete: (table: string, id: string, user: UserSession | null, tx?: Transaction) => Promise<void>;
        query: (sql: string, args?: unknown[]) => Promise<unknown[]>;
    };
    schema: Record<string, unknown>; // Keeping flexible for schema cache
    logSystemEvent: (level: 'INFO' | 'WARN' | 'ERROR', type: string, message: string, meta?: Record<string, unknown>) => Promise<void>;
    query: (sql: string, ...args: unknown[]) => Promise<unknown[]>;
    getSchema: (name: string) => unknown;
}

/**
 * Configuration field definition for action configuration UI
 */
export interface ActionConfigField {
    api_name: string;
    label: string;
    type: string;
    required?: boolean;
    default_value?: string;
    help_text?: string;
    options?: string[];
}

/**
 * Action configuration schema
 * Can be static or dynamic based on trigger object context
 */
export type ActionConfigSchema = ActionConfigField[] | ((triggerObject?: string) => ActionConfigField[]);

/**
 * Action handler function signature
 *
 * @param db - TiDB service instance for database access
 * @param record - The record that triggered the flow
 * @param config - User-configured action parameters from flow definition
 * @param flowName - Name of the flow executing this action
 * @param tx - Active database transaction
 * @param objectApiName - API name of the object that triggered the flow
 * @param currentUser - User session context
 */
export type ActionHandlerFunction = (
    db: TiDBServiceManager,
    record: SObject,
    config: Record<string, unknown>,
    flowName: string,
    tx?: Transaction,
    objectApiName?: string,
    currentUser?: UserSession | null
) => Promise<void>;

/**
 * Action handler definition
 * Combines the handler function with its metadata
 */
export interface ActionHandlerDefinition {
    /** Unique action type identifier (e.g., "CreateTask") */
    actionType: string;

    /** User-friendly display name */
    label: string;

    /** Description of what this action does */
    description?: string;

    /** Category for grouping in UI */
    category: 'Records' | 'Notifications' | 'Integration' | 'Other';

    /** Icon name for UI display */
    icon?: string;

    /** The actual handler implementation */
    handler: ActionHandlerFunction;

    /** Configuration schema for this action */
    configSchema: ActionConfigSchema;

    /** Whether this action requires object context */
    requiresObjectContext?: boolean;

    /** Sort order for UI display */
    sortOrder?: number;
}

/**
 * Action handler metadata from database
 */
export interface ActionHandlerMetadata {
    id: string;
    actionType: string;
    label: string;
    description?: string;
    category: 'Records' | 'Notifications' | 'Integration' | 'Other';
    handlerModule: string;
    icon?: string;
    configSchema: ActionConfigField[] | string; // JSON field
    requiresObjectContext: boolean;
    isActive: boolean;
    sortOrder: number;
}

/**
 * Action handler module export interface
 * All action handler modules must export this structure
 */
export interface ActionHandlerModule {
    /** The action handler definition */
    handler: ActionHandlerDefinition;
}
