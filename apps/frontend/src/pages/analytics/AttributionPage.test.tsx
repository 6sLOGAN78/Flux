import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { AttributionPage } from './AttributionPage';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useAttributionQuery } from '@/hooks/useAttributionQuery';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/hooks/useAttributionQuery', () => ({
  useAttributionQuery: vi.fn(),
}));

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

function renderWithProviders(ui: React.ReactNode) {
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>
  );
}

describe('AttributionPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state correctly', () => {
    vi.mocked(useAttributionQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
    } as any);

    renderWithProviders(<AttributionPage />);
    expect(screen.getByTestId('loading-state')).toBeInTheDocument();
  });

  it('renders error state correctly', () => {
    vi.mocked(useAttributionQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    } as any);

    renderWithProviders(<AttributionPage />);
    expect(screen.getByTestId('error-state')).toBeInTheDocument();
  });

  it('renders empty state correctly', () => {
    vi.mocked(useAttributionQuery).mockReturnValue({
      data: { model: 'linear', total_conversions: 0, total_attributed_revenue: 0, campaigns: [] },
      isLoading: false,
      isError: false,
    } as any);

    renderWithProviders(<AttributionPage />);
    expect(screen.getByTestId('empty-state')).toBeInTheDocument();
  });

  it('renders attribution data correctly', () => {
    const mockData = {
      model: 'linear',
      total_conversions: 10,
      total_attributed_revenue: 1000.5,
      campaigns: [
        {
          campaign_id: 'c1',
          campaign_name: 'Summer Sale',
          attributed_conversions: 5.5,
          attributed_revenue: 500.25,
        },
        {
          campaign_id: 'c2',
          campaign_name: 'Winter Sale',
          attributed_conversions: 4.5,
          attributed_revenue: 500.25,
        }
      ]
    };

    vi.mocked(useAttributionQuery).mockReturnValue({
      data: mockData,
      isLoading: false,
      isError: false,
    } as any);

    renderWithProviders(<AttributionPage />);
    
    expect(screen.getByTestId('total-conversions')).toHaveTextContent('10');
    expect(screen.getByTestId('total-revenue')).toHaveTextContent('$1000.50');
    
    // Check if table contains campaigns
    expect(screen.getByText('Summer Sale')).toBeInTheDocument();
    expect(screen.getByText('Winter Sale')).toBeInTheDocument();
    expect(screen.getAllByText('$500.25')).toHaveLength(2);
    expect(screen.getAllByText('50.0%')).toHaveLength(2); // Contribution percentage
  });

  it('updates model when selector is changed', async () => {
    vi.mocked(useAttributionQuery).mockReturnValue({
      data: { model: 'linear', total_conversions: 0, total_attributed_revenue: 0, campaigns: [] },
      isLoading: false,
      isError: false,
    } as any);

    renderWithProviders(<AttributionPage />);
    
    const selector = screen.getByTestId('model-selector');
    fireEvent.change(selector, { target: { value: 'first_touch' } });

    await waitFor(() => {
      // The hook should have been called again (indirectly checked by ensuring state update)
      expect((selector as HTMLSelectElement).value).toBe('first_touch');
    });
  });
});
