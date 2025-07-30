import { useQuery } from 'react-query';
import { QUERY_KEYS } from '@/types/portfolio';
import { Card, LoadingSpinner, ErrorMessage } from './ui';
import { UI_CONSTANTS } from './ui';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8'];

export default function PortfolioChart() {
  const { data, isLoading, error } = useQuery(
    [QUERY_KEYS.PORTFOLIO_SUMMARY],
    async () => {
      const response = await fetch('/api/v1/portfolio/summary');
      if (!response.ok) throw new Error('Failed to fetch portfolio summary');
      return response.json();
    }
  );

  if (isLoading) {
    return (
      <Card title="Portfolio Allocation" className={UI_CONSTANTS.spacing.section}>
        <div className="flex justify-center py-8">
          <LoadingSpinner size="lg" />
        </div>
      </Card>
    );
  }

  if (error) {
    return (
      <Card title="Portfolio Allocation" className={UI_CONSTANTS.spacing.section}>
        <ErrorMessage message="Failed to load portfolio data" />
      </Card>
    );
  }

  const chartData = data?.asset_allocation?.map((item: any) => ({
    name: item.asset_type,
    value: item.total_value,
    percentage: item.percentage,
  })) || [];

  return (
    <Card title="Portfolio Allocation" className={UI_CONSTANTS.spacing.section}>
      <div className="h-[400px]">
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={chartData}
              cx="50%"
              cy="50%"
              labelLine={false}
              outerRadius={120}
              fill="#8884d8"
              dataKey="value"
            >
              {chartData.map((entry: any, index: number) => (
                <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip 
              formatter={(value: number, name: string, props: any) => [
                `$${value.toLocaleString()}`,
                `${name} (${props.payload.percentage}%)`
              ]}
            />
          </PieChart>
        </ResponsiveContainer>
      </div>
    </Card>
  );
}
