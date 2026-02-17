'use client';

import { useAuthStore } from '@/features/auth/stores/auth-store';

export function useAuth() {
    const user = useAuthStore((s) => s.user);
    const currentOrg = useAuthStore((s) => s.currentOrg);
    const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
    const isLoading = useAuthStore((s) => s.isLoading);
    const setUser = useAuthStore((s) => s.setUser);
    const setCurrentOrg = useAuthStore((s) => s.setCurrentOrg);
    const fetchUser = useAuthStore((s) => s.fetchUser);
    const logout = useAuthStore((s) => s.logout);

    return {
        user,
        currentOrg,
        isAuthenticated,
        isLoading,
        setUser,
        setCurrentOrg,
        fetchUser,
        logout,
        hasRole: (role: string) => user?.role === role,
        isAdmin: user?.role === 'admin',
        isOwner: currentOrg?.owner_id === user?.id
    };
}
