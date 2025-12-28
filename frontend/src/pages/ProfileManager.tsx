import React, { useState, useEffect } from 'react';
import { Users, Plus, AlertCircle, Edit, Trash2, X, Save, Shield } from 'lucide-react';
import { dataAPI } from '../infrastructure/api/data';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';

interface Profile {
    id: string;
    name: string;
    description?: string;
    is_system?: boolean;
    created_date?: string;
    last_modified_date?: string;
}

const ProfileManager: React.FC = () => {
    const [profiles, setProfiles] = useState<Profile[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [showModal, setShowModal] = useState(false);
    const [editingProfile, setEditingProfile] = useState<Profile | null>(null);

    // Form state
    const [formData, setFormData] = useState<Partial<Profile>>({});
    const [saving, setSaving] = useState(false);

    // Delete state
    const [deleteModalOpen, setDeleteModalOpen] = useState(false);
    const [profileToDelete, setProfileToDelete] = useState<Profile | null>(null);

    // Load profiles
    const loadProfiles = async () => {
        try {
            setLoading(true);
            const records = await dataAPI.query<Profile>({
                objectApiName: '_System_Profile',
                sortField: 'name',
                sortDirection: 'ASC'
            });
            setProfiles(records);
            setError(null);
        } catch (err) {
            setError('Failed to load profiles: ' + (err instanceof Error ? err.message : 'Unknown error'));
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadProfiles();
    }, []);

    const handleCreate = () => {
        setEditingProfile(null);
        setFormData({});
        setShowModal(true);
    };

    const handleEdit = (profile: Profile) => {
        setEditingProfile(profile);
        setFormData({ ...profile });
        setShowModal(true);
    };

    const handleDelete = (profile: Profile) => {
        if (profile.is_system) {
            setError('Cannot delete system profiles');
            return;
        }
        setProfileToDelete(profile);
        setDeleteModalOpen(true);
    };

    const confirmDelete = async () => {
        if (!profileToDelete) return;
        try {
            await dataAPI.deleteRecord('_System_Profile', profileToDelete.id);
            setDeleteModalOpen(false);
            setProfileToDelete(null);
            loadProfiles();
        } catch (err) {
            setError('Failed to delete profile: ' + (err instanceof Error ? err.message : 'Unknown error'));
        }
    };

    const handleSave = async () => {
        if (!formData.name) {
            setError('Profile Name is required');
            return;
        }

        try {
            setSaving(true);
            setError(null);

            if (editingProfile) {
                await dataAPI.updateRecord('_System_Profile', editingProfile.id, formData);
            } else {
                await dataAPI.createRecord<Profile>('_System_Profile', formData);
            }

            setShowModal(false);
            loadProfiles();
        } catch (err) {
            setError('Failed to save profile: ' + (err instanceof Error ? err.message : 'Unknown error'));
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="p-6 max-w-7xl mx-auto">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-xl shadow-lg">
                        <Users className="w-6 h-6 text-white" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Profiles</h1>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                            Manage user profiles and access levels
                        </p>
                    </div>
                </div>
                <button
                    onClick={handleCreate}
                    className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-blue-500 to-indigo-600 
                    text-white rounded-lg hover:from-blue-600 hover:to-indigo-700 transition-all shadow-md"
                >
                    <Plus className="w-4 h-4" />
                    New Profile
                </button>
            </div>

            {/* Error */}
            {error && (
                <div className="mb-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 
                rounded-lg flex items-center gap-2 text-red-700 dark:text-red-400">
                    <AlertCircle className="w-5 h-5" />
                    {error}
                </div>
            )}

            {/* Content */}
            {loading ? (
                <div className="flex items-center justify-center py-12">
                    <div className="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
                </div>
            ) : profiles.length === 0 ? (
                <div className="text-center py-12 bg-gray-50 dark:bg-gray-800/50 rounded-xl border border-dashed 
                border-gray-200 dark:border-gray-700">
                    <Users className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                    <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">No Profiles Found</h3>
                    <p className="text-gray-500 dark:text-gray-400 mb-4">
                        Get started by creating a new profile
                    </p>
                    <button
                        onClick={handleCreate}
                        className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
                    >
                        Create Profile
                    </button>
                </div>
            ) : (
                <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
                    <table className="w-full">
                        <thead className="bg-gray-50 dark:bg-gray-900/50">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Profile Name</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Description</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">System</th>
                                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                            {profiles.map((profile) => (
                                <tr key={profile.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors">
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <div className="flex items-center gap-3">
                                            <div className="p-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                                                <BadgeCheck isSystem={profile.is_system} />
                                            </div>
                                            <div className="text-sm font-medium text-gray-900 dark:text-white">{profile.name}</div>
                                        </div>
                                    </td>
                                    <td className="px-6 py-4">
                                        <div className="text-sm text-gray-500 dark:text-gray-400 truncate max-w-md">
                                            {profile.description || '-'}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        {profile.is_system ? (
                                            <span className="px-2 py-1 text-xs font-medium bg-purple-100 dark:bg-purple-900/30 text-purple-800 dark:text-purple-300 rounded-full">
                                                System
                                            </span>
                                        ) : (
                                            <span className="px-2 py-1 text-xs font-medium bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-300 rounded-full">
                                                Custom
                                            </span>
                                        )}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                        <div className="flex items-center justify-end gap-2">
                                            <button
                                                onClick={() => handleEdit(profile)}
                                                className="p-1 text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
                                            >
                                                <Edit className="w-4 h-4" />
                                            </button>
                                            <button
                                                onClick={() => handleDelete(profile)}
                                                disabled={!!profile.is_system}
                                                className={`p-1 ${profile.is_system
                                                    ? 'text-gray-300 cursor-not-allowed dark:text-gray-600'
                                                    : 'text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300'}`}
                                            >
                                                <Trash2 className="w-4 h-4" />
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}

            {/* Create/Edit Modal */}
            {showModal && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
                    <div className="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-lg p-6">
                        <div className="flex items-center justify-between mb-6">
                            <h2 className="text-xl font-bold text-gray-900 dark:text-white">
                                {editingProfile ? 'Edit Profile' : 'New Profile'}
                            </h2>
                            <button onClick={() => setShowModal(false)} className="text-gray-400 hover:text-gray-600">
                                <X className="w-5 h-5" />
                            </button>
                        </div>

                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Profile Name *</label>
                                <input
                                    type="text"
                                    value={formData.name || ''}
                                    onChange={e => setFormData({ ...formData, name: e.target.value })}
                                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600"
                                    placeholder="e.g. Sales Manager"
                                    disabled={!!editingProfile?.is_system} // Cannot rename system profiles usually
                                />
                                {editingProfile?.is_system && (
                                    <p className="text-xs text-amber-600 mt-1">System profile names cannot be changed.</p>
                                )}
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
                                <textarea
                                    value={formData.description || ''}
                                    onChange={e => setFormData({ ...formData, description: e.target.value })}
                                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600"
                                    rows={3}
                                    placeholder="Describe the role and access level for this profile..."
                                />
                            </div>
                        </div>

                        <div className="flex justify-end gap-3 mt-6">
                            <button
                                onClick={() => setShowModal(false)}
                                className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleSave}
                                disabled={saving}
                                className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                            >
                                <Save className="w-4 h-4" />
                                {saving ? 'Saving...' : 'Save'}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            <ConfirmationModal
                isOpen={deleteModalOpen}
                onClose={() => setDeleteModalOpen(false)}
                onConfirm={confirmDelete}
                title="Delete Profile"
                message={`Are you sure you want to delete "${profileToDelete?.name}"? Users assigned to this profile may lose access.`}
                confirmLabel="Delete"
                variant="danger"
            />
        </div>
    );
};

// Helper component
const BadgeCheck: React.FC<{ isSystem?: boolean }> = ({ isSystem }) => {
    if (isSystem) return <Shield className="w-4 h-4 text-purple-600 dark:text-purple-400" />;
    return <Users className="w-4 h-4 text-blue-600 dark:text-blue-400" />;
};

export default ProfileManager;
