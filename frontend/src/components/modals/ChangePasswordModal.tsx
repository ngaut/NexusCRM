import React, { useState } from 'react';
import { X, Lock, Check } from 'lucide-react';
import { Button } from '../ui/Button';
import { dataAPI } from '../../infrastructure/api/data';
import { useSuccessToast, useErrorToast } from '../ui/Toast';
import { formatApiError, getOperationErrorMessage } from '../../core/utils/errorHandling';

interface ChangePasswordModalProps {
    isOpen: boolean;
    onClose: () => void;
    recordId: string;
    objectApiName: string;
    onSuccess?: () => void;
}

export function ChangePasswordModal({
    isOpen,
    onClose,
    recordId,
    objectApiName,
    onSuccess
}: ChangePasswordModalProps) {
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [loading, setLoading] = useState(false);

    const showSuccess = useSuccessToast();
    const showError = useErrorToast();

    if (!isOpen) return null;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (password !== confirmPassword) {
            showError("Passwords do not match");
            return;
        }

        if (password.length < 1) {
            showError("Password cannot be empty");
            return;
        }

        setLoading(true);
        try {
            await dataAPI.updateRecord(objectApiName, recordId, { password });
            showSuccess("Password updated successfully");
            setPassword('');
            setConfirmPassword('');
            if (onSuccess) onSuccess();
            onClose();
        } catch (err) {
            const apiError = formatApiError(err);
            showError(getOperationErrorMessage('update', 'Password', apiError));
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 animate-in fade-in duration-200">
            <div className="bg-white rounded-xl shadow-xl w-full max-w-md overflow-hidden animate-in zoom-in-95 duration-200">
                <div className="px-6 py-4 border-b border-gray-100 flex justify-between items-center bg-gray-50/50">
                    <h2 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                        <Lock className="w-5 h-5 text-gray-500" />
                        Change Password
                    </h2>
                    <button
                        onClick={onClose}
                        className="text-gray-400 hover:text-gray-600 transition-colors p-1 rounded-lg hover:bg-gray-100"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="p-6 space-y-4">
                    <div className="space-y-2">
                        <label className="text-sm font-medium text-gray-700">
                            New Password
                        </label>
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            className="w-full h-10 px-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all placeholder:text-gray-400"
                            placeholder="Enter new password"
                            required
                        />
                    </div>

                    <div className="space-y-2">
                        <label className="text-sm font-medium text-gray-700">
                            Confirm Password
                        </label>
                        <input
                            type="password"
                            value={confirmPassword}
                            onChange={(e) => setConfirmPassword(e.target.value)}
                            className="w-full h-10 px-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all placeholder:text-gray-400"
                            placeholder="Confirm new password"
                            required
                        />
                    </div>

                    <div className="flex justify-end gap-3 pt-4">
                        <Button
                            type="button"
                            variant="ghost"
                            onClick={onClose}
                            disabled={loading}
                        >
                            Cancel
                        </Button>
                        <Button
                            type="submit"
                            variant="primary"
                            loading={loading}
                            disabled={loading}
                            icon={<Check className="w-4 h-4" />}
                        >
                            Update Password
                        </Button>
                    </div>
                </form>
            </div>
        </div>
    );
}
