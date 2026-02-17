import Providers from '@/components/layout/providers';
import type { Metadata } from 'next';
import NextTopLoader from 'nextjs-toploader';
import '../styles/globals.css';

export const metadata: Metadata = {
    title: {
        default: process.env.NEXT_PUBLIC_APP_NAME || 'MyPaaS',
        template: `%s | ${process.env.NEXT_PUBLIC_APP_NAME || 'MyPaaS'}`
    },
    description: 'Cloud platform dashboard'
};

export default function RootLayout({
    children
}: {
    children: React.ReactNode;
}) {
    return (
        <html lang='en' suppressHydrationWarning>
            <body className='bg-background text-foreground overflow-x-hidden font-sans antialiased' suppressHydrationWarning>
                <NextTopLoader color='hsl(var(--primary))' showSpinner={false} />
                <Providers>{children}</Providers>
            </body>
        </html>
    );
}
