import React, { useState } from 'react';
import { Plus, GripVertical, Trash2, Database, LayoutDashboard, Globe, Settings, Users, ChevronDown } from 'lucide-react';
import * as Icons from 'lucide-react';
import type { AppConfig, NavigationItem, ObjectMetadata } from '../../types';
import { AddPageModal } from './AddPageModal';
import { NavigationItemPicker } from '../admin/app/NavigationItemPicker';

interface StudioSidebarProps {
    app: AppConfig;
    selectedObjectApiName?: string;
    availableObjects: ObjectMetadata[];
    availableDashboards: any[];
    onSelectNavItem: (item: NavigationItem) => void;
    onAddObject: (objectDef: { label: string; apiName: string; icon: string }) => Promise<boolean>;
    onAddExistingObject: (object: ObjectMetadata) => void;
    onReorderNavItems: (items: NavigationItem[]) => void;
    onRemoveNavItem: (itemId: string) => void;
    onAddDashboard: (dashboardDef: { label: string; icon: string }) => Promise<boolean>;
    onAddExistingDashboard: (dashboard: { id: string; label: string }) => void;
    onAddWebLink: (webLinkDef: { label: string; url: string; icon: string }) => Promise<boolean>;
    onOpenSettings: () => void;
    onOpenPermissions: () => void;
    openAddModal?: boolean;
}

export const StudioSidebar: React.FC<StudioSidebarProps> = ({
    app,
    selectedObjectApiName,
    availableObjects,
    availableDashboards,
    onSelectNavItem,
    onAddObject,
    onAddExistingObject,
    onAddDashboard,
    onAddExistingDashboard,
    onAddWebLink,
    onReorderNavItems,
    onRemoveNavItem,
    onOpenSettings,
    onOpenPermissions,
    openAddModal,
}) => {
    const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
    const [showPicker, setShowPicker] = useState(false);
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [isNavExpanded, setIsNavExpanded] = useState(true);
    const [objectSearch, setObjectSearch] = useState('');

    // Auto-open picker if requested via prop (instead of modal directly)
    React.useEffect(() => {
        if (openAddModal) {
            setShowPicker(true);
        }
    }, [openAddModal]);

    const handleDragStart = (index: number) => {
        setDraggedIndex(index);
    };

    const handleDragOver = (e: React.DragEvent, index: number) => {
        e.preventDefault();
        if (draggedIndex === null || draggedIndex === index) return;

        const items = [...(app.navigation_items || [])];
        const draggedItem = items[draggedIndex];
        items.splice(draggedIndex, 1);
        items.splice(index, 0, draggedItem);

        onReorderNavItems(items);
        setDraggedIndex(index);
    };

    const handleDragEnd = () => {
        setDraggedIndex(null);
    };

    const getItemIcon = (item: NavigationItem) => {
        if (item.type === 'dashboard') return LayoutDashboard;
        if (item.type === 'web') return Globe;
        const IconComponent = Icons[item.icon as keyof typeof Icons] as React.ComponentType<{ size?: number | string; className?: string }>;
        return IconComponent || Database;
    };

    const handleAddObject = async (objectDef: { label: string; apiName: string; icon: string }) => {
        const success = await onAddObject(objectDef);
        if (success) {
            setShowCreateModal(false);
        }
        return success;
    };

    const handleAddDashboard = async (dashboardDef: { label: string; icon: string }) => {
        const success = await onAddDashboard(dashboardDef);
        if (success) {
            setShowCreateModal(false);
        }
        return success;
    };

    const handleAddWebLink = async (webLinkDef: { label: string; url: string; icon: string }) => {
        const success = await onAddWebLink(webLinkDef);
        if (success) {
            // Note: NavigationItemPicker handles web links directly now, but AddPageModal also supports it.
            // If we use AddPageModal for web links, we close it here.
            // If NavigationItemPicker calls it, it handles its own close.
            setShowCreateModal(false);
            setShowPicker(false);
        }
        return success;
    };

    return (
        <aside className="w-64 bg-white border-r border-slate-200 flex flex-col flex-shrink-0">
            {/* Navigation Section */}
            <div className="flex-1 overflow-y-auto">
                {/* Navigation Header */}
                <button
                    onClick={() => setIsNavExpanded(!isNavExpanded)}
                    className="w-full flex items-center justify-between px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider hover:bg-slate-50"
                >
                    <span>Navigation</span>
                    <ChevronDown
                        size={14}
                        className={`transition-transform ${isNavExpanded ? '' : '-rotate-90'}`}
                    />
                </button>

                {isNavExpanded && (
                    <div className="px-2 pb-2">
                        {/* Navigation Items */}
                        <div className="space-y-0.5">
                            {(app.navigation_items || []).map((item, index) => {
                                const ItemIcon = getItemIcon(item);
                                const isSelected = (item.type === 'object' && item.object_api_name === selectedObjectApiName) ||
                                    (item.type === 'dashboard' && item.dashboard_id === selectedObjectApiName);

                                return (
                                    <div
                                        key={item.id}
                                        draggable
                                        onDragStart={() => handleDragStart(index)}
                                        onDragOver={(e) => handleDragOver(e, index)}
                                        onDragEnd={handleDragEnd}
                                        onClick={() => onSelectNavItem(item)}
                                        className={`group flex items-center gap-2 px-2 py-2 rounded-lg cursor-pointer transition-all ${isSelected
                                            ? 'bg-blue-50 text-blue-700 border border-blue-200'
                                            : 'hover:bg-slate-50 text-slate-700'
                                            } ${draggedIndex === index ? 'opacity-50' : ''}`}
                                    >
                                        <GripVertical
                                            size={14}
                                            className="text-slate-300 opacity-0 group-hover:opacity-100 cursor-grab active:cursor-grabbing flex-shrink-0"
                                        />
                                        <div className={`w-7 h-7 rounded-md flex items-center justify-center flex-shrink-0 ${isSelected ? 'bg-blue-100' : 'bg-slate-100'
                                            }`}>
                                            <ItemIcon size={14} className={isSelected ? 'text-blue-600' : 'text-slate-500'} />
                                        </div>
                                        <span className="flex-1 text-sm font-medium truncate">{item.label}</span>
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                onRemoveNavItem(item.id);
                                            }}
                                            className="p-1 text-slate-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity"
                                        >
                                            <Trash2 size={14} />
                                        </button>
                                    </div>
                                );
                            })}
                        </div>

                        {/* Add Page Button */}
                        <button
                            onClick={() => setShowPicker(true)}
                            className="w-full flex items-center gap-2 px-2 py-2 mt-2 text-blue-600 hover:bg-blue-50 rounded-lg text-sm font-medium transition-colors"
                        >
                            <div className="w-7 h-7 rounded-md bg-blue-100 flex items-center justify-center">
                                <Plus size={14} />
                            </div>
                            Add Navigation Item
                        </button>
                    </div>
                )}
            </div>

            {/* Bottom Settings Section */}
            <div className="border-t border-slate-200 p-2">
                <button
                    onClick={onOpenSettings}
                    className="w-full flex items-center gap-3 px-3 py-2 text-slate-600 hover:bg-slate-50 rounded-lg text-sm transition-colors"
                >
                    <Settings size={16} />
                    App Settings
                </button>
                <button
                    onClick={onOpenPermissions}
                    className="w-full flex items-center gap-3 px-3 py-2 text-slate-600 hover:bg-slate-50 rounded-lg text-sm transition-colors"
                >
                    <Users size={16} />
                    Permissions
                </button>
            </div>

            {/* Navigation Picker */}
            <NavigationItemPicker
                isOpen={showPicker}
                onClose={() => setShowPicker(false)}
                availableObjects={availableObjects}
                availableDashboards={availableDashboards}
                objectSearch={objectSearch}
                setObjectSearch={setObjectSearch}
                onAddObject={(obj) => {
                    onAddExistingObject(obj);
                    setShowPicker(false);
                }}
                onAddDashboard={(dashboard) => {
                    onAddExistingDashboard(dashboard);
                    setShowPicker(false);
                }}
                onAddWeb={(url, label, icon) => {
                    handleAddWebLink({ url, label, icon });
                    setShowPicker(false);
                }}
                onAddStandard={(type) => {
                    // Handle standard items if needed, mostly Home
                    // We can reuse onAddExistingDashboard logic or similar if we have a standard item handler
                    // For now, assuming standard items are handled or we map them to route changes
                    // Or we just add a "Home" link
                    handleAddWebLink({ url: '/', label: 'Home', icon: 'Home' }); // Placeholder for standard home
                }}
                onCreateNewObject={() => {
                    setShowPicker(false);
                    setShowCreateModal(true);
                }}
                onCreateNewDashboard={() => {
                    setShowPicker(false);
                    setShowCreateModal(true);
                }}
            />

            {/* Add Page Modal (Legacy/Creation) */}
            {showCreateModal && (
                <AddPageModal
                    onClose={() => setShowCreateModal(false)}
                    onAddObject={handleAddObject}
                    onAddDashboard={handleAddDashboard}
                    onAddWebLink={handleAddWebLink}
                />
            )}
        </aside>
    );
};
