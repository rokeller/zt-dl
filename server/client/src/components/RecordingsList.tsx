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

    React.useEffect(() => {
        loadRecordings();
    }, []);

    async function loadRecordings() {
        try {
            const resp = await fetch('/api/recordings/');
            if (resp.ok) {
                const rec = (await resp.json()) as Recording[];
                setRecordings(rec);
            }
        } catch (e) {
            console.error('error fetching recordings:', e);
            throw e;
        }
    }

    let lastYear: number | undefined;
    function renderConditionalYearDivider(r: Recording) {
        const curYear = (r.start as Date).getFullYear();
        if (curYear != lastYear) {
            lastYear = curYear;
            return (
                <YearDivider year={curYear} key={'year-' + curYear} />
            );
        }
        return null;
    }

    const listItems = recordings ?
        recordings?.filter(isReady).map((r) => (
            <>
                {renderConditionalYearDivider(r)}
                < RecordingListItem key={r.id} recording={r} />
            </>
        )) :
        Array.from({ length: 12, }).map((_, i) => (
            <RecordingListItemSkeleton key={'skeleton-' + i} />
        ));

    return (
        <List>{listItems}</List>
    );
}
