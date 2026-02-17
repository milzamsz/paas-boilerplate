'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { useFilteredNavItems } from '@/hooks/use-nav';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { Icons, type IconKey } from '@/components/icons';
import { IconChevronDown, IconMenu2, IconX } from '@tabler/icons-react';
import type { NavItem } from '@/types';
import { useState } from 'react';

export default function AppSidebar() {
    const pathname = usePathname();
    const navItems = useFilteredNavItems();
    const { user, currentOrg, logout } = useAuth();
    const [mobileOpen, setMobileOpen] = useState(false);

    return (
        <>
            {/* Mobile toggle */}
            <button
                className='fixed top-4 left-4 z-50 rounded-lg border border-zinc-800 bg-zinc-900 p-2 md:hidden'
                onClick={() => setMobileOpen(!mobileOpen)}
            >
                {mobileOpen ? (
                    <IconX className='h-5 w-5' />
                ) : (
                    <IconMenu2 className='h-5 w-5' />
                )}
            </button>

            {/* Overlay */}
            {mobileOpen && (
                <div
                    className='fixed inset-0 z-40 bg-black/50 md:hidden'
                    onClick={() => setMobileOpen(false)}
                />
            )}

            {/* Sidebar */}
            <aside
                className={cn(
                    'bg-sidebar text-sidebar-foreground border-sidebar-border fixed inset-y-0 left-0 z-40 flex w-64 flex-col border-r transition-transform md:translate-x-0',
                    mobileOpen ? 'translate-x-0' : '-translate-x-full'
                )}
            >
                {/* Brand */}
                <div className='border-sidebar-border flex h-16 items-center gap-2 border-b px-6'>
                    <div className='bg-primary h-8 w-8 rounded-lg' />
                    <span className='text-lg font-semibold'>
                        {process.env.NEXT_PUBLIC_APP_NAME || 'MyPaaS'}
                    </span>
                </div>

                {/* Org context */}
                {currentOrg && (
                    <div className='border-sidebar-border border-b px-4 py-3'>
                        <p className='text-muted-foreground text-xs uppercase tracking-wider'>
                            Organization
                        </p>
                        <p className='text-sm font-medium truncate'>{currentOrg.name}</p>
                    </div>
                )}

                {/* Navigation */}
                <nav className='flex-1 overflow-y-auto px-3 py-4'>
                    <ul className='space-y-1'>
                        {navItems.map((item) => {
                            const Icon =
                                Icons[item.icon as IconKey] || Icons.dashboard;
                            const isActive =
                                pathname === item.url ||
                                pathname.startsWith(item.url + '/');
                            const hasChildren = item.items && item.items.length > 0;

                            if (item.url === '#' && hasChildren) {
                                return (
                                    <li key={item.title}>
                                        <CollapsibleNav
                                            item={item}
                                            pathname={pathname}
                                        />
                                    </li>
                                );
                            }

                            return (
                                <li key={item.title}>
                                    <Link
                                        href={item.url}
                                        onClick={() => setMobileOpen(false)}
                                        className={cn(
                                            'flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors',
                                            isActive
                                                ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
                                                : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground'
                                        )}
                                    >
                                        <Icon className='h-4 w-4 shrink-0' />
                                        {item.title}
                                    </Link>
                                </li>
                            );
                        })}
                    </ul>
                </nav>

                {/* User footer */}
                <div className='border-sidebar-border border-t p-4'>
                    <div className='flex items-center gap-3'>
                        <div className='bg-sidebar-accent flex h-8 w-8 items-center justify-center rounded-full text-xs font-medium'>
                            {user?.name?.charAt(0)?.toUpperCase() || '?'}
                        </div>
                        <div className='flex-1 truncate'>
                            <p className='truncate text-sm font-medium'>
                                {user?.name || 'User'}
                            </p>
                            <p className='text-muted-foreground truncate text-xs'>
                                {user?.email || ''}
                            </p>
                        </div>
                        <button
                            onClick={logout}
                            className='text-muted-foreground hover:text-foreground p-1 transition-colors'
                            title='Sign out'
                        >
                            <Icons.logout className='h-4 w-4' />
                        </button>
                    </div>
                </div>
            </aside>
        </>
    );
}

function CollapsibleNav({
    item,
    pathname
}: {
    item: NavItem;
    pathname: string;
}) {
    const [open, setOpen] = useState(
        item.items?.some(
            (child) =>
                pathname === child.url || pathname.startsWith(child.url + '/')
        ) || false
    );

    const Icon = Icons[item.icon as IconKey] || Icons.dashboard;

    return (
        <div>
            <button
                onClick={() => setOpen(!open)}
                className='text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors'
            >
                <Icon className='h-4 w-4 shrink-0' />
                <span className='flex-1 text-left'>{item.title}</span>
                <IconChevronDown
                    className={cn(
                        'h-4 w-4 transition-transform',
                        open && 'rotate-180'
                    )}
                />
            </button>
            {open && item.items && (
                <ul className='mt-1 ml-4 space-y-1 border-l border-zinc-800 pl-3'>
                    {item.items.map((child) => {
                        const isActive =
                            pathname === child.url ||
                            pathname.startsWith(child.url + '/');
                        return (
                            <li key={child.title}>
                                <Link
                                    href={child.url}
                                    className={cn(
                                        'block rounded-lg px-3 py-1.5 text-sm transition-colors',
                                        isActive
                                            ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
                                            : 'text-sidebar-foreground/60 hover:text-sidebar-foreground'
                                    )}
                                >
                                    {child.title}
                                </Link>
                            </li>
                        );
                    })}
                </ul>
            )}
        </div>
    );
}
