import Badge from '@mui/material/Badge';
import Box from '@mui/material/Box';
import Fab from '@mui/material/Fab';
import Icon from '@mui/material/Icon';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import React from 'react';
import type { DownloadErroredEvent, DownloadStartedEvent, PendingDownload, ProgressUpdatedEvent, QueueEvent, StateUpdatedEvent } from '../models';
import ProgressWithLabel from './ProgressWithLabel';

interface DownloadProgressProps {
    filename: string;
    progress: ProgressUpdatedEvent;
}

function DownloadProgress({ filename, progress }: DownloadProgressProps) {
    return (
        <>
            <Box>
                <Typography>{progress.elapsed}</Typography>
                <Typography variant='caption'>Elapsed</Typography>
            </Box>
            <Box>
                <Typography>{progress.remaining}</Typography>
                <Typography variant='caption'>Remaining</Typography>
            </Box>
            <Box sx={{ flexGrow: 1, }}>
                <Typography variant='caption'>{filename}</Typography>
                <ProgressWithLabel percentage={(progress.completed || 0) * 100} />
            </Box>
        </>
    );
}

export function DownloadQueue() {
    const [pending, setPending] = React.useState<PendingDownload[]>([]);
    const [progress, setProgress] = React.useState<ProgressUpdatedEvent>();
    const [downloading, setDownloading] = React.useState<DownloadStartedEvent>();
    const [error, setError] = React.useState<DownloadErroredEvent>();
    const [state, setState] = React.useState<StateUpdatedEvent>();

    React.useEffect(() => {
        const websocket = new WebSocket('ws://' + window.location.host + '/api/queue/events');

        websocket.onopen = () => console.log('Connected to WebSocket server');
        websocket.onerror = (event) => console.error('websocket error:', event);
        websocket.onmessage = (event) => {
            const e = JSON.parse(event.data) as QueueEvent;
            console.info('received event from server:', e);
            if (e.queueUpdated) {
                setPending(e.queueUpdated.queue);
            } else if (e.downloadStarted) {
                setDownloading(e.downloadStarted);
                setState(undefined);
                setError(undefined);
            } else if (e.progressUpdated) {
                setProgress(e.progressUpdated);
                setState(undefined);
                setError(undefined);
            } else if (e.downloadErrored) {
                setError(e.downloadErrored);
                setState(undefined);
                setProgress(undefined);
            } else if (e.stateUpdated) {
                setState(e.stateUpdated);
                setError(undefined);
                setProgress(undefined);
            }
        };
        websocket.onclose = () => console.log('Disconnected from WebSocket server');

        return () => websocket.close();
    }, []);

    return (
        <Toolbar sx={{ gap: 2, }}>
            <Fab size='small' color='default'>
                <Badge badgeContent={pending?.length} color='secondary'>
                    <Icon>format_list_bulleted</Icon>
                </Badge>
            </Fab>
            {downloading ?
                progress ?
                    <DownloadProgress filename={downloading.filename} progress={progress} /> :
                    error ?
                        <Typography color='error'>{error.reason}</Typography> :
                        <Box>
                            <Typography variant='caption'>{downloading.filename}</Typography>
                            <Typography>{state?.state} - {state?.reason}</Typography>
                        </Box>
                : <Typography>Not downloading anything right now.</Typography>
            }
        </Toolbar >
    );
}
