import React, { createContext, useContext, useState, useEffect, ReactNode, useMemo } from 'react';
import { metadataAPI } from '../infrastructure/api/metadata';
import { useRuntime } from './RuntimeContext';
import type { AppConfig, NavigationItem } from '../types';
import { STORAGE_KEYS } from '../core/constants/ApplicationDefaults';

interface AppContextValue {
    apps: AppConfig[]; // All apps (for admin views)
    visibleApps: AppConfig[]; // Apps visible to current user based on profile
    currentAppId: string | null;
    currentApp: AppConfig | null;
    currentAppNavigationItems: NavigationItem[];
    loading: boolean;
    setCurrentAppId: (appId: string) => void;
    refreshApps: () => Promise<AppConfig[]>;
}

const AppContext = createContext<AppContextValue | null>(null);

interface AppProviderProps {
    children: ReactNode;
}

export function AppProvider({ children }: AppProviderProps) {
    const { user } = useRuntime();
    const [apps, setApps] = useState<AppConfig[]>([]);
    const [currentAppId, setCurrentAppIdState] = useState<string | null>(() => {
        // Restore from localStorage
        return localStorage.getItem(STORAGE_KEYS.CURRENT_APP) || null;
    });
    const [loading, setLoading] = useState(true);

    const refreshApps = async () => {
        try {
            const response = await metadataAPI.getApps();
            setApps(response.apps || []);
            return response.apps || [];
        } catch (error) {
            // Silently fail for app metadata issues to avoid UI blocking
            return [];
        }
    };

    useEffect(() => {
        const initialize = async () => {
            setLoading(true);
            try {
                const appsResponse = await metadataAPI.getApps();
                const loadedApps = appsResponse.apps || [];
                setApps(loadedApps);
            } catch (error) {
                // Initialize with empty state on failure
            } finally {
                setLoading(false);
            }
        };

        initialize();
    }, []);

    // Filter apps by user's profile
    const visibleApps = useMemo(() => {
        if (!user?.profile_id) return apps;

        return apps.filter(app => {
            // If no assignedProfiles or empty array, app is visible to all
            if (!app.assigned_profiles || app.assigned_profiles.length === 0) {
                return true;
            }
            // Otherwise, check if user's profile is in the list
            return app.assigned_profiles.includes(user.profile_id);
        });
    }, [apps, user?.profile_id]);

    // Default to first visible app if current app is not accessible
    // Note: currentAppId is intentionally NOT in dependencies to avoid infinite loop
    // This effect should only run when loading completes or visibleApps changes
    useEffect(() => {
        if (!loading && visibleApps.length > 0) {
            const currentIsVisible = visibleApps.some(app => app.id === currentAppId);
            if (!currentAppId || !currentIsVisible) {
                setCurrentAppIdState(visibleApps[0].id);
                localStorage.setItem(STORAGE_KEYS.CURRENT_APP, visibleApps[0].id);
            }
        }
    }, [loading, visibleApps]); // Removed currentAppId to prevent infinite re-render loop

    const setCurrentAppId = (appId: string) => {
        setCurrentAppIdState(appId);
        localStorage.setItem(STORAGE_KEYS.CURRENT_APP, appId);
    };

    const currentApp = visibleApps.find(app => app.id === currentAppId) || null;

    // Get navigation items for the current app
    const currentAppNavigationItems = currentApp?.navigation_items || [];

    const value: AppContextValue = {
        apps,
        visibleApps,
        currentAppId,
        currentApp,
        currentAppNavigationItems,
        loading,
        setCurrentAppId,
        refreshApps
    };

    return (
        <AppContext.Provider value={value}>
            {children}
        </AppContext.Provider>
    );
}

export function useApp() {
    const context = useContext(AppContext);
    if (!context) {
        throw new Error('useApp must be used within an AppProvider');
    }
    return context;
}
