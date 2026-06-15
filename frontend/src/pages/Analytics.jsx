import React, { useState, useEffect } from 'react';

export default function Analytics({ userRole, apiHeaders, triggerToast }) {
  const [report, setReport] = useState({
    items: [],
    totalFifo: 0,
    totalLifo: 0,
    totalAverageCost: 0
  });
  const [loading, setLoading] = useState(true);

  const fetchValuations = async () => {
    try {
      setLoading(true);
      const res = await fetch('http://localhost:8080/api/reports/valuation', { headers: apiHeaders });
      if (res.ok) {
        const data = await res.json();
        setReport(data);
      }
    } catch (err) {
      console.error(err);
      triggerToast('Error loading valuation reports');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchValuations();
  }, [userRole]);

  const formatPrice = (price) => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(price);
  };

  // Find max value for SVG bar sizing calculations
  const maxValue = Math.max(report.totalFifo, report.totalLifo, report.totalAverageCost, 1000);
  const getPercentage = (value) => {
    return (value / maxValue) * 100;
  };

  return (
    <div>
      <div className="actions-panel">
        <div>
          <h2>Financial Valuation & Analytics Reports</h2>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            Comparative audit valuations for FIFO, LIFO, and Average Cost accounting.
          </p>
        </div>
        
        <button className="btn btn-outline" onClick={fetchValuations}>
          🔄 Refresh
        </button>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-muted)' }}>Loading financial valuations...</div>
      ) : (
        <div>
          {/* SVG Visual Chart */}
          <div className="chart-container">
            <h3 style={{ fontSize: '1.1rem', marginBottom: '0.5rem' }}>Total Asset Valuation Breakdown</h3>
            
            <div className="chart-bar-group">
              {/* FIFO */}
              <div className="chart-bar-row">
                <span className="chart-bar-label">FIFO Valuation</span>
                <div className="chart-bar-track">
                  <div 
                    className="chart-bar-fill fill-fifo" 
                    style={{ width: `${getPercentage(report.totalFifo)}%` }}
                  />
                </div>
                <span style={{ fontWeight: '700', width: '120px', textAlign: 'right' }}>
                  {formatPrice(report.totalFifo)}
                </span>
              </div>

              {/* LIFO */}
              <div className="chart-bar-row">
                <span className="chart-bar-label">LIFO Valuation</span>
                <div className="chart-bar-track">
                  <div 
                    className="chart-bar-fill fill-lifo" 
                    style={{ width: `${getPercentage(report.totalLifo)}%` }}
                  />
                </div>
                <span style={{ fontWeight: '700', width: '120px', textAlign: 'right', color: 'var(--warning)' }}>
                  {formatPrice(report.totalLifo)}
                </span>
              </div>

              {/* Average Cost */}
              <div className="chart-bar-row">
                <span className="chart-bar-label">Average Cost</span>
                <div className="chart-bar-track">
                  <div 
                    className="chart-bar-fill fill-avg" 
                    style={{ width: `${getPercentage(report.totalAverageCost)}%` }}
                  />
                </div>
                <span style={{ fontWeight: '700', width: '120px', textAlign: 'right', color: 'var(--success)' }}>
                  {formatPrice(report.totalAverageCost)}
                </span>
              </div>
            </div>

            <div className="chart-legend">
              <div className="legend-item">
                <div className="legend-color fill-fifo" />
                <span>First In, First Out (FIFO)</span>
              </div>
              <div className="legend-item">
                <div className="legend-color fill-lifo" />
                <span>Last In, First Out (LIFO)</span>
              </div>
              <div className="legend-item">
                <div className="legend-color fill-avg" />
                <span>Weighted Average Cost</span>
              </div>
            </div>
          </div>

          {/* Valuations Table per Product */}
          <div className="table-panel">
            <h3 style={{ padding: '1.2rem 1.5rem 0.5rem', fontSize: '1.2rem' }}>SKU Valuation Ledger</h3>
            <div className="table-wrapper">
              {report.items.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-muted)' }}>No items to value.</div>
              ) : (
                <table className="table">
                  <thead>
                    <tr>
                      <th>SKU</th>
                      <th>Product Name</th>
                      <th>Stock On Hand</th>
                      <th>Catalog Base Cost</th>
                      <th>FIFO Value</th>
                      <th>LIFO Value</th>
                      <th>Avg Cost Value</th>
                    </tr>
                  </thead>
                  <tbody>
                    {report.items.map((item) => (
                      <tr key={item.sku}>
                        <td style={{ fontFamily: 'monospace', fontWeight: '600', color: 'var(--text-primary)' }}>{item.sku}</td>
                        <td>{item.productName}</td>
                        <td style={{ fontWeight: '500' }}>{item.stockOnHand}</td>
                        <td style={{ color: 'var(--text-muted)' }}>{formatPrice(item.baseCost)}</td>
                        <td style={{ fontWeight: '600', color: 'var(--accent-color)' }}>{formatPrice(item.fifoValuation)}</td>
                        <td style={{ fontWeight: '600', color: 'var(--warning)' }}>{formatPrice(item.lifoValuation)}</td>
                        <td style={{ fontWeight: '600', color: 'var(--success)' }}>{formatPrice(item.avgCostVal)}</td>
                      </tr>
                    ))}
                    {/* Summary row */}
                    <tr style={{ backgroundColor: 'rgba(0,0,0,0.15)', fontWeight: '700', borderTop: '2px solid rgba(255,255,255,0.05)' }}>
                      <td colSpan="2">TOTAL LEDGER VALUATION</td>
                      <td>
                        {report.items.reduce((sum, item) => sum + item.stockOnHand, 0)}
                      </td>
                      <td>-</td>
                      <td style={{ color: 'var(--accent-color)', fontSize: '1.05rem' }}>{formatPrice(report.totalFIFO)}</td>
                      <td style={{ color: 'var(--warning)', fontSize: '1.05rem' }}>{formatPrice(report.totalLIFO)}</td>
                      <td style={{ color: 'var(--success)', fontSize: '1.05rem' }}>{formatPrice(report.totalAverageCost)}</td>
                    </tr>
                  </tbody>
                </table>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
