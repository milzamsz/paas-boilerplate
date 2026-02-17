'use client';

import PageContainer from '@/components/layout/page-container';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { useState } from 'react';
import { apiClient } from '@/lib/api-client';

export default function ProfilePage() {
    const { user, setUser } = useAuth();
    const [name, setName] = useState(user?.name || '');
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState('');

    async function handleSave(e: React.FormEvent) {
        e.preventDefault();
        setSaving(true);
        setMessage('');
        try {
            const res = await apiClient.put('/users/me', { name });
            setUser(res.data?.data || res.data);
            setMessage('Profile updated successfully');
        } catch {
            setMessage('Failed to update profile');
        } finally {
            setSaving(false);
        }
    }

    return (
        <PageContainer title='Profile' description='Manage your account settings'>
            <div className='border-border bg-card max-w-lg rounded-xl border p-6'>
                <form onSubmit={handleSave} className='space-y-5'>
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
                            disabled
                            value={user?.email || ''}
                            className='border-input bg-muted text-muted-foreground w-full rounded-lg border px-4 py-2.5 text-sm'
                        />
                    </div>

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
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            className='border-input bg-background text-foreground focus:ring-ring w-full rounded-lg border px-4 py-2.5 text-sm transition focus:ring-2 focus:outline-none'
                        />
                    </div>

                    <div className='space-y-2'>
                        <label className='text-foreground text-sm font-medium'>Role</label>
                        <p className='text-muted-foreground text-sm'>{user?.role || '—'}</p>
                    </div>

                    {message && (
                        <p
                            className={`text-sm ${message.includes('success') ? 'text-emerald-500' : 'text-red-400'}`}
                        >
                            {message}
                        </p>
                    )}

                    <button
                        type='submit'
                        disabled={saving}
                        className='bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg px-4 py-2.5 text-sm font-medium transition disabled:opacity-50'
                    >
                        {saving ? 'Saving…' : 'Save Changes'}
                    </button>
                </form>
            </div>
        </PageContainer>
    );
}
