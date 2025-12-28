import { dataAPI } from '../../infrastructure/api/data';
import { metadataAPI } from '../../infrastructure/api/metadata'; // Assuming this exists or will use dataAPI for everything
import { TiDBServiceManager } from '../actions/ActionHandlerTypes';

/**
 * Client-side adapter for TiDBServiceManager.
 * Allows action handlers designed for backend execution (conceptually) to run on the frontend
 * by proxying database calls to API endpoints.
 */
export const clientDBAdapter: TiDBServiceManager = {
    persistence: {
        // Map persistence.create/update/delete to dataAPI
        insert: async (objectName: string, data: Record<string, unknown>) => {
            return dataAPI.createRecord(objectName, data);
        },
        update: async (objectName: string, id: string, data: Record<string, unknown>) => {
            return dataAPI.updateRecord(objectName, id, data);
        },
        async delete(objectName: string, id: string) {
            return dataAPI.deleteRecord(objectName, id);
        },
        async query(sql: string, args: unknown[] = []) {
            // Placeholder: client-side SQL not fully supported, direct implementation or generic fallback
            console.warn('persistence.query called on client - SQL support limited');
            return [];
        }
    },
    schema: {
        // Map schema operations
        getObject: async (objectName: string) => {
            return metadataAPI.getSchema(objectName);
        }
    },
    // Log system event - in production, this could POST to /api/logs
    logSystemEvent: async (_level: string, _source: string, _message: string, _details?: unknown) => {
        // No-op in client - system logging should be handled server-side
    },
    // Generic query support
    query: async (objectName: string, criteria: Record<string, unknown>) => {
        // Convert simplified criteria object to array format expected by dataAPI.query
        // Criteria object: { field: value, ... }
        const terms = Object.entries(criteria).map(([field, val]) => {
            if (typeof val === 'string') {
                return `${field} == '${val}'`;
            }
            return `${field} == ${val}`;
        });
        const filterExpr = terms.join(' && ');

        return dataAPI.query({
            objectApiName: objectName,
            filterExpr: filterExpr,
            limit: 1000 // Default limit
        });
    },
    // Synchronous schema retrieval (mocked or cache-based)
    getSchema: (name: string) => {
        // This is synchronous in the interface but API is async.
        // For client-side, this might need pre-loading or architectural change.
        // Returning null/undefined as basic fallback, handlers should handle this or use async methods.
        console.warn('getSchema (sync) called on clientDBAdapter - not fully supported');
        return null;
    }
};
