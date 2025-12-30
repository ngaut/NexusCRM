
import React from 'react';
import { RegistryBase } from '@shared/utils';
import { dataAPI } from '../infrastructure/api/data';
import { TableUIComponent } from '../constants/tables';
import { Dashboard } from '../pages/Dashboard';
import { ObjectView } from '../pages/ObjectView';
import FlowsPage from '../pages/FlowsPage';
import { ObjectManager } from '../pages/ObjectManager';
// Admin/Setup Components
import { UserManager } from '../pages/UserManager';
import { AppManager } from '../pages/AppManager';
import { LayoutEditor } from '../components/admin/LayoutEditor';
import { PermissionSetManager } from '../pages/PermissionSetManager';
import { GroupManager } from '../pages/GroupManager';
import { SharingRuleManager } from '../pages/SharingRuleManager';
import { RecycleBin } from '../pages/RecycleBin';
import { LayoutSelection } from '../pages/LayoutSelection';
import { ObjectDetail } from '../pages/ObjectDetail';

import ProfileManager from '../pages/ProfileManager';
import DataImport from '../pages/DataImport';
import DataExport from '../pages/DataExport';

/**
 * ComponentRegistry - Extends RegistryBase for type-safe component registration
 *
 * Maps component names (strings) to React functional components.
 * Used by dynamic rendering systems (layouts, setup pages, etc.)
 *
 * METADATA INTEGRATION:
 * Component metadata is defined in _System_UIComponent table (bootstrapped in backend/internal/bootstrap/system_data.go).
 * The metadata documents all components for:
 * - Admin UIs (component picker, component library)
 * - Documentation and help systems
 * - Permission and embeddability checking
 * - Future plugin/extension system for custom components
 *
 * Note: Actual component registration remains code-based due to webpack/vite bundling requirements.
 * The registerDefaults() method should be kept in sync with _System_UIComponent seed data.
 */
class ComponentRegistryClass extends RegistryBase<React.ComponentType<any>> {
    /**
     * Non-embeddable components (full pages, admin tools, etc.)
     * These cannot be embedded in layouts or dashboards.
     * Controlled by _System_UIComponent metadata.
     */
    private nonEmbeddableComponents = new Set<string>();
    constructor() {
        super({
            // Validate that registered values are React components
            validator: (key, value) => {
                if (typeof value !== 'function' && typeof value !== 'object') {
                    return `Component "${key}" must be a function or React element`;
                }
                if (typeof value === 'object') {
                    if (value === null) {
                        return `Component "${key}" is null, must be a function or React element`;
                    }
                    // value is now narrowed to non-null object
                    const objValue = value!;
                    // Check for React component or element
                    if (!('$$typeof' in objValue)) {
                        return `Component "${key}" is an object but not a valid React component`;
                    }
                }
                return true;
            },
            enableEvents: true,
            allowOverwrite: true
        });

        this.registerDefaults();
        this.loadFromDatabase();
    }

    /**
     * Load component metadata from the database
     */
    private async loadFromDatabase() {
        try {
            const results = await dataAPI.query({
                objectApiName: TableUIComponent
            });

            if (results && results.length > 0) {
                this.nonEmbeddableComponents.clear();
                results.forEach((record: Record<string, unknown>) => {
                    // is_embeddable can be boolean or 0/1 depending on DB driver result mapping
                    if (record.is_embeddable === false || record.is_embeddable === 0) {
                        this.nonEmbeddableComponents.add(String(record.name));
                    }
                });
            }
        } catch (error) {
            console.warn('Failed to load UI component metadata, using fallbacks:', error);
            // Fallback to critical defaults to ensure system usability
            ['Dashboard', 'ObjectView', 'Flows', 'ObjectManager', 'Setup', 'AppManager', 'UserManager']
                .forEach(c => this.nonEmbeddableComponents.add(c));
        }
    }

    /**
     * Register all built-in components
     * IMPORTANT: Keep this in sync with SEED_UI_COMPONENTS in metadata/seeds.ts
     * Each component registered here should have a corresponding metadata record.
     */
    private registerDefaults() {
        // Page Components (non-embeddable)
        this.register('Dashboard', Dashboard);
        this.register('ObjectView', ObjectView);
        this.register('Flows', FlowsPage);
        this.register('FlowsPage', FlowsPage); // Alias
        this.register('ObjectManager', ObjectManager);
        // Note: Setup is NOT registered here to avoid circular dependency
        // Setup uses ComponentRegistry for dynamic page rendering


        // Admin/Setup Components (non-embeddable)
        this.register('AppManager', AppManager);
        this.register('UserManager', UserManager);
        this.register('LayoutEditor', LayoutEditor);
        this.register('PermissionSetManager', PermissionSetManager);
        this.register('GroupManager', GroupManager);
        this.register('SharingRuleManager', SharingRuleManager);
        this.register('RecycleBin', RecycleBin);
        this.register('LayoutSelection', LayoutSelection);
        this.register('ObjectDetail', ObjectDetail);

        this.register('ProfileManager', ProfileManager);
        this.register('DataImport', DataImport);
        this.register('DataExport', DataExport);
    }

    /**
     * Override get to return null instead of undefined for React rendering safety
     */
    public get(name: string): React.ComponentType<any> | null {
        return super.get(name) ?? null;
    }

    /**
     * Get list of components that can be embedded in layouts
     * Excludes full-page and admin components
     */
    public getEmbeddableComponents(): string[] {
        return this.getKeys().filter(k => !this.nonEmbeddableComponents.has(k));
    }

    /**
     * Check if a component is embeddable
     */
    public isEmbeddable(name: string): boolean {
        return this.has(name) && !this.nonEmbeddableComponents.has(name);
    }

    /**
     * Register a component as non-embeddable (e.g., for plugins adding full pages)
     */
    public registerAsNonEmbeddable(name: string, component: React.ComponentType<any>): void {
        this.register(name, component);
        this.nonEmbeddableComponents.add(name);
    }
}

export const ComponentRegistry = new ComponentRegistryClass();
