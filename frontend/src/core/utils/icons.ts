/**
 * Icon utilities for dynamic icon loading.
 * Extracted from utils.ts
 */

import * as Icons from 'lucide-react';

/**
 * Get a Lucide icon component by name.
 * Falls back to Box icon if not found.
 */
export const getIcon = (iconName: string) => {
    const icon = (Icons as unknown as Record<string, React.ComponentType | undefined>)[iconName];
    return icon || Icons.Box;
};
