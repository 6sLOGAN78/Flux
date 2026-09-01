import { useState, useMemo } from 'react';
import { useAttributionQuery } from '@/hooks/useAttributionQuery';
import { DataTable } from '@/components/ui/DataTable';

const MODELS = [
  { value: 'first_touch', label: 'First Touch' },
  { value: 'last_touch', label: 'Last Touch' },
  { value: 'linear', label: 'Linear' },
  { value: 'time_decay', label: 'Time Decay' },
  { value: 'position_based', label: 'Position Based (U-Shaped)' },
];

const RANGES = [
  { value: '7d', label: 'Last 7 Days' },
  { value: '30d', label: 'Last 30 Days' },
  { value: '90d', label: 'Last 90 Days' },
];

export function AttributionPage() {
  const [activeRange, setActiveRange] = useState('30d');
  const [activeModel, setActiveModel] = useState('linear');

  const { from, to } = useMemo(() => {
    const now = new Date();
    let fromDate = new Date();
    switch (activeRange) {
      case '7d':
        fromDate.setDate(now.getDate() - 7);
        break;
      case '30d':
        fromDate.setDate(now.getDate() - 30);
        break;
      case '90d':
        fromDate.setDate(now.getDate() - 90);
        break;
      default:
        fromDate.setDate(now.getDate() - 30);
    }
    return { from: fromDate.toISOString(), to: now.toISOString() };
  }, [activeRange]);

  const { data, isLoading, isError } = useAttributionQuery(from, to, activeModel);

  // Table columns
  const columns = [
    { key: 'campaign', header: 'Campaign / Touchpoint', accessor: 'campaign_name' },
    { 
      key: 'conversions',
      header: 'Attributed Conversions', 
      accessor: (row: any) => row.attributed_conversions.toFixed(2)
    },
    { 
      key: 'revenue',
      header: 'Attributed Revenue', 
      accessor: (row: any) => `$${row.attributed_revenue.toFixed(2)}`
    },
    {
      key: 'contribution',
      header: 'Contribution',
      accessor: (row: any) => {
        if (!data || data.total_attributed_revenue === 0) return '0%';
        const perc = (row.attributed_revenue / data.total_attributed_revenue) * 100;
        return `${perc.toFixed(1)}%`;
      }
    }
  ];

  return (
    <div className="space-y-6 p-6" data-testid="attribution-page">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center space-y-4 sm:space-y-0">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Multi-Touch Attribution</h1>
          <p className="text-sm text-gray-500 mt-1">Analyze revenue and conversions across your campaigns.</p>
        </div>
      </div>
      
      <div className="flex flex-wrap items-center gap-4 bg-white p-4 rounded-lg shadow-sm border border-gray-200">
        <div className="flex items-center space-x-2">
          <span className="text-sm font-medium text-gray-700">Model:</span>
          <select 
            value={activeModel}
            onChange={(e) => setActiveModel(e.target.value)}
            className="border-gray-300 rounded-md text-sm focus:ring-blue-500 focus:border-blue-500"
            data-testid="model-selector"
          >
            {MODELS.map(m => (
              <option key={m.value} value={m.value}>{m.label}</option>
            ))}
          </select>
        </div>

        <div className="flex items-center space-x-2">
          <span className="text-sm font-medium text-gray-700">Date Range:</span>
          <div className="flex bg-gray-100 p-1 rounded-md" data-testid="date-range-selector">
            {RANGES.map((r) => (
              <button
                key={r.value}
                onClick={() => setActiveRange(r.value)}
                className={`px-3 py-1 text-xs font-medium rounded-sm transition-colors ${
                  activeRange === r.value 
                    ? 'bg-white text-gray-900 shadow-sm' 
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                {r.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        {isLoading ? (
          <div className="py-12 flex justify-center items-center" data-testid="loading-state">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        ) : isError ? (
          <div className="py-12 text-center text-red-600" data-testid="error-state">
            Failed to load attribution data.
          </div>
        ) : !data || data.campaigns.length === 0 ? (
          <div className="py-12 text-center text-gray-500" data-testid="empty-state">
            No attribution data available for this period.
          </div>
        ) : (
          <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="bg-gray-50 p-4 rounded-md">
                <p className="text-sm text-gray-500 font-medium">Total Conversions</p>
                <p className="text-2xl font-bold text-gray-900" data-testid="total-conversions">{data.total_conversions}</p>
              </div>
              <div className="bg-gray-50 p-4 rounded-md">
                <p className="text-sm text-gray-500 font-medium">Attributed Revenue</p>
                <p className="text-2xl font-bold text-gray-900" data-testid="total-revenue">${data.total_attributed_revenue.toFixed(2)}</p>
              </div>
            </div>
            <DataTable 
              columns={columns} 
              data={data.campaigns} 
              keyExtractor={(item) => item.campaign_id}
            />
          </div>
        )}
      </div>
    </div>
  );
}
