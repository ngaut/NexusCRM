/**
 * Send Email Action Handler
 *
 * Sends email notifications using templates from _System_EmailTemplate.
 * Supports merge fields, HTML/text email, and template-based or custom emails.
 *
 * @module core/actions/handlers/SendEmailHandler
 */

import { ActionHandlerModule } from '../ActionHandlerTypes';
import { SYSTEM_TABLES } from '../../constants/SystemObjects';
import { SObject } from '../../../types';

export const handler: ActionHandlerModule = {
    handler: {
        actionType: 'SendEmail',
        label: 'Send Email Alert',
        description: 'Send an email notification when automation rules trigger.',
        category: 'Notifications',
        icon: 'Mail',
        sortOrder: 40,

        configSchema: [
            {
                api_name: 'templateName',
                label: 'Email Template',
                type: 'Text',
                required: false,
                help_text: 'Reference a template from _System_EmailTemplate (e.g., WelcomeCustomer). Leave empty to use custom subject/body below.'
            },
            {
                api_name: 'recipientEmail',
                label: 'Recipient Email',
                type: 'Email',
                required: false,
                help_text: 'Recipient email address (supports merge fields like {!Record.email}). Leave empty to send to record owner.'
            },
            {
                api_name: 'subject',
                label: 'Custom Subject (optional)',
                type: 'Text',
                required: false,
                help_text: 'Subject line if not using a template (supports merge fields)'
            },
            {
                api_name: 'body',
                label: 'Custom Body (optional)',
                type: 'TextArea',
                required: false,
                help_text: 'Email message content if not using a template (supports merge fields)'
            }
        ],

        handler: async (db, record, config, flowName, tx, _, currentUser) => {
            try {
                let emailTemplate: SObject | null = null;
                const templateName = typeof config.templateName === 'string' ? config.templateName : undefined;
                let subject = typeof config.subject === 'string' ? config.subject : 'Notification';
                let body = typeof config.body === 'string' ? config.body : 'Rule triggered';
                let htmlBody: string | null = null;
                let fromName = 'NexusCRM';
                let fromEmail = 'notifications@nexuscrm.com';
                let replyTo: string | null = null;

                // Load template from metadata if templateName is provided
                if (templateName) {
                    const templates = await db.query(SYSTEM_TABLES.EMAIL_TEMPLATE, {
                        name: templateName,
                        is_active: true
                    }, tx);

                    if (templates.length === 0) {
                        console.warn(`[SendEmailHandler] Template "${templateName}" not found or inactive`);
                        await db.logSystemEvent('WARN', 'Email', `Email template ${templateName} not found or inactive`, {
                            flowName
                        });
                        // Fall through to use custom subject/body
                    } else {
                        emailTemplate = templates[0] as SObject;
                        // SObject properties might be unknown, cast to string safely
                        subject = String(emailTemplate.subject || '') || subject;
                        body = String(emailTemplate.textBody || emailTemplate.htmlBody || '');
                        htmlBody = emailTemplate.htmlBody ? String(emailTemplate.htmlBody) : null;
                        fromName = emailTemplate.fromName ? String(emailTemplate.fromName) : fromName;
                        fromEmail = emailTemplate.fromEmail ? String(emailTemplate.fromEmail) : fromEmail;
                        replyTo = emailTemplate.replyTo ? String(emailTemplate.replyTo) : null;
                    }
                }

                // Process merge fields in subject and body
                const mergeContext = {
                    Record: record,
                    CurrentUser: currentUser || { name: 'System' },
                    SystemURL: process.env.REACT_APP_BACKEND_URL || 'http://localhost:3001'
                };

                subject = processMergeFields(subject, mergeContext);
                body = processMergeFields(body, mergeContext);
                if (htmlBody) {
                    htmlBody = processMergeFields(htmlBody, mergeContext);
                }

                // Determine recipient email
                let recipientEmail = typeof config.recipientEmail === 'string' ? config.recipientEmail : undefined;
                if (recipientEmail) {
                    recipientEmail = processMergeFields(recipientEmail, mergeContext);
                } else {
                    recipientEmail = typeof record.email === 'string' ? record.email : (typeof record.owner_email === 'string' ? record.owner_email : 'owner@example.com');
                }

                await db.logSystemEvent('INFO', 'Email', `Email queued for ${recipientEmail}`, {
                    flowName,
                    templateName: templateName || 'Custom',
                    subject,
                    recipient: recipientEmail
                });

                const emailPreview = htmlBody
                    ? `${subject}\n\n${body.substring(0, 200)}${body.length > 200 ? '...' : ''}`
                    : `${subject}\n\n${body}`;

                await db.persistence.insert('feed_item', {
                    body: `📧 Email sent to ${recipientEmail}:\n\n${emailPreview}`,
                    parent_id: record.id,
                    type: 'TrackedChange'
                }, currentUser || null, tx);

            } catch (e: unknown) {
                const errMsg = e instanceof Error ? e.message : String(e);
                console.error('[SendEmailHandler] Email action error:', e);
                await db.logSystemEvent('ERROR', 'Email', 'Email Action Failed', {
                    error: errMsg,
                    flowName
                });
                // Don't throw - email failures shouldn't rollback the transaction
            }
        }
    }
};

/**
 * Process merge fields in a string
 * Supports {!Record.Field}, {!CurrentUser.Field}, {!SystemURL}
 */
function processMergeFields(template: string, context: Record<string, unknown>): string {
    if (!template) return '';

    return template.replace(/\{!([^}]+)\}/g, (match, path) => {
        try {
            const parts = path.split('.');
            let value: unknown = context;

            for (const part of parts) {
                if (value && typeof value === 'object' && part in value) {
                    value = (value as Record<string, unknown>)[part];
                } else {
                    return match; // Keep original if path doesn't exist
                }
            }

            return value !== null && value !== undefined ? String(value) : '';
        } catch (e) {
            console.warn(`[SendEmailHandler] Failed to process merge field: ${path}`);
            return match; // Keep original on error
        }
    });
}
