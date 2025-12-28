import { apiClient } from './client';
import type { SObject } from '../../types';

// ============================================================================
// Types
// ============================================================================

export interface ApprovalWorkItem {
    id: string;
    process_id: string;
    object_api_name: string;
    record_id: string;
    status: 'Pending' | 'Approved' | 'Rejected';
    submitted_by_id: string;
    submitted_date: string;
    approver_id?: string;
    approved_by_id?: string;
    approved_date?: string;
    comments?: string;
    created_date?: string;
    // Multi-step flow fields
    flow_instance_id?: string;
    flow_step_id?: string;
}

export interface SubmitApprovalRequest {
    object_api_name: string;
    record_id: string;
    comments?: string;
}

export interface ApprovalActionRequest {
    comments?: string;
}

export interface SubmitApprovalResponse {
    success: boolean;
    message: string;
    work_item_id: string;
}

export interface ApprovalActionResponse {
    success: boolean;
    message: string;
}

// ============================================================================
// API Client
// ============================================================================

export const approvalsAPI = {
    /**
     * Submit a record for approval
     */
    async submit(request: SubmitApprovalRequest): Promise<SubmitApprovalResponse> {
        return apiClient.post<SubmitApprovalResponse>('/api/approvals/submit', request);
    },

    /**
     * Approve a pending work item
     */
    async approve(workItemId: string, comments?: string): Promise<ApprovalActionResponse> {
        return apiClient.post<ApprovalActionResponse>(
            `/api/approvals/${encodeURIComponent(workItemId)}/approve`,
            { comments }
        );
    },

    /**
     * Reject a pending work item
     */
    async reject(workItemId: string, comments?: string): Promise<ApprovalActionResponse> {
        return apiClient.post<ApprovalActionResponse>(
            `/api/approvals/${encodeURIComponent(workItemId)}/reject`,
            { comments }
        );
    },

    /**
     * Get pending approvals for current user
     */
    async getPending(): Promise<ApprovalWorkItem[]> {
        const response = await apiClient.get<{ work_items: ApprovalWorkItem[] }>(
            '/api/approvals/pending'
        );
        return response.work_items || [];
    },

    /**
     * Get approval history for a specific record
     */
    async getHistory(objectApiName: string, recordId: string): Promise<ApprovalWorkItem[]> {
        const response = await apiClient.get<{ work_items: ApprovalWorkItem[] }>(
            `/api/approvals/history/${encodeURIComponent(objectApiName)}/${encodeURIComponent(recordId)}`
        );
        return response.work_items || [];
    },

    /**
     * Get flow instance details including step progress
     */
    async getFlowInstanceProgress(flowInstanceId: string): Promise<FlowInstanceProgress | null> {
        try {
            const response = await apiClient.get<FlowInstanceProgress>(
                `/api/approvals/flow-progress/${encodeURIComponent(flowInstanceId)}`
            );
            return response;
        } catch {
            return null;
        }
    },

    /**
     * Check if an approval process exists for an object
     * Used to conditionally show/hide the "Submit for Approval" button
     */
    async hasProcessForObject(objectApiName: string): Promise<boolean> {
        try {
            const response = await apiClient.get<{ has_process: boolean; process_name?: string }>(
                `/api/approvals/check/${encodeURIComponent(objectApiName)}`
            );
            return response.has_process ?? false;
        } catch {
            return false;
        }
    },
};

// Flow Instance Progress for multi-step approvals
export interface FlowInstanceProgress {
    id: string;
    flow_id: string;
    status: 'Running' | 'Paused' | 'Completed' | 'Failed';
    current_step_id?: string;
    current_step_order?: number;
    total_steps?: number;
    steps?: FlowStepProgress[];
}

export interface FlowStepProgress {
    id: string;
    step_order: number;
    step_name: string;
    step_type: 'action' | 'approval' | 'decision';
    status: 'pending' | 'completed' | 'current' | 'skipped';
}
