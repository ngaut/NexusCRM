import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { dataAPI } from '../infrastructure/api/data';
import { useNotification } from '../contexts/NotificationContext';
import { evaluateExpression, substituteTemplate } from '../core/utils/expressionEvaluator';
import type { ActionConfig, SObject } from '../types';
import * as Icons from 'lucide-react';
import { Loader2 } from 'lucide-react';
import { ConfirmationModal } from './modals/ConfirmationModal';
import { actionHandlerRegistry } from '../core/actions/ActionHandlerRegistry';
import { clientDBAdapter } from '../core/api/ClientDBAdapter';

interface ActionButtonProps {
    action: ActionConfig;
    record?: SObject;
    objectApiName: string;
    variant?: 'header' | 'quick';
    onActionComplete?: () => void;
}

export const ActionButton: React.FC<ActionButtonProps> = ({
    action,
    record,
    objectApiName,
    variant = 'header',
    onActionComplete
}) => {
    const navigate = useNavigate();
    const { success, error } = useNotification();
    const [loading, setLoading] = useState(false);

    // Delete confirmation modal state
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
    const [deleting, setDeleting] = useState(false);

    // Check visibility condition using safe evaluator
    const isVisible = () => {
        if (!action.visibility_condition) return true;
        if (!record) return true;
        return evaluateExpression(action.visibility_condition, record);
    };

    const handleAction = async () => {
        if (loading) return;
        setLoading(true);

        try {
            switch (action.type) {
                case 'Standard':
                    await handleStandardAction();
                    break;
                case 'CreateRecord':
                    await handleCreateRecord();
                    break;
                case 'UpdateRecord':
                    await handleUpdateRecord();
                    break;
                case 'Url':
                    handleUrlAction();
                    break;
                case 'Custom':
                    // Look up handler in registry by type or name
                    // Some actions might store the type in action.name if action.type is 'Custom'
                    // Or action.type itself is the custom type identifier (e.g. 'LeadConvertModal')
                    // The standard defines 'Custom' as the generic type, so specific custom type is likely in 'component' or 'name'.

                    // Strategy: 
                    // 1. Check if 'action.component' matches a registered handler
                    // 2. Check if 'action.name' matches a registered handler

                    let handlerDef = undefined;
                    if (action.component && actionHandlerRegistry.hasHandler(action.component)) {
                        handlerDef = actionHandlerRegistry.getHandler(action.component);
                    } else if (actionHandlerRegistry.hasHandler(action.name)) {
                        handlerDef = actionHandlerRegistry.getHandler(action.name);
                    }

                    if (handlerDef) {
                        // Execute the handler
                        await handlerDef.handler(
                            clientDBAdapter,
                            record || {},
                            action.config || {},
                            'ManualAction',
                            undefined, // Transaction (not supported on client)
                            objectApiName,
                            undefined // Current User (could get from context if needed)
                        );
                    } else {
                        error('Configuration Error', `Handler not found for action: ${action.label}`);
                    }
                    break;
                default:
                // Unknown action type - silently ignore
            }

            onActionComplete?.();
        } catch (err: unknown) {
            const message = err instanceof Error ? err.message : 'Failed to execute action';
            error('Action Failed', message);
        } finally {
            setLoading(false);
        }
    };

    const handleStandardAction = async () => {
        switch (action.name) {
            case 'Delete':
                if (!record) return;
                setShowDeleteConfirm(true);
                break;

            case 'Clone':
                if (!record) return;
                // Create a copy without id and system fields
                const { id, created_date, last_modified_date, created_by_id, last_modified_by_id, ...recordData } = record;
                const newRecord = await dataAPI.createRecord(objectApiName, {
                    ...recordData,
                    name: `${record.name || 'Record'} (Clone)`
                });
                success('Cloned', 'Record cloned successfully');
                navigate(`/object/${objectApiName}/${newRecord.id}`);
                break;

            case 'Edit':
                if (!record) return;
                navigate(`/object/${objectApiName}/${record.id}?edit=true`);
                break;

            case 'Refresh':
                window.location.reload();
                break;

            default:
            // Unknown standard action - silently ignore
        }
    };

    const confirmDelete = async () => {
        if (!record) return;
        setDeleting(true);
        try {
            await dataAPI.deleteRecord(objectApiName, record.id as string);
            success('Deleted', 'Record moved to recycle bin');
            navigate(`/object/${objectApiName}`);
        } catch (err: unknown) {
            const message = err instanceof Error ? err.message : 'Failed to delete record';
            error('Delete Failed', message);
        } finally {
            setDeleting(false);
            setShowDeleteConfirm(false);
        }
    };

    const handleCreateRecord = async () => {
        if (!action.target_object) {
            error('Configuration Error', 'CreateRecord action requires target_object');
            return;
        }

        // Prepare default field values from config
        const defaults: Record<string, unknown> = {};
        if (action.config && record) {
            Object.entries(action.config).forEach(([key, value]) => {
                // Simple template replacement
                if (typeof value === 'string' && value.startsWith('{') && value.endsWith('}')) {
                    const fieldName = value.slice(1, -1);
                    defaults[key] = record[fieldName];
                } else {
                    defaults[key] = value;
                }
            });
        }

        // Navigate to create page with defaults in URL
        // Explicitly cast to Record<string, string> for URLSearchParams if needed, or stringify values
        const paramsObj: Record<string, string> = {};
        Object.entries(defaults).forEach(([k, v]) => paramsObj[k] = String(v));
        const params = new URLSearchParams(paramsObj).toString();

        navigate(`/object/${action.target_object}/new${params ? `?${params}` : ''}`);
    };

    const handleUpdateRecord = async () => {
        if (!record) return;
        if (!action.config) {
            error('Configuration Error', 'UpdateRecord action requires config with field updates');
            return;
        }

        // Apply updates from config
        await dataAPI.updateRecord(objectApiName, record.id as string, action.config);
        success('Updated', 'Record updated successfully');
        window.location.reload(); // Refresh to show changes
    };

    const handleUrlAction = () => {
        if (!action.config?.url) {
            error('Configuration Error', 'Url action requires url in config');
            return;
        }

        // Use safe template substitution
        const url = record ? substituteTemplate(String(action.config.url), record) : String(action.config.url);

        // Open in new tab or same tab based on config
        if (action.config.openInNewTab) {
            window.open(url, '_blank');
        } else {
            window.location.href = url;
        }
    };

    if (!isVisible()) return null;

    // Get icon
    const IconComponent = action.icon ? (Icons as unknown as Record<string, React.ComponentType<{ size?: number }>>)[action.icon] : null;

    // Styling based on variant
    const buttonClasses = variant === 'quick'
        ? 'flex items-center gap-2 px-3 py-2 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-lg hover:bg-slate-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed'
        : 'flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-slate-600 rounded-lg hover:bg-slate-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed';

    return (
        <>
            <button
                onClick={handleAction}
                disabled={loading}
                className={buttonClasses}
                title={action.label}
            >
                {loading ? (
                    <Loader2 size={16} className="animate-spin" />
                ) : IconComponent ? (
                    <IconComponent size={16} />
                ) : null}
                <span>{action.label}</span>
            </button>


            {/* Delete Record Confirmation Modal */}
            <ConfirmationModal
                isOpen={showDeleteConfirm}
                onClose={() => setShowDeleteConfirm(false)}
                onConfirm={confirmDelete}
                title="Delete Record"
                message={`Are you sure you want to delete this ${objectApiName}? This action cannot be undone.`}
                confirmLabel="Delete"
                cancelLabel="Cancel"
                variant="danger"
                loading={deleting}
            />
        </>
    );
};
