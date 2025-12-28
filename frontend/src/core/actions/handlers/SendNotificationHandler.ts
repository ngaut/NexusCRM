/**
 * Send Notification Action Handler
 *
 * Sends an in-app notification to users.
 *
 * @module core/actions/handlers/SendNotificationHandler
 */

import { ActionHandlerModule } from '../ActionHandlerTypes';

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
            await db.persistence.insert('feed_item', {
                body: message,
                parent_id: record.id,
                type: 'TextPost'
            }, currentUser || null, tx);
        }
    }
};
