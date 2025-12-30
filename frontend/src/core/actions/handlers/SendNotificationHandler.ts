/**
 * Send Notification Action Handler
 *
 * Sends an in-app notification to users.
 *
 * @module core/actions/handlers/SendNotificationHandler
 */

import { ActionHandlerModule } from '../ActionHandlerTypes';
import { SYSTEM_TABLE_NAMES, FIELDS_SYSTEM_FEEDITEM } from '../../../generated-schema';

export const handler: ActionHandlerModule = {
    handler: {
        actionType: 'SendNotification',
        label: 'Send In-App Notification',
        description: 'Send a notification that appears in the user interface.',
        category: 'Notifications',
        icon: 'Bell',
        sortOrder: 30,

        configSchema: [
            {
                api_name: 'message',
                label: 'Notification Text',
                type: 'Text',
                required: true,
                help_text: 'The notification message to display to users'
            }
        ],

        handler: async (db, record, config, flowName, tx, _, currentUser) => {
            const message = `🔔 [ALERT] ${config.message || 'System Notification'}`;
            await db.persistence.insert(SYSTEM_TABLE_NAMES.SYSTEM_FEEDITEM, {
                [FIELDS_SYSTEM_FEEDITEM.BODY]: message,
                [FIELDS_SYSTEM_FEEDITEM.PARENT_ID]: record.id,
                [FIELDS_SYSTEM_FEEDITEM.TYPE]: 'TextPost'
            }, currentUser || null, tx);
        }
    }
};
