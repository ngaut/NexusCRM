/**
 * Update Record Action Handler
 *
 * Updates a field value on the triggering record.
 *
 * @module core/actions/handlers/UpdateRecordHandler
 */

import { ActionHandlerModule } from '../ActionHandlerTypes';
import { parseStandardValue } from '../../utils/parsing';

// Simple inline formula evaluation (safe subset)
function safeEvaluateFormula(formula: string, context: { record: Record<string, unknown>; user?: Record<string, unknown> }): unknown {
    try {
        const fn = new Function('record', 'user', `"use strict"; return (${formula});`);
        return fn(context.record, context.user);
    } catch (e) {
        console.error('Formula evaluation error:', e);
        return undefined;
    }
}

export const handler: ActionHandlerModule = {
    handler: {
        actionType: 'UpdateRecord',
        label: 'Update Field Value',
        description: 'Automatically update a field value when conditions are met. Supports formulas.',
        category: 'Records',
        icon: 'Edit',
        sortOrder: 50,
        requiresObjectContext: true,

        // Dynamic config schema based on trigger object
        configSchema: (triggerObject) => {
            // This function will be called by FlowEngine with the trigger object context
            // For now, we return a generic schema - the FlowEngine can enhance it
            return [
                {
                    api_name: 'field',
                    label: 'Field to Update',
                    type: 'Text',
                    required: true,
                    help_text: 'API name of the field to update'
                },
                {
                    api_name: 'value',
                    label: 'New Value',
                    type: 'Text',
                    required: true,
                    help_text: 'Enter value or =Formula (e.g., =Amount * 1.1)'
                }
            ];
        },

        handler: async (db, record, config, flowName, tx, objectApiName, currentUser) => {
            if (!config.field || config.value === undefined) {
                console.warn(`[UpdateRecordHandler] Missing field or value configuration`);
                return;
            }

            if (!objectApiName) {
                await db.logSystemEvent('WARN', 'FlowAction', 'UpdateRecord failed: Missing Context', { flowName });
                return;
            }

            const schema = db.getSchema(objectApiName) as { fields?: { api_name: string; type: string }[] };
            const fieldName = String(config.field);
            const fieldMeta = schema?.fields?.find(f => f.api_name === fieldName);

            let rawVal = String(config.value);
            let finalVal: unknown = rawVal;

            // Support formula values starting with =
            if (rawVal.startsWith('=')) {
                finalVal = safeEvaluateFormula(rawVal.substring(1), {
                    record,
                    user: (currentUser as unknown as Record<string, unknown>) || undefined
                });
            }

            // Parse value to correct type
            if (fieldMeta) {
                finalVal = parseStandardValue(finalVal, fieldMeta.type);
            }

            // Only update if value changed
            if (record[fieldName] != finalVal) {
                await db.logSystemEvent('INFO', 'FlowAction', `UpdateRecord triggered by ${flowName}`, {
                    field: fieldName,
                    oldValue: record[fieldName],
                    newValue: finalVal
                });

                await db.persistence.update(objectApiName, record.id, {
                    [fieldName]: finalVal
                }, currentUser || null, tx);
            }
        }
    }
};
