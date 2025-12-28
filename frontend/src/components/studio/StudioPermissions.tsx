import React, { useState, useEffect } from 'react';
import { Shield, Save, AlertCircle, Check } from 'lucide-react';
import { usersAPI } from '../../infrastructure/api/users';
import type { ObjectMetadata, Profile, ObjectPermission } from '../../types';

interface StudioPermissionsProps {
    appObjects: ObjectMetadata[]; // Objects associated with this app
}

export const StudioPermissions: React.FC<StudioPermissionsProps> = ({ appObjects }) => {
    const [profiles, setProfiles] = useState<Profile[]>([]);
    const [permissions, setPermissions] = useState<Record<string, ObjectPermission[]>>({}); // profileId -> perms[]
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [hasChanges, setHasChanges] = useState(false);

    useEffect(() => {
        loadData();
    }, []);

    const loadData = async () => {
        setLoading(true);
        setError(null);
        try {
            // 1. Load Profiles
            const loadedProfiles = await usersAPI.getProfiles();
            setProfiles(loadedProfiles);

            // 2. Load Permissions for each profile
            const permsMap: Record<string, ObjectPermission[]> = {};
            await Promise.all(loadedProfiles.map(async (p: Profile) => {
                const { permissions: perms } = await usersAPI.getProfilePermissions(p.id);
                permsMap[p.id] = perms;
            }));

            setPermissions(permsMap);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load permissions');
        } finally {
            setLoading(false);
        }
    };

    const getPermission = (profileId: string, objectApiName: string): ObjectPermission => {
        const profilePerms = permissions[profileId] || [];
        const existing = profilePerms.find(p => p.object_api_name === objectApiName);
        if (existing) return existing;

        // Default empty permission if not found
        return {
            profile_id: profileId,
            object_api_name: objectApiName,
            allow_read: false,
            allow_create: false,
            allow_edit: false,
            allow_delete: false,
            view_all: false,
            modify_all: false
        };
    };

    const handleToggle = (profileId: string, objectApiName: string, field: keyof ObjectPermission) => {
        // We need to work with the raw Permission object structure
        const currentPerms = [...(permissions[profileId] || [])];
        let permIndex = currentPerms.findIndex(p => p.object_api_name === objectApiName);

        // Create if doesn't exist
        if (permIndex === -1) {
            const newPerm: ObjectPermission = {
                profile_id: profileId,
                object_api_name: objectApiName,
                allow_read: false,
                allow_create: false,
                allow_edit: false,
                allow_delete: false,
                view_all: false,
                modify_all: false
            };
            currentPerms.push(newPerm);
            permIndex = currentPerms.length - 1;
        }

        const perm = { ...currentPerms[permIndex] };

        // Toggle the boolean field
        // Toggle the boolean field
        if (typeof perm[field] === 'boolean') {
            (perm as unknown as Record<string, boolean>)[field] = !perm[field];
        }

        // Logic: if Modify All -> View All, Edit, Read, Delete, Create
        if (field === 'modify_all' && perm.modify_all) {
            perm.view_all = true;
            perm.allow_edit = true;
            perm.allow_delete = true;
            perm.allow_create = true;
            perm.allow_read = true;
        }
        // Logic: if View All -> Read
        if (field === 'view_all' && perm.view_all) {
            perm.allow_read = true;
        }
        // Logic: Edit/Create/Delete implies Read
        if ((field === 'allow_edit' || field === 'allow_create' || field === 'allow_delete') && perm[field]) {
            perm.allow_read = true;
        }

        currentPerms[permIndex] = perm;
        setPermissions({
            ...permissions,
            [profileId]: currentPerms
        });
        setHasChanges(true);
    };

    const handleSave = async () => {
        setSaving(true);
        setError(null);
        try {
            // Save each profile's permissions
            await Promise.all(Object.entries(permissions).map(async ([profileId, perms]) => {
                // Only send perms for displayed objects (optimization)?
                // Or properly just send the changed ones?
                // For now, send all perms for the profile to be safe/simple
                await usersAPI.updateProfilePermissions(profileId, perms);
            }));
            setHasChanges(false);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to save');
        } finally {
            setSaving(false);
        }
    };

    if (loading) return <div className="p-8 text-center text-slate-500">Loading permissions...</div>;

    if (appObjects.length === 0) {
        return (
            <div className="bg-white rounded-xl border border-slate-200 p-8 text-center">
                <Shield className="mx-auto text-slate-300 mb-3" size={48} />
                <h3 className="text-lg font-medium text-slate-700">No Objects Found</h3>
                <p className="text-slate-500">Add objects to your app to configure permissions.</p>
            </div>
        );
    }

    return (
        <div className="bg-white rounded-xl border border-slate-200 flex flex-col h-full shadow-sm">
            {/* Header */}
            <div className="px-6 py-4 border-b border-slate-200 flex justify-between items-center bg-slate-50/50 rounded-t-xl">
                <div>
                    <h2 className="text-lg font-bold text-slate-800 flex items-center gap-2">
                        <Shield className="text-blue-600" size={20} />
                        Permissions Matrix
                    </h2>
                    <p className="text-sm text-slate-500">Configure object access for this App</p>
                </div>
                <button
                    onClick={handleSave}
                    disabled={!hasChanges || saving}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 font-medium transition-all shadow-sm"
                >
                    {saving ? (
                        <>
                            <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                            Saving...
                        </>
                    ) : (
                        <>
                            <Save size={16} />
                            Save Changes
                        </>
                    )}
                </button>
            </div>

            {error && (
                <div className="bg-red-50 border-b border-red-100 text-red-700 px-6 py-3 text-sm flex items-center gap-2">
                    <AlertCircle size={16} />
                    {error}
                </div>
            )}

            {!error && !hasChanges && (
                <div className="bg-blue-50/50 border-b border-blue-100 text-blue-700 px-6 py-3 text-sm flex items-center gap-2">
                    <AlertCircle size={16} />
                    Changes are saved per profile. Remember to save after edits.
                </div>
            )}

            {/* Matrix */}
            <div className="flex-1 overflow-auto">
                <table className="w-full text-left border-collapse">
                    <thead className="bg-slate-50 sticky top-0 z-10 shadow-sm">
                        <tr>
                            <th className="px-6 py-3 border-b border-slate-200 font-semibold text-slate-700 text-sm whitespace-nowrap bg-slate-50">
                                Object
                            </th>
                            {profiles.map(profile => (
                                <th key={profile.id} className="px-6 py-3 border-b border-slate-200 font-semibold text-slate-700 text-sm whitespace-nowrap bg-slate-50 min-w-[200px]">
                                    {profile.name}
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                        {appObjects.map(obj => (
                            <tr key={obj.api_name} className="hover:bg-slate-50/50 group">
                                <td className="px-6 py-4 border-r border-slate-100 bg-white group-hover:bg-slate-50/50 sticky left-0 z-10 font-medium text-slate-800">
                                    <div className="flex items-center gap-2">
                                        {/* Icon placehoder */}
                                        <span>{obj.label}</span>
                                    </div>
                                    <div className="text-xs text-slate-400 font-mono mt-0.5">{obj.api_name}</div>
                                </td>
                                {profiles.map(profile => {
                                    const perm = getPermission(profile.id, obj.api_name);
                                    const p = perm;

                                    return (
                                        <td key={`${obj.api_name}-${profile.id}`} className="px-6 py-4 border-r border-slate-100 last:border-r-0">
                                            <div className="grid grid-cols-3 gap-2 text-xs">
                                                {/* Read */}
                                                <label className="flex items-center gap-1.5 cursor-pointer" title="Read Access">
                                                    <input
                                                        type="checkbox"
                                                        checked={!!p.allow_read}
                                                        onChange={() => handleToggle(profile.id, obj.api_name, 'allow_read')}
                                                        className="rounded text-blue-600 focus:ring-blue-500 border-slate-300"
                                                    />
                                                    <span>Read</span>
                                                </label>
                                                {/* Create */}
                                                <label className="flex items-center gap-1.5 cursor-pointer" title="Create Records">
                                                    <input
                                                        type="checkbox"
                                                        checked={!!p.allow_create}
                                                        onChange={() => handleToggle(profile.id, obj.api_name, 'allow_create')}
                                                        className="rounded text-blue-600 focus:ring-blue-500 border-slate-300"
                                                    />
                                                    <span>Create</span>
                                                </label>
                                                {/* Edit */}
                                                <label className="flex items-center gap-1.5 cursor-pointer" title="Edit Records">
                                                    <input
                                                        type="checkbox"
                                                        checked={!!p.allow_edit}
                                                        onChange={() => handleToggle(profile.id, obj.api_name, 'allow_edit')}
                                                        className="rounded text-blue-600 focus:ring-blue-500 border-slate-300"
                                                    />
                                                    <span>Edit</span>
                                                </label>
                                                {/* Delete */}
                                                <label className="flex items-center gap-1.5 cursor-pointer" title="Delete Records">
                                                    <input
                                                        type="checkbox"
                                                        checked={!!p.allow_delete}
                                                        onChange={() => handleToggle(profile.id, obj.api_name, 'allow_delete')}
                                                        className="rounded text-blue-600 focus:ring-blue-500 border-slate-300"
                                                    />
                                                    <span>Delete</span>
                                                </label>
                                                {/* View All */}
                                                <label className="flex items-center gap-1.5 cursor-pointer text-slate-500" title="View All Records (Overrides Sharing)">
                                                    <input
                                                        type="checkbox"
                                                        checked={!!p.view_all}
                                                        onChange={() => handleToggle(profile.id, obj.api_name, 'view_all')}
                                                        className="rounded text-purple-600 focus:ring-purple-500 border-slate-300"
                                                    />
                                                    <span>View All</span>
                                                </label>
                                                {/* Modify All */}
                                                <label className="flex items-center gap-1.5 cursor-pointer text-slate-500" title="Modify All Records (Overrides Sharing)">
                                                    <input
                                                        type="checkbox"
                                                        checked={!!p.modify_all}
                                                        onChange={() => handleToggle(profile.id, obj.api_name, 'modify_all')}
                                                        className="rounded text-purple-600 focus:ring-purple-500 border-slate-300"
                                                    />
                                                    <span>Mod All</span>
                                                </label>
                                            </div>
                                        </td>
                                    );
                                })}
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};
