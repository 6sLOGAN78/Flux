import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AIInsightsPage } from './AIInsightsPage';
import {
  CTRPredictionChart,
  CTRPredictionPoint,
} from '@/components/ai/CTRPredictionChart';
import {
  AnomalyEventStream,
  AnomalyEvent,
} from '@/components/ai/AnomalyEventStream';
import {
  OptimizationTipsCard,
  OptimizationTip,
} from '@/components/ai/OptimizationTipsCard';

describe('Predictive AI Forecasting & Anomaly Detection Center', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockChartData: CTRPredictionPoint[] = [
    { hour: '12:00', actual: 4.2, predicted: 4.1 },
    { hour: '14:00', actual: 5.8, predicted: 5.7 },
    { hour: '16:00', actual: null, predicted: 6.4 },
    { hour: '18:00', actual: null, predicted: 7.1 },
  ];

  const mockAnomalies: AnomalyEvent[] = [
    {
      id: 'anom_1',
      type: 'traffic_spike',
      slug: 'summer-sale',
      zScore: 3.84,
      description: 'Sudden +480% referral surge from Hacker News',
      timestamp: '2026-08-19T22:40:00Z',
    },
    {
      id: 'anom_2',
      type: 'bot_surge',
      slug: 'checkout-v2',
      zScore: 4.12,
      description: 'Anomalous crawler requests detected from IP subnet',
      timestamp: '2026-08-19T22:35:00Z',
    },
  ];

  const mockTips: OptimizationTip[] = [
    {
      id: 'tip_1',
      title: 'Optimal Distribution Window',
      description: 'Publishing Twitter/X short links at 14:00 UTC yields 2.4x higher conversion.',
      impact: 'high',
      actionLabel: 'Schedule Campaign',
    },
  ];

  it('renders CTRPredictionChart with actual and forecast metrics', () => {
    const html = renderToString(
      <CTRPredictionChart data={mockChartData} />
    );

    expect(html).toContain('CTR Forecasting &amp; Trend Trajectory');
    expect(html).toContain('AI Predictive Model');
  });

  it('renders AnomalyEventStream with Z-scores and type badges', () => {
    const html = renderToString(
      <AnomalyEventStream
        anomalies={mockAnomalies}
        onResolveAnomaly={() => {}}
      />
    );

    expect(html).toContain('Real-Time Anomaly Stream');
    expect(html).toContain('summer-sale');
    expect(html).toContain('Z: +3.84');
    expect(html).toContain('bot_surge');
  });

  it('renders OptimizationTipsCard with AI recommendation items', () => {
    const html = renderToString(
      <OptimizationTipsCard tips={mockTips} />
    );

    expect(html).toContain('AI Optimization Recommendations');
    expect(html).toContain('Optimal Distribution Window');
    expect(html).toContain('High Impact');
  });

  it('renders full AIInsightsPage with prediction chart, anomaly feed, and KPI cards', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <AIInsightsPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Predictive AI Insights');
    expect(html).toContain('CTR Forecasting &amp; Trend Trajectory');
    expect(html).toContain('Real-Time Anomaly Stream');
  });
});
