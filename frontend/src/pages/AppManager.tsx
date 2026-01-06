import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Plus, Edit, Trash2, Layers, Sparkles } from 'lucide-react';
import * as Icons from 'lucide-react';
import { metadataAPI } from '../infrastructure/api/metadata';
import type { AppConfig } from '../types';
import { AppBuilderModal } from '../components/admin/AppBuilderModal';
import { getColorClasses } from '../core/utils/colorClasses';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';
import { useErrorToast } from '../components/ui/Toast';
import { formatApiError } from '../core/utils/errorHandling';

export const AppManager: React.FC = () => {
    const errorToast = useErrorToast();
    const navigate = useNavigate();
    const [apps, setApps] = useState<AppConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingApp, setEditingApp] = useState<AppConfig | null>(null);

    // Delete confirmation state
    const [deleteModalOpen, setDeleteModalOpen] = useState(false);
    const [appToDelete, setAppToDelete] = useState<AppConfig | null>(null);
    const [deleting, setDeleting] = useState(false);

    const loadApps = async () => {
        try {
            const response = await metadataAPI.getApps();
            setApps(response.apps || []);
        } catch {
            // Apps loading failure - handled via empty state
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadApps();
    }, []);

    const handleDelete = (app: AppConfig) => {
        setAppToDelete(app);
        setDeleteModalOpen(true);
    };

    const confirmDelete = async () => {
        if (!appToDelete) return;
        setDeleting(true);
        try {
            await metadataAPI.deleteApp(appToDelete.id);
            setApps(apps.filter(app => app.id !== appToDelete.id));
            setDeleteModalOpen(false);
            setAppToDelete(null);
        } catch (error: unknown) {
            errorToast(`Failed to delete app: ${formatApiError(error).message}`);
        } finally {
            setDeleting(false);
        }
    };

    const handleEdit = (app: AppConfig) => {
        setEditingApp(app);
        setIsModalOpen(true);
    };

    const handleCreate = () => {
        setEditingApp(null);
        setIsModalOpen(true);
    };

    const handleSave = () => {
        setIsModalOpen(false);
        setEditingApp(null);
        loadApps();
    };

    const getIconComponent = (iconName: string) => {
        const IconComponent = Icons[iconName as keyof typeof Icons] as React.ComponentType<{ size?: number | string; className?: string }>;
        return IconComponent || Icons.Layers;
    };

    if (loading) {
        return <div className="p-8">Loading apps...</div>;
    }

    return (
        <div className="p-6">
            <div className="flex justify-between items-center mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-slate-800">App Manager</h1>
                    <p className="text-slate-600 mt-1">Create and manage application configurations</p>
                </div>
                <button
                    onClick={() => navigate('/studio/new')}
                    className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-lg hover:from-blue-700 hover:to-indigo-700 shadow-lg shadow-blue-500/25"
                >
                    <Sparkles size={18} />
                    New App (Studio)
                </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {apps.map(app => {
                    const IconComponent = getIconComponent(app.icon);
                    const colors = getColorClasses(app.color);
                    return (
                        <div
                            key={app.id}
                            className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-2xl p-6 hover:shadow-xl transition-all group"
                        >
                            <div className="flex items-start justify-between mb-4">
                                <div className={`w-12 h-12 ${colors.bg} rounded-lg flex items-center justify-center ${colors.text}`}>
                                    <IconComponent size={24} />
                                </div>
                                <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                    <button
                                        onClick={() => handleEdit(app)}
                                        className="p-2 text-slate-400 hover:text-blue-600 hover:bg-blue-50 rounded"
                                    >
                                        <Edit size={16} />
                                    </button>
                                    <button
                                        onClick={() => handleDelete(app)}
                                        className="p-2 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded"
                                    >
                                        <Trash2 size={16} />
                                    </button>
                                </div>
                            </div>

                            <h3 className="text-lg font-semibold text-slate-800 mb-1">{app.label}</h3>
                            <p className="text-sm text-slate-500 mb-3">{app.description}</p>

                            <div className="flex items-center justify-between">
                                <div className="text-xs text-slate-400">
                                    {app.navigation_items?.length || 0} navigation items
                                </div>
                                <Link
                                    to={`/studio/${app.id}`}
                                    className="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-700 font-medium"
                                >
                                    <Sparkles size={12} />
                                    Open in Studio
                                </Link>
                            </div>
                        </div>
                    );
                })}
            </div>

            {apps.length === 0 && (
                <div className="text-center py-12">
                    <Layers className="mx-auto text-slate-300 mb-3" size={48} />
                    <p className="text-slate-500">No apps found. Create your first app to get started.</p>
                </div>
            )}

            {isModalOpen && (
                <AppBuilderModal
                    app={editingApp}
                    onSave={handleSave}
                    onClose={() => {
                        setIsModalOpen(false);
                        setEditingApp(null);
                    }}
                />
            )}

            {/* Delete App Confirmation Modal */}
            <ConfirmationModal
                isOpen={deleteModalOpen}
                onClose={() => {
                    setDeleteModalOpen(false);
                    setAppToDelete(null);
                }}
                onConfirm={confirmDelete}
                title="Delete App"
                message={`Are you sure you want to delete the app "${appToDelete?.label}"?`}
                confirmLabel="Delete"
                cancelLabel="Cancel"
                variant="danger"
                loading={deleting}
            />
        </div>
    );
};
