import { cn } from '@/lib/utils';

interface PageContainerProps {
    title: string;
    description?: string;
    children: React.ReactNode;
    className?: string;
    actions?: React.ReactNode;
}

export default function PageContainer({
    title,
    description,
    children,
    className,
    actions
}: PageContainerProps) {
    return (
        <div className={cn('flex flex-1 flex-col gap-6 p-6', className)}>
            <div className='flex items-center justify-between'>
                <div>
                    <h1 className='text-foreground text-2xl font-bold tracking-tight'>
                        {title}
                    </h1>
                    {description && (
                        <p className='text-muted-foreground mt-1 text-sm'>
                            {description}
                        </p>
                    )}
                </div>
                {actions && <div className='flex items-center gap-2'>{actions}</div>}
            </div>
            {children}
        </div>
    );
}
