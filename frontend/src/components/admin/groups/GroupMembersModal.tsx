import React from 'react';
import { Users, UserPlus, UserMinus } from 'lucide-react';
import type { Group, GroupMember, User } from '../../../types';

interface GroupMembersModalProps {
    group: Group | null;
    members: GroupMember[];
    availableUsers: User[];
    memberUsers: User[];
    onClose: () => void;
    onAddMember: (userId: string) => Promise<void>;
    onRemoveMember: (memberId: string) => Promise<void>;
}

export const GroupMembersModal: React.FC<GroupMembersModalProps> = ({
    group,
    members,
    availableUsers,
    memberUsers,
    onClose,
    onAddMember,
    onRemoveMember,
}) => {
    if (!group) return null;

    const nonMemberUsers = availableUsers.filter(
        u => !memberUsers.some(mu => mu.id === u.id)
    );

    return (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-white rounded-xl shadow-2xl w-full max-w-2xl mx-4 max-h-[80vh] overflow-hidden flex flex-col">
                <div className="p-6 border-b border-slate-200">
                    <h2 className="text-xl font-semibold text-slate-800">
                        Members of "{group.label}"
                    </h2>
                </div>
                <div className="p-6 overflow-auto flex-1">
                    <div className="grid grid-cols-2 gap-6">
                        {/* Current Members */}
                        <div>
                            <h3 className="font-semibold text-slate-700 mb-3 flex items-center gap-2">
                                <Users size={16} /> Current Members ({memberUsers.length})
                            </h3>
                            <div className="space-y-2">
                                {memberUsers.length === 0 ? (
                                    <p className="text-slate-500 text-sm">No members yet</p>
                                ) : (
                                    memberUsers.map(user => {
                                        const membership = members.find(m => m.user_id === user.id);
                                        return (
                                            <div key={user.id} className="flex items-center justify-between p-2 bg-slate-50 rounded-lg">
                                                <div>
                                                    <div className="font-medium text-slate-800">{user.name}</div>
                                                    <div className="text-xs text-slate-500">{user.email}</div>
                                                </div>
                                                <button
                                                    onClick={() => membership && onRemoveMember(membership.id)}
                                                    className="p-1 text-red-600 hover:bg-red-50 rounded"
                                                    title="Remove member"
                                                >
                                                    <UserMinus size={16} />
                                                </button>
                                            </div>
                                        );
                                    })
                                )}
                            </div>
                        </div>

                        {/* Available Users */}
                        <div>
                            <h3 className="font-semibold text-slate-700 mb-3 flex items-center gap-2">
                                <UserPlus size={16} /> Add Members ({nonMemberUsers.length})
                            </h3>
                            <div className="space-y-2 max-h-64 overflow-auto">
                                {nonMemberUsers.length === 0 ? (
                                    <p className="text-slate-500 text-sm">All users are members</p>
                                ) : (
                                    nonMemberUsers.map(user => (
                                        <div key={user.id} className="flex items-center justify-between p-2 border border-slate-200 rounded-lg hover:border-blue-300">
                                            <div>
                                                <div className="font-medium text-slate-800">{user.name}</div>
                                                <div className="text-xs text-slate-500">{user.email}</div>
                                            </div>
                                            <button
                                                onClick={() => onAddMember(user.id)}
                                                className="p-1 text-teal-600 hover:bg-teal-50 rounded"
                                                title="Add member"
                                            >
                                                <UserPlus size={16} />
                                            </button>
                                        </div>
                                    ))
                                )}
                            </div>
                        </div>
                    </div>
                </div>
                <div className="p-6 border-t border-slate-200 flex justify-end">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-200 font-medium"
                    >
                        Done
                    </button>
                </div>
            </div>
        </div>
    );
};
