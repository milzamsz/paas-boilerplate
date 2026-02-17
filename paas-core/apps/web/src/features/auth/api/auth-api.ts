import { apiClient, setAccessToken, clearTokens } from '@/lib/api-client';
import { supabase, isSupabaseEnabled } from '@/lib/supabase-client';
import type { ApiResponse, User } from '@/types';

export interface LoginPayload {
    email: string;
    password: string;
}

export interface RegisterPayload {
    email: string;
    password: string;
    name: string;
}

export interface AuthTokens {
    access_token: string;
    refresh_token: string;
    user: User;
}

/**
 * Login via the Go API. When Supabase is enabled, the Go API proxies to
 * Supabase GoTrue; when local, it uses bcrypt + HS256 JWT.
 */
export async function loginApi(
    payload: LoginPayload
): Promise<ApiResponse<AuthTokens>> {
    if (isSupabaseEnabled()) {
        // Use Supabase JS client directly for auth session management
        const { data: sbData, error } = await supabase.auth.signInWithPassword({
            email: payload.email,
            password: payload.password,
        });
        if (error) throw new Error(error.message);

        const session = sbData.session;
        if (!session) throw new Error('No session returned');

        setAccessToken(session.access_token);

        // Also call the Go API to sync the profile
        const profile = await getProfile();

        return {
            data: {
                access_token: session.access_token,
                refresh_token: session.refresh_token,
                user: profile.data,
            },
        } as ApiResponse<AuthTokens>;
    }

    // Default: local auth via Go API
    const { data } = await apiClient.post('/auth/login', payload);
    const tokens = data.data || data;
    setAccessToken(tokens.access_token);
    if (typeof window !== 'undefined' && tokens.refresh_token) {
        localStorage.setItem('refresh_token', tokens.refresh_token);
    }
    return data;
}

/**
 * Register via the Go API. Same proxy behavior as login.
 */
export async function registerApi(
    payload: RegisterPayload
): Promise<ApiResponse<AuthTokens>> {
    if (isSupabaseEnabled()) {
        const { data: sbData, error } = await supabase.auth.signUp({
            email: payload.email,
            password: payload.password,
            options: {
                data: { name: payload.name }
            },
        });
        if (error) throw new Error(error.message);

        const session = sbData.session;
        if (!session) throw new Error('Signup successful but no session (check email confirmation settings)');

        setAccessToken(session.access_token);

        const profile = await getProfile();

        return {
            data: {
                access_token: session.access_token,
                refresh_token: session.refresh_token,
                user: profile.data,
            },
        } as ApiResponse<AuthTokens>;
    }

    // Default: local auth via Go API
    const { data } = await apiClient.post('/auth/register', payload);
    const tokens = data.data || data;
    setAccessToken(tokens.access_token);
    if (typeof window !== 'undefined' && tokens.refresh_token) {
        localStorage.setItem('refresh_token', tokens.refresh_token);
    }
    return data;
}

export async function getProfile(): Promise<ApiResponse<User>> {
    const { data } = await apiClient.get('/users/me');
    return data;
}

export async function logoutApi() {
    if (isSupabaseEnabled()) {
        await supabase.auth.signOut();
    }
    clearTokens();
    if (typeof window !== 'undefined') {
        window.location.href = '/auth/sign-in';
    }
}
