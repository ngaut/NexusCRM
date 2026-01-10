import { RegistryBase } from '@shared/utils';
import { SYSTEM_TABLE_NAMES } from '../generated-schema';
import { dataAPI } from '../infrastructure/api/data';
import * as Icons from 'lucide-react';
import { Logger } from '../core/services/Logger';

/**
 * Setup Page Definition (matches _System_SetupPage schema)
 */
export interface SetupPageDefinition {
    id: string;
    label: string;
    icon: string;
    component_name: string;
    category: string; // 'Objects' | 'Security' | 'System' | 'Automation' | 'Integrations'
    order?: number;
    permissionRequired?: string;
    enabled?: boolean;
    description?: string;
    path?: string; // Derived for routing
}

/**
 * SetupRegistry - Loads pages from _System_SetupPage metadata table
 *
 * This registry loads from database metadata for complete configurability.
 */
class SetupRegistryClass extends RegistryBase<SetupPageDefinition> {
    private loaded: boolean = false;

    constructor() {
        super({
            validator: (key, value) => {
                if (!value.id || !value.label || !value.component_name) {
                    return `Setup page "${key}" is missing required fields (id, label, component_name)`;
                }
                return true;
            },
            enableEvents: true,
            allowOverwrite: true
        });
    }

    /**
     * Load pages from database metadata
     * Called automatically on first access, or manually to refresh
     */
    async loadFromDatabase(): Promise<void> {
        try {
            const rows = await dataAPI.query({
                objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_SETUPPAGE,
                filterExpr: 'is_enabled == true',
                sortField: 'page_order',
                sortDirection: 'ASC'
            });

            // Clear existing before loading new
            this.clear();

            rows.forEach((row) => {
                const componentName = String(row.component_name || row.componentName);
                // Fix: Backend returns __sys_gen_id, not id
                const id = String(row.id || row.__sys_gen_id || row.ID);
                const page: SetupPageDefinition = {
                    id: id,
                    label: String(row.label),
                    icon: String(row.icon),
                    component_name: componentName,
                    category: String(row.category),
                    order: Number(row.page_order || row.pageOrder || row.order),
                    permissionRequired: row.permission_required ? String(row.permission_required) : undefined,
                    enabled: row.is_enabled === 1 || row.is_enabled === true,
                    description: row.description ? String(row.description) : undefined,
                    path: row.path ? String(row.path) : this.deriveFallbackPath(componentName)
                };
                this.register(page.id, page);
            });



            this.loaded = true;
        } catch (error) {
            Logger.warn('Failed to load setup pages from metadata, using fallback empty array:', error);
            this.clear();
            this.loaded = true;
        }
    }

    /**
     * Derive path from component name when path is not provided in metadata.
     */
    private deriveFallbackPath(componentName: string): string {
        return `/setup/${componentName.toLowerCase()}`;
    }

    /**
     * Register a new setup page programmatically
     * Useful for plugins or dynamic registration
     */
    register(key: string, page: SetupPageDefinition): void {
        super.register(key, page);
    }

    /**
     * Get all setup pages
     * Returns empty array if not loaded yet (will be loaded by Setup component)
     */
    getPages(): SetupPageDefinition[] {
        const pages = this.getValues();
        // Sort by order then label
        return pages.sort((a, b) => {
            const orderDiff = (a.order || 99) - (b.order || 99);
            if (orderDiff !== 0) return orderDiff;
            return a.label.localeCompare(b.label);
        });
    }

    /**
     * Check if pages have been loaded from database
     */
    isLoaded(): boolean {
        return this.loaded;
    }

    /**
     * Clear all pages (useful for testing)
     */
    clear() {
        super.clear();
        this.loaded = false;
    }
}

export const SetupRegistry = new SetupRegistryClass();

