import React from 'react';
import { Users } from 'lucide-react';
import type { AppConfig } from '../../../types';
import type { Profile } from '../../../types';

interface AppFormUserAccessProps {
    assignedProfiles: string[];
    availableProfiles: Profile[];
    visibleToAll: boolean;
    onChangeVisibleToAll: (visible: boolean) => void;
    onChangeProfiles: (profiles: string[]) => void;
}

export const AppFormUserAccess: React.FC<AppFormUserAccessProps> = ({
    assignedProfiles,
    availableProfiles,
    visibleToAll,
    onChangeVisibleToAll,
    onChangeProfiles,
}) => {
    return (
        <div>
            <div className="flex items-center gap-2 mb-3">
                <Users size={16} className="text-slate-600" />
                <label className="text-sm font-medium text-slate-700">User Access</label>
            </div>

            {/* Visible to all toggle */}
            <label className="flex items-center gap-3 p-3 border rounded-lg cursor-pointer hover:bg-slate-50 mb-3">
                <input
                    type="checkbox"
                    checked={visibleToAll}
                    onChange={(e) => onChangeVisibleToAll(e.target.checked)}
                    className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                />
                <div>
                    <div className="font-medium text-slate-800 text-sm">Visible to all profiles</div>
                    <div className="text-xs text-slate-500">All users can access this app</div>
                </div>
            </label>

            {/* Profile selection - only show when not visible to all */}
            {!visibleToAll && (
                <div className="border rounded-lg overflow-hidden">
                    <div className="px-3 py-2 bg-slate-50 border-b text-xs font-medium text-slate-600">
                        Select profiles that can access this app
                    </div>
                    <div className="max-h-40 overflow-y-auto divide-y">
                        {availableProfiles.map(profile => (
                            <label
                                key={profile.id}
                                className="flex items-center gap-3 px-3 py-2.5 cursor-pointer hover:bg-slate-50"
                            >
                                <input
                                    type="checkbox"
                                    checked={assignedProfiles.includes(profile.id)}
                                    onChange={(e) => {
                                        const current = assignedProfiles;
                                        const updated = e.target.checked
                                            ? [...current, profile.id]
                                            : current.filter(id => id !== profile.id);
                                        onChangeProfiles(updated);
                                    }}
                                    className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                                />
                                <div className="flex-1">
                                    <div className="font-medium text-slate-800 text-sm">{profile.name}</div>
                                    {profile.description && (
                                        <div className="text-xs text-slate-500">{profile.description}</div>
                                    )}
                                </div>
                                {profile.is_system && (
                                    <span className="text-xs px-2 py-0.5 bg-slate-100 text-slate-600 rounded">System</span>
                                )}
                            </label>
                        ))}
                    </div>
                </div>
            )}
            <p className="text-xs text-slate-500 mt-1">Control which user profiles can see this app in their App Launcher.</p>
        </div>
    );
};
