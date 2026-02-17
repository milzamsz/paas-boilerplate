import {
    IconBuildingSkyscraper,
    IconCreditCard,
    IconDashboard,
    IconFolder,
    IconUser,
    IconLogout,
    IconSettings
} from '@tabler/icons-react';

export const Icons = {
    dashboard: IconDashboard,
    building: IconBuildingSkyscraper,
    folder: IconFolder,
    creditCard: IconCreditCard,
    user: IconUser,
    logout: IconLogout,
    settings: IconSettings
} as const;

export type IconKey = keyof typeof Icons;
