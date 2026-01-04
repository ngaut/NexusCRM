/**
 * Create Task Action Handler
 *
 * Creates a follow-up task related to the triggering record.
 *
 * @module core/actions/handlers/CreateTaskHandler
 */

import { ActionHandlerModule } from '../ActionHandlerTypes';
import { TIME_MS } from '../../constants';

export const handler: ActionHandlerModule = {
    handler: {
        actionType: 'CreateTask',
        label: 'Create Follow-up Task',
        description: 'Automatically create a task related to the record when conditions are met.',
        category: 'Records',
        icon: 'CheckSquare',
        sortOrder: 10,

        configSchema: [
            {
                api_name: 'subject',
                label: 'Task Subject',
                type: 'Text',
                required: true,
                default_value: 'Follow up',
                help_text: 'The subject line for the created task'
            },
            {
                api_name: 'priority',
                label: 'Priority',
                type: 'Picklist',
                options: ['High', 'Normal', 'Low'],
                default_value: 'Normal',
                help_text: 'Task priority level'
            },
            {
                api_name: 'daysDue',
                label: 'Due in (Days)',
                type: 'Number',
                default_value: '1',
                help_text: 'Number of days until the task is due'
            }
        ],

        handler: async (db, record, config, flowName, tx, _, currentUser) => {
            const subject = (config.subject as string) || `Auto: Follow up on ${record['name'] || 'Record'}`;
            const priority = (config.priority as string) || 'Normal';

            await db.persistence.insert('task', {
                'subject': subject,
                'status': 'Not Started',
                'priority': priority,
                'what_id': record.id,
                // Default due date: tomorrow
                'due_date': new Date(Date.now() + (parseInt(String(config.daysDue || '1')) || 1) * TIME_MS.DAY)
                    .toISOString()
                    .split('T')[0]
            }, currentUser || null, tx);
        }
    }
};
