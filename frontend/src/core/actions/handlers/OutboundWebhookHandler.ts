/**
 * Outbound Webhook Action Handler
 *
 * Sends HTTP requests to external webhooks (e.g., Zapier, Slack).
 * Supports both direct URL configuration and metadata-driven webhook references.
 *
 * @module core/actions/handlers/OutboundWebhookHandler
 */

import { ActionHandlerModule, TiDBServiceManager } from '../ActionHandlerTypes';
import { SObject } from '../../../types';
import { SYSTEM_TABLE_NAMES } from '../../../generated-schema';

export const handler: ActionHandlerModule = {
    handler: {
        actionType: 'OutboundWebhook',
        label: 'Send Outbound Webhook',
        description: 'Send data to external services via HTTP webhook when automation triggers.',
        category: 'Integration',
        icon: 'Zap',
        sortOrder: 70,

        configSchema: [
            {
                api_name: 'webhookName',
                label: 'Webhook Configuration',
                type: 'Text',
                required: false,
                help_text: 'Reference a webhook from _System_Webhook (e.g., SlackNotifier). Leave empty to use direct URL below.'
            },
            {
                api_name: 'url',
                label: 'Direct URL (optional)',
                type: 'Url',
                required: false,
                help_text: 'Direct webhook URL if not using a named webhook configuration above'
            }
        ],

        handler: async (db, record, config, flowName, tx, _, currentUser) => {
            try {
                let webhookConfig: SObject | Record<string, unknown> | null = null;

                // Load webhook from metadata if webhookName is provided
                if (config.webhookName) {
                    const webhooks = await db.query(SYSTEM_TABLE_NAMES.SYSTEM_WEBHOOK, {
                        name: config.webhookName,
                        is_active: true
                    }, tx) as unknown[]; // db.query returns unknown[], safe to cast to array access, but items are unknown

                    if (webhooks.length === 0) {
                        console.warn(`[OutboundWebhookHandler] Webhook "${config.webhookName}" not found or inactive`);
                        await db.logSystemEvent('WARN', 'Integration', `Webhook ${config.webhookName} not found or inactive`, {
                            flowName
                        });
                        return;
                    }

                    webhookConfig = webhooks[0] as Record<string, unknown>;
                }
                // Fallback to direct URL
                else if (config.url) {
                    webhookConfig = {
                        url: config.url,
                        method: 'POST',
                        headers: null,
                        authType: 'None',
                        authConfig: null,
                        timeout: 30000,
                        retryAttempts: 3,
                        retryDelay: 1000
                    };
                } else {
                    console.warn('[OutboundWebhookHandler] No webhook name or URL provided');
                    return;
                }

                const { url, method, headers, authType, authConfig, timeout, retryAttempts, retryDelay } = webhookConfig;

                // Prepare payload
                const payload = {
                    event: 'FlowTrigger',
                    flow: flowName,
                    timestamp: new Date().toISOString(),
                    recordId: record.id,
                    data: record,
                    user: currentUser?.name || 'System'
                };

                // Prepare headers
                const requestHeaders: Record<string, string> = {
                    'Content-Type': 'application/json',
                    'User-Agent': 'NexusCRM-Webhook/2.5.0',
                    ...(headers && typeof headers === 'string' ? JSON.parse(headers) : {})
                };

                // Add authentication headers
                if (authType && authType !== 'None' && authConfig) {
                    const auth = typeof authConfig === 'string' ? JSON.parse(authConfig) : authConfig;

                    if (authType === 'Bearer') {
                        requestHeaders['Authorization'] = `Bearer ${auth.token}`;
                    } else if (authType === 'Basic') {
                        const credentials = Buffer.from(`${auth.username}:${auth.password}`).toString('base64');
                        requestHeaders['Authorization'] = `Basic ${credentials}`;
                    } else if (authType === 'ApiKey') {
                        requestHeaders[auth.headerName || 'X-API-Key'] = auth.apiKey;
                    }
                }

                await db.logSystemEvent('INFO', 'Integration', `Webhook triggered for ${url}`, {
                    flowName,
                    webhookName: config.webhookName || 'Direct URL',
                    status: 'Pending'
                });

                if (url && typeof url === 'string' && url.startsWith('http')) {
                    // Fire and forget with retry logic - do not await to prevent blocking transaction
                    // Webhook calls are async and don't affect transaction success
                    executeWebhookWithRetry(
                        url,
                        (method as string) || 'POST',
                        requestHeaders,
                        payload,
                        typeof timeout === 'number' ? timeout : 30000,
                        typeof retryAttempts === 'number' ? retryAttempts : 3,
                        typeof retryDelay === 'number' ? retryDelay : 1000,
                        typeof config.webhookName === 'string' ? config.webhookName : (url || 'Webhook'),
                        db,
                        flowName
                    );
                } else {
                    console.warn(`[OutboundWebhookHandler] Invalid webhook URL: ${url}`);
                }
            } catch (e: unknown) {
                const errMsg = e instanceof Error ? e.message : String(e);
                console.error('[OutboundWebhookHandler] Webhook action error:', e);
                await db.logSystemEvent('ERROR', 'Integration', 'Webhook Action Failed', {
                    error: errMsg,
                    flowName
                });
                // Don't throw - webhook failures shouldn't rollback the transaction
            }
        }
    }
};

/**
 * Execute webhook with retry logic
 */
async function executeWebhookWithRetry(
    url: string,
    method: string,
    headers: Record<string, string>,
    payload: unknown,
    timeout: number,
    maxRetries: number,
    retryDelay: number,
    webhookIdentifier: string,
    db: TiDBServiceManager,
    flowName: string
) {
    let attempt = 0;

    const tryRequest = async (): Promise<void> => {
        attempt++;
        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), timeout);

            const response = await fetch(url, {
                method,
                headers,
                body: JSON.stringify(payload),
                signal: controller.signal
            });

            clearTimeout(timeoutId);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }



        } catch (err: unknown) {
            const errMsg = err instanceof Error ? err.message : String(err);
            const isLastAttempt = attempt >= maxRetries;
            console.warn(`[OutboundWebhookHandler] Webhook attempt ${attempt}/${maxRetries} failed: ${errMsg}`);

            if (!isLastAttempt) {
                // Wait before retry
                await new Promise(resolve => setTimeout(resolve, retryDelay));
                return tryRequest();
            } else {
                console.error(`[OutboundWebhookHandler] Webhook failed after ${maxRetries} attempts: ${webhookIdentifier}`);
                // Could log final failure to database here
            }
        }
    };

    // Start the retry chain (fire and forget)
    tryRequest().catch(err => {
        console.error('[OutboundWebhookHandler] Unexpected error in retry chain:', err);
    });
}
