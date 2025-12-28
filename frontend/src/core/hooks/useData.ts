import { useState, useEffect, useCallback } from 'react';
import { dataAPI, QueryRequest } from '../../infrastructure/api/data';
import type { SObject, SearchResult } from '../../types';

export function useRecord(objectApiName: string, recordId?: string) {
    const [record, setRecord] = useState<SObject | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    const fetchRecord = useCallback(async () => {
        if (!recordId) {
            setRecord(null);
            return;
        }
        setLoading(true);
        setError(null);
        try {
            const data = await dataAPI.getRecord(objectApiName, recordId);
            setRecord(data);
        } catch (err) {
            setError(err instanceof Error ? err : new Error('An unknown error occurred'));
        } finally {
            setLoading(false);
        }
    }, [objectApiName, recordId]);

    useEffect(() => {
        fetchRecord();
    }, [fetchRecord]);

    return { record, loading, error, refetch: fetchRecord };
}

export function useQuery(request: QueryRequest) {
    const [records, setRecords] = useState<SObject[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    const fetchRecords = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await dataAPI.query(request);
            setRecords(data);
        } catch (err) {
            setError(err instanceof Error ? err : new Error('An unknown error occurred'));
        } finally {
            setLoading(false);
        }
    }, [JSON.stringify(request)]); // Safe JSON dependency for object comparison

    useEffect(() => {
        fetchRecords();
    }, [fetchRecords]);

    return { records, loading, error, refetch: fetchRecords };
}

export function useSearch(term: string) {
    const [results, setResults] = useState<SearchResult[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    useEffect(() => {
        if (!term) {
            setResults([]);
            return;
        }
        const timer = setTimeout(async () => {
            setLoading(true);
            try {
                const data = await dataAPI.search(term);
                setResults(data);
            } catch (err) {
                setError(err instanceof Error ? err : new Error('An unknown error occurred'));
            } finally {
                setLoading(false);
            }
        }, 300); // Debounce

        return () => clearTimeout(timer);
    }, [term]);

    return { results, loading, error };
}
