import { apiClient } from './client';
import { API_ENDPOINTS } from './endpoints';
import { COMMON_FIELDS } from '../../core/constants';
import type { SObject } from '../../types';

// ============================================================================
// Types
// ============================================================================

export interface ApprovalWorkItem {
    [COMMON_FIELDS.ID]: string;
    [COMMON_FIELDS.PROCESS_ID]: string;
    [COMMON_FIELDS.OBJECT_API_NAME]: string;
    [COMMON_FIELDS.RECORD_ID]: string;
    status: 'Pending' | 'Approved' | 'Rejected';
    submitted_by_id: string;
    [COMMON_FIELDS.SUBMITTED_DATE]: string;
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
        return apiClient.post<SubmitApprovalResponse>(API_ENDPOINTS.APPROVALS.SUBMIT, request);
    },

    /**
     * Approve a pending work item
     */
    async approve(workItemId: string, comments?: string): Promise<ApprovalActionResponse> {
        return apiClient.post<ApprovalActionResponse>(
            API_ENDPOINTS.APPROVALS.APPROVE(workItemId),
            { comments }
        );
    },

    /**
     * Reject a pending work item
     */
    async reject(workItemId: string, comments?: string): Promise<ApprovalActionResponse> {
        return apiClient.post<ApprovalActionResponse>(
            API_ENDPOINTS.APPROVALS.REJECT(workItemId),
            { comments }
        );
    },

    /**
     * Get pending approvals for current user
     */
    async getPending(): Promise<ApprovalWorkItem[]> {
        const response = await apiClient.get<{ work_items: ApprovalWorkItem[] }>(
            API_ENDPOINTS.APPROVALS.PENDING
        );
        return response.work_items || [];
    },

    /**
     * Get approval history for a specific record
     */
    async getHistory(objectApiName: string, recordId: string): Promise<ApprovalWorkItem[]> {
        const response = await apiClient.get<{ work_items: ApprovalWorkItem[] }>(
            API_ENDPOINTS.APPROVALS.HISTORY(objectApiName, recordId)
        );
        return response.work_items || [];
    },

    /**
     * Get flow instance details including step progress
     */
    async getFlowInstanceProgress(flowInstanceId: string): Promise<FlowInstanceProgress | null> {
        try {
            const response = await apiClient.get<FlowInstanceProgress>(
                API_ENDPOINTS.APPROVALS.FLOW_PROGRESS(flowInstanceId)
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
                API_ENDPOINTS.APPROVALS.CHECK(objectApiName)
            );
            return response.has_process ?? false;
        } catch {
            return false;
        }
    },
};

// Flow Instance Progress for multi-step approvals
export interface FlowInstanceProgress {
    [COMMON_FIELDS.ID]: string;
    [COMMON_FIELDS.FLOW_ID]: string;
    [COMMON_FIELDS.STATUS]: 'Running' | 'Paused' | 'Completed' | 'Failed';
    [COMMON_FIELDS.CURRENT_STEP_ID]?: string;
    current_step_order?: number;
    total_steps?: number;
    steps?: FlowStepProgress[];
}

export interface FlowStepProgress {
    [COMMON_FIELDS.ID]: string;
    [COMMON_FIELDS.STEP_ORDER]: number;
    [COMMON_FIELDS.STEP_NAME]: string;
    [COMMON_FIELDS.STEP_TYPE]: 'action' | 'approval' | 'decision';
    [COMMON_FIELDS.STATUS]: 'pending' | 'completed' | 'current' | 'skipped';
}
