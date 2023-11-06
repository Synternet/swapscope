import { Box, MenuItem, Select, SelectChangeEvent, Typography } from '@mui/material';
import { OperationType } from '../../types';

const values: { value: OperationType; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'add', label: 'Add' },
  { value: 'remove', label: 'Remove' },
];

interface OperationTypeFilterProps {
  value: OperationType;
  onChange: (value: OperationType) => void;
}

export function OperationTypeFilter(props: OperationTypeFilterProps) {
  const { value, onChange } = props;

  const handleChange = (event: SelectChangeEvent<OperationType>) => {
    onChange(event.target.value as OperationType);
  };

  return (
    <Box>
      <Typography gutterBottom sx={{ whiteSpace: 'nowrap' }}>
        Operation type
      </Typography>
      <Select value={value} onChange={handleChange} size="small" data-testid="OperationTypeFilter">
        {values.map((x) => (
          <MenuItem key={x.value} value={x.value} data-testid={`OperationTypeFilter-${x.value}`}>
            {x.label}
          </MenuItem>
        ))}
      </Select>
    </Box>
  );
}
