import { api } from './client';
import { API_ENDPOINTS } from './endpoints';
import { COMMON_FIELDS } from '../../core/constants';
import type { User, Profile, ObjectPermission, FieldPermission } from '../../types';

export interface CreateUserPayload {
    [COMMON_FIELDS.NAME]: string;
    [COMMON_FIELDS.EMAIL]: string;
    [COMMON_FIELDS.PASSWORD]?: string;
    [COMMON_FIELDS.PROFILE_ID]: string;
}

export const usersAPI = {
    getUsers: async () => {
        const response = await api.get<{ users: User[] }>(API_ENDPOINTS.AUTH.USERS);
        return response.users;
    },

    getProfiles: async () => {
        const response = await api.get<{ profiles: Profile[] }>(API_ENDPOINTS.AUTH.PROFILES);
        return response.profiles;
    },

    createUser: (user: Partial<User>) => api.post<{ message: string; user: User }>(API_ENDPOINTS.AUTH.REGISTER, user),
    updateUser: (id: string, user: Partial<User>) => api.put<{ message: string }>(`${API_ENDPOINTS.AUTH.USERS}/${id}`, user),
    deleteUser: (id: string) => api.delete<{ message: string }>(`${API_ENDPOINTS.AUTH.USERS}/${id}`),

    // Profile CRUD
    createProfile: (profile: { name: string; description?: string }) =>
        api.post<{ message: string; profile: Profile }>(API_ENDPOINTS.AUTH.PROFILES, profile),

    // Permission operations (Return full response object to access .permissions property)
    getProfilePermissions: (profileId: string) => api.get<{ permissions: ObjectPermission[] }>(API_ENDPOINTS.AUTH.PROFILE_PERMISSIONS(profileId)),
    updateProfilePermissions: (profileId: string, permissions: ObjectPermission[]) => api.put<{ message: string }>(API_ENDPOINTS.AUTH.PROFILE_PERMISSIONS(profileId), permissions),

    getProfileFieldPermissions: (profileId: string) => api.get<{ permissions: FieldPermission[] }>(API_ENDPOINTS.AUTH.PROFILE_FIELD_PERMISSIONS(profileId)),
    updateProfileFieldPermissions: (profileId: string, permissions: FieldPermission[]) => api.put<{ message: string }>(API_ENDPOINTS.AUTH.PROFILE_FIELD_PERMISSIONS(profileId), permissions),

    // Permission Set permission operations
    getPermissionSetPermissions: (permSetId: string) => api.get<{ permissions: ObjectPermission[] }>(API_ENDPOINTS.AUTH.PERM_SET_PERMISSIONS(permSetId)),
    updatePermissionSetPermissions: (permSetId: string, permissions: ObjectPermission[]) => api.put<{ message: string }>(API_ENDPOINTS.AUTH.PERM_SET_PERMISSIONS(permSetId), permissions),

    getPermissionSetFieldPermissions: (permSetId: string) => api.get<{ permissions: FieldPermission[] }>(API_ENDPOINTS.AUTH.PERM_SET_FIELD_PERMISSIONS(permSetId)),
    updatePermissionSetFieldPermissions: (permSetId: string, permissions: FieldPermission[]) => api.put<{ message: string }>(API_ENDPOINTS.AUTH.PERM_SET_FIELD_PERMISSIONS(permSetId), permissions),

    // Effective Permissions (User)
    getUserEffectivePermissions: (userId: string) => api.get<{ permissions: ObjectPermission[] }>(API_ENDPOINTS.AUTH.USER_EFFECTIVE_PERMISSIONS(userId)),
    getUserEffectiveFieldPermissions: (userId: string) => api.get<{ permissions: FieldPermission[] }>(API_ENDPOINTS.AUTH.USER_EFFECTIVE_FIELD_PERMISSIONS(userId)),
};
