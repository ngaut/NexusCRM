/**
 * Action Handlers Index
 *
 * Exports all available action handlers for easy registration.
 *
 * @module core/actions/handlers
 */

export { handler as CreateTaskHandler } from './CreateTaskHandler';
export { handler as PostToChatterHandler } from './PostToChatterHandler';
export { handler as SendNotificationHandler } from './SendNotificationHandler';
export { handler as SendEmailHandler } from './SendEmailHandler';
export { handler as UpdateRecordHandler } from './UpdateRecordHandler';
export { handler as CreateRecordHandler } from './CreateRecordHandler';
export { handler as OutboundWebhookHandler } from './OutboundWebhookHandler';

/**
 * Get all standard action handlers as an array
 */
import * as Handlers from './index';

export const ALL_HANDLERS = [
    Handlers.CreateTaskHandler,
    Handlers.PostToChatterHandler,
    Handlers.SendNotificationHandler,
    Handlers.SendEmailHandler,
    Handlers.UpdateRecordHandler,
    Handlers.CreateRecordHandler,
    Handlers.OutboundWebhookHandler,
];
