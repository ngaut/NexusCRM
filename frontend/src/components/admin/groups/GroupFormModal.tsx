import React, { useState, useEffect } from 'react';

export interface GroupFormData {
    name: string;
    label: string;
    type: 'Queue' | 'Regular';
    email: string;
}

interface GroupFormModalProps {
    isOpen: boolean;
    initialData: GroupFormData | null;
    isEditing: boolean;
    onClose: () => void;
    onSave: (data: GroupFormData) => Promise<void>;
}

export const GroupFormModal: React.FC<GroupFormModalProps> = ({
    isOpen,
    initialData,
    isEditing,
    onClose,
    onSave,
}) => {
    const [formData, setFormData] = useState({
        name: '',
        label: '',
        type: 'Queue' as 'Queue' | 'Regular',
        email: ''
    });

    useEffect(() => {
        if (isOpen && initialData) {
            setFormData(initialData);
        } else if (isOpen) {
            setFormData({ name: '', label: '', type: 'Queue', email: '' });
        }
    }, [isOpen, initialData]);

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[100]">
            <div className="bg-white rounded-xl shadow-2xl w-full max-w-md mx-4">
                <div className="p-6 border-b border-slate-200">
                    <h2 className="text-xl font-semibold text-slate-800">
                        {isEditing ? 'Edit Group' : 'Create Group'}
                    </h2>
                </div>
                <div className="p-6 space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">
                            API Name *
                        </label>
                        <input
                            type="text"
                            value={formData.name}
                            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            placeholder="e.g., sales_queue"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">
                            Label *
                        </label>
                        <input
                            type="text"
                            value={formData.label}
                            onChange={(e) => setFormData({ ...formData, label: e.target.value })}
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            placeholder="e.g., Sales Queue"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">
                            Type
                        </label>
                        <select
                            value={formData.type}
                            onChange={(e) => setFormData({ ...formData, type: e.target.value as 'Queue' | 'Regular' })}
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                            <option value="Queue">Queue</option>
                            <option value="Regular">Regular Group</option>
                        </select>
                    </div>
                    {formData.type === 'Queue' && (
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1">
                                Queue Email
                            </label>
                            <input
                                type="email"
                                value={formData.email}
                                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                placeholder="queue@company.com"
                            />
                        </div>
                    )}
                </div>
                <div className="p-6 border-t border-slate-200 flex justify-end gap-3">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-slate-600 hover:text-slate-800 font-medium"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={() => onSave(formData)}
                        className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium"
                    >
                        {isEditing ? 'Save Changes' : 'Create'}
                    </button>
                </div>
            </div>
        </div>
    );
};
