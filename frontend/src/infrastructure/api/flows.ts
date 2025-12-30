import { apiClient } from './client';
import { API_ENDPOINTS } from './endpoints';
import { COMMON_FIELDS } from '../../core/constants';
import { FLOW_STATUS, FlowStatus } from '../../core/constants/FlowConstants';

export interface FlowStep {
    [COMMON_FIELDS.ID]: string;
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
        const response = await apiClient.get<{ flows: Flow[] }>(API_ENDPOINTS.METADATA.FLOWS);
        return response.flows || [];
    },

    async getById(flowId: string): Promise<Flow> {
        return apiClient.get(API_ENDPOINTS.METADATA.FLOW(flowId));
    },

    async create(flow: Omit<Flow, 'id' | 'lastModified'>): Promise<Flow> {
        return apiClient.post(API_ENDPOINTS.METADATA.FLOWS, flow);
    },

    async update(flowId: string, updates: Partial<Flow>): Promise<Flow> {
        return apiClient.patch(API_ENDPOINTS.METADATA.FLOW(flowId), updates);
    },

    async delete(flowId: string): Promise<void> {
        return apiClient.delete(API_ENDPOINTS.METADATA.FLOW(flowId));
    },

    async toggleStatus(flowId: string, currentStatus: string): Promise<Flow> {
        const newStatus = currentStatus === FLOW_STATUS.ACTIVE ? FLOW_STATUS.DRAFT : FLOW_STATUS.ACTIVE;
        return apiClient.patch(API_ENDPOINTS.METADATA.FLOW(flowId), { status: newStatus });
    },

    /**
     * Execute an auto-launched flow (admin only)
     * @param flowId - ID of the flow to execute
     * @param request - Execution context including record_id and field values
     */
    async execute(flowId: string, request: ExecuteFlowRequest = {}): Promise<ExecuteFlowResponse> {
        return apiClient.post<ExecuteFlowResponse>(
            API_ENDPOINTS.FLOWS.EXECUTE(flowId),
            request
        );
    },
};

