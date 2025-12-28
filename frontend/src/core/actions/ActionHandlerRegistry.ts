/**
 * Action Handler Registry
 *
 * Manages the registration and lookup of flow action handlers.
 * Supports both static (code-based) and dynamic (metadata-driven) registration.
 *
 * @module core/actions/ActionHandlerRegistry
 */

import { ActionHandlerDefinition, ActionHandlerMetadata, ActionHandlerModule } from './ActionHandlerTypes';


/**
 * Registry for flow action handlers
 *
 * Provides a centralized location for registering and retrieving action handlers.
 * Supports:
 * - Static registration via registerAction()
 * - Dynamic loading from metadata via loadFromMetadata()
 * - Module-based registration via registerModule()
 */
export class ActionHandlerRegistry {
    private handlers: Map<string, ActionHandlerDefinition> = new Map();
    private loadedModules: Set<string> = new Set();

    /**
     * Register an action handler
     *
     * @param definition - The action handler definition
     */
    registerAction(definition: ActionHandlerDefinition): void {
        if (this.handlers.has(definition.actionType)) {
            console.warn(`[ActionHandlerRegistry] Overwriting existing handler: ${definition.actionType}`);
        }

        this.handlers.set(definition.actionType, definition);

    }

    /**
     * Register an action handler from a module
     *
     * @param modulePath - Path to the module (for tracking)
     * @param module - The action handler module
     */
    registerModule(modulePath: string, module: ActionHandlerModule): void {
        if (this.loadedModules.has(modulePath)) {
            console.warn(`[ActionHandlerRegistry] Module already loaded: ${modulePath}`);
            return;
        }

        this.registerAction(module.handler);
        this.loadedModules.add(modulePath);
    }

    /**
     * Get an action handler by action type
     *
     * @param actionType - The action type identifier
     * @returns The action handler definition or undefined if not found
     */
    getHandler(actionType: string): ActionHandlerDefinition | undefined {
        return this.handlers.get(actionType);
    }

    /**
     * Get all registered action handlers
     *
     * @returns Array of all action handler definitions
     */
    getAllHandlers(): ActionHandlerDefinition[] {
        return Array.from(this.handlers.values());
    }

    /**
     * Get action handlers filtered by category
     *
     * @param category - The category to filter by
     * @returns Array of action handler definitions in the category
     */
    getHandlersByCategory(category: string): ActionHandlerDefinition[] {
        return this.getAllHandlers()
            .filter(h => h.category === category)
            .sort((a, b) => (a.sortOrder || 100) - (b.sortOrder || 100));
    }

    /**
     * Check if an action handler is registered
     *
     * @param actionType - The action type to check
     * @returns true if registered, false otherwise
     */
    hasHandler(actionType: string): boolean {
        return this.handlers.has(actionType);
    }

    /**
     * Get all registered action types
     *
     * @returns Array of action type identifiers
     */
    getActionTypes(): string[] {
        return Array.from(this.handlers.keys());
    }

    /**
     * Clear all registered handlers (primarily for testing)
     */
    clear(): void {
        this.handlers.clear();
        this.loadedModules.clear();

    }

    /**
     * Get registry statistics
     *
     * @returns Statistics about registered handlers
     */
    getStats(): {
        totalHandlers: number;
        byCategory: Record<string, number>;
        loadedModules: number;
    } {
        const byCategory: Record<string, number> = {};

        for (const handler of this.handlers.values()) {
            byCategory[handler.category] = (byCategory[handler.category] || 0) + 1;
        }

        return {
            totalHandlers: this.handlers.size,
            byCategory,
            loadedModules: this.loadedModules.size,
        };
    }

    /**
     * Load action handlers from metadata
     *
     * This method would be called by the FlowEngine to load handlers
     * dynamically based on _System_ActionHandler records.
     *
     * @param metadata - Array of action handler metadata from database
     * @param moduleLoader - Function to dynamically import handler modules
     */
    async loadFromMetadata(
        metadata: ActionHandlerMetadata[],
        moduleLoader?: (modulePath: string) => Promise<ActionHandlerModule>
    ): Promise<void> {

        for (const meta of metadata) {
            if (!meta.isActive) {
                continue;
            }

            // If a module loader is provided, dynamically import the handler
            if (moduleLoader && meta.handlerModule) {
                try {
                    const module = await moduleLoader(meta.handlerModule);
                    this.registerModule(meta.handlerModule, module);
                } catch (error: unknown) {
                    const msg = error instanceof Error ? error.message : String(error);
                    console.error(`[ActionHandlerRegistry] Failed to load module ${meta.handlerModule}:`, msg);
                }
            }
        }

        const stats = this.getStats();

    }
}

/**
 * Global action handler registry instance
 */
export const actionHandlerRegistry = new ActionHandlerRegistry();
