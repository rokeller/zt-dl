import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import Slide from '@mui/material/Slide';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import useScrollTrigger from '@mui/material/useScrollTrigger';
import React from 'react';
import { DownloadQueue } from './components/DownloadQueue';
import { RecordingsList } from './components/RecordingsList';

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

export function Layout() {
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
                    <RecordingsList />
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
