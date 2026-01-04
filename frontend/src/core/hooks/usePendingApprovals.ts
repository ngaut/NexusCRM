import { useState, useEffect, useCallback } from 'react';
import { approvalsAPI, ApprovalWorkItem } from '../../infrastructure/api/approvals';
import { UI_TIMING } from '../constants';

interface UsePendingApprovalsReturn {
    count: number;
    items: ApprovalWorkItem[];
    loading: boolean;
    error: string | null;
    refresh: () => void;
}

/**
 * Hook to fetch pending approval count for current user.
 * Auto-refreshes every 60 seconds.
 */
export function usePendingApprovals(): UsePendingApprovalsReturn {
    const [items, setItems] = useState<ApprovalWorkItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchPending = useCallback(async () => {
        try {
            const result = await approvalsAPI.getPending();
            setItems(result);
            setError(null);
        } catch (err: unknown) {
            // Silently fail for non-critical badge
            setError(err instanceof Error ? err.message : String(err));
            setItems([]);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchPending();

        // Auto-refresh every 60 seconds
        const interval = setInterval(fetchPending, UI_TIMING.POLLING_NORMAL_MS);

        return () => clearInterval(interval);
    }, [fetchPending]);

    return {
        count: items.length,
        items,
        loading,
        error,
        refresh: fetchPending,
    };
}
