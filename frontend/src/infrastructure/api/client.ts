import { API_CONFIG } from '../../core/constants/EnvironmentConfig';
import { STORAGE_KEYS } from '../../core/constants/ApplicationDefaults';

export class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public data?: unknown
  ) {
    super(message);
    this.name = 'APIError';
  }
}

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  body?: unknown;
  headers?: Record<string, string>;
  requiresAuth?: boolean;
}

// Event bus for auth errors to avoid circular dependencies
export const authEvents = new EventTarget();
export const AUTH_EVENT_UNAUTHORIZED = 'auth:unauthorized';

class APIClient {
  private baseURL: string;
  private token: string | null = null;

  constructor() {
    this.baseURL = API_CONFIG.BACKEND_URL;
    // Try to restore token from localStorage
    this.token = localStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
  }

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, token);
    } else {
      localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
    }
  }

  getToken(): string | null {
    return this.token;
  }

  async request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
    const {
      method = 'GET',
      body,
      headers = {},
      requiresAuth = true
    } = options;

    const url = `${this.baseURL}${endpoint}`;

    const requestHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
      ...headers
    };

    // Add auth header if required and token exists
    if (requiresAuth && this.token) {
      requestHeaders['Authorization'] = `Bearer ${this.token}`;
    }

    const requestOptions: RequestInit = {
      method,
      headers: requestHeaders,
      credentials: 'include' // Include cookies for CORS
    };

    if (body && method !== 'GET') {
      requestOptions.body = JSON.stringify(body);
    }

    try {
      const response = await fetch(url, requestOptions);

      // Handle non-JSON responses
      const contentType = response.headers.get('content-type');
      const isJSON = contentType?.includes('application/json');

      if (!response.ok) {
        // Global handler for 401 Unauthorized
        if (response.status === 401) {
          authEvents.dispatchEvent(new Event(AUTH_EVENT_UNAUTHORIZED));
        }

        const errorData = isJSON ? await response.json() : { message: response.statusText };
        throw new APIError(
          errorData.message || errorData.error || `Request failed with status ${response.status}`,
          response.status,
          errorData
        );
      }

      // Return parsed JSON or null for 204 No Content
      if (response.status === 204) {
        return null as T;
      }

      return isJSON ? await response.json() : null as T;
    } catch (error) {
      if (error instanceof APIError) {
        throw error;
      }
      // Network error or other fetch failure
      throw new APIError(
        error instanceof Error ? error.message : 'Network request failed',
        0
      );
    }
  }

  // Convenience methods
  async get<T>(endpoint: string, requiresAuth = true): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET', requiresAuth });
  }

  async post<T>(endpoint: string, body?: unknown, requiresAuth = true): Promise<T> {
    return this.request<T>(endpoint, { method: 'POST', body, requiresAuth });
  }

  async put<T>(endpoint: string, body?: unknown, requiresAuth = true): Promise<T> {
    return this.request<T>(endpoint, { method: 'PUT', body, requiresAuth });
  }

  async patch<T>(endpoint: string, body?: unknown, requiresAuth = true): Promise<T> {
    return this.request<T>(endpoint, { method: 'PATCH', body, requiresAuth });
  }

  async delete<T>(endpoint: string, requiresAuth = true): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE', requiresAuth });
  }
}

// Singleton instance
export const apiClient = new APIClient();

// Also export as 'api' for convenience
export const api = apiClient;
