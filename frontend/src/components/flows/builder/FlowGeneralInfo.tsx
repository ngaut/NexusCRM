import React from 'react';
import { FLOW_STATUS, FlowStatus } from '../../../core/constants/FlowConstants';

interface FlowGeneralInfoProps {
    name: string;
    setName: (name: string) => void;
    status: FlowStatus;
    setStatus: (status: FlowStatus) => void;
}

export const FlowGeneralInfo: React.FC<FlowGeneralInfoProps> = ({
    name,
    setName,
    status,
    setStatus
}) => {
    return (
        <div className="space-y-4">
            <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Flow Name *
                </label>
                <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="e.g., Create Task on New Lead"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg 
                  bg-white dark:bg-gray-700 text-gray-900 dark:text-white 
                  focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                />
            </div>

            <div className="flex items-center gap-4">
                <label className="flex items-center gap-2 cursor-pointer">
                    <input
                        type="radio"
                        checked={status === FLOW_STATUS.DRAFT}
                        onChange={() => setStatus(FLOW_STATUS.DRAFT)}
                        className="text-purple-500 focus:ring-purple-500"
                    />
                    <span className="text-sm text-gray-700 dark:text-gray-300">Draft</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                    <input
                        type="radio"
                        checked={status === FLOW_STATUS.ACTIVE}
                        onChange={() => setStatus(FLOW_STATUS.ACTIVE)}
                        className="text-green-500 focus:ring-green-500"
                    />
                    <span className="text-sm text-gray-700 dark:text-gray-300">{FLOW_STATUS.ACTIVE}</span>
                </label>
            </div>
        </div>
    );
};
