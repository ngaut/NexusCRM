import { apiClient } from './client';
import { API_ENDPOINTS } from './endpoints';
import { COMMON_FIELDS } from '../../core/constants';
import { FLOW_STATUS, FlowStatus } from '../../core/constants/FlowConstants';

export interface FlowStep {
    [COMMON_FIELDS.ID]: string;
    id?: string; // Alias for [COMMON_FIELDS.ID] which computes to __sys_gen_id
    [COMMON_FIELDS.FLOW_ID]?: string;
    [COMMON_FIELDS.STEP_ORDER]: number;
    [COMMON_FIELDS.STEP_NAME]: string;
    [COMMON_FIELDS.STEP_TYPE]: string;
    [COMMON_FIELDS.ACTION_TYPE]?: string;
    action_config?: Record<string, unknown>;
    entry_condition?: string;
    [COMMON_FIELDS.ON_SUCCESS_STEP]?: string;
    [COMMON_FIELDS.ON_FAILURE_STEP]?: string;
}

export interface Flow {
    [COMMON_FIELDS.ID]: string;
    id?: string; // Alias for [COMMON_FIELDS.ID] which computes to __sys_gen_id
    [COMMON_FIELDS.NAME]: string;
    [COMMON_FIELDS.STATUS]: FlowStatus;
    [COMMON_FIELDS.TRIGGER_OBJECT]: string;
    [COMMON_FIELDS.TRIGGER_TYPE]: string;
    trigger_condition: string;
    [COMMON_FIELDS.ACTION_TYPE]: string;
    action_config: Record<string, unknown>;
    flow_type: 'simple' | 'multistep';
    steps?: FlowStep[];
    [COMMON_FIELDS.LAST_MODIFIED_DATE]: string;
}

// ============================================================================
// Execute Flow Types
// ============================================================================

export interface ExecuteFlowRequest {
    [COMMON_FIELDS.RECORD_ID]?: string;
    [COMMON_FIELDS.OBJECT_API_NAME]?: string;
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
        const response = await apiClient.get<{ data: Flow[] }>(API_ENDPOINTS.METADATA.FLOWS);
        return response.data;
    },

    async getById(flowId: string): Promise<Flow> {
        const response = await apiClient.get<{ data: Flow }>(API_ENDPOINTS.METADATA.FLOW(flowId));
        return response.data;
    },

    async create(flow: Omit<Flow, 'id' | 'lastModified'>): Promise<Flow> {
        const response = await apiClient.post<{ data: Flow }>(API_ENDPOINTS.METADATA.FLOWS, flow);
        return response.data;
    },

    async update(flowId: string, updates: Partial<Flow>): Promise<Flow> {
        const response = await apiClient.patch<{ data: Flow }>(API_ENDPOINTS.METADATA.FLOW(flowId), updates);
        return response.data;
    },

    async delete(flowId: string): Promise<void> {
        return apiClient.delete(API_ENDPOINTS.METADATA.FLOW(flowId));
    },

    async toggleStatus(flowId: string, currentStatus: string): Promise<Flow> {
        const newStatus = currentStatus === FLOW_STATUS.ACTIVE ? FLOW_STATUS.DRAFT : FLOW_STATUS.ACTIVE;
        const response = await apiClient.patch<{ data: Flow }>(API_ENDPOINTS.METADATA.FLOW(flowId), { status: newStatus });
        return response.data;
    },

    /**
     * Execute an auto-launched flow (admin only)
     * @param flowId - ID of the flow to execute
     * @param request - Execution context including record_id and field values
     */
    async execute(flowId: string, request: ExecuteFlowRequest = {}): Promise<ExecuteFlowResponse> {
        const response = await apiClient.post<{ data: ExecuteFlowResponse }>(
            API_ENDPOINTS.FLOWS.EXECUTE(flowId),
            request
        );
        return response.data;
    },
};

