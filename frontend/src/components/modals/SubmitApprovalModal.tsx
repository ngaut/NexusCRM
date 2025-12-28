import React, { useState } from 'react';
import { Send } from 'lucide-react';
import { approvalsAPI } from '../../infrastructure/api/approvals';
import { Button } from '../ui/Button';
import { useSuccessToast, useErrorToast } from '../ui/Toast';

interface SubmitApprovalModalProps {
    isOpen: boolean;
    onClose: () => void;
    objectApiName: string;
    recordId: string;
    recordName?: string;
    onSuccess?: () => void;
}

export function SubmitApprovalModal({
    isOpen,
    onClose,
    objectApiName,
    recordId,
    recordName,
    onSuccess,
}: SubmitApprovalModalProps) {
    const showSuccess = useSuccessToast();
    const showError = useErrorToast();

    const [comments, setComments] = useState('');
    const [submitting, setSubmitting] = useState(false);

    if (!isOpen) return null;

    const handleSubmit = async () => {
        setSubmitting(true);
        try {
            await approvalsAPI.submit({
                object_api_name: objectApiName,
                record_id: recordId,
                comments: comments || undefined,
            });
            showSuccess('Record submitted for approval');
            setComments('');
            onClose();
            onSuccess?.();
        } catch (err) {
            showError('Failed to submit for approval: ' + (err instanceof Error ? err.message : 'Unknown error'));
        } finally {
            setSubmitting(false);
        }
    };

    const handleCancel = () => {
        setComments('');
        onClose();
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
            {/* Backdrop */}
            <div
                className="absolute inset-0 bg-black/50 backdrop-blur-sm"
                onClick={handleCancel}
            />

            {/* Modal */}
            <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-md mx-4 overflow-hidden animate-fade-in">
                {/* Header */}
                <div className="px-6 py-4 border-b border-gray-200 bg-gradient-to-r from-amber-500 to-orange-600">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-white/20 rounded-lg">
                            <Send className="w-5 h-5 text-white" />
                        </div>
                        <div>
                            <h2 className="text-lg font-semibold text-white">Submit for Approval</h2>
                            {recordName && (
                                <p className="text-sm text-white/80">{recordName}</p>
                            )}
                        </div>
                    </div>
                </div>

                {/* Body */}
                <div className="p-6 space-y-4">
                    <p className="text-sm text-gray-600">
                        Submit this {objectApiName.replace('_', ' ')} record for approval.
                        An approver will review and either approve or reject it.
                    </p>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                            Comments (optional)
                        </label>
                        <textarea
                            value={comments}
                            onChange={(e) => setComments(e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-amber-500 focus:border-amber-500 transition-colors"
                            rows={3}
                            placeholder="Add any notes for the approver..."
                            disabled={submitting}
                        />
                    </div>
                </div>

                {/* Footer */}
                <div className="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end gap-3">
                    <Button
                        variant="ghost"
                        onClick={handleCancel}
                        disabled={submitting}
                    >
                        Cancel
                    </Button>
                    <Button
                        variant="primary"
                        onClick={handleSubmit}
                        disabled={submitting}
                        icon={<Send className="w-4 h-4" />}
                    >
                        {submitting ? 'Submitting...' : 'Submit for Approval'}
                    </Button>
                </div>
            </div>
        </div>
    );
}
