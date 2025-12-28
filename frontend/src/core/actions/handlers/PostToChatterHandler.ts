/**
 * Post to Chatter Action Handler
 *
 * Posts a message to the record's Chatter feed.
 *
 * @module core/actions/handlers/PostToChatterHandler
 */

import { ActionHandlerModule } from '../ActionHandlerTypes';

export const handler: ActionHandlerModule = {
    handler: {
        actionType: 'PostToChatter',
        label: 'Post to Chatter',
        description: 'Post a message to the record feed when automation rules trigger.',
        category: 'Notifications',
        icon: 'MessageSquare',
        sortOrder: 20,

        configSchema: [
            {
                api_name: 'message',
                label: 'Message Template',
                type: 'TextArea',
                required: true,
                help_text: 'The message to post to the record feed. Use {Field__c} syntax for merge fields.'
            }
        ],

        handler: async (db, record, config, flowName, tx, _, currentUser) => {
            await db.persistence.insert('feed_item', {
                body: config.message || `System Notification: Automation rule "${flowName}" was triggered.`,
                parent_id: record.id,
                type: 'TrackedChange'
            }, currentUser || null, tx);
        }
    }
};
