import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { CheckCircle, XCircle, ExternalLink, RefreshCw, Inbox, Calendar } from 'lucide-react';
import { approvalsAPI, ApprovalWorkItem } from '../infrastructure/api/approvals';
import { COMMON_FIELDS } from '../core/constants';
import { Button } from '../components/ui/Button';
import { useSuccessToast, useErrorToast } from '../components/ui/Toast';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';

export function ApprovalQueue() {
    const navigate = useNavigate();
    const showSuccess = useSuccessToast();
    const showError = useErrorToast();

    const [workItems, setWorkItems] = useState<ApprovalWorkItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [processingId, setProcessingId] = useState<string | null>(null);

    // Modal state for approve/reject with comments
    const [actionModal, setActionModal] = useState<{
        isOpen: boolean;
        action: 'approve' | 'reject';
        workItem: ApprovalWorkItem | null;
    }>({ isOpen: false, action: 'approve', workItem: null });
    const [comments, setComments] = useState('');

    const loadPendingApprovals = async () => {
        setLoading(true);
        setError(null);
        try {
            const items = await approvalsAPI.getPending();
            setWorkItems(items);
        } catch (err) {
            setError('Failed to load pending approvals: ' + (err instanceof Error ? err.message : 'Unknown error'));
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadPendingApprovals();
    }, []);

    const handleApprove = (workItem: ApprovalWorkItem) => {
        setActionModal({ isOpen: true, action: 'approve', workItem });
        setComments('');
    };

    const handleReject = (workItem: ApprovalWorkItem) => {
        setActionModal({ isOpen: true, action: 'reject', workItem });
        setComments('');
    };

    const confirmAction = async () => {
        if (!actionModal.workItem) return;

        setProcessingId(actionModal.workItem[COMMON_FIELDS.ID] as string);
        try {
            if (actionModal.action === 'approve') {
                await approvalsAPI.approve(actionModal.workItem[COMMON_FIELDS.ID] as string, comments);
                showSuccess('Approval submitted successfully');
            } else {
                await approvalsAPI.reject(actionModal.workItem[COMMON_FIELDS.ID] as string, comments);
                showSuccess('Rejection submitted successfully');
            }
            setActionModal({ isOpen: false, action: 'approve', workItem: null });
            loadPendingApprovals();
        } catch (err) {
            showError(`Failed to ${actionModal.action}: ` + (err instanceof Error ? err.message : 'Unknown error'));
        } finally {
            setProcessingId(null);
        }
    };

    const viewRecord = (workItem: ApprovalWorkItem) => {
        navigate(`/object/${workItem[COMMON_FIELDS.OBJECT_API_NAME]}/${workItem[COMMON_FIELDS.RECORD_ID]}`);
    };

    const formatDate = (dateStr: string) => {
        return new Date(dateStr).toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    return (
        <div className="p-6 max-w-7xl mx-auto">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-gradient-to-br from-amber-500 to-orange-600 rounded-xl shadow-lg">
                        <Inbox className="w-6 h-6 text-white" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-gray-900">Approval Queue</h1>
                        <p className="text-sm text-gray-500">
                            {workItems.length} pending approval{workItems.length !== 1 ? 's' : ''}
                        </p>
                    </div>
                </div>
                <Button
                    variant="secondary"
                    onClick={loadPendingApprovals}
                    icon={<RefreshCw className="w-4 h-4" />}
                    disabled={loading}
                >
                    Refresh
                </Button>
            </div>

            {/* Error */}
            {error && (
                <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700">
                    <XCircle className="w-5 h-5" />
                    {error}
                </div>
            )}

            {/* Loading */}
            {loading ? (
                <div className="flex items-center justify-center py-12">
                    <div className="w-8 h-8 border-2 border-amber-500 border-t-transparent rounded-full animate-spin" />
                </div>
            ) : workItems.length === 0 ? (
                /* Empty State */
                <div className="text-center py-16 bg-gray-50 rounded-xl border border-dashed border-gray-200">
                    <CheckCircle className="w-16 h-16 text-green-400 mx-auto mb-4" />
                    <h3 className="text-lg font-medium text-gray-900 mb-2">All Caught Up!</h3>
                    <p className="text-gray-500">
                        You have no pending approvals at this time.
                    </p>
                </div>
            ) : (
                /* Approval Items List */
                <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
                    <table className="min-w-full divide-y divide-gray-200">
                        <thead className="bg-gray-50">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Record
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Object
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Submitted
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Comments
                                </th>
                                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Actions
                                </th>
                            </tr>
                        </thead>
                        <tbody className="bg-white divide-y divide-gray-200">
                            {workItems.map((item) => (
                                <tr key={item[COMMON_FIELDS.ID] as string} className="hover:bg-gray-50 transition-colors">
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <button
                                            onClick={() => viewRecord(item)}
                                            className="flex items-center gap-2 text-blue-600 hover:text-blue-800 font-medium"
                                        >
                                            {(item[COMMON_FIELDS.RECORD_ID] as string).substring(0, 8)}...
                                            <ExternalLink className="w-3 h-3" />
                                        </button>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <span className="px-2 py-1 text-xs font-medium bg-gray-100 text-gray-700 rounded">
                                            {item[COMMON_FIELDS.OBJECT_API_NAME]}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                        <div className="flex items-center gap-1">
                                            <div className="sr-only">Calendar</div>
                                            <Calendar className="w-4 h-4" />
                                            {formatDate(item[COMMON_FIELDS.SUBMITTED_DATE] as string)}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4">
                                        <span className="text-sm text-gray-600 truncate max-w-xs block">
                                            {item.comments || '—'}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-right">
                                        <div className="flex items-center justify-end gap-2">
                                            <Button
                                                variant="primary"
                                                size="sm"
                                                onClick={() => handleApprove(item)}
                                                disabled={processingId === item[COMMON_FIELDS.ID]}
                                                icon={<CheckCircle className="w-4 h-4" />}
                                            >
                                                Approve
                                            </Button>
                                            <Button
                                                variant="danger"
                                                size="sm"
                                                onClick={() => handleReject(item)}
                                                disabled={processingId === item[COMMON_FIELDS.ID]}
                                                icon={<XCircle className="w-4 h-4" />}
                                            >
                                                Reject
                                            </Button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}

            {/* Action Confirmation Modal */}
            <ConfirmationModal
                isOpen={actionModal.isOpen}
                onClose={() => setActionModal({ isOpen: false, action: 'approve', workItem: null })}
                onConfirm={confirmAction}
                title={actionModal.action === 'approve' ? 'Approve Request' : 'Reject Request'}
                message={
                    <div className="space-y-4">
                        <p>
                            Are you sure you want to {actionModal.action} this approval request?
                        </p>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Comments (optional)
                            </label>
                            <textarea
                                value={comments}
                                onChange={(e) => setComments(e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                rows={3}
                                placeholder="Add any comments..."
                            />
                        </div>
                    </div>
                }
                confirmLabel={actionModal.action === 'approve' ? 'Approve' : 'Reject'}
                variant={actionModal.action === 'approve' ? 'info' : 'danger'}
                loading={processingId !== null}
            />
        </div>
    );
}

export default ApprovalQueue;
