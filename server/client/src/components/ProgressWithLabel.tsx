import Box from '@mui/material/Box';
import type { LinearProgressProps } from '@mui/material/LinearProgress';
import LinearProgress from '@mui/material/LinearProgress';
import Typography from '@mui/material/Typography';

function LinearProgressWithLabel(props: LinearProgressProps & { value: number }) {
    return (
        <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <Box sx={{ width: '100%', mr: 1 }}>
                <LinearProgress variant='determinate' {...props} />
            </Box>
            <Box sx={{ minWidth: 35 }}>
                <Typography
                    variant='body2'
                    sx={{ color: 'text.secondary' }}
                >{`${Math.round(props.value * 10) / 10}%`}</Typography>
            </Box>
        </Box>
    );
}

interface ProgressWithLabelProps {
    percentage: number;
}

export default function ProgressWithLabel({ percentage }: ProgressWithLabelProps) {
    return (
        <Box sx={{ width: '100%' }}>
            <LinearProgressWithLabel value={percentage} />
        </Box>
    );
}
