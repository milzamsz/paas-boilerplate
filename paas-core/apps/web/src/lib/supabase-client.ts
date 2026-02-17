import { createClient } from '@supabase/supabase-js';

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL || '';
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY || '';

/**
 * Supabase client instance.
 * Only used when NEXT_PUBLIC_AUTH_PROVIDER=supabase.
 * Import lazily or guard with isSupabaseEnabled() to avoid errors
 * when Supabase env vars are not set.
 */
export const supabase = createClient(supabaseUrl, supabaseAnonKey);

/**
 * Returns true when the app is configured to use Supabase auth.
 */
export function isSupabaseEnabled(): boolean {
    return process.env.NEXT_PUBLIC_AUTH_PROVIDER === 'supabase';
}
