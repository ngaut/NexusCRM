import { apiClient } from './client';
import { FLOW_STATUS, FlowStatus } from '../../core/constants/FlowConstants';

export interface FlowStep {
    id: string;
    flow_id?: string;
    step_order: number;
    step_name: string;
    step_type: string;
    action_type?: string;
    action_config?: Record<string, unknown>;
    entry_condition?: string;
    on_success_step?: string;
    on_failure_step?: string;
}

export interface Flow {
    id: string;
    name: string;
    status: FlowStatus;
    trigger_object: string;
    trigger_type: string;
    trigger_condition: string;
    action_type: string;
    action_config: Record<string, unknown>;
    flow_type: 'simple' | 'multistep';
    steps?: FlowStep[];
    last_modified: string;
}

// ============================================================================
// Execute Flow Types
// ============================================================================

export interface ExecuteFlowRequest {
    record_id?: string;
    object_api_name?: string;
    context?: Record<string, any>;
}

export interface ExecuteFlowResponse {
    success: boolean;
    flow_id: string;
    message: string;
    result?: {
        created_id?: string;
        updated_id?: string;
        target_object?: string;
        action_type?: string;
        [key: string]: unknown;
    };
}

// ============================================================================
// API Client
// ============================================================================

export const flowsApi = {
    async getAll(): Promise<Flow[]> {
        const response = await apiClient.get<{ flows: Flow[] }>('/api/metadata/flows');
        return response.flows || [];
    },

    async getById(flowId: string): Promise<Flow> {
        return apiClient.get(`/api/metadata/flows/${flowId}`);
    },

    async create(flow: Omit<Flow, 'id' | 'lastModified'>): Promise<Flow> {
        return apiClient.post('/api/metadata/flows', flow);
    },

    async update(flowId: string, updates: Partial<Flow>): Promise<Flow> {
        return apiClient.patch(`/api/metadata/flows/${flowId}`, updates);
    },

    async delete(flowId: string): Promise<void> {
        return apiClient.delete(`/api/metadata/flows/${flowId}`);
    },

    async toggleStatus(flowId: string, currentStatus: string): Promise<Flow> {
        const newStatus = currentStatus === FLOW_STATUS.ACTIVE ? FLOW_STATUS.DRAFT : FLOW_STATUS.ACTIVE;
        return apiClient.patch(`/api/metadata/flows/${flowId}`, { status: newStatus });
    },

    /**
     * Execute an auto-launched flow (admin only)
     * @param flowId - ID of the flow to execute
     * @param request - Execution context including record_id and field values
     */
    async execute(flowId: string, request: ExecuteFlowRequest = {}): Promise<ExecuteFlowResponse> {
        return apiClient.post<ExecuteFlowResponse>(
            `/api/flows/${encodeURIComponent(flowId)}/execute`,
            request
        );
    },
};

