'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
    registerApi,
    type RegisterPayload
} from '@/features/auth/api/auth-api';
import { useAuthStore } from '@/features/auth/stores/auth-store';
import { ensureCsrf } from '@/lib/api-client';
import Link from 'next/link';

export default function SignUpPage() {
    const router = useRouter();
    const setUser = useAuthStore((s) => s.setUser);
    const [form, setForm] = useState<RegisterPayload>({
        email: '',
        password: '',
        name: ''
    });
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    useEffect(() => { ensureCsrf(); }, []);

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        setError('');
        setLoading(true);
        try {
            const res = await registerApi(form);
            const tokens = res.data || res;
            setUser(tokens.user);
            router.push('/dashboard/overview');
        } catch (err: unknown) {
            const axiosErr = err as { response?: { data?: { error?: string } } };
            setError(axiosErr.response?.data?.error || 'Registration failed');
        } finally {
            setLoading(false);
        }
    }

    return (
        <div className='flex min-h-screen items-center justify-center bg-gradient-to-br from-zinc-900 via-zinc-950 to-black p-4'>
            <div className='w-full max-w-md space-y-8'>
                <div className='text-center'>
                    <h1 className='text-foreground text-3xl font-bold tracking-tight'>
                        {process.env.NEXT_PUBLIC_APP_NAME || 'MyPaaS'}
                    </h1>
                    <p className='text-muted-foreground mt-2 text-sm'>
                        Create your account
                    </p>
                </div>

                <div className='border-border bg-card rounded-xl border p-8 shadow-2xl shadow-black/20'>
                    <form onSubmit={handleSubmit} className='space-y-5'>
                        {error && (
                            <div className='rounded-lg border border-red-500/20 bg-red-500/10 px-4 py-3 text-sm text-red-400'>
                                {error}
                            </div>
                        )}

                        <div className='space-y-2'>
                            <label
                                htmlFor='name'
                                className='text-foreground text-sm font-medium'
                            >
                                Name
                            </label>
                            <input
                                id='name'
                                type='text'
                                required
                                value={form.name}
                                onChange={(e) => setForm({ ...form, name: e.target.value })}
                                className='border-input bg-background text-foreground placeholder:text-muted-foreground focus:ring-ring w-full rounded-lg border px-4 py-2.5 text-sm transition focus:ring-2 focus:outline-none'
                                placeholder='Your name'
                            />
                        </div>

                        <div className='space-y-2'>
                            <label
                                htmlFor='email'
                                className='text-foreground text-sm font-medium'
                            >
                                Email
                            </label>
                            <input
                                id='email'
                                type='email'
                                required
                                value={form.email}
                                onChange={(e) => setForm({ ...form, email: e.target.value })}
                                className='border-input bg-background text-foreground placeholder:text-muted-foreground focus:ring-ring w-full rounded-lg border px-4 py-2.5 text-sm transition focus:ring-2 focus:outline-none'
                                placeholder='you@example.com'
                            />
                        </div>

                        <div className='space-y-2'>
                            <label
                                htmlFor='password'
                                className='text-foreground text-sm font-medium'
                            >
                                Password
                            </label>
                            <input
                                id='password'
                                type='password'
                                required
                                minLength={8}
                                value={form.password}
                                onChange={(e) => setForm({ ...form, password: e.target.value })}
                                className='border-input bg-background text-foreground placeholder:text-muted-foreground focus:ring-ring w-full rounded-lg border px-4 py-2.5 text-sm transition focus:ring-2 focus:outline-none'
                                placeholder='••••••••'
                            />
                        </div>

                        <button
                            type='submit'
                            disabled={loading}
                            className='bg-primary text-primary-foreground hover:bg-primary/90 w-full rounded-lg px-4 py-2.5 text-sm font-medium transition disabled:opacity-50'
                        >
                            {loading ? 'Creating account…' : 'Create Account'}
                        </button>
                    </form>

                    <p className='text-muted-foreground mt-6 text-center text-sm'>
                        Already have an account?{' '}
                        <Link
                            href='/auth/sign-in'
                            className='text-primary hover:underline'
                        >
                            Sign In
                        </Link>
                    </p>
                </div>
            </div>
        </div>
    );
}
