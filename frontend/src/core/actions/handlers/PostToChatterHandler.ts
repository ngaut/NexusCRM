/**
 * Post to Chatter Action Handler
 *
 * Posts a message to the record's Chatter feed.
 *
 * @module core/actions/handlers/PostToChatterHandler
 */

import { ActionHandlerModule } from '../ActionHandlerTypes';
import { SYSTEM_TABLE_NAMES, FIELDS_SYSTEM_FEEDITEM } from '../../../generated-schema';

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
            await db.persistence.insert(SYSTEM_TABLE_NAMES.SYSTEM_FEEDITEM, {
                [FIELDS_SYSTEM_FEEDITEM.BODY]: config.message || `System Notification: Automation rule "${flowName}" was triggered.`,
                [FIELDS_SYSTEM_FEEDITEM.PARENT_ID]: record.id,
                [FIELDS_SYSTEM_FEEDITEM.TYPE]: 'TrackedChange'
            }, currentUser || null, tx);
        }
    }
};
