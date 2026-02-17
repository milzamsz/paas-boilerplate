import AppSidebar from '@/components/layout/app-sidebar';
import Header from '@/components/layout/header';

export default function DashboardLayout({
    children
}: {
    children: React.ReactNode;
}) {
    return (
        <div className='flex min-h-screen'>
            <AppSidebar />
            <div className='flex flex-1 flex-col md:ml-64'>
                <Header />
                <main className='flex-1'>{children}</main>
            </div>
        </div>
    );
}
