
import { ComponentType } from 'react';
import { SObject } from '../../types';

export enum PluginType {
    ACTION = 'ACTION',
    FIELD = 'FIELD',
    VIEW = 'VIEW'
}

export interface BasePlugin {
    name: string;
    type: PluginType;
    description?: string;
}

export interface ActionPlugin extends BasePlugin {
    type: PluginType.ACTION;
    component: ComponentType<ActionPluginProps>;
}

export interface ActionPluginProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
    record: SObject;
    // Allow extra props for flexibility
    [key: string]: unknown;
}

export type Plugin = ActionPlugin; // Union with others later (FieldPlugin | ViewPlugin)
