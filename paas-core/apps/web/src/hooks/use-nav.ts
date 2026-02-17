'use client';

import { useMemo } from 'react';
import { navItems } from '@/config/nav-config';
import { useAuth } from '@/features/auth/hooks/use-auth';
import type { NavItem } from '@/types';

export function useFilteredNavItems(): NavItem[] {
    const { currentOrg, user } = useAuth();

    return useMemo(() => {
        function filterItems(items: NavItem[]): NavItem[] {
            return items
                .filter((item) => {
                    if (!item.access) return true;
                    if (item.access.requireOrg && !currentOrg) return false;
                    if (item.access.roles && user) {
                        return item.access.roles.includes(user.role);
                    }
                    return true;
                })
                .map((item) => ({
                    ...item,
                    items: item.items ? filterItems(item.items) : []
                }));
        }

        return filterItems(navItems);
    }, [currentOrg, user]);
}
