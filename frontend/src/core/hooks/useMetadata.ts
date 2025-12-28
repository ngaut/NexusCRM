import { useState, useEffect } from 'react';
import { metadataAPI } from '../../infrastructure/api/metadata';
import type { ObjectMetadata, PageLayout, ActionMetadata } from '../../types';

export function useObjectMetadata(objectApiName: string) {
    const [metadata, setMetadata] = useState<ObjectMetadata | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    const load = async () => {
        if (!objectApiName) return;
        setLoading(true);
        try {
            const data = await metadataAPI.getSchema(objectApiName);
            setMetadata(data.schema);
        } catch (err) {
            setError(err instanceof Error ? err : new Error('Unknown error'));
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        load();
    }, [objectApiName]);

    return { metadata, loading, error, refresh: load };
}

export function useLayout(objectApiName: string, type: string = 'Detail') {
    const [layout, setLayout] = useState<PageLayout | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    useEffect(() => {
        if (!objectApiName) return;
        const load = async () => {
            setLoading(true);
            try {
                const data = await metadataAPI.getLayout(objectApiName);
                setLayout(data.layout);
            } catch (err) {
                setError(err instanceof Error ? err : new Error('Unknown error'));
            } finally {
                setLoading(false);
            }
        };
        load();
    }, [objectApiName]);

    return { layout, loading, error };
}

export function useSchemas() {
    const [schemas, setSchemas] = useState<ObjectMetadata[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    const load = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await metadataAPI.getSchemas();
            setSchemas(data.schemas || []);
        } catch (err) {
            setError(err instanceof Error ? err : new Error('Unknown error'));
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        load();
    }, []);

    return { schemas, loading, error, refresh: load };
}

export function useActions(objectApiName: string) {
    const [actions, setActions] = useState<ActionMetadata[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);
    const [refreshKey, setRefreshKey] = useState(0);

    const load = async () => {
        if (!objectApiName) return;
        setLoading(true);
        try {
            const data = await metadataAPI.getActions(objectApiName);
            setActions(data.actions || []);
        } catch (err) {
            setError(err instanceof Error ? err : new Error('Unknown error'));
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        load();
    }, [objectApiName, refreshKey]);

    return { actions, loading, error, refresh: () => setRefreshKey(k => k + 1) };
}
