
import React, { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { X, Save, AlertCircle } from 'lucide-react';
import { usersAPI } from '../../infrastructure/api/users';
import type { User, Profile } from '../../types';

interface UserEditorModalProps {
    user?: User | null;
    profiles: Profile[];
    onClose: () => void;
    onSave: () => void;
}

export const UserEditorModal: React.FC<UserEditorModalProps> = ({ user, profiles, onClose, onSave }) => {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [showPasswordChange, setShowPasswordChange] = useState(false);

    const [formData, setFormData] = useState({
        name: user?.name || '',
        email: user?.email || '',
        profile_id: user?.profile_id || profiles[0]?.id || '',
        password: '',
        confirmPassword: ''
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setLoading(true);

        try {
            if (!user) {
                // Create
                if (formData.password !== formData.confirmPassword) {
                    throw new Error("Passwords do not match");
                }
                if (!formData.password) {
                    throw new Error("Password is required for new users");
                }

                await usersAPI.createUser({
                    name: formData.name,
                    email: formData.email,
                    profile_id: formData.profile_id,
                    password: formData.password
                });
            } else {
                // Update
                if (formData.password && formData.password !== formData.confirmPassword) {
                    throw new Error("Passwords do not match");
                }

                await usersAPI.updateUser(user.id, {
                    name: formData.name,
                    email: formData.email,
                    profile_id: formData.profile_id,
                    ...(formData.password ? { password: formData.password } : {})
                });
            }
            onSave();
            onClose();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to save user');
        } finally {
            setLoading(false);
        }
    };

    return createPortal(
        <div className="fixed inset-0 bg-slate-900/50 backdrop-blur-sm flex items-center justify-center p-4 z-[100]">
            <div className="bg-white rounded-xl shadow-2xl w-full max-w-md overflow-hidden">
                <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50">
                    <h3 className="font-semibold text-slate-800">
                        {user ? 'Edit User' : 'New User'}
                    </h3>
                    <button
                        onClick={onClose}
                        className="text-slate-400 hover:text-slate-600 transition-colors"
                    >
                        <X size={20} />
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="p-6 space-y-4">
                    {error && (
                        <div className="p-3 bg-red-50 text-red-600 text-sm rounded-lg flex items-center gap-2">
                            <AlertCircle size={16} />
                            {error}
                        </div>
                    )}

                    <div className="space-y-1">
                        <label className="text-sm font-medium text-slate-700">Full Name</label>
                        <input
                            type="text"
                            required
                            value={formData.name}
                            onChange={e => setFormData({ ...formData, name: e.target.value })}
                            className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500"
                            onFocus={(e) => e.target.select()}
                        />
                    </div>

                    <div className="space-y-1">
                        <label className="text-sm font-medium text-slate-700">Email</label>
                        <input
                            type="email"
                            required
                            value={formData.email}
                            onChange={e => setFormData({ ...formData, email: e.target.value })}
                            className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500"
                            onFocus={(e) => e.target.select()}
                        />
                    </div>

                    <div className="space-y-1">
                        <label className="text-sm font-medium text-slate-700">Profile</label>
                        <select
                            value={formData.profile_id}
                            onChange={e => setFormData({ ...formData, profile_id: e.target.value })}
                            className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500"
                        >
                            {profiles.map(p => (
                                <option key={p.id} value={p.id}>{p.name}</option>
                            ))}
                        </select>
                    </div>

                    {user && (
                        <div className="flex items-center gap-2 pt-2">
                            <input
                                type="checkbox"
                                id="changePassword"
                                checked={showPasswordChange}
                                onChange={e => {
                                    setShowPasswordChange(e.target.checked);
                                    if (!e.target.checked) {
                                        setFormData({ ...formData, password: '', confirmPassword: '' });
                                    }
                                }}
                                className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                            />
                            <label htmlFor="changePassword" className="text-sm text-slate-700">Change Password</label>
                        </div>
                    )}

                    {(!user || showPasswordChange) && (
                        <>
                            <div className="space-y-1">
                                <label className="text-sm font-medium text-slate-700">
                                    {user ? 'New Password' : 'Password'}
                                </label>
                                <input
                                    type="password"
                                    required={!user || showPasswordChange}
                                    value={formData.password}
                                    onChange={e => setFormData({ ...formData, password: e.target.value })}
                                    className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500"
                                    placeholder={user ? "Leave empty to keep current" : ""}
                                    onFocus={(e) => e.target.select()}
                                />
                            </div>

                            <div className="space-y-1">
                                <label className="text-sm font-medium text-slate-700">Confirm Password</label>
                                <input
                                    type="password"
                                    required={!user || showPasswordChange}
                                    value={formData.confirmPassword}
                                    onChange={e => setFormData({ ...formData, confirmPassword: e.target.value })}
                                    className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500"
                                    onFocus={(e) => e.target.select()}
                                />
                            </div>
                        </>
                    )}

                    <div className="pt-4 flex justify-end gap-3">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-slate-600 hover:bg-slate-50 rounded-lg font-medium transition-colors"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={loading}
                            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium shadow-sm transition-colors flex items-center gap-2 disabled:opacity-50"
                        >
                            <Save size={18} />
                            {loading ? 'Saving...' : 'Save User'}
                        </button>
                    </div>
                </form>
            </div>
        </div>,
        document.body
    );
};
