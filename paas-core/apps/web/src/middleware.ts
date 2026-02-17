import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const PUBLIC_PATHS = ['/auth/sign-in', '/auth/sign-up', '/api/health'];

export function middleware(request: NextRequest) {
    const { pathname } = request.nextUrl;

    // Allow public paths
    if (PUBLIC_PATHS.some((p) => pathname.startsWith(p))) {
        return NextResponse.next();
    }

    // Allow root / and static assets
    if (
        pathname === '/' ||
        pathname.startsWith('/_next') ||
        pathname.startsWith('/favicon')
    ) {
        return NextResponse.next();
    }

    // Check for access token in cookie or header
    const token =
        request.cookies.get('access_token')?.value ||
        request.headers.get('authorization')?.replace('Bearer ', '');

    if (!token && pathname.startsWith('/dashboard')) {
        const signInUrl = new URL('/auth/sign-in', request.url);
        signInUrl.searchParams.set('callbackUrl', pathname);
        return NextResponse.redirect(signInUrl);
    }

    return NextResponse.next();
}

export const config = {
    matcher: ['/((?!_next/static|_next/image|favicon.ico).*)']
};
