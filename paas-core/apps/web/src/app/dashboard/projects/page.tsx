'use client';

import PageContainer from '@/components/layout/page-container';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { useEffect, useState } from 'react';
import { apiClient } from '@/lib/api-client';
import type { Project } from '@/types';
import Link from 'next/link';
import { IconPlus, IconExternalLink } from '@tabler/icons-react';

export default function ProjectsPage() {
    const { currentOrg } = useAuth();
    const [projects, setProjects] = useState<Project[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        if (!currentOrg) {
            setLoading(false);
            return;
        }
        apiClient
            .get(`/orgs/${currentOrg.id}/projects`)
            .then((res) => setProjects(res.data?.data || res.data || []))
            .catch(() => setProjects([]))
            .finally(() => setLoading(false));
    }, [currentOrg]);

    if (!currentOrg) {
        return (
            <PageContainer
                title='Projects'
                description='Select an organization first'
            >
                <div className='border-border bg-card flex flex-col items-center justify-center rounded-xl border py-16'>
                    <p className='text-muted-foreground mb-4 text-sm'>
                        Please select an organization to view projects
                    </p>
                    <Link
                        href='/dashboard/orgs'
                        className='bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg px-4 py-2 text-sm font-medium'
                    >
                        Go to Organizations
                    </Link>
                </div>
            </PageContainer>
        );
    }

    return (
        <PageContainer
            title='Projects'
            description={`Projects in ${currentOrg.name}`}
            actions={
                <Link
                    href='/dashboard/projects/new'
                    className='bg-primary text-primary-foreground hover:bg-primary/90 inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition'
                >
                    <IconPlus className='h-4 w-4' />
                    New Project
                </Link>
            }
        >
            {loading ? (
                <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-3'>
                    {[1, 2, 3].map((i) => (
                        <div
                            key={i}
                            className='border-border bg-card h-40 animate-pulse rounded-xl border'
                        />
                    ))}
                </div>
            ) : projects.length === 0 ? (
                <div className='border-border bg-card flex flex-col items-center justify-center rounded-xl border py-16'>
                    <p className='text-muted-foreground mb-4 text-sm'>
                        No projects yet
                    </p>
                    <Link
                        href='/dashboard/projects/new'
                        className='bg-primary text-primary-foreground hover:bg-primary/90 inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium'
                    >
                        <IconPlus className='h-4 w-4' />
                        Create your first project
                    </Link>
                </div>
            ) : (
                <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-3'>
                    {projects.map((project) => (
                        <Link
                            key={project.id}
                            href={`/dashboard/projects/${project.id}`}
                            className='border-border bg-card hover:border-primary/50 group rounded-xl border p-6 transition-all hover:shadow-md'
                        >
                            <div className='flex items-start justify-between'>
                                <div>
                                    <p className='text-foreground font-medium'>{project.name}</p>
                                    <p className='text-muted-foreground mt-1 text-xs'>
                                        {project.region}
                                    </p>
                                </div>
                                <span
                                    className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${project.status === 'active'
                                            ? 'bg-emerald-500/10 text-emerald-500'
                                            : 'bg-zinc-500/10 text-zinc-500'
                                        }`}
                                >
                                    {project.status}
                                </span>
                            </div>
                            {project.description && (
                                <p className='text-muted-foreground mt-3 line-clamp-2 text-sm'>
                                    {project.description}
                                </p>
                            )}
                            <div className='mt-4 flex items-center gap-1 text-xs opacity-0 transition group-hover:opacity-100'>
                                <span className='text-primary'>Open project</span>
                                <IconExternalLink className='text-primary h-3 w-3' />
                            </div>
                        </Link>
                    ))}
                </div>
            )}
        </PageContainer>
    );
}
