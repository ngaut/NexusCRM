
import React from 'react';
import { LeadConvertModal as ModalComponent } from '../../components/modals/LeadConvertModal';
import { ActionPlugin, ActionPluginProps, PluginType } from '../../core/plugins/types';

export const LeadConvertPlugin: ActionPlugin = {
    name: 'LeadConvertModal',
    type: PluginType.ACTION,
    description: 'Converts a Lead into an Account, Contact, and Opportunity',
    component: (props: ActionPluginProps) => {
        // Adapt generic props to specific component props if needed
        // LeadConvertModal expects 'lead', generic passes 'record'
        return <ModalComponent
            lead={props.record}
            isOpen={props.isOpen}
            onClose={props.onClose}
            onSuccess={props.onSuccess}
        />;
    }
};
