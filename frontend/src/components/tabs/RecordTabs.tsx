import React from 'react';
import { SObject } from '../../types';

export interface TabProps {
    recordId?: string;
    objectApiName?: string;
    record?: SObject;
}

export const RecordFieldsTab: React.FC<TabProps> = ({ record }) => {
    return (
        <div className="p-4">
            <h3 className="text-lg font-medium mb-4">Record Details</h3>
            {/* Placeholder for record fields */}
            <div className="bg-slate-50 p-4 rounded border border-slate-200">
                <pre className="text-xs overflow-auto">{JSON.stringify(record, null, 2)}</pre>
            </div>
        </div>
    );
};

export const RelatedListsTab: React.FC<TabProps> = ({ objectApiName, recordId }) => {
    return (
        <div className="p-4">
            <h3 className="text-lg font-medium mb-4">Related Lists</h3>
            <div className="text-slate-500">Related items for {objectApiName} ({recordId})</div>
        </div>
    );
};

export const RecordFeedTab: React.FC<TabProps> = ({ recordId }) => {
    return (
        <div className="p-4">
            <h3 className="text-lg font-medium mb-4">Activity & Feed</h3>
            <div className="text-slate-500">Feed for record {recordId}</div>
        </div>
    );
};
