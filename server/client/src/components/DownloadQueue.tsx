import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import { useSnackbar } from 'notistack';
import React from 'react';
import type {
    ClientEvent, DownloadStartedEvent, PendingDownload, ProgressUpdatedEvent,
    ServerEvent, SourceStream, StateUpdatedEvent
} from '../models';
import { DownloadProgress } from './DownloadProgress';
import { QueueFabMenu } from './QueueFabMenu';
import { StreamSelectionDialog } from './StreamSelectionDialog';

type SourceStreamsSelectedHandler = (streams: SourceStream[]) => void;

function noopSourceStreamSelectionHandler() { }

function sendEvent(ws: WebSocket, e: ClientEvent) {
    if (!ws) {
        return;
    }
    const json = JSON.stringify(e, null, 0);
    ws.send(json);
}

function selectStreams(ws: WebSocket, correlation: string, streams: SourceStream[]) {
    streams = streams.map((s) => ({ index: s.index }));
    sendEvent(ws, { correlation, streamsSelected: { streams } });
}

export function DownloadQueue() {
    const { enqueueSnackbar } = useSnackbar();
    const [pending, setPending] = React.useState<PendingDownload[]>([]);
    const [progress, setProgress] = React.useState<ProgressUpdatedEvent>();
    const [downloading, setDownloading] = React.useState<DownloadStartedEvent>();
    const [state, setState] = React.useState<StateUpdatedEvent>();
    const [sourceStreams, setSourceStreams] = React.useState<SourceStream[]>([]);
    const [onSourceStreamsSelected, setOnSourceStreamsSelected] = React.useState<SourceStreamsSelectedHandler>();

    React.useEffect(() => {
        const websocket = new WebSocket('ws://' + window.location.host + '/api/queues/events');
        let connected = false;

        websocket.onopen = () => {
            connected = true;
            enqueueSnackbar(
                'Successfully connected to zt-dl event stream.',
                { variant: 'info', });
        };
        websocket.onerror = (event) => {
            if (!connected) {
                return;
            }
            enqueueSnackbar(
                'Error from zt-dl event stream.',
                { variant: 'error', });
            console.error('websocket error:', event);
        };
        websocket.onmessage = (event) => {
            const e = JSON.parse(event.data) as ServerEvent;
            if (e.queueUpdated) {
                setPending(e.queueUpdated.queue);
            } else if (e.downloadStarted) {
                enqueueSnackbar(
                    `Started download of "${e.downloadStarted.filename}" ...`,
                    { variant: 'info', });
                setDownloading(e.downloadStarted);
                setState(undefined);
            } else if (e.progressUpdated) {
                setProgress(e.progressUpdated);
                setState(undefined);
            } else if (e.downloadErrored) {
                enqueueSnackbar(
                    `Download of "${e.downloadErrored.filename}" failed: ${e.downloadErrored.reason}`,
                    { variant: 'error', });
                setState(undefined);
                setProgress(undefined);
            } else if (e.stateUpdated) {
                setState(e.stateUpdated);
                setProgress(undefined);
            } else if (e.selectStreams) {
                setSourceStreams(e.selectStreams.streams);
                const handler = (streams: SourceStream[]) => {
                    selectStreams(websocket, e.correlation || '', streams);
                    setOnSourceStreamsSelected(undefined);
                }
                setOnSourceStreamsSelected(() => handler);
            } else {
                console.info('received unknown/unsupported event from server:', e, event);
            }
        };
        websocket.onclose = (event) => {
            if (!connected) {
                return;
            }

            enqueueSnackbar(
                'Disconnected from zt-dl event stream unexpectedly: ' + event.code,
                { variant: 'error', });
            console.warn('disconnected from websocket server:', event);
            connected = false;
        };

        return () => websocket.close();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    return (
        <Toolbar sx={{ gap: 2, }}>
            <StreamSelectionDialog
                open={onSourceStreamsSelected != undefined}
                sourceStreams={sourceStreams}
                onClose={onSourceStreamsSelected || noopSourceStreamSelectionHandler} />
            <QueueFabMenu queue={pending} />
            {downloading ?
                progress ?
                    <DownloadProgress filename={downloading.filename} progress={progress} /> :
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
