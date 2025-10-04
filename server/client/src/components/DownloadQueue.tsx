import Badge from '@mui/material/Badge';
import Box from '@mui/material/Box';
import Fab from '@mui/material/Fab';
import Icon from '@mui/material/Icon';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import { useSnackbar } from 'notistack';
import React from 'react';
import type { DownloadErroredEvent, DownloadStartedEvent, PendingDownload, ProgressUpdatedEvent, QueueEvent, StateUpdatedEvent } from '../models';
import { DownloadProgress } from './DownloadProgress';

export function DownloadQueue() {
    const { enqueueSnackbar } = useSnackbar();
    const [pending, setPending] = React.useState<PendingDownload[]>([]);
    const [progress, setProgress] = React.useState<ProgressUpdatedEvent>();
    const [downloading, setDownloading] = React.useState<DownloadStartedEvent>();
    const [error, setError] = React.useState<DownloadErroredEvent>();
    const [state, setState] = React.useState<StateUpdatedEvent>();

    React.useEffect(() => {
        const websocket = new WebSocket('ws://' + window.location.host + '/api/queue/events');

        websocket.onopen = () => {
            enqueueSnackbar(
                'Successfully connected to zt-dl event stream.',
                { variant: 'info', });
        };
        websocket.onerror = (event) => {
            enqueueSnackbar(
                'Error from zt-dl event stream.',
                { variant: 'error', });
            console.error('websocket error:', event);
        };
        websocket.onmessage = (event) => {
            const e = JSON.parse(event.data) as QueueEvent;
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
            } else {
                console.info('received unknown/unsupported event from server:', e, event);
            }
        };
        websocket.onclose = (event) => {
            enqueueSnackbar(
                'Disconnected from zt-dl event stream unexpectedly: ' + event.code,
                { variant: 'error', });
            console.warn('Disconnected from WebSocket server:', event);
        };

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
                            {state ?
                                <Typography>{state.state} - {state.reason}</Typography> :
                                <Typography>Please be patient while the download is prepared ...</Typography>
                            }
                        </Box>
                : <Typography>Not downloading anything right now.</Typography>
            }
        </Toolbar >
    );
}
