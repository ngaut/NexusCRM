import { apiClient } from './client';
import { API_ENDPOINTS } from './endpoints';
import { COMMON_FIELDS } from '../../core/constants';
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
    [COMMON_FIELDS.ID]: string;
    id?: string; // Alias for [COMMON_FIELDS.ID]
    [COMMON_FIELDS.OBJECT_API_NAME]: string;
    [COMMON_FIELDS.NAME]: string;
    [COMMON_FIELDS.LABEL]: string;
    [COMMON_FIELDS.TYPE]: string;
    icon: string;
    target_object?: string;
    [COMMON_FIELDS.CONFIG]?: Record<string, any>;
}

export const actionAPI = {
    getActions: async (objectName: string): Promise<ActionMetadata[]> => {
        const response = await apiClient.get<{ data: ActionMetadata[] }>(
            API_ENDPOINTS.METADATA.ACTIONS(objectName)
        );
        return response.data;
    },

    executeAction: async (
        actionId: string,
        request: ExecuteActionRequest
    ): Promise<ExecuteActionResponse> => {
        const response = await apiClient.post<ExecuteActionResponse>(
            API_ENDPOINTS.ACTIONS.EXECUTE(actionId),
            request
        );
        return response;
    }
};
