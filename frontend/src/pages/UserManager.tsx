import React, { useState, useEffect } from 'react';
import { Users, Shield, Plus, Mail, User as UserIcon } from 'lucide-react';
import { usersAPI } from '../infrastructure/api/users';
import type { User, Profile } from '../types';
import { COMMON_FIELDS } from '../core/constants';
import { PermissionEditorModal } from '../components/modals/PermissionEditorModal';
import { UserPermissionSetsSection } from '../components/UserPermissionSetsSection';
import { EffectivePermissionsModal } from '../components/modals/EffectivePermissionsModal';
import { UserEditorModal } from '../components/modals/UserEditorModal';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';
import { useErrorToast } from '../components/ui/Toast';
import { RecordListSkeleton } from '../components/ui/LoadingSkeleton';
import { EmptyState } from '../components/ui/EmptyState';

export const UserManager: React.FC = () => {
    const errorToast = useErrorToast();
    const [activeTab, setActiveTab] = useState<'users' | 'profiles'>('users');
    const [users, setUsers] = useState<User[]>([]);
    const [profiles, setProfiles] = useState<Profile[]>([]);
    const [loading, setLoading] = useState(true);
    const [editingPermissions, setEditingPermissions] = useState<Profile | null>(null);
    const [creatingUser, setCreatingUser] = useState(false);
    const [editingUser, setEditingUser] = useState<User | null>(null);
    const [managingPermissionsUser, setManagingPermissionsUser] = useState<User | null>(null);
    const [viewingEffectivePermissionsUser, setViewingEffectivePermissionsUser] = useState<User | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [deletingUser, setDeletingUser] = useState<User | null>(null);

    useEffect(() => {
        loadData();
    }, [activeTab]);

    const loadData = async () => {
        setLoading(true);
        setError(null);
        try {
            if (activeTab === 'users') {
                const [fetchedUsers, fetchedProfiles] = await Promise.all([
                    usersAPI.getUsers(),
                    usersAPI.getProfiles()
                ]);
                setUsers(fetchedUsers);
                setProfiles(fetchedProfiles);
            } else {
                const fetchedProfiles = await usersAPI.getProfiles();
                setProfiles(fetchedProfiles);
            }
        } catch {
            setError('Failed to load data. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="max-w-7xl mx-auto p-6">
            <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-4">
                    <div className="p-3 bg-purple-100 rounded-lg">
                        <Users className="text-purple-600" size={32} />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-slate-800">Users & Profiles</h1>
                        <p className="text-slate-500">Manage user access and security settings.</p>
                    </div>
                </div>
                <button
                    onClick={() => {
                        if (activeTab === 'users') setCreatingUser(true);
                    }}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium shadow-sm transition-colors"
                >
                    <Plus size={18} />
                    {activeTab === 'users' ? 'Add User' : 'Create Profile'}
                </button>
            </div>

            <div className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden">
                <div className="border-b border-slate-200">
                    <nav className="flex -mb-px">
                        <button
                            onClick={() => setActiveTab('users')}
                            className={`px-6 py-4 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 ${activeTab === 'users'
                                ? 'border-blue-500 text-blue-600'
                                : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                                }`}
                        >
                            <UserIcon size={18} />
                            Users
                        </button>
                        <button
                            onClick={() => setActiveTab('profiles')}
                            className={`px-6 py-4 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 ${activeTab === 'profiles'
                                ? 'border-blue-500 text-blue-600'
                                : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                                }`}
                        >
                            <Shield size={18} />
                            Profiles
                        </button>
                    </nav>
                </div>

                <div className="p-6">
                    {loading ? (
                        <RecordListSkeleton rows={5} columns={6} />
                    ) : error ? (
                        <EmptyState
                            variant="error"
                            title="Error Loading Users"
                            description={error}
                            action={{ label: 'Retry', onClick: loadData }}
                        />
                    ) : activeTab === 'users' ? (
                        <div className="overflow-x-auto">
                            <table className="w-full text-left text-sm">
                                <thead className="bg-slate-50 border-b border-slate-200">
                                    <tr>
                                        <th className="px-6 py-3 font-semibold text-slate-700">Name</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700">Email</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700">Role</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700">Status</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700">Last Login</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700 text-right">Actions</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-slate-100">
                                    {users.map(user => (
                                        <tr key={user[COMMON_FIELDS.ID] as string} className="hover:bg-slate-50">
                                            <td className="px-6 py-4 font-medium text-slate-900">
                                                {user.name}
                                            </td>
                                            <td className="px-6 py-4 text-slate-600 flex items-center gap-2">
                                                <Mail size={14} className="text-slate-400" />
                                                {user.email}
                                            </td>
                                            <td className="px-6 py-4">
                                                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-50 text-purple-700 border border-purple-100">
                                                    {profiles.find(p => p.id === user.profile_id)?.name || user.profile_id}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4">
                                                {user.is_active ? (
                                                    <span className="text-green-600 flex items-center gap-1.5">
                                                        <span className="w-1.5 h-1.5 rounded-full bg-green-600" />
                                                        Active
                                                    </span>
                                                ) : (
                                                    <span className="text-slate-500 flex items-center gap-1.5">
                                                        <span className="w-1.5 h-1.5 rounded-full bg-slate-400" />
                                                        Inactive
                                                    </span>
                                                )}
                                            </td>
                                            <td className="px-6 py-4 text-slate-500">
                                                {user.last_login ? new Date(user.last_login).toLocaleString() : 'Never'}
                                            </td>
                                            <td className="px-6 py-4 text-right whitespace-nowrap">
                                                <button
                                                    onClick={() => setManagingPermissionsUser(user)}
                                                    className="text-sm font-medium text-purple-600 hover:text-purple-800 mr-3"
                                                >
                                                    Assign Perm Sets
                                                </button>
                                                <button
                                                    onClick={() => setViewingEffectivePermissionsUser(user)}
                                                    className="text-sm font-medium text-teal-600 hover:text-teal-800 mr-3"
                                                >
                                                    View Effective
                                                </button>
                                                <button
                                                    onClick={() => setEditingUser(user)}
                                                    className="text-sm font-medium text-blue-600 hover:text-blue-800 mr-3"
                                                >
                                                    Edit
                                                </button>
                                                <button
                                                    onClick={async () => {
                                                        try {
                                                            await usersAPI.updateUser(user[COMMON_FIELDS.ID] as string, { is_active: !user.is_active });
                                                            loadData();
                                                        } catch {
                                                            errorToast("Failed to update user status");
                                                        }
                                                    }}
                                                    className={`text-sm font-medium mr-3 ${user.is_active ? 'text-amber-600 hover:text-amber-800' : 'text-green-600 hover:text-green-800'}`}
                                                >
                                                    {user.is_active ? 'Deactivate' : 'Activate'}
                                                </button>
                                                <button
                                                    onClick={() => setDeletingUser(user)}
                                                    className="text-sm font-medium text-red-600 hover:text-red-800"
                                                >
                                                    Delete
                                                </button>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    ) : (
                        <div className="overflow-x-auto">
                            <table className="w-full text-left text-sm">
                                <thead className="bg-slate-50 border-b border-slate-200">
                                    <tr>
                                        <th className="px-6 py-3 font-semibold text-slate-700">Profile Name</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700">Description</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700">Type</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700 text-right">Actions</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-slate-100">
                                    {profiles.map(profile => (
                                        <tr key={profile[COMMON_FIELDS.ID] as string} className="hover:bg-slate-50">
                                            <td className="px-6 py-4 font-medium text-slate-900">
                                                {profile.name}
                                            </td>
                                            <td className="px-6 py-4 text-slate-600">
                                                {profile.description}
                                            </td>
                                            <td className="px-6 py-4">
                                                {profile.is_system ? (
                                                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-800 border border-slate-200">
                                                        Standard
                                                    </span>
                                                ) : (
                                                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-50 text-blue-700 border border-blue-100">
                                                        Custom
                                                    </span>
                                                )}
                                            </td>
                                            <td className="px-6 py-4 text-right">
                                                <div className="flex gap-2 justify-end items-center">
                                                    {profile.id === 'system_admin' ? (
                                                        <span className="text-sm text-slate-400 italic">
                                                            Full access (not editable)
                                                        </span>
                                                    ) : (
                                                        <button
                                                            onClick={() => setEditingPermissions(profile)}
                                                            className="text-sm font-medium text-blue-600 hover:text-blue-800"
                                                        >
                                                            Edit Permissions
                                                        </button>
                                                    )}
                                                </div>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            </div>
            {editingPermissions && (
                <PermissionEditorModal
                    entity={{ id: editingPermissions.id, name: editingPermissions.name, type: 'profile' }}
                    onClose={() => setEditingPermissions(null)}
                    onSave={() => {
                        setEditingPermissions(null);
                        loadData();
                    }}
                />
            )}

            {managingPermissionsUser && (
                <UserPermissionSetsSection
                    userId={managingPermissionsUser[COMMON_FIELDS.ID] as string}
                    userName={managingPermissionsUser.name}
                    isOpen={!!managingPermissionsUser}
                    onClose={() => setManagingPermissionsUser(null)}
                />
            )}

            {viewingEffectivePermissionsUser && (
                <EffectivePermissionsModal
                    user={viewingEffectivePermissionsUser}
                    onClose={() => setViewingEffectivePermissionsUser(null)}
                />
            )}

            {creatingUser && (
                <UserEditorModal
                    profiles={profiles}
                    onClose={() => setCreatingUser(false)}
                    onSave={loadData}
                />
            )}
            {editingUser && (
                <UserEditorModal
                    user={editingUser}
                    profiles={profiles}
                    onClose={() => setEditingUser(null)}
                    onSave={loadData}
                />
            )}
            <ConfirmationModal
                isOpen={!!deletingUser}
                onClose={() => setDeletingUser(null)}
                onConfirm={async () => {
                    if (!deletingUser) return;
                    try {
                        await usersAPI.deleteUser(deletingUser[COMMON_FIELDS.ID] as string);
                        loadData();
                    } catch {
                        errorToast("Failed to delete user");
                    } finally {
                        setDeletingUser(null);
                    }
                }}
                title="Delete User"
                message={`Are you sure you want to delete user ${deletingUser?.name}? This action cannot be undone.`}
                confirmLabel="Delete"
                variant="danger"
            />
        </div>
    );
};
