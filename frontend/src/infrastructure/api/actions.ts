import { apiClient } from './client';
import type { SObject } from '../../types';

export interface ExecuteActionRequest {
    recordId?: string;
    objectApiName?: string;
    contextRecord?: SObject;
    input?: Record<string, any>;
}

export interface ExecuteActionResponse {
    message: string;
    results?: Record<string, any>;
}

export interface ActionMetadata {
    id: string;
    object_api_name: string;
    name: string;
    label: string;
    type: string;
    icon: string;
    target_object?: string;
    config?: Record<string, any>;
}

export const actionAPI = {
    getActions: async (objectName: string): Promise<ActionMetadata[]> => {
        const response = await apiClient.get<{ actions: ActionMetadata[] }>(
            `/api/metadata/actions/${objectName}`
        );
        return response.actions;
    },

    executeAction: async (
        actionId: string,
        request: ExecuteActionRequest
    ): Promise<ExecuteActionResponse> => {
        const response = await apiClient.post<ExecuteActionResponse>(
            `/api/actions/execute/${actionId}`,
            request
        );
        return response;
    }
};
