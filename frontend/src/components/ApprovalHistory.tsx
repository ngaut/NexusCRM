import React, { useState, useEffect } from 'react';
import { Clock, CheckCircle, XCircle, Send, User, MessageSquare } from 'lucide-react';
import { approvalsAPI, ApprovalWorkItem } from '../infrastructure/api/approvals';

interface ApprovalHistoryProps {
    objectApiName: string;
    recordId: string;
}

export function ApprovalHistory({ objectApiName, recordId }: ApprovalHistoryProps) {
    const [history, setHistory] = useState<ApprovalWorkItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchHistory = async () => {
            setLoading(true);
            try {
                const items = await approvalsAPI.getHistory(objectApiName, recordId);
                setHistory(items);
                setError(null);
            } catch (err: unknown) {
                const errMsg = err instanceof Error ? err.message : String(err);
                setError(errMsg);
            } finally {
                setLoading(false);
            }
        };
        fetchHistory();
    }, [objectApiName, recordId]);

    const formatDate = (dateStr: string) => {
        return new Date(dateStr).toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    const getStatusInfo = (status: string) => {
        switch (status) {
            case 'Approved':
                return { icon: CheckCircle, color: 'text-green-500', bg: 'bg-green-50' };
            case 'Rejected':
                return { icon: XCircle, color: 'text-red-500', bg: 'bg-red-50' };
            case 'Pending':
            default:
                return { icon: Clock, color: 'text-amber-500', bg: 'bg-amber-50' };
        }
    };

    if (loading) {
        return (
            <div className="text-sm text-gray-500 animate-pulse">
                Loading approval history...
            </div>
        );
    }

    if (error) {
        return (
            <div className="text-sm text-gray-400">
                No approval history available
            </div>
        );
    }

    if (history.length === 0) {
        return (
            <div className="text-sm text-gray-500 flex items-center gap-2">
                <Clock className="w-4 h-4" />
                No approval history
            </div>
        );
    }

    return (
        <div className="space-y-3">
            {history.map((item) => {
                const statusInfo = getStatusInfo(item.status);
                const StatusIcon = statusInfo.icon;

                return (
                    <div
                        key={item.id}
                        className={`p-3 rounded-lg border ${statusInfo.bg} border-gray-100`}
                    >
                        <div className="flex items-center justify-between mb-2">
                            <div className="flex items-center gap-2">
                                <StatusIcon className={`w-4 h-4 ${statusInfo.color}`} />
                                <span className={`text-sm font-medium ${statusInfo.color}`}>
                                    {item.status}
                                </span>
                            </div>
                            <span className="text-xs text-gray-400">
                                {formatDate(item.submitted_date)}
                            </span>
                        </div>

                        {item.comments && (
                            <div className="flex items-start gap-2 mt-2 text-xs text-gray-600">
                                <MessageSquare className="w-3 h-3 mt-0.5 text-gray-400" />
                                <span className="italic">"{item.comments}"</span>
                            </div>
                        )}

                        {item.approved_date && (
                            <div className="text-xs text-gray-400 mt-1">
                                {item.status === 'Approved' ? 'Approved' : 'Rejected'} on {formatDate(item.approved_date)}
                            </div>
                        )}
                    </div>
                );
            })}
        </div>
    );
}

/**
 * Get current approval status for display as a badge.
 * Also returns the pending work item if one exists (for inline actions).
 */
export function useApprovalStatus(objectApiName: string, recordId: string) {
    const [status, setStatus] = useState<'Pending' | 'Approved' | 'Rejected' | null>(null);
    const [pendingItem, setPendingItem] = useState<ApprovalWorkItem | null>(null);
    const [loading, setLoading] = useState(true);
    const [refreshKey, setRefreshKey] = useState(0);

    useEffect(() => {
        const fetchStatus = async () => {
            try {
                const items = await approvalsAPI.getHistory(objectApiName, recordId);
                if (items.length > 0) {
                    // Get most recent item
                    const latest = items[0];
                    setStatus(latest.status);
                    // If pending, expose the item for inline actions
                    if (latest.status === 'Pending') {
                        setPendingItem(latest);
                    } else {
                        setPendingItem(null);
                    }
                } else {
                    setStatus(null);
                    setPendingItem(null);
                }
            } catch {
                setStatus(null);
                setPendingItem(null);
            } finally {
                setLoading(false);
            }
        };
        fetchStatus();
    }, [objectApiName, recordId, refreshKey]);

    const refresh = () => setRefreshKey(k => k + 1);

    return { status, pendingItem, loading, refresh };
}
