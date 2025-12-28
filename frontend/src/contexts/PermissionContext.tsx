import React, { createContext, useContext, useState, useEffect, ReactNode, useCallback } from 'react';
import { authAPI } from '../infrastructure/api/auth';
import type { ObjectPermission, FieldPermission } from '../types';
import { useRuntime } from './RuntimeContext';

interface PermissionContextValue {
    objectPermissions: Record<string, ObjectPermission>;
    fieldPermissions: Record<string, FieldPermission>;
    loading: boolean;
    refreshPermissions: () => Promise<void>;
    hasObjectPermission: (objectApiName: string, action: 'read' | 'create' | 'edit' | 'delete' | 'view_all' | 'modify_all') => boolean;
    hasFieldPermission: (objectApiName: string, fieldApiName: string, action: 'readable' | 'editable') => boolean;
}

const PermissionContext = createContext<PermissionContextValue | null>(null);

export function PermissionProvider({ children }: { children: ReactNode }) {
    const { user, authStatus } = useRuntime();
    // Use singular ObjectPermission type directly
    const [objectPermissions, setObjectPermissions] = useState<Record<string, ObjectPermission>>({});
    const [fieldPermissions, setFieldPermissions] = useState<Record<string, FieldPermission>>({});
    const [loading, setLoading] = useState(false);

    const refreshPermissions = useCallback(async () => {
        if (authStatus !== 'authenticated') {
            setObjectPermissions({});
            setFieldPermissions({});
            return;
        }

        setLoading(true);
        try {
            const data = await authAPI.getMyPermissions();

            // Transform array to map for faster lookup
            const objPermMap: Record<string, ObjectPermission> = {};
            if (data.objectPermissions) {
                data.objectPermissions.forEach((p: ObjectPermission) => {
                    objPermMap[p.object_api_name] = p;
                });
            }
            setObjectPermissions(objPermMap);

            // Transform field permissions
            const fieldPermMap: Record<string, FieldPermission> = {};
            if (data.fieldPermissions) {
                // Key format: Object.Field
                data.fieldPermissions.forEach((p: FieldPermission) => {
                    fieldPermMap[`${p.object_api_name}.${p.field_api_name}`] = p;
                });
            }
            setFieldPermissions(fieldPermMap);

        } catch (err) {
            // Silently fail for permissions to avoid blocking the UI
            // In a real app, we might want to log this to a monitoring service
            setObjectPermissions({});
            setFieldPermissions({});
        } finally {
            setLoading(false);
        }
    }, [authStatus]);

    useEffect(() => {
        refreshPermissions();
    }, [refreshPermissions]);

    const hasObjectPermission = useCallback((objectApiName: string, action: 'read' | 'create' | 'edit' | 'delete' | 'view_all' | 'modify_all'): boolean => {
        const perm = objectPermissions[objectApiName];
        if (!perm) return false;

        if (perm.modify_all) return true; // Modify All grants everything

        switch (action) {
            case 'read': return perm.allow_read || perm.view_all;
            case 'create': return perm.allow_create;
            case 'edit': return perm.allow_edit;
            case 'delete': return perm.allow_delete;
            case 'view_all': return perm.view_all;
            case 'modify_all': return perm.modify_all;
            default: return false;
        }
    }, [objectPermissions]);

    const hasFieldPermission = useCallback((objectApiName: string, fieldApiName: string, action: 'readable' | 'editable'): boolean => {
        const key = `${objectApiName}.${fieldApiName}`;
        const fieldPerm = fieldPermissions[key];
        const objPerm = objectPermissions[objectApiName];

        // Fallback logic matches backend visibility check
        if (!objPerm || !objPerm.allow_read) return false; // Must have object read access

        if (fieldPerm) {
            return action === 'readable' ? fieldPerm.readable : fieldPerm.editable;
        }

        // Default behavior if no FLS record:
        if (action === 'readable') return true;
        if (action === 'editable') return objPerm.allow_edit; // Can't edit field if can't edit object

        return false;
    }, [objectPermissions, fieldPermissions]);

    const value = {
        objectPermissions,
        fieldPermissions,
        loading,
        refreshPermissions,
        hasObjectPermission,
        hasFieldPermission
    };

    return (
        <PermissionContext.Provider value={value}>
            {children}
        </PermissionContext.Provider>
    );
}

export const usePermissions = () => {
    const context = useContext(PermissionContext);
    if (!context) {
        throw new Error('usePermissions must be used within a PermissionProvider');
    }
    return context;
};
