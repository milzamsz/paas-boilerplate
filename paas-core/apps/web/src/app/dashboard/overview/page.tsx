'use client';

import PageContainer from '@/components/layout/page-container';
import { useAuth } from '@/features/auth/hooks/use-auth';
import {
    IconBuildingSkyscraper,
    IconFolder,
    IconCloudComputing,
    IconActivity
} from '@tabler/icons-react';

const stats = [
    {
        title: 'Organizations',
        value: '—',
        icon: IconBuildingSkyscraper,
        color: 'text-blue-500'
    },
    {
        title: 'Projects',
        value: '—',
        icon: IconFolder,
        color: 'text-emerald-500'
    },
    {
        title: 'Deployments',
        value: '—',
        icon: IconCloudComputing,
        color: 'text-purple-500'
    },
    {
        title: 'Uptime',
        value: '99.9%',
        icon: IconActivity,
        color: 'text-amber-500'
    }
];

export default function OverviewPage() {
    const { user } = useAuth();

    return (
        <PageContainer
            title='Dashboard'
            description={`Welcome back, ${user?.name || 'there'}!`}
        >
            {/* Stats grid */}
            <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
                {stats.map((stat) => (
                    <div
                        key={stat.title}
                        className='border-border bg-card rounded-xl border p-6 transition-shadow hover:shadow-md'
                    >
                        <div className='flex items-center justify-between'>
                            <p className='text-muted-foreground text-sm font-medium'>
                                {stat.title}
                            </p>
                            <stat.icon className={`h-5 w-5 ${stat.color}`} />
                        </div>
                        <p className='text-foreground mt-2 text-3xl font-bold'>
                            {stat.value}
                        </p>
                    </div>
                ))}
            </div>

            {/* Activity placeholder */}
            <div className='border-border bg-card rounded-xl border p-6'>
                <h2 className='text-foreground mb-4 text-lg font-semibold'>
                    Recent Activity
                </h2>
                <div className='text-muted-foreground flex h-48 items-center justify-center text-sm'>
                    Activity feed will appear here once connected to the API.
                </div>
            </div>

            {/* Charts placeholder */}
            <div className='grid gap-4 lg:grid-cols-2'>
                <div className='border-border bg-card rounded-xl border p-6'>
                    <h2 className='text-foreground mb-4 text-lg font-semibold'>
                        Deployments
                    </h2>
                    <div className='text-muted-foreground flex h-48 items-center justify-center text-sm'>
                        Deployment chart (Recharts) — coming soon
                    </div>
                </div>
                <div className='border-border bg-card rounded-xl border p-6'>
                    <h2 className='text-foreground mb-4 text-lg font-semibold'>
                        Resource Usage
                    </h2>
                    <div className='text-muted-foreground flex h-48 items-center justify-center text-sm'>
                        Usage chart (Recharts) — coming soon
                    </div>
                </div>
            </div>
        </PageContainer>
    );
}
