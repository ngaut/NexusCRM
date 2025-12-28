import React from 'react';
import { Download, FileDown } from 'lucide-react';

const DataExport: React.FC = () => {
    return (
        <div className="p-6 max-w-7xl mx-auto">
            <div className="flex items-center gap-3 mb-6">
                <div className="p-2 bg-gradient-to-br from-blue-500 to-cyan-600 rounded-xl shadow-lg">
                    <Download className="w-6 h-6 text-white" />
                </div>
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Data Export</h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                        Export system data for backup or reporting
                    </p>
                </div>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-12 text-center">
                <div className="w-16 h-16 bg-gray-100 dark:bg-gray-700 rounded-full flex items-center justify-center mx-auto mb-4">
                    <FileDown className="w-8 h-8 text-gray-400" />
                </div>
                <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">Export Wizard Coming Soon</h3>
                <p className="text-gray-500 dark:text-gray-400 max-w-md mx-auto">
                    The data export wizard is currently under development. Soon you will be able to schedule weekly or monthly exports of all your organization's data.
                </p>
            </div>
        </div>
    );
};

export default DataExport;
