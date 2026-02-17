'use client';

import { ThemeProvider as NextThemesProvider } from 'next-themes';
import { useAuthStore } from '@/features/auth/stores/auth-store';
import { useEffect } from 'react';
import { TooltipProvider } from '@/components/ui/tooltip';

export default function Providers({ children }: { children: React.ReactNode }) {
    const hydrate = useAuthStore((s) => s.hydrate);

    useEffect(() => {
        hydrate();
    }, [hydrate]);

    return (
        <NextThemesProvider
            attribute='class'
            defaultTheme='dark'
            enableSystem
            disableTransitionOnChange
        >
            <TooltipProvider>
                {children}
            </TooltipProvider>
        </NextThemesProvider>
    );
}
