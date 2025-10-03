import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import Divider from '@mui/material/Divider';
import List from '@mui/material/List';
import Slide from '@mui/material/Slide';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import useScrollTrigger from '@mui/material/useScrollTrigger';
import React from 'react';
import { DownloadQueue } from './components/DownloadQueue';
import { RecordingListItem, RecordingListItemSkeleton } from './components/RecordingListItem';
import { fixRecording, type Recording } from './models';

function isReady(r: Recording): boolean {
    fixRecording(r);
    return ((r.end as Date).valueOf() < new Date().valueOf());
}

interface HideOnScrollProps {
    children?: React.ReactElement<unknown>;
}

function HideOnScroll({ children }: HideOnScrollProps) {
    const trigger = useScrollTrigger();

    return (
        <Slide appear={false} direction='down' in={!trigger}>
            {children ?? <div />}
        </Slide>
    );
}

interface YearDividerProps {
    year: number;
}

function YearDivider({ year }: YearDividerProps) {
    return (<Divider textAlign='left'>{year}</Divider>);
}

export function Layout() {
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
            return <YearDivider year={curYear} key={'year-' + curYear} />;
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
        Array.from({ length: 12, }).map((_, i) => <RecordingListItemSkeleton key={'skeleton-' + i} />);

    return (
        <Box>
            <HideOnScroll>
                <AppBar position='fixed'>
                    <Toolbar>
                        <Typography variant='h6' component='div'>
                            zt-dl - Zattoo Downloader
                        </Typography>
                    </Toolbar>
                </AppBar>
            </HideOnScroll>
            <Toolbar />{/* placeholder for top toolbar */}
            <Container>
                <Box sx={{ my: 2, }}>
                    <List>{listItems}</List>
                </Box>
            </Container>
            <Toolbar />{/* placeholder for bottom toolbar */}
            <AppBar position='fixed' color='primary' sx={{
                top: 'auto',
                bottom: 0,
            }}>
                <DownloadQueue />
            </AppBar>
        </Box>
    );
}
