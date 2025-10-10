import Badge from '@mui/material/Badge';
import Fab from '@mui/material/Fab';
import Icon from '@mui/material/Icon';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import Typography from '@mui/material/Typography';
import { useSnackbar } from 'notistack';
import React from 'react';
import type { PendingDownload } from '../models';

interface QueueFabMenuProps {
    queue?: PendingDownload[];
}

function ellipsisStart(txt: string, maxLen: number) {
    if (txt.length <= maxLen) {
        return txt;
    }

    return '…' + txt.substring(txt.length - maxLen + 1);
}

export function QueueFabMenu({ queue }: QueueFabMenuProps) {
    const { enqueueSnackbar } = useSnackbar();
    const [menuAnchorEl, setMenuAnchorEl] = React.useState<HTMLElement | null>(null);
    const fabId = React.useId();
    const menuId = React.useId();
    const menuOpen = Boolean(menuAnchorEl);

    function onClickFab(event: React.MouseEvent<HTMLElement>) {
        setMenuAnchorEl(event.currentTarget);
    }

    function closeMenu() {
        setMenuAnchorEl(null);
    }

    async function dequeueRecording(item: PendingDownload) {
        try {
            const resp = await fetch('/api/recordings/' + item.recordingId + '/dequeue', {
                method: 'POST',
            })
            if (resp.ok) {
                enqueueSnackbar(
                    `Successfully removed download of "${item.filename}" from queue.`,
                    { variant: 'success', });
            } else {
                enqueueSnackbar(
                    `Failed to remove download of "${item.filename}" from queue.`,
                    { variant: 'error', });
            }
        } catch (e) {
            enqueueSnackbar(
                `Failed to remove download of "${item.filename}" from queue: ${e}`,
                { variant: 'error', });
            console.error('failed to remove download of recording from queue:', e);
        }
    }

    const menuItems = queue && queue.length > 0 ?
        queue.map((item) => (
            <MenuItem key={item.recordingId} title='Click to remove'
                onClick={() => dequeueRecording(item)}>
                <Icon color='error'>cancel</Icon>
                <Typography>{ellipsisStart(item.filename, 50)}</Typography>
            </MenuItem>
        )) :
        [(
            <MenuItem key='no_pending_downloads' onClick={closeMenu}>
                <Typography>No pending downloads in queue</Typography>
            </MenuItem>
        )];

    return (
        <>
            <Fab size='small' color='default' aria-label='queue' id={fabId}
                aria-controls={menuOpen ? menuId : undefined}
                aria-expanded={menuOpen ? 'true' : undefined}
                aria-haspopup='true' onClick={onClickFab}>
                <Badge badgeContent={queue?.length} color='secondary'>
                    <Icon>format_list_bulleted</Icon>
                </Badge>
            </Fab>
            <Menu id={menuId} anchorEl={menuAnchorEl} open={menuOpen}
                anchorOrigin={{
                    vertical: 'top',
                    horizontal: 'left',
                }}
                transformOrigin={{
                    vertical: 'bottom',
                    horizontal: 'left',
                }}
                onClose={closeMenu} slotProps={{
                    list: {
                        'aria-labelledby': fabId,
                    },
                }}>
                {menuItems}
            </Menu>
        </>
    );
}
