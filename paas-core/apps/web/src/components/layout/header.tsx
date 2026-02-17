'use client';

import { usePathname } from 'next/navigation';
import Link from 'next/link';
import { IconSun, IconMoon } from '@tabler/icons-react';
import { useTheme } from 'next-themes';

export default function Header() {
    const pathname = usePathname();
    const { theme, setTheme } = useTheme();

    // Generate breadcrumbs from pathname
    const segments = pathname
        .split('/')
        .filter(Boolean)
        .map((seg, i, arr) => ({
            label: seg.charAt(0).toUpperCase() + seg.slice(1),
            href: '/' + arr.slice(0, i + 1).join('/')
        }));

    return (
        <header className='border-border bg-background/80 sticky top-0 z-30 flex h-16 items-center justify-between border-b px-6 backdrop-blur-sm'>
            {/* Breadcrumbs */}
            <nav className='flex items-center gap-1.5 text-sm'>
                {segments.map((seg, i) => (
                    <span key={seg.href} className='flex items-center gap-1.5'>
                        {i > 0 && (
                            <span className='text-muted-foreground'>/</span>
                        )}
                        {i === segments.length - 1 ? (
                            <span className='text-foreground font-medium'>
                                {seg.label}
                            </span>
                        ) : (
                            <Link
                                href={seg.href}
                                className='text-muted-foreground hover:text-foreground transition-colors'
                            >
                                {seg.label}
                            </Link>
                        )}
                    </span>
                ))}
            </nav>

            {/* Actions */}
            <div className='flex items-center gap-2'>
                <button
                    onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
                    className='text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg p-2 transition-colors'
                    title='Toggle theme'
                >
                    {theme === 'dark' ? (
                        <IconSun className='h-4 w-4' />
                    ) : (
                        <IconMoon className='h-4 w-4' />
                    )}
                </button>
            </div>
        </header>
    );
}
