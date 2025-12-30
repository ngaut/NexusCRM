import React, { useState, useEffect } from 'react';
import { Users, Plus, Search, Trash2, Edit2, Mail } from 'lucide-react';
import { dataAPI } from '../infrastructure/api/data';
import { SYSTEM_TABLE_NAMES } from '../generated-schema';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';
import { useErrorToast, useSuccessToast } from '../components/ui/Toast';
import type { Group, GroupMember, User } from '../types';

// Sub-components
import { GroupFormModal, GroupFormData } from '../components/admin/groups/GroupFormModal';
import { GroupMembersModal } from '../components/admin/groups/GroupMembersModal';

export const GroupManager: React.FC = () => {
    const errorToast = useErrorToast();
    const successToast = useSuccessToast();
    const [groups, setGroups] = useState<Group[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [editingGroup, setEditingGroup] = useState<Group | null>(null);
    const [deletingGroup, setDeletingGroup] = useState<Group | null>(null);
    const [managingMembersFor, setManagingMembersFor] = useState<Group | null>(null);
    const [members, setMembers] = useState<GroupMember[]>([]);
    const [availableUsers, setAvailableUsers] = useState<User[]>([]);
    const [memberUsers, setMemberUsers] = useState<User[]>([]);

    useEffect(() => {
        loadGroups();
    }, []);

    const loadGroups = async () => {
        setLoading(true);
        try {
            const records = await dataAPI.query<Group>({ objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_GROUP });
            setGroups(records);
        } catch {
            errorToast('Failed to load groups');
        } finally {
            setLoading(false);
        }
    };

    const loadMembers = async (groupId: string) => {
        try {
            // Load group members
            const membersRecords = await dataAPI.query<GroupMember>({
                objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_GROUPMEMBER,
                filterExpr: `group_id == '${groupId}'`
            });
            setMembers(membersRecords);

            // Load all users
            const allUsers = await dataAPI.query<User>({ objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_USER });
            setAvailableUsers(allUsers);

            // Map member user IDs to user objects
            const memberUserIds = membersRecords.map((m: GroupMember) => m.user_id);
            setMemberUsers(allUsers.filter((u: User) => memberUserIds.includes(u.id)));
        } catch {
            errorToast('Failed to load group members');
        }
    };

    const handleCreate = async (formData: GroupFormData) => {
        try {
            await dataAPI.createRecord(SYSTEM_TABLE_NAMES.SYSTEM_GROUP, formData as unknown as Record<string, unknown>);
            successToast('Group created successfully');
            setShowCreateModal(false);
            loadGroups();
        } catch {
            errorToast('Failed to create group');
        }
    };

    const handleUpdate = async (formData: GroupFormData) => {
        if (!editingGroup) return;
        try {
            await dataAPI.updateRecord(SYSTEM_TABLE_NAMES.SYSTEM_GROUP, editingGroup.id, formData as unknown as Record<string, unknown>);
            successToast('Group updated successfully');
            setEditingGroup(null);
            loadGroups();
        } catch {
            errorToast('Failed to update group');
        }
    };

    const handleDelete = async () => {
        if (!deletingGroup) return;
        try {
            await dataAPI.deleteRecord(SYSTEM_TABLE_NAMES.SYSTEM_GROUP, deletingGroup.id);
            successToast('Group deleted successfully');
            setDeletingGroup(null);
            loadGroups();
        } catch {
            errorToast('Failed to delete group');
        }
    };

    const addMember = async (userId: string) => {
        if (!managingMembersFor) return;
        try {
            await dataAPI.createRecord(SYSTEM_TABLE_NAMES.SYSTEM_GROUPMEMBER, {
                group_id: managingMembersFor.id,
                user_id: userId
            });
            successToast('Member added');
            loadMembers(managingMembersFor.id);
        } catch {
            errorToast('Failed to add member');
        }
    };

    const removeMember = async (memberId: string) => {
        try {
            await dataAPI.deleteRecord(SYSTEM_TABLE_NAMES.SYSTEM_GROUPMEMBER, memberId);
            successToast('Member removed');
            if (managingMembersFor) {
                loadMembers(managingMembersFor.id);
            }
        } catch {
            errorToast('Failed to remove member');
        }
    };

    const openMembersModal = (group: Group) => {
        setManagingMembersFor(group);
        loadMembers(group.id);
    };

    const filteredGroups = groups.filter(g =>
        g.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        g.label.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="max-w-7xl mx-auto p-6">
            <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-4">
                    <div className="p-3 bg-teal-100 rounded-lg">
                        <Users className="text-teal-600" size={32} />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-slate-800">Groups & Queues</h1>
                        <p className="text-slate-500">Manage user groups for record ownership and collaboration.</p>
                    </div>
                </div>
                <button
                    onClick={() => setShowCreateModal(true)}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium shadow-sm transition-colors"
                >
                    <Plus size={18} />
                    New Group
                </button>
            </div>

            {/* Search Bar */}
            <div className="mb-6">
                <div className="relative">
                    <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                    <input
                        type="text"
                        placeholder="Search groups..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full pl-10 pr-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                </div>
            </div>

            {/* Groups Table */}
            <div className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden">
                {loading ? (
                    <div className="text-center py-8 text-slate-500">Loading...</div>
                ) : filteredGroups.length === 0 ? (
                    <div className="text-center py-8 text-slate-500">
                        {searchQuery ? 'No groups match your search.' : 'No groups created yet.'}
                    </div>
                ) : (
                    <table className="w-full text-left text-sm">
                        <thead className="bg-slate-50 border-b border-slate-200">
                            <tr>
                                <th className="px-6 py-3 font-semibold text-slate-700">Name</th>
                                <th className="px-6 py-3 font-semibold text-slate-700">Label</th>
                                <th className="px-6 py-3 font-semibold text-slate-700">Type</th>
                                <th className="px-6 py-3 font-semibold text-slate-700">Email</th>
                                <th className="px-6 py-3 font-semibold text-slate-700 text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100">
                            {filteredGroups.map(group => (
                                <tr key={group.id} className="hover:bg-slate-50">
                                    <td className="px-6 py-4 font-medium text-slate-900">{group.name}</td>
                                    <td className="px-6 py-4 text-slate-600">{group.label}</td>
                                    <td className="px-6 py-4">
                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${group.type === 'Queue'
                                            ? 'bg-blue-50 text-blue-700 border border-blue-100'
                                            : 'bg-slate-100 text-slate-700 border border-slate-200'
                                            }`}>
                                            {group.type}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4 text-slate-500">
                                        {group.email ? (
                                            <span className="flex items-center gap-1">
                                                <Mail size={14} />
                                                {group.email}
                                            </span>
                                        ) : '-'}
                                    </td>
                                    <td className="px-6 py-4 text-right whitespace-nowrap">
                                        <button
                                            onClick={() => openMembersModal(group)}
                                            className="text-sm font-medium text-teal-600 hover:text-teal-800 mr-3"
                                        >
                                            <Users size={14} className="inline mr-1" />
                                            Members
                                        </button>
                                        <button
                                            onClick={() => setEditingGroup(group)}
                                            className="text-sm font-medium text-blue-600 hover:text-blue-800 mr-3"
                                        >
                                            <Edit2 size={14} className="inline mr-1" />
                                            Edit
                                        </button>
                                        <button
                                            onClick={() => setDeletingGroup(group)}
                                            className="text-sm font-medium text-red-600 hover:text-red-800"
                                        >
                                            <Trash2 size={14} className="inline mr-1" />
                                            Delete
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>

            {/* Create/Edit Modal */}
            <GroupFormModal
                isOpen={showCreateModal || !!editingGroup}
                initialData={editingGroup ? {
                    name: editingGroup.name,
                    label: editingGroup.label,
                    type: editingGroup.type as 'Queue' | 'Regular',
                    email: editingGroup.email || ''
                } : null}
                isEditing={!!editingGroup}
                onClose={() => {
                    setShowCreateModal(false);
                    setEditingGroup(null);
                }}
                onSave={editingGroup ? handleUpdate : handleCreate}
            />

            {/* Members Modal */}
            <GroupMembersModal
                group={managingMembersFor}
                members={members}
                availableUsers={availableUsers}
                memberUsers={memberUsers}
                onClose={() => setManagingMembersFor(null)}
                onAddMember={addMember}
                onRemoveMember={removeMember}
            />

            {/* Delete Confirmation */}
            <ConfirmationModal
                isOpen={!!deletingGroup}
                onClose={() => setDeletingGroup(null)}
                onConfirm={handleDelete}
                title="Delete Group"
                message={`Are you sure you want to delete "${deletingGroup?.label}"? This will also remove all member associations.`}
                confirmLabel="Delete"
                variant="danger"
            />
        </div>
    );
};
