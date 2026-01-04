import { ActionMetadata, SObject } from '../types';
import { pluginRegistry } from '../core/plugins/PluginRegistry';
import { dataAPI } from '../infrastructure/api/data';
import { ROUTES } from '../core/constants/Routes';

export interface ActionContext {
    navigate: (path: string) => void;
    showSuccess: (msg: string) => void;
    showError: (msg: string) => void;
    refreshRecord: () => void;
    openModal: (component: React.ComponentType<unknown>, props: Record<string, unknown>) => void;
    closeModal: () => void;
    confirm: (options: { title: string; message: string; onConfirm: () => Promise<void> }) => void;
}

export interface ActionHandler {
    execute: (action: ActionMetadata, record: SObject, context: ActionContext) => Promise<void>;
}

class ActionRegistry {
    private handlers: Map<string, ActionHandler> = new Map();

    register(type: string, handler: ActionHandler) {
        this.handlers.set(type, handler);
    }

    get(type: string): ActionHandler | undefined {
        return this.handlers.get(type);
    }

    has(type: string): boolean {
        return this.handlers.has(type);
    }
}

export const actionRegistry = new ActionRegistry();

// Register Action Handlers

// 1. Link Action
actionRegistry.register('link', {
    execute: async (action, record, context) => {
        if (action.config?.url) {
            window.open(String(action.config.url), '_blank');
        }
    }
});

// 2. Flow Action
actionRegistry.register('flow', {
    execute: async (action, record, context) => {
        if (action.config?.flowId) {
            context.navigate(`/flow/${action.config.flowId}?recordId=${record.id}`);
        }
    }
});

// 3. Modal/Custom Action
actionRegistry.register('modal', {
    execute: async (action, record, context) => {
        if (action.config?.component) {
            const plugin = pluginRegistry.getAction(String(action.config.component));
            if (plugin) {
                context.openModal(plugin.component, {
                    record: record,
                    onSuccess: () => {
                        context.closeModal();
                        context.refreshRecord();
                    },
                    onClose: context.closeModal
                });
            } else {
                console.warn(`Action plugin ${action.config.component} not found`);
                context.showError(`Action plugin '${action.config.component}' not found`);
            }
        }
    }
});

// Map 'Custom' to 'modal' handling
actionRegistry.register('Custom', {
    execute: async (action, record, context) => {
        const modalHandler = actionRegistry.get('modal');
        if (modalHandler) {
            await modalHandler.execute(action, record, context);
        }
    }
});

// 4. API Action
actionRegistry.register('api', {
    execute: async (action, record, context) => {
        context.confirm({
            title: 'Confirm Action',
            message: `Are you sure you want to ${action.label.toLowerCase()}?`,
            onConfirm: async () => {
                await dataAPI.executeAction(action.id, { record_id: record.id });
                context.showSuccess(`${action.label} executed successfully`);
                context.refreshRecord();
            }
        });
    }
});

// 5. UpdateRecord (Edit) Action - Standard action for editing records
actionRegistry.register('UpdateRecord', {
    execute: async (action, record, context) => {
        // Navigate to edit mode
        context.navigate(ROUTES.OBJECT.EDIT(action.object_api_name, record.id as string));
    }
});

// 6. DeleteRecord Action - Standard action for deleting records
actionRegistry.register('DeleteRecord', {
    execute: async (action, record, context) => {
        context.confirm({
            title: 'Delete Record',
            message: 'Are you sure you want to delete this record? This action cannot be undone.',
            onConfirm: async () => {
                await dataAPI.deleteRecord(action.object_api_name, record.id);
                context.showSuccess('Record deleted successfully');
                // Navigate back to list view
                context.navigate(ROUTES.OBJECT.LIST(action.object_api_name));
            }
        });
    }
});
