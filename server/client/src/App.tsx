import CssBaseline from '@mui/material/CssBaseline';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { SnackbarProvider } from 'notistack';
import { Layout } from './Layout';

const darkTheme = createTheme({
    palette: {
        mode: 'dark',
    },
});


function App() {
    return (
        <ThemeProvider theme={darkTheme}>
            <CssBaseline />
            <SnackbarProvider maxSnack={10}>
                <Layout />
            </SnackbarProvider>
        </ThemeProvider>
    )
}

export default App
