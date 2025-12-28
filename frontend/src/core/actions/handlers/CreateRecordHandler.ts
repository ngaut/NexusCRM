/**
 * Create Record Action Handler
 *
 * Creates a new record of any object type with field mappings.
 *
 * @module core/actions/handlers/CreateRecordHandler
 */

import { ActionHandlerModule } from '../ActionHandlerTypes';

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
        actionType: 'CreateRecord',
        label: 'Create New Record',
        description: 'Create a record of any object type with field mappings from the trigger record.',
        category: 'Records',
        icon: 'Plus Circle',
        sortOrder: 60,

        // Dynamic config schema - could be enhanced to show available objects
        configSchema: [
            {
                api_name: 'target_object',
                label: 'Target Object',
                type: 'Text',
                required: true,
                help_text: 'API name of the object to create (e.g., contact, account)'
            },
            {
                api_name: 'mappings',
                label: 'Field Mappings (JSON)',
                type: 'TextArea',
                required: true,
                default_value: '{"name": "record.name + \' Copy\'"}',
                help_text: 'JSON object mapping target fields to formulas. Example: {"name": "record.name", "amount": "record.amount * 1.1"}',
            }
        ],

        handler: async (db, record, config, flowName, tx, _, currentUser) => {
            if (!config.target_object) {
                console.warn(`[CreateRecordHandler] Missing target_object configuration`);
                return;
            }

            try {
                const data: Record<string, unknown> = {};
                let mappings: Record<string, string>;

                // Parse mappings (handle both string and object formats)
                if (typeof config.mappings === 'string') {
                    mappings = JSON.parse(config.mappings);
                } else {
                    mappings = config.mappings as Record<string, string>;
                }

                const targetObject = String(config.target_object);

                // Evaluate each formula mapping
                for (const [field, formula] of Object.entries(mappings)) {
                    try {
                        const val = safeEvaluateFormula(String(formula), {
                            record,
                            user: (currentUser as unknown as Record<string, unknown>) || undefined
                        });

                        if (val !== undefined) {
                            data[field] = val;
                        }
                    } catch (e: unknown) {
                        const errMsg = e instanceof Error ? e.message : String(e);
                        await db.logSystemEvent('WARN', 'FlowAction', `Mapping Failed: ${field}`, {
                            error: errMsg,
                            flowName,
                            formula
                        });
                    }
                }

                // Create the new record
                await db.persistence.insert(targetObject, data, currentUser || null, tx);

                await db.logSystemEvent('INFO', 'FlowAction', `CreateRecord completed by ${flowName}`, {
                    target_object: targetObject,
                    recordId: record.id
                });
            } catch (e: unknown) {
                const errMsg = e instanceof Error ? e.message : String(e);
                await db.logSystemEvent('ERROR', 'FlowAction', 'CreateRecord Failed', {
                    error: errMsg,
                    flowName,
                    target_object: String(config.target_object)
                });
                throw e; // Re-throw to trigger transaction rollback
            }
        }
    }
};
