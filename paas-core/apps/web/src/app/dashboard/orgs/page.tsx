'use client';

import PageContainer from '@/components/layout/page-container';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { useEffect, useState } from 'react';
import { apiClient } from '@/lib/api-client';
import type { Organization } from '@/types';
import Link from 'next/link';
import { IconPlus } from '@tabler/icons-react';

export default function OrgsPage() {
    const { setCurrentOrg } = useAuth();
    const [orgs, setOrgs] = useState<Organization[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        apiClient
            .get('/orgs')
            .then((res) => setOrgs(res.data?.data || res.data || []))
            .catch(() => setOrgs([]))
            .finally(() => setLoading(false));
    }, []);

    return (
        <PageContainer
            title='Organizations'
            description='Manage your organizations and teams'
            actions={
                <Link
                    href='/dashboard/orgs/new'
                    className='bg-primary text-primary-foreground hover:bg-primary/90 inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition'
                >
                    <IconPlus className='h-4 w-4' />
                    New Organization
                </Link>
            }
        >
            {loading ? (
                <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-3'>
                    {[1, 2, 3].map((i) => (
                        <div
                            key={i}
                            className='border-border bg-card h-32 animate-pulse rounded-xl border'
                        />
                    ))}
                </div>
            ) : orgs.length === 0 ? (
                <div className='border-border bg-card flex flex-col items-center justify-center rounded-xl border py-16'>
                    <p className='text-muted-foreground mb-4 text-sm'>
                        No organizations yet
                    </p>
                    <Link
                        href='/dashboard/orgs/new'
                        className='bg-primary text-primary-foreground hover:bg-primary/90 inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium'
                    >
                        <IconPlus className='h-4 w-4' />
                        Create your first organization
                    </Link>
                </div>
            ) : (
                <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-3'>
                    {orgs.map((org) => (
                        <button
                            key={org.id}
                            onClick={() => setCurrentOrg(org)}
                            className='border-border bg-card hover:border-primary/50 rounded-xl border p-6 text-left transition-all hover:shadow-md'
                        >
                            <div className='flex items-center gap-3'>
                                <div className='flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 text-sm font-bold text-white'>
                                    {org.name.charAt(0).toUpperCase()}
                                </div>
                                <div>
                                    <p className='text-foreground font-medium'>{org.name}</p>
                                    <p className='text-muted-foreground text-xs'>
                                        {org.slug} · {org.plan}
                                    </p>
                                </div>
                            </div>
                        </button>
                    ))}
                </div>
            )}
        </PageContainer>
    );
}
