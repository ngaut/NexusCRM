/**
 * Color class utilities for consistent styling.
 * Provides Tailwind CSS color mappings for UI components.
 */

interface ColorClasses {
    bg: string;
    text: string;
    border: string;
    hover: string;
}

const colorMap: Record<string, ColorClasses> = {
    blue: {
        bg: 'bg-blue-500',
        text: 'text-blue-600',
        border: 'border-blue-500',
        hover: 'hover:bg-blue-600',
    },
    green: {
        bg: 'bg-green-500',
        text: 'text-green-600',
        border: 'border-green-500',
        hover: 'hover:bg-green-600',
    },
    red: {
        bg: 'bg-red-500',
        text: 'text-red-600',
        border: 'border-red-500',
        hover: 'hover:bg-red-600',
    },
    yellow: {
        bg: 'bg-yellow-500',
        text: 'text-yellow-600',
        border: 'border-yellow-500',
        hover: 'hover:bg-yellow-600',
    },
    purple: {
        bg: 'bg-purple-500',
        text: 'text-purple-600',
        border: 'border-purple-500',
        hover: 'hover:bg-purple-600',
    },
    pink: {
        bg: 'bg-pink-500',
        text: 'text-pink-600',
        border: 'border-pink-500',
        hover: 'hover:bg-pink-600',
    },
    indigo: {
        bg: 'bg-indigo-500',
        text: 'text-indigo-600',
        border: 'border-indigo-500',
        hover: 'hover:bg-indigo-600',
    },
    teal: {
        bg: 'bg-teal-500',
        text: 'text-teal-600',
        border: 'border-teal-500',
        hover: 'hover:bg-teal-600',
    },
    orange: {
        bg: 'bg-orange-500',
        text: 'text-orange-600',
        border: 'border-orange-500',
        hover: 'hover:bg-orange-600',
    },
    cyan: {
        bg: 'bg-cyan-500',
        text: 'text-cyan-600',
        border: 'border-cyan-500',
        hover: 'hover:bg-cyan-600',
    },
    gray: {
        bg: 'bg-gray-500',
        text: 'text-gray-600',
        border: 'border-gray-500',
        hover: 'hover:bg-gray-600',
    },
    slate: {
        bg: 'bg-slate-500',
        text: 'text-slate-600',
        border: 'border-slate-500',
        hover: 'hover:bg-slate-600',
    },
};

/**
 * Get Tailwind CSS classes for a given color name.
 * Returns default gray classes if color not found.
 */
export const getColorClasses = (color: string = 'blue'): ColorClasses => {
    return colorMap[color.toLowerCase()] || colorMap.blue;
};
