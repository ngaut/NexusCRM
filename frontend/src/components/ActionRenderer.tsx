
import React, { useState } from 'react';
import { Button } from './ui/Button';
import { ActionMetadata, SObject } from '../types';
import * as Icons from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useErrorToast, useSuccessToast } from '../components/ui/Toast';
import { formatApiError } from '../core/utils/errorHandling';
import { ConfirmationModal } from './modals/ConfirmationModal';
import { actionRegistry, ActionContext } from '../registries/ActionRegistry';

interface ActionRendererProps {
    action: ActionMetadata;
    record: SObject;
    onActionComplete?: () => void;
}

export const ActionRenderer: React.FC<ActionRendererProps> = ({ action, record, onActionComplete }) => {
    const navigate = useNavigate();
    const showError = useErrorToast();
    const showSuccess = useSuccessToast();
    const [loading, setLoading] = useState(false);

    // Dynamic Modal State
    const [modalConfig, setModalConfig] = useState<{
        isOpen: boolean;
        component: React.ComponentType<any> | null;
        props: Record<string, unknown>;
    }>({ isOpen: false, component: null, props: {} });

    // Confirmation Modal State
    const [confirmConfig, setConfirmConfig] = useState<{
        isOpen: boolean;
        title: string;
        message: string;
        onConfirm: () => Promise<void>;
    }>({ isOpen: false, title: '', message: '', onConfirm: async () => { } });

    // Dynamic Icon
    const IconComponent = (Icons[action.icon as keyof typeof Icons] || Icons.Zap) as React.ComponentType<{ className?: string }>;

    // Action Context Implementation
    const context: ActionContext = {
        navigate: (path) => navigate(path),
        showSuccess: (msg) => showSuccess(msg),
        showError: (msg) => showError(msg),
        refreshRecord: () => onActionComplete?.(),
        openModal: (component, props) => {
            setModalConfig({
                isOpen: true,
                component,
                props: props || {}
            });
        },
        closeModal: () => {
            setModalConfig(prev => ({ ...prev, isOpen: false }));
        },
        confirm: (options) => {
            setConfirmConfig({
                isOpen: true,
                title: options.title,
                message: options.message,
                onConfirm: options.onConfirm
            });
        }
    };

    // Handle Click
    const handleClick = async () => {
        const handler = actionRegistry.get(action.type);
        if (handler) {
            try {
                await handler.execute(action, record, context);
            } catch (err: unknown) {
                showError(formatApiError(err).message);
            }
        } else {
            showError(`Action type '${action.type}' is not supported.`);
        }
    };

    const handleConfirm = async () => {
        setLoading(true);
        try {
            await confirmConfig.onConfirm();
        } catch (err: unknown) {
            showError(formatApiError(err).message);
        } finally {
            setLoading(false);
            setConfirmConfig(prev => ({ ...prev, isOpen: false }));
        }
    };

    const ModalComponent = modalConfig.component;

    return (
        <>
            <Button
                variant="secondary"
                size="sm"
                onClick={handleClick}
                loading={loading}
                icon={<IconComponent className="w-4 h-4" />}
            >
                {action.label}
            </Button>

            {/* Dynamic Custom Action Modal */}
            {modalConfig.isOpen && ModalComponent && (
                <ModalComponent
                    isOpen={modalConfig.isOpen}
                    onClose={context.closeModal}
                    {...modalConfig.props}
                />
            )}

            {/* Confirmation Modal for API Actions */}
            <ConfirmationModal
                isOpen={confirmConfig.isOpen}
                onClose={() => setConfirmConfig(prev => ({ ...prev, isOpen: false }))}
                onConfirm={handleConfirm}
                title={confirmConfig.title}
                message={confirmConfig.message}
                confirmLabel="Confirm"
                loading={loading}
            />
        </>
    );
};
