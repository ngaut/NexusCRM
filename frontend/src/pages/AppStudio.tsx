import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { ArrowLeft, Save, Box, ExternalLink } from 'lucide-react';
import { metadataAPI } from '../infrastructure/api/metadata';
import type { AppConfig, ObjectMetadata, NavigationItem } from '../types';
import { StudioSidebar } from '../components/studio/StudioSidebar';
import { StudioObjectEditor } from '../components/studio/StudioObjectEditor';
import { StudioPermissions } from '../components/studio/StudioPermissions';
import { StudioDashboardEditor } from '../components/studio/StudioDashboardEditor';
import { AppSettingsEditor } from '../components/studio/AppSettingsEditor';
import { getColorClasses } from '../core/utils/colorClasses';
import { useErrorToast } from '../components/ui/Toast';
import { Logger } from '../core/services/Logger';

type EditorMode = 'object' | 'dashboard' | 'settings' | 'permissions' | null;

interface EditorState {
    mode: EditorMode;
    objectApiName?: string;
    dashboardId?: string;
}

export const AppStudio: React.FC = () => {
    const errorToast = useErrorToast();
    const { appId } = useParams<{ appId: string }>();
    const navigate = useNavigate();
    const isNewApp = appId === 'new';

    const [searchParams] = useSearchParams();

    // App state
    const [app, setApp] = useState<AppConfig>({
        id: '',
        label: '',
        description: '',
        icon: 'Layers',
        color: 'blue',
        navigation_items: [],
    });
    const [loading, setLoading] = useState(!isNewApp);
    const [saving, setSaving] = useState(false);
    const [hasChanges, setHasChanges] = useState(false);

    // Available objects for adding to nav
    const [availableObjects, setAvailableObjects] = useState<ObjectMetadata[]>([]);

    // Editor state - what's currently being edited in the right panel
    const [editor, setEditor] = useState<EditorState>({ mode: null });

    // Load app data
    useEffect(() => {
        const init = async () => {
            if (!isNewApp && appId) {
                await loadApp(appId);
            } else {
                setLoading(false);
            }
            await loadObjects();
        };
        init();
    }, [appId, isNewApp]);

    // Handle Deep Linking
    useEffect(() => {
        const dashboardId = searchParams.get('dashboardId');
        if (dashboardId) {
            setEditor({ mode: 'dashboard', dashboardId });
        }
    }, [searchParams]);

    const shouldOpenAddModal = searchParams.get('action') === 'add_page';

    const loadApp = async (id: string) => {
        try {
            const response = await metadataAPI.getApps();
            const foundApp = (response.apps || []).find((a: AppConfig) => a.id === id);
            if (foundApp) {
                setApp(foundApp);
                // Auto-select first nav item for editing if no editor selected
                if (!editor.mode && foundApp.navigation_items?.length > 0) {
                    const firstItem = foundApp.navigation_items[0];
                    if (firstItem.type === 'object' && firstItem.object_api_name) {
                        setEditor({ mode: 'object', objectApiName: firstItem.object_api_name });
                    } else if (firstItem.type === 'dashboard' && firstItem.dashboard_id) {
                        setEditor({ mode: 'dashboard', dashboardId: firstItem.dashboard_id });
                    }
                }
            }
        } catch (error) {
            // App loading failure - log for debugging, handled via empty state
            Logger.warn('AppStudio: Failed to load app:', error instanceof Error ? error.message : error);
        } finally {
            setLoading(false);
        }
    };

    const loadObjects = async () => {
        try {
            const response = await metadataAPI.getSchemas();
            setAvailableObjects(response.schemas || []);
        } catch (error) {
            // Objects loading failure - log for debugging, handled via empty state
            Logger.warn('AppStudio: Failed to load objects:', error instanceof Error ? error.message : error);
        }
    };

    const handleSaveApp = async (appConfig: AppConfig) => {
        setSaving(true);
        try {
            if (isNewApp) {
                await metadataAPI.createApp(appConfig);
                navigate(`/studio/${appConfig.id}`, { replace: true });
            } else {
                await metadataAPI.updateApp(appConfig.id, appConfig);
            }
            setHasChanges(false);
        } catch (error: unknown) {
            const errorMessage = error instanceof Error ? error.message : 'Unknown error';
            // Extract API error message from response if available
            const apiError = typeof error === 'object' && error !== null && 'response' in error
                ? ((error as { response?: { data?: { error?: string } } }).response?.data?.error)
                : undefined;
            errorToast(apiError || errorMessage);
        } finally {
            setSaving(false);
        }
    };

    const updateApp = useCallback((updates: Partial<AppConfig>) => {
        setApp(prev => ({ ...prev, ...updates }));
        setHasChanges(true);
    }, []);

    const handleNavItemSelect = (item: NavigationItem) => {
        if (item.type === 'object' && item.object_api_name) {
            setEditor({ mode: 'object', objectApiName: item.object_api_name });
        } else if (item.type === 'dashboard' && item.dashboard_id) {
            setEditor({ mode: 'dashboard', dashboardId: item.dashboard_id });
        }
    };

    const handleAddObject = async (objectDef: { label: string; apiName: string; icon: string }) => {
        try {
            // Create the object via API
            await metadataAPI.createSchema({
                api_name: objectDef.apiName,
                label: objectDef.label,
                plural_label: objectDef.label + 's',
                icon: objectDef.icon,
                description: `Custom object: ${objectDef.label}`,
                is_custom: true,
                sharing_model: 'Private',
                fields: [],
                app_id: !isNewApp ? app.id : undefined, // Link to app if it exists
            });

            if (!isNewApp) {
                // If app exists, backend added the nav item. Reload app to get the new nav item ID.
                await loadApp(app.id);
            } else {
                // For new apps, manually add to local state
                const newNavItem: NavigationItem = {
                    id: `nav-${objectDef.apiName}-${Date.now()}`,
                    type: 'object',
                    object_api_name: objectDef.apiName,
                    label: objectDef.label + 's',
                    icon: objectDef.icon,
                };
                const updatedNavItems = [...(app.navigation_items || []), newNavItem];
                updateApp({ navigation_items: updatedNavItems });
            }

            // Refresh objects list and select the new object
            await loadObjects();
            setEditor({ mode: 'object', objectApiName: objectDef.apiName });

            return true;
        } catch (error: unknown) {
            throw error;
        }
    };

    const handleAddDashboard = async (def: { label: string; icon: string }) => {
        if (!appId) return false;
        try {
            // 1. Create Dashboard
            const newDashboardId = crypto.randomUUID();
            await metadataAPI.createDashboard({
                id: newDashboardId,
                label: def.label,
                widgets: []
            });

            // 2. Add to App Navigation (backend doesn't auto-add dashboards yet unlike objects)
            const newItem: NavigationItem = {
                id: crypto.randomUUID(),
                type: 'dashboard',
                label: def.label,
                icon: def.icon,
                dashboard_id: newDashboardId,
            };
            const updatedItems = [...(app.navigation_items || []), newItem];

            if (!isNewApp) {
                await metadataAPI.updateApp(app.id, { navigation_items: updatedItems });
                await loadApp(appId);
            } else {
                updateApp({ navigation_items: updatedItems });
            }

            setEditor({ mode: 'dashboard', dashboardId: newDashboardId });
            return true;
        } catch (error) {
            Logger.warn('AppStudio: Failed to add dashboard:', error instanceof Error ? error.message : error);
            errorToast('Failed to add dashboard. Please try again.');
            return false;
        }
    };

    const handleAddWebLink = async (def: { label: string; url: string; icon: string }) => {
        if (!app.id && !isNewApp) return false; // Should have app ID unless new

        try {
            const newLink: NavigationItem = {
                id: crypto.randomUUID(),
                type: 'web',
                label: def.label,
                page_url: def.url,
                icon: def.icon,
            };
            const updatedItems = [...(app.navigation_items || []), newLink];

            if (!isNewApp) {
                await metadataAPI.updateApp(app.id, { navigation_items: updatedItems });
                await loadApp(app.id);
            } else {
                updateApp({ navigation_items: updatedItems });
            }
            return true;
        } catch (error) {
            Logger.warn('AppStudio: Failed to add web link:', error instanceof Error ? error.message : error);
            errorToast('Failed to add web link. Please try again.');
            return false;
        }
    };

    const handleReorderNavItems = (items: NavigationItem[]) => {
        updateApp({ navigation_items: items });
        // If not new app, auto-save order
        if (!isNewApp) {
            metadataAPI.updateApp(app.id, { navigation_items: items }).catch((error) => {
                Logger.warn('AppStudio: Failed to save nav order:', error instanceof Error ? error.message : error);
                errorToast('Failed to save navigation order. Please try again.');
            });
        }
    };

    const handleRemoveNavItem = (itemId: string) => {
        const updatedItems = (app.navigation_items || []).filter(item => item.id !== itemId);
        updateApp({ navigation_items: updatedItems });

        if (!isNewApp) {
            metadataAPI.updateApp(app.id, { navigation_items: updatedItems }).catch((error) => {
                Logger.warn('AppStudio: Failed to remove nav item:', error instanceof Error ? error.message : error);
                errorToast('Failed to remove item. Please try again.');
            });
        }

        // Clear editor if removed item was selected
        if (editor.mode === 'object') {
            const removedItem = app.navigation_items?.find(item => item.id === itemId);
            if (removedItem?.object_api_name === editor.objectApiName) {
                setEditor({ mode: null });
            }
        } else if (editor.mode === 'dashboard') {
            const removedItem = app.navigation_items?.find(item => item.id === itemId);
            if (removedItem?.dashboard_id === editor.dashboardId) {
                setEditor({ mode: null });
            }
        }
    };

    if (loading) {
        return (
            <div className="h-screen flex items-center justify-center bg-slate-50">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (!app && !isNewApp) {
        return (
            <div className="h-screen flex items-center justify-center bg-slate-50">
                <div className="text-slate-500">App not found</div>
            </div>
        );
    }

    return (
        <div className="h-screen flex flex-col bg-slate-50 overflow-hidden">
            {/* Top Bar */}
            <header className="bg-white border-b border-slate-200 h-14 flex items-center justify-between px-4 flex-shrink-0 z-10">
                <div className="flex items-center gap-4">
                    <button onClick={() => navigate('/setup/appmanager')} className="p-2 hover:bg-slate-100 rounded-lg text-slate-500">
                        <ArrowLeft size={20} />
                    </button>
                    <div className="flex items-center gap-3">
                        <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${getColorClasses(app.color)}`}>
                            <Box size={18} />
                        </div>
                        <div>
                            <h1 className="font-semibold text-slate-800">{app.label || 'New App'}</h1>
                            <div className="text-xs text-slate-500">App Studio</div>
                        </div>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    {hasChanges && (
                        <span className="text-xs text-amber-600 bg-amber-50 px-2 py-1 rounded border border-amber-200">
                            Unsaved changes
                        </span>
                    )}
                    {!isNewApp && (
                        <button
                            onClick={() => window.open(`/app/${app.id}`, '_blank')}
                            className="flex items-center gap-2 px-3 py-1.5 text-slate-600 hover:bg-slate-100 rounded-lg text-sm font-medium transition-colors"
                        >
                            <ExternalLink size={16} />
                            Launch App
                        </button>
                    )}
                    <button
                        onClick={() => handleSaveApp(app)}
                        disabled={saving || (!hasChanges && !isNewApp)}
                        className="flex items-center gap-2 px-4 py-1.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm font-medium shadow-sm"
                    >
                        <Save size={16} />
                        {saving ? 'Saving...' : 'Save'}
                    </button>
                </div>
            </header>

            <div className="flex-1 flex overflow-hidden">
                {/* Left Sidebar */}
                <StudioSidebar
                    app={app}
                    openAddModal={shouldOpenAddModal}
                    selectedObjectApiName={editor.mode === 'object' ? editor.objectApiName : (editor.mode === 'dashboard' ? editor.dashboardId : undefined)}
                    onSelectNavItem={handleNavItemSelect}
                    onAddObject={handleAddObject}
                    onAddDashboard={handleAddDashboard}
                    onAddWebLink={handleAddWebLink}
                    onReorderNavItems={handleReorderNavItems}
                    onRemoveNavItem={handleRemoveNavItem}
                    onOpenSettings={() => setEditor({ mode: 'settings' })}
                    onOpenPermissions={() => setEditor({ mode: 'permissions' })}
                />

                {/* Main Content Area */}
                <main className="flex-1 p-6 overflow-y-auto relative">
                    {/* Editor Panels */}
                    {editor.mode === 'object' && editor.objectApiName && (
                        <div className="h-full">
                            <StudioObjectEditor
                                key={editor.objectApiName} // Force remount on change
                                objectApiName={editor.objectApiName}
                                onObjectUpdated={() => !isNewApp && loadApp(app.id)}
                            />
                        </div>
                    )}

                    {editor.mode === 'dashboard' && editor.dashboardId && (
                        <div className="h-full">
                            <StudioDashboardEditor
                                key={editor.dashboardId}
                                dashboardId={editor.dashboardId}
                                onDashboardUpdated={() => !isNewApp && loadApp(app.id)}
                            />
                        </div>
                    )}

                    {editor.mode === 'settings' && (
                        <AppSettingsEditor
                            app={app}
                            isNewApp={isNewApp}
                            onChange={updateApp}
                        />
                    )}

                    {editor.mode === 'permissions' && (
                        <div className="h-full">
                            <StudioPermissions appObjects={availableObjects} />
                        </div>
                    )}

                    {/* Empty State */}
                    {!editor.mode && (
                        <div className="h-full flex flex-col items-center justify-center text-slate-400">
                            <Box size={64} className="mb-4 text-slate-300" />
                            <h3 className="text-lg font-medium text-slate-600">Welcome to App Studio</h3>
                            <p>Select an item from the sidebar to edit</p>
                        </div>
                    )}
                </main>
            </div>
        </div>
    );
};
