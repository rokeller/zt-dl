import { Typography } from '@mui/material';
import Divider from '@mui/material/Divider';
import List from '@mui/material/List';
import React from 'react';
import { fixRecording, type Recording } from '../models';
import { RecordingListItem, RecordingListItemSkeleton } from './RecordingListItem';

function isReady(r: Recording): boolean {
    fixRecording(r);
    return ((r.end as Date).valueOf() < new Date().valueOf());
}

interface YearDividerProps {
    year: number;
}

function YearDivider({ year }: YearDividerProps) {
    return (<Divider textAlign='left'>{year}</Divider>);
}

export function RecordingsList() {
    const [recordings, setRecordings] = React.useState<Recording[]>();
    const [fetchError, setFetchError] = React.useState<string>();

    async function loadRecordings() {
        try {
            const resp = await fetch('/api/recordings/');
            if (resp.ok) {
                const rec = (await resp.json()) as Recording[];
                setRecordings(rec);
            } else {
                setFetchError("status: " + resp.status);
            }
        } catch (e) {
            console.error('error fetching recordings:', e);
            setFetchError(String(e));
        }
    }

    React.useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        loadRecordings();
    }, []);

    if (fetchError) {
        return (
            <Typography variant='h6' color='error'>
                Failed to get recordings ({fetchError}). Is the server running?
            </Typography>
        );
    }

    let lastYear: number | undefined;
    const listItems = recordings ?
        recordings?.filter(isReady).map((r) => {
            const curYear = (r.start as Date).getFullYear();
            const showDivider = curYear != lastYear;
            // oxlint-disable-next-line react-hooks/immutability
            lastYear = curYear;

            return (
                <>
                    {showDivider && (<YearDivider year={curYear} key={'year-' + curYear} />)}
                    < RecordingListItem key={'rec-' + r.id} recording={r} />
                </>
            );
        }) :
        Array.from({ length: 12, }).map((_, i) => (
            <RecordingListItemSkeleton key={'skeleton-' + i} />
        ));

    return (
        <List>{listItems}</List>
    );
}
