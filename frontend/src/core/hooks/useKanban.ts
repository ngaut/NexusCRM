import { useState, useEffect, useMemo } from 'react';
import { metadataAPI } from '../../infrastructure/api/metadata';
import { dataAPI } from '../../infrastructure/api/data';
import { ObjectMetadata, SObject } from '../../types';
import { getSafeString } from '../utils/recordUtils';
import { COMMON_FIELDS } from '../constants';
import { FIELD_TYPES } from '@shared/generated/constants';

interface UseKanbanProps {
    objectApiName?: string;
    config?: Record<string, unknown>; // Widget config
}

interface UseKanbanResult {
    schema: ObjectMetadata | null;
    columns: string[];
    groupedRecords: Record<string, SObject[]>;
    loading: boolean;
    error: string | null;
}

export function useKanban({ objectApiName, config }: UseKanbanProps): UseKanbanResult {
    const [schema, setSchema] = useState<ObjectMetadata | null>(null);
    const [records, setRecords] = useState<SObject[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!objectApiName) {
            setLoading(false);
            return;
        }

        setLoading(true);
        const fetchData = async () => {
            try {
                // 1. Get Schema
                const schemaRes = await metadataAPI.getSchema(objectApiName);
                const schemaData = schemaRes.schema;
                setSchema(schemaData);

                // 2. Fetch Data
                const queryConfig = config?.query as { filterExpr?: string, group_by?: string } | undefined;
                const filterExpr = queryConfig?.filterExpr;

                const recordData = await dataAPI.query({
                    objectApiName,
                    filterExpr,
                    limit: 50
                });

                // Explicit cast if needed, but dataAPI usually returns SObject[] (proxied)
                setRecords(recordData as SObject[]);
                setLoading(false);
            } catch (err: unknown) {
                setError(err instanceof Error ? err.message : "Failed to load data");
                setLoading(false);
            }
        };

        fetchData();
    }, [objectApiName, config]);

    // Grouping Logic
    const { columns, groupedRecords } = useMemo(() => {
        if (!schema) return { columns: [], groupedRecords: {} };

        const queryConfig = config?.query as { filterExpr?: string, group_by?: string } | undefined;
        const groupByField = queryConfig?.group_by || schema.kanban_group_by || COMMON_FIELDS.STATUS;
        const fields = schema.fields || [];
        const fieldMeta = fields.find(f => f.api_name === groupByField);

        // Define Columns
        let derivedColumns: string[] = [];
        if (fieldMeta?.type === 'Picklist' && fieldMeta.options && fieldMeta.options.length > 0) {
            derivedColumns = fieldMeta.options;
        } else {
            // Deduce columns from data (fallback)
            const values = new Set(records.map(r => getSafeString(r[groupByField], 'Unassigned')));
            derivedColumns = Array.from(values).sort();
        }

        const groups: Record<string, SObject[]> = {};
        derivedColumns.forEach(c => groups[c] = []);

        // Distribute records
        records.forEach(r => {
            const val = getSafeString(r[groupByField], 'Unassigned');
            // If value matches a column, add it. If not, add to existing group or create new if valid?
            // The original logic was: if not in groups, init array.
            if (!groups[val]) groups[val] = [];
            groups[val].push(r);
        });

        // Ensure we iterate defined columns + any extra found
        const allKeys = Array.from(new Set([...derivedColumns, ...Object.keys(groups)]));

        return { columns: allKeys, groupedRecords: groups };

    }, [schema, records, config]);

    return { schema, columns, groupedRecords, loading, error };
}
