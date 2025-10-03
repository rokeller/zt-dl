import Icon from '@mui/material/Icon';
import IconButton from '@mui/material/IconButton';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Skeleton from '@mui/material/Skeleton';
import { styled } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import type React from 'react';
import type { Recording } from '../models';
import { formatDate } from '../utils';

const Thumbnail = styled('img')({
    maxWidth: 100,
    maxHeight: 50,
    marginRight: 8,
});

interface RecordingListItemProps {
    recording: Recording;
}

function DownloadRecording({ recording }: React.PropsWithChildren<RecordingListItemProps>) {
    async function startDownload() {
        const filename = recording.episode_title && recording.episode_title.length > 0 ?
            `${recording.title} - ${recording.episode_title}.mp4` :
            `${recording.title}.mp4`;

        try {
            const resp = await fetch('/api/recordings/' + recording.id + '/enqueue', {
                method: 'POST',
                body: new URLSearchParams({
                    filename,
                }),
            })
            if (resp.ok) {
                // alert('recording enqueued: ' + filename)
            } else {
                alert('could not enqueue: ' + filename);
            }
        } catch (e) {
            alert('failed to enqueue');
            console.error('failed to enqueue recording download:', e);
        }
    }

    return (
        <IconButton edge='end' aria-label='comments' onClick={startDownload}>
            <Icon>download</Icon>
        </IconButton>
    );
}

export function RecordingListItem({ recording }: React.PropsWithChildren<RecordingListItemProps>) {
    const r = recording;

    return (
        <ListItem secondaryAction={
            <DownloadRecording recording={r} />
        }>
            <ListItemIcon>
                <Thumbnail src={r.image_url} />
            </ListItemIcon>
            <ListItemText>
                <Typography variant='body2'>{r.cid}</Typography>
                {r.episode_title && r.episode_title.length ?
                    <Typography variant='body1'>{r.title}: {r.episode_title}</Typography> :
                    <Typography variant='body1'>{r.title}</Typography>
                }
                <Typography variant='caption'>
                    {formatDate(r.start)}
                    {' - '}
                    {formatDate(r.end)}
                </Typography>
            </ListItemText>
        </ListItem>
    );
}

export function RecordingListItemSkeleton() {
    return (
        <ListItem secondaryAction={
            <Skeleton variant="circular"><IconButton /></Skeleton>
        }>
            <ListItemIcon>
                <Skeleton sx={{
                    width: 66,
                    height: 50,
                    marginRight: 8,
                }} />
            </ListItemIcon>
            <ListItemText>
                <Typography variant='body2'><Skeleton /></Typography>
                <Typography variant='body1'><Skeleton /></Typography>
                <Typography variant='caption'><Skeleton /></Typography>
            </ListItemText>
        </ListItem>
    )
}
