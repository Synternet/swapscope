import { Box } from '@mui/material';

export function Loader() {
  return (
    <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', flex: 1 }}>
      <h3>Loading...</h3>
    </Box>
  );
}
