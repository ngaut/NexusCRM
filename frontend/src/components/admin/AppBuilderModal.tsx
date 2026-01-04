import React, { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { X, Plus, GripVertical, Trash2 } from 'lucide-react';
import * as Icons from 'lucide-react';
import { metadataAPI } from '../../infrastructure/api/metadata';
import { usersAPI } from '../../infrastructure/api/users';
import type { Profile } from '../../types';
import type { AppConfig, NavigationItem, ObjectMetadata } from '../../types';
import { NavigationItemPicker } from './app/NavigationItemPicker';
import { UtilityBarBuilder } from './app/UtilityBarBuilder';
import { AppFormBasicInfo } from './app/AppFormBasicInfo';
import { AppFormUserAccess } from './app/AppFormUserAccess';
import { AppIconPicker } from './app/AppIconPicker';

interface AppBuilderModalProps {
    app: AppConfig | null;
    onSave: () => void;
    onClose: () => void;
}

export const AppBuilderModal: React.FC<AppBuilderModalProps> = ({ app, onSave, onClose }) => {
    const [formData, setFormData] = useState<AppConfig>({
        id: '',
        label: '',
        description: '',
        icon: 'Layers',
        color: 'blue',
        navigation_items: [],
        utility_items: [],
        assigned_profiles: []
    });

    const [availableObjects, setAvailableObjects] = useState<ObjectMetadata[]>([]);
    const [availableProfiles, setAvailableProfiles] = useState<Profile[]>([]);
    const [isSaving, setIsSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [objectSearch, setObjectSearch] = useState('');
    const [showObjectPicker, setShowObjectPicker] = useState(false);
    const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
    const [visibleToAll, setVisibleToAll] = useState(true);
    const [availableDashboards, setAvailableDashboards] = useState<{ id: string; label: string; description?: string }[]>([]);

    useEffect(() => {
        if (app) {
            setFormData(app);
        }
        // Set visibleToAll based on assigned profiles
        if (app?.assigned_profiles && app.assigned_profiles.length > 0) {
            setVisibleToAll(false);
        } else {
            setVisibleToAll(true);
        }
        loadObjects();
        loadProfiles();
        loadDashboards();
    }, [app]);

    const loadObjects = async () => {
        try {
            const response = await metadataAPI.getSchemas();
            const objects = response.schemas || [];
            setAvailableObjects(objects);
        } catch {
            // Objects loading failure - handled via empty state
        }
    };

    const loadProfiles = async () => {
        try {
            const profiles = await usersAPI.getProfiles();
            setAvailableProfiles(profiles || []);
        } catch {
            // Profiles loading failure - handled via empty state
        }
    };

    const loadDashboards = async () => {
        try {
            const response = await metadataAPI.getDashboards();
            setAvailableDashboards(response.dashboards || []);
        } catch {
            // Dashboards loading failure - handled via empty state
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setIsSaving(true);

        try {
            // Prepare data - set profile assignment
            const dataToSave = {
                ...formData,
                assigned_profiles: visibleToAll ? [] : (formData.assigned_profiles || []),
            };

            if (app) {
                await metadataAPI.updateApp(app.id, dataToSave);
            } else {
                await metadataAPI.createApp(dataToSave);
            }
            onSave();
        } catch (err: unknown) {
            const msg = err instanceof Error ? err.message : 'An error occurred';
            if (err && typeof err === 'object' && 'response' in err) {
                const apiError = err as { response?: { data?: { error?: string } } };
                setError(apiError.response?.data?.error || msg);
            } else {
                setError(msg);
            }
        } finally {
            setIsSaving(false);
        }
    };

    const addNavigationItem = (obj: ObjectMetadata) => {
        const newItem: NavigationItem = {
            id: `nav-${obj.api_name}-${Date.now()}`,
            type: 'object',
            object_api_name: obj.api_name,
            label: obj.plural_label || obj.label,
            icon: obj.icon || 'Database',
        };
        setFormData(prev => ({
            ...prev,
            navigation_items: [...(prev.navigation_items || []), newItem],
        }));
        setShowObjectPicker(false);
        setObjectSearch('');
    };

    const addWebNavigationItem = (url: string, label: string, icon: string) => {
        const newItem: NavigationItem = {
            id: `nav-web-${Date.now()}`,
            type: 'web',
            page_url: url.startsWith('http') ? url : `https://${url}`,
            label: label,
            icon: icon || 'ExternalLink',
        };
        setFormData(prev => ({
            ...prev,
            navigation_items: [...(prev.navigation_items || []), newItem],
        }));
        setShowObjectPicker(false);
    };

    const addDashboardNavigationItem = (dashboard: { id: string; label: string }) => {
        const newItem: NavigationItem = {
            id: `nav-dashboard-${dashboard.id}`,
            type: 'dashboard',
            dashboard_id: dashboard.id,
            label: dashboard.label,
            icon: 'LayoutDashboard',
        };

        setFormData(prev => ({
            ...prev,
            navigation_items: [...(prev.navigation_items || []), newItem],
        }));
        setShowObjectPicker(false);
    };

    const removeNavigationItem = (id: string) => {
        setFormData(prev => ({
            ...prev,
            navigation_items: (prev.navigation_items || []).filter(item => item.id !== id),
        }));
    };

    const handleDragStart = (index: number) => {
        setDraggedIndex(index);
    };

    const handleDragOver = (e: React.DragEvent, index: number) => {
        e.preventDefault();
        if (draggedIndex === null || draggedIndex === index) return;

        const items = [...(formData.navigation_items || [])];
        const draggedItem = items[draggedIndex];
        items.splice(draggedIndex, 1);
        items.splice(index, 0, draggedItem);

        setFormData(prev => ({ ...prev, navigation_items: items }));
        setDraggedIndex(index);
    };

    const handleDragEnd = () => {
        setDraggedIndex(null);
    };

    return createPortal(
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[100] p-4">
            <div className="bg-white rounded-xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col">
                {/* Header */}
                <div className="flex items-center justify-between p-6 border-b bg-slate-50 rounded-t-xl">
                    <h2 className="text-xl font-bold text-slate-800">
                        {app ? 'Edit App' : 'Create New App'}
                    </h2>
                    <button onClick={onClose} className="text-slate-400 hover:text-slate-600">
                        <X size={20} />
                    </button>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto">
                    <div className="p-6 space-y-6">
                        {error && (
                            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
                                {error}
                            </div>
                        )}

                        <AppFormBasicInfo
                            formData={formData}
                            isEditMode={!!app}
                            onChange={(updates) => setFormData(prev => ({ ...prev, ...updates }))}
                        />

                        <AppIconPicker
                            icon={formData.icon || 'Layers'}
                            color={formData.color || 'blue'}
                            onChange={(updates) => setFormData(prev => ({ ...prev, ...updates }))}
                        />

                        {/* Navigation Items */}
                        <div>
                            <div className="flex items-center justify-between mb-2">
                                <label className="block text-sm font-medium text-slate-700">
                                    Navigation Items ({formData.navigation_items?.length || 0})
                                </label>
                                <button
                                    type="button"
                                    onClick={() => setShowObjectPicker(true)}
                                    className="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700 font-medium"
                                >
                                    <Plus size={16} />
                                    Add Item
                                </button>
                            </div>

                            {/* Navigation Items List */}
                            <div className="border rounded-lg overflow-hidden">
                                {(formData.navigation_items || []).length === 0 ? (
                                    <div className="px-4 py-8 text-center text-slate-500 text-sm bg-slate-50">
                                        No navigation items. Click "Add Item" to add objects to this app.
                                    </div>
                                ) : (
                                    <div className="divide-y">
                                        {(formData.navigation_items || []).map((item, index) => {
                                            const ItemIcon = Icons[item.icon as keyof typeof Icons] as React.ComponentType<{ size?: number; className?: string }> || Icons.Database;
                                            return (
                                                <div
                                                    key={item.id}
                                                    draggable
                                                    onDragStart={() => handleDragStart(index)}
                                                    onDragOver={(e) => handleDragOver(e, index)}
                                                    onDragEnd={handleDragEnd}
                                                    className={`flex items-center gap-3 px-3 py-2.5 bg-white hover:bg-slate-50 cursor-grab active:cursor-grabbing transition-colors ${draggedIndex === index ? 'bg-blue-50 border-blue-200' : ''
                                                        }`}
                                                >
                                                    <GripVertical size={16} className="text-slate-400 flex-shrink-0" />
                                                    <div className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center flex-shrink-0">
                                                        <ItemIcon size={16} className="text-slate-600" />
                                                    </div>
                                                    <div className="flex-1 min-w-0">
                                                        <div className="font-medium text-slate-800 text-sm truncate">{item.label}</div>
                                                        <div className="text-xs text-slate-500 truncate">{item.object_api_name || item.page_url || 'Dashboard'}</div>
                                                    </div>
                                                    <button
                                                        type="button"
                                                        onClick={() => removeNavigationItem(item.id)}
                                                        className="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors flex-shrink-0"
                                                    >
                                                        <Trash2 size={14} />
                                                    </button>
                                                </div>
                                            );
                                        })}
                                    </div>
                                )}
                            </div>
                            <p className="text-xs text-slate-500 mt-1">Drag items to reorder. Items appear in the left navigation when this app is selected.</p>
                        </div>

                        <AppFormUserAccess
                            assignedProfiles={formData.assigned_profiles || []}
                            availableProfiles={availableProfiles}
                            visibleToAll={visibleToAll}
                            onChangeVisibleToAll={(val) => {
                                setVisibleToAll(val);
                                if (val) setFormData(prev => ({ ...prev, assigned_profiles: [] }));
                            }}
                            onChangeProfiles={(profiles) => setFormData(prev => ({ ...prev, assigned_profiles: profiles }))}
                        />

                        <UtilityBarBuilder
                            utilityItems={formData.utility_items || []}
                            onChange={(items) => setFormData(prev => ({ ...prev, utility_items: items }))}
                        />
                    </div>

                    {/* Footer */}
                    <div className="flex justify-end gap-3 p-6 border-t bg-slate-50">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-slate-700 hover:bg-slate-200 rounded-lg font-medium"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={isSaving}
                            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 font-medium"
                        >
                            {isSaving ? 'Saving...' : app ? 'Update App' : 'Create App'}
                        </button>
                    </div>
                </form>
            </div>

            {/* Navigation Picker Component */}
            <NavigationItemPicker
                isOpen={showObjectPicker}
                onClose={() => setShowObjectPicker(false)}
                availableObjects={availableObjects.filter(obj => !(formData.navigation_items || []).some(item => item.object_api_name === obj.api_name))}
                availableDashboards={availableDashboards}
                objectSearch={objectSearch}
                setObjectSearch={setObjectSearch}
                onAddObject={addNavigationItem}
                onAddWeb={addWebNavigationItem}
                onAddDashboard={addDashboardNavigationItem}
                onAddStandard={() => {
                    const homeItem: NavigationItem = {
                        id: `nav-home-${Date.now()}`,
                        type: 'page',
                        label: 'Home',
                        icon: 'Home',
                    };
                    setFormData(prev => ({
                        ...prev,
                        navigation_items: [...(prev.navigation_items || []), homeItem],
                    }));
                    setShowObjectPicker(false);
                }}
            />
        </div >,
        document.body
    );
};
