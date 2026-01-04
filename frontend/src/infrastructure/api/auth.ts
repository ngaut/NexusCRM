import { apiClient, APIError } from './client';
import { API_ENDPOINTS } from './endpoints';
import type { ObjectPermission, FieldPermission } from '../../types';
import type { UserSession } from '../../types';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: UserSession;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export const authAPI = {
  /**
   * Login with email and password
   * @returns JWT token and user session
   */
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    try {
      const response = await apiClient.post<LoginResponse>(
        API_ENDPOINTS.AUTH.LOGIN,
        credentials,
        false // Login doesn't require auth
      );

      // Store token in client
      if (response.token) {
        apiClient.setToken(response.token);
      }

      return response;
    } catch (error) {
      if (error instanceof APIError) {
        // Rethrow with user-friendly message
        const apiErrorData = error.data as { message?: string } | undefined;
        throw new Error(apiErrorData?.message || 'Login failed. Please check your credentials.');
      }
      throw error;
    }
  },

  /**
   * Logout current user
   */
  async logout(): Promise<void> {
    try {
      await apiClient.post(API_ENDPOINTS.AUTH.LOGOUT, {});
    } finally {
      // Always clear token, even if request fails
      apiClient.setToken(null);
    }
  },

  /**
   * Register new user
   */
  async register(data: RegisterRequest): Promise<LoginResponse> {
    const response = await apiClient.post<LoginResponse>(
      API_ENDPOINTS.AUTH.REGISTER,
      data,
      false // Registration doesn't require auth
    );

    // Store token in client
    if (response.token) {
      apiClient.setToken(response.token);
    }

    return response;
  },

  /**
   * Verify if current token is valid
   */
  async verify(): Promise<UserSession> {
    const response = await apiClient.get<{ success: boolean; user: UserSession }>(API_ENDPOINTS.AUTH.ME);
    return response.user;
  },

  /**
   * Get permissions for the current user
   */
  async getMyPermissions(): Promise<{ objectPermissions: ObjectPermission[], fieldPermissions: FieldPermission[] }> {
    const response = await apiClient.get<{ objectPermissions: ObjectPermission[], fieldPermissions?: FieldPermission[] }>(API_ENDPOINTS.AUTH.PERMISSIONS);
    return {
      objectPermissions: response.objectPermissions || [],
      fieldPermissions: response.fieldPermissions || []
    };
  },

  /**
   * Check if user is authenticated (has valid token)
   */
  isAuthenticated(): boolean {
    return apiClient.getToken() !== null;
  }
};
