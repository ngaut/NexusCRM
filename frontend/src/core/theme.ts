/**
 * NexusCRM Design System
 * Centralized design tokens for consistent UI/UX
 */

export const theme = {
    // Color System
    colors: {
        // Primary - Blue (for actions, links, focus)
        primary: {
            50: '#eff6ff',
            100: '#dbeafe',
            200: '#bfdbfe',
            300: '#93c5fd',
            400: '#60a5fa',
            500: '#3b82f6',
            600: '#2563eb',
            700: '#1d4ed8',
            800: '#1e40af',
            900: '#1e3a8a',
        },

        // Semantic Colors
        success: {
            50: '#f0fdf4',
            500: '#10b981',
            600: '#059669',
            700: '#047857',
        },
        warning: {
            50: '#fffbeb',
            500: '#f59e0b',
            600: '#d97706',
            700: '#b45309',
        },
        error: {
            50: '#fef2f2',
            500: '#ef4444',
            600: '#dc2626',
            700: '#b91c1c',
        },
        info: {
            50: '#eff6ff',
            500: '#3b82f6',
            600: '#2563eb',
            700: '#1d4ed8',
        },

        // Neutral (Gray scale)
        neutral: {
            50: '#f9fafb',
            100: '#f3f4f6',
            200: '#e5e7eb',
            300: '#d1d5db',
            400: '#9ca3af',
            500: '#6b7280',
            600: '#4b5563',
            700: '#374151',
            800: '#1f2937',
            900: '#111827',
        },
    },

    // Typography
    typography: {
        fonts: {
            sans: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
            mono: '"Fira Code", "Courier New", monospace',
        },
        sizes: {
            xs: '0.75rem',    // 12px - Captions, labels
            sm: '0.875rem',   // 14px - Body, default
            base: '1rem',     // 16px - Emphasized text
            lg: '1.125rem',   // 18px - Subsections
            xl: '1.25rem',    // 20px - Headings
            '2xl': '1.5rem',  // 24px - Section titles
            '3xl': '2rem',    // 32px - Page titles
            '4xl': '2.5rem',  // 40px - Hero text
        },
        weights: {
            normal: 400,
            medium: 500,
            semibold: 600,
            bold: 700,
        },
        lineHeights: {
            tight: 1.25,
            normal: 1.5,
            relaxed: 1.75,
        },
    },

    // Spacing (4px base unit)
    spacing: {
        0: '0',
        1: '0.25rem',   // 4px
        2: '0.5rem',    // 8px
        3: '0.75rem',   // 12px
        4: '1rem',      // 16px
        5: '1.25rem',   // 20px
        6: '1.5rem',    // 24px
        8: '2rem',      // 32px
        10: '2.5rem',   // 40px
        12: '3rem',     // 48px
        16: '4rem',     // 64px
        20: '5rem',     // 80px
    },

    // Border Radius
    radius: {
        none: '0',
        sm: '0.25rem',   // 4px
        md: '0.375rem',  // 6px
        lg: '0.5rem',    // 8px
        xl: '0.75rem',   // 12px
        '2xl': '1rem',   // 16px
        full: '9999px',
    },

    // Shadows (Elevation system)
    shadows: {
        none: 'none',
        sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
        md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
        lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
        xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
        inner: 'inset 0 2px 4px 0 rgba(0, 0, 0, 0.06)',
    },

    // Transitions
    transitions: {
        fast: '150ms cubic-bezier(0.4, 0, 0.2, 1)',
        base: '200ms cubic-bezier(0.4, 0, 0.2, 1)',
        slow: '300ms cubic-bezier(0.4, 0, 0.2, 1)',
    },

    // Z-index layers
    zIndex: {
        base: 0,
        dropdown: 10,
        sticky: 20,
        overlay: 30,
        modal: 40,
        popover: 50,
        tooltip: 60,
    },

    // Breakpoints
    breakpoints: {
        sm: '640px',
        md: '768px',
        lg: '1024px',
        xl: '1280px',
        '2xl': '1536px',
    },
} as const;

// Status color mappings
export const statusColors = {
    // Lead statuses
    new: theme.colors.primary[500],
    contacted: theme.colors.info[500],
    qualified: theme.colors.success[500],
    unqualified: theme.colors.neutral[400],

    // Opportunity stages
    prospecting: theme.colors.primary[500],
    qualification: theme.colors.info[500],
    proposal: theme.colors.warning[500],
    negotiation: theme.colors.warning[600],
    won: theme.colors.success[500],
    lost: theme.colors.error[500],

    // Generic statuses
    active: theme.colors.success[500],
    inactive: theme.colors.neutral[400],
    draft: theme.colors.warning[500],
    pending: theme.colors.info[500],
    completed: theme.colors.success[500],
    cancelled: theme.colors.error[500],
} as const;

// Priority colors
export const priorityColors = {
    critical: theme.colors.error[500],
    high: theme.colors.warning[500],
    medium: theme.colors.info[500],
    low: theme.colors.neutral[400],
} as const;

export type Theme = typeof theme;
export type StatusColor = keyof typeof statusColors;
export type PriorityColor = keyof typeof priorityColors;
