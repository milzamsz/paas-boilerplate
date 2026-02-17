'use client';

import PageContainer from '@/components/layout/page-container';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { useEffect, useState } from 'react';
import { apiClient } from '@/lib/api-client';
import type { BillingPlan, Subscription } from '@/types';
import { formatCurrency } from '@/lib/utils';
import { IconCheck } from '@tabler/icons-react';

export default function BillingPage() {
    const { currentOrg } = useAuth();
    const [plans, setPlans] = useState<BillingPlan[]>([]);
    const [subscription, setSubscription] = useState<Subscription | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        Promise.all([
            apiClient.get('/billing/plans').catch(() => ({ data: { data: [] } })),
            currentOrg
                ? apiClient
                    .get(`/orgs/${currentOrg.id}/billing/subscription`)
                    .catch(() => ({ data: { data: null } }))
                : Promise.resolve({ data: { data: null } })
        ])
            .then(([plansRes, subRes]) => {
                setPlans(plansRes.data?.data || plansRes.data || []);
                setSubscription(subRes.data?.data || null);
            })
            .finally(() => setLoading(false));
    }, [currentOrg]);

    return (
        <PageContainer
            title='Billing'
            description={
                currentOrg
                    ? `Manage billing for ${currentOrg.name}`
                    : 'Select an organization to manage billing'
            }
        >
            {/* Current subscription */}
            {subscription && (
                <div className='border-border bg-card rounded-xl border p-6'>
                    <h2 className='text-foreground mb-2 text-lg font-semibold'>
                        Current Plan
                    </h2>
                    <div className='flex items-center gap-4'>
                        <span className='bg-primary/10 text-primary rounded-full px-3 py-1 text-sm font-medium'>
                            {subscription.plan?.name || subscription.plan_id}
                        </span>
                        <span className='text-muted-foreground text-sm'>
                            Status: {subscription.status}
                        </span>
                    </div>
                </div>
            )}

            {/* Plans grid */}
            {loading ? (
                <div className='grid gap-6 lg:grid-cols-3'>
                    {[1, 2, 3].map((i) => (
                        <div
                            key={i}
                            className='border-border bg-card h-64 animate-pulse rounded-xl border'
                        />
                    ))}
                </div>
            ) : (
                <div className='grid gap-6 lg:grid-cols-3'>
                    {plans.map((plan) => {
                        const isCurrent = subscription?.plan_id === plan.id;
                        return (
                            <div
                                key={plan.id}
                                className={`border-border bg-card relative rounded-xl border p-6 transition-shadow hover:shadow-md ${isCurrent ? 'border-primary ring-primary/20 ring-2' : ''
                                    }`}
                            >
                                {isCurrent && (
                                    <span className='bg-primary text-primary-foreground absolute -top-3 left-4 rounded-full px-3 py-0.5 text-xs font-medium'>
                                        Current
                                    </span>
                                )}
                                <h3 className='text-foreground text-lg font-semibold'>
                                    {plan.name}
                                </h3>
                                <p className='text-foreground mt-3 text-3xl font-bold'>
                                    {formatCurrency(plan.price_monthly)}
                                    <span className='text-muted-foreground text-sm font-normal'>
                                        /mo
                                    </span>
                                </p>
                                <ul className='mt-6 space-y-3'>
                                    {plan.features?.map((feature) => (
                                        <li
                                            key={feature}
                                            className='flex items-center gap-2 text-sm'
                                        >
                                            <IconCheck className='h-4 w-4 text-emerald-500' />
                                            <span className='text-muted-foreground'>{feature}</span>
                                        </li>
                                    ))}
                                </ul>
                                <button
                                    disabled={isCurrent}
                                    className={`mt-6 w-full rounded-lg px-4 py-2 text-sm font-medium transition ${isCurrent
                                            ? 'bg-muted text-muted-foreground cursor-default'
                                            : 'bg-primary text-primary-foreground hover:bg-primary/90'
                                        }`}
                                >
                                    {isCurrent ? 'Current Plan' : 'Upgrade'}
                                </button>
                            </div>
                        );
                    })}
                </div>
            )}
        </PageContainer>
    );
}
