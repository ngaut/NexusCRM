import React, { useState, useEffect } from 'react';
import { Key, Plus, X, Loader2, Shield, Trash2, Check } from 'lucide-react';
import { dataAPI } from '../infrastructure/api/data';
import { useErrorToast, useSuccessToast } from './ui/Toast';

interface PermissionSet {
    id: string;
    name: string;
    label: string;
    description?: string;
    is_active: boolean;
}

interface PermissionSetAssignment {
    id: string;
    assignee_id: string;
    permission_set_id: string;
    created_date?: string;
}

interface UserPermissionSetsSectionProps {
    userId: string;
    userName?: string;
    isOpen: boolean;
    onClose: () => void;
}

export const UserPermissionSetsSection: React.FC<UserPermissionSetsSectionProps> = ({
    userId,
    userName,
    isOpen,
    onClose
}) => {
    const errorToast = useErrorToast();
    const successToast = useSuccessToast();
    const [loading, setLoading] = useState(true);
    const [assignments, setAssignments] = useState<(PermissionSetAssignment & { permissionSet?: PermissionSet })[]>([]);
    const [allPermissionSets, setAllPermissionSets] = useState<PermissionSet[]>([]);
    const [showAddModal, setShowAddModal] = useState(false);
    const [adding, setAdding] = useState(false);
    const [removingId, setRemovingId] = useState<string | null>(null);

    useEffect(() => {
        if (isOpen && userId) {
            loadData();
        }
    }, [isOpen, userId]);

    const loadData = async () => {
        setLoading(true);
        try {
            // Load user's current assignments
            const assignmentsRecords = await dataAPI.query({
                objectApiName: '_system_permissionsetassignment',
                filterExpr: `assignee_id == '${userId}'`
            });

            // Load all permission sets
            const permSets = await dataAPI.query({ objectApiName: '_system_permissionset' }) as unknown as PermissionSet[];
            setAllPermissionSets(permSets);

            // Merge permission set data into assignments
            const assignmentsWithDetails = (assignmentsRecords as unknown as PermissionSetAssignment[]).map((a: PermissionSetAssignment) => ({
                ...a,
                permissionSet: permSets.find((ps: PermissionSet) => ps.id === a.permission_set_id)
            }));

            setAssignments(assignmentsWithDetails);
        } catch {
            errorToast('Failed to load permission sets');
        } finally {
            setLoading(false);
        }
    };

    const handleAdd = async (permSetId: string) => {
        setAdding(true);
        try {
            await dataAPI.createRecord('_system_permissionsetassignment', {
                assignee_id: userId,
                permission_set_id: permSetId
            });
            successToast('Permission Set assigned');
            setShowAddModal(false);
            loadData();
        } catch (err: unknown) {
            const msg = err instanceof Error ? err.message : 'Failed to assign permission set';
            errorToast(msg);
        } finally {
            setAdding(false);
        }
    };

    const handleRemove = async (assignmentId: string) => {
        setRemovingId(assignmentId);
        try {
            await dataAPI.deleteRecord('_system_permissionsetassignment', assignmentId);
            successToast('Permission Set removed');
            loadData();
        } catch {
            errorToast('Failed to remove permission set');
        } finally {
            setRemovingId(null);
        }
    };

    // Get available permission sets (not already assigned)
    const assignedIds = new Set(assignments.map(a => a.permission_set_id));
    const availablePermSets = allPermissionSets.filter(ps => !assignedIds.has(ps.id) && ps.is_active);

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-white rounded-xl shadow-2xl w-full max-w-lg mx-4">
                <div className="flex items-center justify-between p-6 border-b border-slate-200">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-amber-100 rounded-lg">
                            <Key className="text-amber-600" size={20} />
                        </div>
                        <div>
                            <h2 className="text-lg font-semibold text-slate-800">Permission Sets</h2>
                            {userName && <p className="text-sm text-slate-500">for {userName}</p>}
                        </div>
                    </div>
                    <button onClick={onClose} className="p-2 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-50">
                        <X size={20} />
                    </button>
                </div>

                <div className="p-6">
                    {loading ? (
                        <div className="flex items-center justify-center py-8 text-slate-500">
                            <Loader2 className="animate-spin mr-2" size={20} />
                            Loading...
                        </div>
                    ) : (
                        <>
                            {/* Current Assignments */}
                            <div className="mb-6">
                                <div className="flex items-center justify-between mb-3">
                                    <h3 className="text-sm font-medium text-slate-700">Assigned Permission Sets</h3>
                                    <button
                                        onClick={() => setShowAddModal(true)}
                                        disabled={availablePermSets.length === 0}
                                        className="flex items-center gap-1 px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                                    >
                                        <Plus size={14} />
                                        Assign
                                    </button>
                                </div>

                                {assignments.length === 0 ? (
                                    <div className="text-center py-6 text-slate-400 bg-slate-50 rounded-lg">
                                        No permission sets assigned
                                    </div>
                                ) : (
                                    <div className="space-y-2">
                                        {assignments.map(assignment => (
                                            <div
                                                key={assignment.id}
                                                className="flex items-center justify-between p-3 bg-slate-50 rounded-lg"
                                            >
                                                <div className="flex items-center gap-3">
                                                    <Shield size={18} className="text-amber-500" />
                                                    <div>
                                                        <div className="font-medium text-slate-800">
                                                            {assignment.permissionSet?.label || 'Unknown'}
                                                        </div>
                                                        <div className="text-xs text-slate-500">
                                                            {assignment.permissionSet?.name}
                                                        </div>
                                                    </div>
                                                </div>
                                                <button
                                                    onClick={() => handleRemove(assignment.id)}
                                                    disabled={removingId === assignment.id}
                                                    className="p-1.5 text-red-500 hover:text-red-700 hover:bg-red-50 rounded-lg disabled:opacity-50"
                                                >
                                                    {removingId === assignment.id ? (
                                                        <Loader2 className="animate-spin" size={16} />
                                                    ) : (
                                                        <Trash2 size={16} />
                                                    )}
                                                </button>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>

                            {/* Add Permission Set Modal */}
                            {showAddModal && (
                                <div className="border-t border-slate-200 pt-4">
                                    <h3 className="text-sm font-medium text-slate-700 mb-3">Available Permission Sets</h3>
                                    {availablePermSets.length === 0 ? (
                                        <div className="text-center py-4 text-slate-400">
                                            No more permission sets available
                                        </div>
                                    ) : (
                                        <div className="space-y-2 max-h-48 overflow-y-auto">
                                            {availablePermSets.map(ps => (
                                                <button
                                                    key={ps.id}
                                                    onClick={() => handleAdd(ps.id)}
                                                    disabled={adding}
                                                    className="w-full flex items-center justify-between p-3 bg-slate-50 hover:bg-blue-50 rounded-lg text-left transition-colors disabled:opacity-50"
                                                >
                                                    <div className="flex items-center gap-3">
                                                        <Shield size={18} className="text-slate-400" />
                                                        <div>
                                                            <div className="font-medium text-slate-800">{ps.label}</div>
                                                            <div className="text-xs text-slate-500">{ps.description || ps.name}</div>
                                                        </div>
                                                    </div>
                                                    <Plus size={16} className="text-blue-600" />
                                                </button>
                                            ))}
                                        </div>
                                    )}
                                    <button
                                        onClick={() => setShowAddModal(false)}
                                        className="mt-3 w-full py-2 text-sm text-slate-600 hover:text-slate-800"
                                    >
                                        Cancel
                                    </button>
                                </div>
                            )}
                        </>
                    )}
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
