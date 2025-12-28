/**
 * File Upload API Client
 * 
 * Provides file upload functionality with proper auth handling.
 */

import { API_CONFIG } from '../../core/constants/EnvironmentConfig';

export interface UploadedFile {
    path: string;
    name: string;
    size: number;
    mime: string;
}

export const filesAPI = {
    /**
     * Upload a file to the server
     * Uses FormData for multipart upload with auth token
     */
    async upload(file: File): Promise<UploadedFile> {
        const formData = new FormData();
        formData.append('file', file);

        const token = localStorage.getItem('auth_token');
        const headers: Record<string, string> = {};

        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        const response = await fetch(`${API_CONFIG.BACKEND_URL}/api/files/upload`, {
            method: 'POST',
            headers,
            body: formData,
            credentials: 'include'
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({ error: 'Upload failed' }));
            throw new Error(errorData.error || `Upload failed with status ${response.status}`);
        }

        return response.json();
    }
};
