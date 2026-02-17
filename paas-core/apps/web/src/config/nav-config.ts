import type { NavItem } from '@/types';

export const navItems: NavItem[] = [
    {
        title: 'Dashboard',
        url: '/dashboard/overview',
        icon: 'dashboard',
        isActive: false,
        shortcut: ['d', 'd'],
        items: []
    },
    {
        title: 'Organizations',
        url: '/dashboard/orgs',
        icon: 'building',
        isActive: false,
        items: []
    },
    {
        title: 'Projects',
        url: '/dashboard/projects',
        icon: 'folder',
        isActive: false,
        shortcut: ['p', 'p'],
        items: [],
        access: { requireOrg: true }
    },
    {
        title: 'Billing',
        url: '/dashboard/billing',
        icon: 'creditCard',
        isActive: false,
        shortcut: ['b', 'b'],
        items: [],
        access: { requireOrg: true }
    },
    {
        title: 'Account',
        url: '#',
        icon: 'user',
        isActive: true,
        items: [
            {
                title: 'Profile',
                url: '/dashboard/profile',
                icon: 'user',
                shortcut: ['m', 'm']
            }
        ]
    }
];
