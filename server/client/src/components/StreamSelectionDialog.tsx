import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import ListSubheader from '@mui/material/ListSubheader';
import React from 'react';
import type { SourceStream } from '../models';

function compareSourceStreams(a: SourceStream, b: SourceStream): number {
    let res = a.type?.localeCompare(b.type || '');
    if (res !== undefined && res !== 0) {
        return res;
    }
    res = a.index - b.index;
    return res;
}

function idForSourceStream(s: SourceStream): string {
    return 'source-stream-' + s.index;
}

interface SourceStreamsListItemsProps {
    sourceStreams: SourceStream[];
    selectedStreams: SourceStream[];
    onToggleItem: (s: SourceStream) => void;
}

function SourceStreamsListItems({ sourceStreams, selectedStreams, onToggleItem }: SourceStreamsListItemsProps) {
    let lastType: string | undefined;

    return (
        sourceStreams.map((s) => {
            const id = idForSourceStream(s);
            const needTypeSection = lastType != s.type;
            lastType = s.type;
            return (
                <>
                    {needTypeSection ? <ListSubheader>{s.type}</ListSubheader> : null}
                    <ListItem key={id} disablePadding>
                        <ListItemButton role={undefined} onClick={() => onToggleItem(s)}>
                            <ListItemIcon>
                                <Checkbox
                                    edge='start'
                                    checked={selectedStreams.includes(s)}
                                    tabIndex={-1}
                                    disableRipple
                                    aria-labelledby={id} />
                            </ListItemIcon>
                            <ListItemText id={id} primary={s.desc} />
                        </ListItemButton>
                    </ListItem>
                </>
            )
        })
    );
}

interface StreamSelectionDialogProps {
    open: boolean;
    onClose: (selected: SourceStream[]) => void;
    sourceStreams: SourceStream[];
}

export function StreamSelectionDialog({
    open, onClose,
    sourceStreams,
}: StreamSelectionDialogProps) {
    const [selected, setSelected] = React.useState<SourceStream[]>([]);

    function handleClose() {
        onClose(selected);
    };

    function onToggleItem(s: SourceStream) {
        const curIdx = selected.indexOf(s);
        const newSelected = [...selected];

        if (curIdx === -1) {
            newSelected.push(s);
        } else {
            newSelected.splice(curIdx, 1);
        }
        setSelected(newSelected);
    }

    const sortedStreams = sourceStreams.sort(compareSourceStreams);
    const items = (
        <SourceStreamsListItems sourceStreams={sortedStreams}
            selectedStreams={selected} onToggleItem={onToggleItem} />
    );
    const numSelected = selected.length;

    return (
        <Dialog open={open}>
            <DialogTitle>Select Streams to Download</DialogTitle>
            <DialogContent>
                <List sx={{ pt: 0 }}>
                    {items}
                </List>
            </DialogContent>
            <DialogActions>
                <Button disabled={numSelected <= 0} onClick={handleClose}>Submit</Button>
            </DialogActions>
        </Dialog>
    );
}
