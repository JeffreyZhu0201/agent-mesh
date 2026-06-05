import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { Home } from './pages/Home';

export default function App() {
  return (
    <ThemeProvider>
      <CssBaseline />
      <Home />
    </ThemeProvider>
  );
}
