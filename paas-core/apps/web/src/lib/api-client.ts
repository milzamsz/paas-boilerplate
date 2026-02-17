import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios';

const API_BASE_URL =
    process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const apiClient = axios.create({
    baseURL: `${API_BASE_URL}/api/v1`,
    headers: { 'Content-Type': 'application/json' },
    withCredentials: true,
    timeout: 15000
});

// ── CSRF helpers ─────────────────────────────────────────────────
function getCookie(name: string): string | null {
    if (typeof document === 'undefined') return null;
    const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
    return match ? decodeURIComponent(match[1]) : null;
}

/** Pre-fetch a CSRF cookie from the API so the first POST succeeds. */
export async function ensureCsrf() {
    if (!getCookie('__csrf_token')) {
        await apiClient.get('/billing/plans'); // lightweight public GET
    }
}

// ── Token management ──────────────────────────────────────────────
let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
    accessToken = token;
    if (typeof window !== 'undefined') {
        if (token) {
            localStorage.setItem('access_token', token);
        } else {
            localStorage.removeItem('access_token');
        }
    }
}

export function getAccessToken(): string | null {
    if (accessToken) return accessToken;
    if (typeof window !== 'undefined') {
        accessToken = localStorage.getItem('access_token');
    }
    return accessToken;
}

export function clearTokens() {
    accessToken = null;
    if (typeof window !== 'undefined') {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
    }
}

// ── Request interceptor — attach access token + CSRF ──────────────
apiClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    const token = getAccessToken();
    if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    // Double-submit CSRF: read cookie → send as header
    const csrf = getCookie('__csrf_token');
    if (csrf && config.headers) {
        config.headers['X-CSRF-Token'] = csrf;
    }
    return config;
});

// ── Response interceptor — auto-refresh on 401 ────────────────────
let isRefreshing = false;
let failedQueue: Array<{
    resolve: (token: string) => void;
    reject: (err: unknown) => void;
}> = [];

function processQueue(error: unknown, token: string | null = null) {
    failedQueue.forEach((p) => {
        if (error) {
            p.reject(error);
        } else if (token) {
            p.resolve(token);
        }
    });
    failedQueue = [];
}

apiClient.interceptors.response.use(
    (response) => response,
    async (error: AxiosError) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & {
            _retry?: boolean;
        };

        if (error.response?.status === 401 && !originalRequest._retry) {
            if (isRefreshing) {
                return new Promise((resolve, reject) => {
                    failedQueue.push({
                        resolve: (token: string) => {
                            if (originalRequest.headers) {
                                originalRequest.headers.Authorization = `Bearer ${token}`;
                            }
                            resolve(apiClient(originalRequest));
                        },
                        reject
                    });
                });
            }

            originalRequest._retry = true;
            isRefreshing = true;

            try {
                const refreshToken =
                    typeof window !== 'undefined'
                        ? localStorage.getItem('refresh_token')
                        : null;

                if (!refreshToken) {
                    throw new Error('No refresh token');
                }

                const { data } = await axios.post(
                    `${API_BASE_URL}/api/v1/auth/refresh`,
                    { refresh_token: refreshToken },
                    { headers: { 'Content-Type': 'application/json' } }
                );

                const newAccessToken = data.data?.access_token || data.access_token;
                const newRefreshToken = data.data?.refresh_token || data.refresh_token;

                setAccessToken(newAccessToken);
                if (newRefreshToken && typeof window !== 'undefined') {
                    localStorage.setItem('refresh_token', newRefreshToken);
                }

                processQueue(null, newAccessToken);

                if (originalRequest.headers) {
                    originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
                }
                return apiClient(originalRequest);
            } catch (refreshError) {
                processQueue(refreshError, null);
                clearTokens();
                if (typeof window !== 'undefined') {
                    window.location.href = '/auth/sign-in';
                }
                return Promise.reject(refreshError);
            } finally {
                isRefreshing = false;
            }
        }

        return Promise.reject(error);
    }
);
