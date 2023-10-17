import { Link } from '@mui/material';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import { formatDate, formatPoolLimit, truncateAddress } from '@src/utils';
import { LiquidityPoolItem } from '../../types';
import { TokenInfo } from './TokenInfo';

interface LiquidityPoolTableProps {
  list: LiquidityPoolItem[];
}

export function LiquidityPoolTable({ list }: LiquidityPoolTableProps) {
  return (
    <TableContainer component={Paper}>
      <Table data-testid="LiquidityPoolTable">
        <TableHead>
          <TableRow>
            <TableCell>Timestamp</TableCell>
            <TableCell>Token A</TableCell>
            <TableCell>Token B</TableCell>
            <TableCell>Liquidity add range</TableCell>
            <TableCell>Transaction</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {list.map((row) => (
            <TableRow key={row.id} sx={{ '&:last-child td, &:last-child th': { border: 0 } }}>
              <TableCell>{formatDate(row.timestamp)}</TableCell>
              <TableCell>
                <TokenInfo token={row.pair[0]} />
              </TableCell>
              <TableCell>
                <TokenInfo token={row.pair[1]} />
              </TableCell>
              <TableCell>
                {formatPoolLimit(row.lowerTokenRatio)} - {formatPoolLimit(row.upperTokenRatio)}
              </TableCell>
              <TableCell>
                <Link href={`https://etherscan.io/tx/${row.txHash}`} target="_blank">
                  {truncateAddress(row.txHash)}
                </Link>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}
