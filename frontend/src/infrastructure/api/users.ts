import { api } from './client';
import type { User, Profile, ObjectPermission, FieldPermission } from '../../types';

export interface CreateUserPayload {
    name: string;
    email: string;
    password?: string;
    profile_id: string;
}

export const usersAPI = {
    getUsers: async () => {
        const response = await api.get<{ users: User[] }>('/api/auth/users');
        return response.users;
    },

    getProfiles: async () => {
        const response = await api.get<{ profiles: Profile[] }>('/api/auth/profiles');
        return response.profiles;
    },

    createUser: (user: Partial<User>) => api.post<{ message: string; user: User }>('/api/auth/register', user),
    updateUser: (id: string, user: Partial<User>) => api.put<{ message: string }>(`/api/auth/users/${id}`, user),
    deleteUser: (id: string) => api.delete<{ message: string }>(`/api/auth/users/${id}`),

    // Permission operations (Return full response object to access .permissions property)
    getProfilePermissions: (profileId: string) => api.get<{ permissions: ObjectPermission[] }>(`/api/auth/profiles/${profileId}/permissions`),
    updateProfilePermissions: (profileId: string, permissions: ObjectPermission[]) => api.put<{ message: string }>(`/api/auth/profiles/${profileId}/permissions`, permissions),

    getProfileFieldPermissions: (profileId: string) => api.get<{ permissions: FieldPermission[] }>(`/api/auth/profiles/${profileId}/permissions/fields`),
    updateProfileFieldPermissions: (profileId: string, permissions: FieldPermission[]) => api.put<{ message: string }>(`/api/auth/profiles/${profileId}/permissions/fields`, permissions),

    // Permission Set permission operations
    getPermissionSetPermissions: (permSetId: string) => api.get<{ permissions: ObjectPermission[] }>(`/api/auth/permission-sets/${permSetId}/permissions`),
    updatePermissionSetPermissions: (permSetId: string, permissions: ObjectPermission[]) => api.put<{ message: string }>(`/api/auth/permission-sets/${permSetId}/permissions`, permissions),

    getPermissionSetFieldPermissions: (permSetId: string) => api.get<{ permissions: FieldPermission[] }>(`/api/auth/permission-sets/${permSetId}/permissions/fields`),
    updatePermissionSetFieldPermissions: (permSetId: string, permissions: FieldPermission[]) => api.put<{ message: string }>(`/api/auth/permission-sets/${permSetId}/permissions/fields`, permissions),

    // Effective Permissions (User)
    getUserEffectivePermissions: (userId: string) => api.get<{ permissions: ObjectPermission[] }>(`/api/auth/users/${userId}/permissions/effective`),
    getUserEffectiveFieldPermissions: (userId: string) => api.get<{ permissions: FieldPermission[] }>(`/api/auth/users/${userId}/permissions/fields/effective`),
};
