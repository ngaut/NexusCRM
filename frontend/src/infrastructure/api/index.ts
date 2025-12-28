/**
 * Centralized API client for backend REST API
 *
 * Architecture:
 * - client.ts: Core fetch wrapper with JWT authentication
 * - auth.ts: Login, logout, register endpoints
 * - metadata.ts: Schema, app, layout endpoints
 * - data.ts: Query, CRUD, search, recycle bin endpoints
 *
 * Usage:
 * import { authAPI, metadataAPI, dataAPI } from '@/infrastructure/api';
 *
 * const user = await authAPI.login({ email, password });
 * const schemas = await metadataAPI.getSchemas();
 * const records = await dataAPI.query({ objectApiName: 'account', ... });
 */

export { apiClient, APIError } from './client';
export * from './flows';
export * from './feed';
export * from './analytics';
export type { RequestOptions } from './client';

export { authAPI } from './auth';
export type { LoginRequest, LoginResponse, RegisterRequest } from './auth';

export { metadataAPI } from './metadata';

export { dataAPI } from './data';
export type { QueryRequest } from './data';

export { filesAPI } from './files';
export type { UploadedFile } from './files';

export { agentApi } from './agent';
export type { ChatRequest, ChatResponse, ChatMessage } from './agent';
