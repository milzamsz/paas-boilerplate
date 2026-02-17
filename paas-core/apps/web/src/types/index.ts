export interface NavItem {
    title: string;
    url: string;
    icon?: string;
    isActive?: boolean;
    shortcut?: string[];
    items?: NavItem[];
    access?: NavAccess;
}

export interface NavAccess {
    requireOrg?: boolean;
    roles?: string[];
    permission?: string;
}

export interface User {
    id: string;
    email: string;
    name: string;
    role: string;
    avatar_url?: string;
    created_at: string;
    updated_at: string;
}

export interface Organization {
    id: string;
    name: string;
    slug: string;
    owner_id: string;
    plan: string;
    created_at: string;
    updated_at: string;
}

export interface OrgMember {
    id: string;
    org_id: string;
    user_id: string;
    role: string;
    user?: User;
    created_at: string;
}

export interface Project {
    id: string;
    org_id: string;
    name: string;
    description?: string;
    region: string;
    status: string;
    created_at: string;
    updated_at: string;
}

export interface BillingPlan {
    id: string;
    name: string;
    slug: string;
    price_monthly: number;
    price_yearly: number;
    features: string[];
    is_active: boolean;
}

export interface Subscription {
    id: string;
    org_id: string;
    plan_id: string;
    plan?: BillingPlan;
    status: string;
    current_period_start: string;
    current_period_end: string;
}

export interface Invoice {
    id: string;
    subscription_id: string;
    amount: number;
    currency: string;
    status: string;
    paid_at?: string;
    created_at: string;
}

export interface ApiResponse<T> {
    data: T;
    message?: string;
}

export interface ApiError {
    error: string;
    code: string;
    details?: Record<string, string>;
}

export interface PaginatedResponse<T> {
    data: T[];
    total: number;
    page: number;
    per_page: number;
}
