import React, { useState, useEffect } from 'react';
import { Clock, CheckCircle, XCircle, User, Send, AlertCircle, Layers } from 'lucide-react';
import { Button } from './ui/Button';
import { approvalsAPI, ApprovalWorkItem } from '../infrastructure/api/approvals';
import { useSuccessToast, useErrorToast } from './ui/Toast';
import { useRuntime } from '../contexts/RuntimeContext';
import { StepProgressCompact } from './StepProgressIndicator';
import { APPROVAL_STATUS } from '../core/constants';

interface ApprovalBannerProps {
    objectApiName: string;
    recordId: string;
    recordName: string;
    /** Approval work items (from useApprovalStatus or parent) */
    pendingItem: ApprovalWorkItem | null;
    onActionComplete: () => void;
}

/**
 * Banner component shown on record detail when approval is pending.
 * - Shows pending status with submitter info
 * - If current user is the approver, shows inline Approve/Reject buttons
 */
export function ApprovalBanner({
    objectApiName,
    recordId,
    recordName,
    pendingItem,
    onActionComplete,
}: ApprovalBannerProps) {
    const { user } = useRuntime();
    const showSuccess = useSuccessToast();
    const showError = useErrorToast();

    const [processing, setProcessing] = useState(false);
    const [showCommentInput, setShowCommentInput] = useState(false);
    const [actionType, setActionType] = useState<'approve' | 'reject' | null>(null);
    const [comments, setComments] = useState('');
    const [flowProgress, setFlowProgress] = useState<{ currentStep: number; totalSteps: number; stepName?: string } | null>(null);

    // Fetch flow progress if this is a multi-step flow
    useEffect(() => {
        const fetchFlowProgress = async () => {
            if (pendingItem?.flow_instance_id) {
                try {
                    const instance = await approvalsAPI.getFlowInstanceProgress(pendingItem.flow_instance_id);
                    if (instance && instance.current_step_order && instance.total_steps) {
                        setFlowProgress({
                            currentStep: instance.current_step_order,
                            totalSteps: instance.total_steps,
                            stepName: instance.steps?.find(s => s.id === instance.current_step_id)?.step_name,
                        });
                    }
                } catch {
                    // Ignore errors - flow progress is optional
                }
            }
        };
        fetchFlowProgress();
    }, [pendingItem?.flow_instance_id]);

    // No pending item = no banner
    if (!pendingItem || pendingItem.status !== APPROVAL_STATUS.PENDING) {
        return null;
    }

    const isApprover = user?.id === pendingItem.approver_id;
    const isSubmitter = user?.id === pendingItem.submitted_by_id;

    const handleAction = async (action: 'approve' | 'reject') => {
        if (showCommentInput && actionType === action) {
            // Execute the action
            setProcessing(true);
            try {
                if (action === 'approve') {
                    await approvalsAPI.approve(pendingItem.id, comments);
                    showSuccess('Approval granted successfully');
                } else {
                    await approvalsAPI.reject(pendingItem.id, comments);
                    showSuccess('Approval rejected');
                }
                setShowCommentInput(false);
                setComments('');
                onActionComplete();
            } catch (err: unknown) {
                const errMsg = err instanceof Error ? err.message : String(err);
                showError(`Failed to ${action}: ${errMsg}`);
            } finally {
                setProcessing(false);
            }
        } else {
            // Show comment input
            setActionType(action);
            setShowCommentInput(true);
            setComments('');
        }
    };

    const handleCancel = () => {
        setShowCommentInput(false);
        setActionType(null);
        setComments('');
    };

    const formatDate = (dateStr: string) => {
        return new Date(dateStr).toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
            year: 'numeric',
        });
    };

    return (
        <div className={`rounded-lg border p-4 mb-6 ${isApprover
            ? 'bg-amber-50 border-amber-200'
            : 'bg-blue-50 border-blue-200'
            }`}>
            <div className="flex items-start justify-between gap-4">
                <div className="flex items-start gap-3">
                    <div className={`p-2 rounded-full ${isApprover ? 'bg-amber-100' : 'bg-blue-100'
                        }`}>
                        {isApprover ? (
                            <AlertCircle className="w-5 h-5 text-amber-600" />
                        ) : (
                            <Clock className="w-5 h-5 text-blue-600" />
                        )}
                    </div>
                    <div>
                        <h3 className={`font-semibold ${isApprover ? 'text-amber-900' : 'text-blue-900'
                            }`}>
                            {isApprover
                                ? '🔔 Approval Request for You'
                                : '⏳ Pending Approval'}
                        </h3>
                        <p className="text-sm text-gray-600 mt-1">
                            {isSubmitter
                                ? 'You submitted this for approval'
                                : `Submitted by ${pendingItem.submitted_by_id || 'Unknown'}`
                            }
                            {pendingItem.created_date && (
                                <span> on {formatDate(pendingItem.created_date)}</span>
                            )}
                        </p>
                        {pendingItem.comments && (
                            <p className="text-sm text-gray-500 mt-2 italic">
                                "{pendingItem.comments}"
                            </p>
                        )}
                        {!isApprover && (
                            <p className="text-sm text-gray-500 mt-2">
                                Waiting for: <span className="font-medium">{pendingItem.approver_id || 'Any approver'}</span>
                            </p>
                        )}
                        {/* Multi-step flow progress */}
                        {flowProgress && (
                            <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-600">
                                <div className="flex items-center gap-2 text-xs text-gray-500 mb-2">
                                    <Layers className="w-3.5 h-3.5" />
                                    <span>Multi-Step Approval Flow</span>
                                </div>
                                <StepProgressCompact
                                    currentStep={flowProgress.currentStep}
                                    totalSteps={flowProgress.totalSteps}
                                    currentStepName={flowProgress.stepName}
                                />
                            </div>
                        )}
                    </div>
                </div>

                {/* Action buttons for approvers */}
                {isApprover && !showCommentInput && (
                    <div className="flex items-center gap-2 flex-shrink-0">
                        <Button
                            variant="primary"
                            size="sm"
                            icon={<CheckCircle className="w-4 h-4" />}
                            onClick={() => handleAction('approve')}
                            disabled={processing}
                        >
                            Approve
                        </Button>
                        <Button
                            variant="danger"
                            size="sm"
                            icon={<XCircle className="w-4 h-4" />}
                            onClick={() => handleAction('reject')}
                            disabled={processing}
                        >
                            Reject
                        </Button>
                    </div>
                )}
            </div>

            {/* Comment input for approve/reject */}
            {showCommentInput && (
                <div className="mt-4 pt-4 border-t border-gray-200">
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        {actionType === 'approve' ? 'Approval' : 'Rejection'} Comments (optional)
                    </label>
                    <textarea
                        value={comments}
                        onChange={(e) => setComments(e.target.value)}
                        placeholder={`Add comments for your ${actionType}...`}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        rows={2}
                        autoFocus
                    />
                    <div className="flex justify-end gap-2 mt-3">
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={handleCancel}
                            disabled={processing}
                        >
                            Cancel
                        </Button>
                        <Button
                            variant={actionType === 'approve' ? 'primary' : 'danger'}
                            size="sm"
                            onClick={() => handleAction(actionType!)}
                            disabled={processing}
                            icon={actionType === 'approve'
                                ? <CheckCircle className="w-4 h-4" />
                                : <XCircle className="w-4 h-4" />
                            }
                        >
                            {processing
                                ? 'Processing...'
                                : actionType === 'approve' ? 'Confirm Approval' : 'Confirm Rejection'
                            }
                        </Button>
                    </div>
                </div>
            )}
        </div>
    );
}
