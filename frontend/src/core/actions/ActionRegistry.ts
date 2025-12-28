import { RegistryBase } from '@shared/utils';
import { ActionConfig, SObject, FieldMetadata, ActionType } from '../../types';

export interface ActionDefinition {
    name: string;
    label: string;
    type: ActionType;
    icon?: string;
    description?: string;
    component?: string;
    target_object?: string;
    config?: Record<string, unknown>;
    handler?: ActionHandler;
    params?: FieldMetadata[];
    category?: string;
}

export interface ActionContext {
    record: SObject;
    schema: Record<string, unknown>;
    onNavigate: (obj: string, id: string | null) => void;
    onEdit: () => void;
    onRefresh: () => void;
    showToast: (type: 'success' | 'error' | 'info', msg: string) => void;
}

export type ActionHandler = (context: ActionContext, config?: Record<string, unknown>) => Promise<void>;

/**
 * Registry for system and custom actions
 * Centralizes action definitions and handlers for UI buttons, flows, and automation
 */
export class ActionRegistry extends RegistryBase<ActionDefinition> {
    constructor() {
        super({
            allowOverwrite: false,
            // enableEvents: true, // Optional if RegistryBase supports it
            validator: (key, value) => {
                if (!value.name || !value.label) {
                    return `Action ${key} must have name and label`;
                }
                return true;
            }
        });
    }

    /**
     * Register an action with its handler
     */
    registerAction(definition: ActionDefinition) {
        this.register(definition.name, definition);
    }

    /**
     * Get actions by checking values
     */
    getByType(type: ActionType): ActionDefinition[] {
        return this.getValues().filter(action => action.type === type);
    }

    /**
     * Get actions by category
     */
    getByCategory(category: string): ActionDefinition[] {
        return this.getValues().filter(action => action.category === category);
    }

    /**
     * Get handler for an action
     */
    getHandler(name: string): ActionHandler | undefined {
        return this.get(name)?.handler;
    }

    /**
     * Get all actions suitable for a specific object
     */
    getForObject(objectApiName: string): ActionDefinition[] {
        return this.getValues().filter(action =>
            !action.target_object || action.target_object === objectApiName
        );
    }
}

export const Actions = new ActionRegistry();
