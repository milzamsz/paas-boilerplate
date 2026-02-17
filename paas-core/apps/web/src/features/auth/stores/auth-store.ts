import { create } from 'zustand';
import type { User, Organization } from '@/types';
import { getProfile, logoutApi } from '@/features/auth/api/auth-api';
import { getAccessToken, setAccessToken } from '@/lib/api-client';
import { supabase, isSupabaseEnabled } from '@/lib/supabase-client';

interface AuthState {
    user: User | null;
    currentOrg: Organization | null;
    isAuthenticated: boolean;
    isLoading: boolean;

    setUser: (user: User | null) => void;
    setCurrentOrg: (org: Organization | null) => void;
    fetchUser: () => Promise<void>;
    logout: () => void;
    hydrate: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
    user: null,
    currentOrg: null,
    isAuthenticated: false,
    isLoading: true,

    setUser: (user) =>
        set({ user, isAuthenticated: !!user, isLoading: false }),

    setCurrentOrg: (org) => set({ currentOrg: org }),

    fetchUser: async () => {
        try {
            set({ isLoading: true });
            const { data } = await getProfile();
            set({ user: data, isAuthenticated: true, isLoading: false });
        } catch {
            set({ user: null, isAuthenticated: false, isLoading: false });
        }
    },

    logout: () => {
        set({ user: null, currentOrg: null, isAuthenticated: false });
        logoutApi();
    },

    hydrate: () => {
        if (isSupabaseEnabled()) {
            // Listen for Supabase auth state changes
            supabase.auth.getSession().then(({ data }) => {
                if (data.session) {
                    setAccessToken(data.session.access_token);
                    useAuthStore.getState().fetchUser();
                } else {
                    set({ isLoading: false });
                }
            });

            // Subscribe to future auth changes (token refresh, sign out, etc.)
            supabase.auth.onAuthStateChange((_event, session) => {
                if (session) {
                    setAccessToken(session.access_token);
                    useAuthStore.getState().fetchUser();
                } else {
                    set({ user: null, isAuthenticated: false, isLoading: false });
                }
            });
        } else {
            // Local auth: check for existing token
            const token = getAccessToken();
            if (token) {
                useAuthStore.getState().fetchUser();
            } else {
                set({ isLoading: false });
            }
        }
    },
}));
