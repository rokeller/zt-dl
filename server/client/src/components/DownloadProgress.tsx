import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import type { ProgressUpdatedEvent } from '../models';
import ProgressWithLabel from './ProgressWithLabel';

interface DownloadProgressProps {
    filename: string;
    progress: ProgressUpdatedEvent;
}

export function DownloadProgress({ filename, progress }: DownloadProgressProps) {
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
