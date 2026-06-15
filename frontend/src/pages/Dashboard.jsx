import React, { useState, useEffect } from 'react';

export default function Dashboard({ userRole, setSession, triggerToast, apiHeaders }) {
  const [stats, setStats] = useState({
    totalSKUs: 0,
    lowStockCount: 0,
    totalPhysicalItems: 0
  });
  const [lowStockAlerts, setLowStockAlerts] = useState([]);
  const [recentLogs, setRecentLogs] = useState([]);
  const [loading, setLoading] = useState(true);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [productsRes, stockRes, lowRes, logsRes] = await Promise.all([
        fetch('http://localhost:8080/api/products', { headers: apiHeaders }),
        fetch('http://localhost:8080/api/stock', { headers: apiHeaders }),
        fetch('http://localhost:8080/api/stock/low-alerts', { headers: apiHeaders }),
        fetch('http://localhost:8080/api/audit', { headers: apiHeaders })
      ]);

      if (productsRes.ok && stockRes.ok && lowRes.ok && logsRes.ok) {
        const products = await productsRes.json();
        const stock = await stockRes.json();
        const lowStock = await lowRes.json();
        const logs = await logsRes.json();

        // Calculate summary counts
        const totalSKUs = products.length;
        const lowStockCount = lowStock.length;
        const totalPhysicalItems = stock.reduce((sum, item) => sum + item.stockOnHand, 0);

        setStats({ totalSKUs, lowStockCount, totalPhysicalItems });
        setLowStockAlerts(lowStock);
        setRecentLogs(logs.slice(0, 8)); // Top 8 newest logs
      }
    } catch (err) {
      console.error('Error loading dashboard stats:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [userRole]);

  const handleRoleChange = async (roleKey) => {
    let username = 'admin';
    let password = 'adminpass';

    if (roleKey === 'WarehouseStaff') {
      username = 'staff';
      password = 'staffpass';
    } else if (roleKey === 'SalesRep') {
      username = 'sales';
      password = 'salespass';
    }

    try {
      const res = await fetch('http://localhost:8080/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });

      if (res.ok) {
        const session = await res.json();
        setSession(session);
        triggerToast(`Switched credentials to simulated role: ${session.role}`);
      } else {
        throw new Error('Simulation authentication failed');
      }
    } catch (err) {
      console.error(err);
      triggerToast('Error switching roles. Make sure backend is running.');
    }
  };

  const formatTime = (timeStr) => {
    return new Date(timeStr).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  const getActionBadgeClass = (action) => {
    switch (action) {
      case 'Create': return 'alert-success';
      case 'Receive': return 'alert-success';
      case 'Ship': return 'alert-danger';
      case 'Adjust': return 'alert-warning';
      case 'Reserve': return 'alert-warning';
      default: return 'alert-success';
    }
  };

  if (loading && recentLogs.length === 0) {
    return <div style={{ textAlign: 'center', padding: '4rem 0', color: 'var(--text-secondary)' }}>Loading overview statistics...</div>;
  }

  return (
    <div>
      <div className="actions-panel" style={{ marginBottom: '1.5rem' }}>
        <div>
          <h2>Operational Overview</h2>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            Real-time stock indicators and transaction trails.
          </p>
        </div>
        
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.8rem' }}>
          <span style={{ fontSize: '0.9rem', color: 'var(--text-secondary)' }}>Simulate User Credentials:</span>
          <select 
            className="form-select" 
            value={userRole} 
            onChange={(e) => handleRoleChange(e.target.value)}
            style={{ width: '180px' }}
          >
            <option value="Admin">Admin (Full Control)</option>
            <option value="WarehouseStaff">Warehouse Staff</option>
            <option value="SalesRep">Sales Representative</option>
          </select>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="metrics-row">
        <div className="metric-box">
          <span className="metric-box-title">Catalog SKUs</span>
          <span className="metric-box-value">{stats.totalSKUs}</span>
          <span className="metric-box-desc">Unique catalog products listed</span>
        </div>
        <div className="metric-box">
          <span className="metric-box-title" style={{ color: stats.lowStockCount > 0 ? 'var(--danger)' : 'inherit' }}>
            Low Stock Alerts
          </span>
          <span className="metric-box-value" style={{ color: stats.lowStockCount > 0 ? 'var(--danger)' : 'var(--success)' }}>
            {stats.lowStockCount}
          </span>
          <span className="metric-box-desc">SKUs below safe reorder thresholds</span>
        </div>
        <div className="metric-box">
          <span className="metric-box-title">Physical Inventory</span>
          <span className="metric-box-value">{stats.totalPhysicalItems.toLocaleString()}</span>
          <span className="metric-box-desc">Total stock on hand in warehouse</span>
        </div>
      </div>

      {/* Main Grid */}
      <div className="dashboard-grid">
        {/* Left Side: Audit Trail */}
        <div className="table-panel" style={{ padding: '1.5rem' }}>
          <div className="panel-heading">
            <h3>Recent Inventory Actions (Audit Trail)</h3>
            <button className="btn btn-outline btn-sm" onClick={fetchData}>Refresh</button>
          </div>
          
          <div className="table-wrapper">
            {recentLogs.length === 0 ? (
              <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>
                No operations logged yet.
              </div>
            ) : (
              <table className="table">
                <thead>
                  <tr>
                    <th>Time</th>
                    <th>User</th>
                    <th>Action</th>
                    <th>SKU</th>
                    <th>Quantity Change</th>
                    <th>Audit Note</th>
                  </tr>
                </thead>
                <tbody>
                  {recentLogs.map((log) => (
                    <tr key={log.id}>
                      <td style={{ fontSize: '0.85rem' }}>{formatTime(log.timestamp)}</td>
                      <td style={{ fontWeight: '500', color: 'var(--text-primary)' }}>{log.username}</td>
                      <td>
                        <span className={`stock-alert-pill ${getActionBadgeClass(log.action)}`}>
                          {log.action}
                        </span>
                      </td>
                      <td style={{ fontFamily: 'monospace' }}>{log.sku || '-'}</td>
                      <td style={{ 
                        fontWeight: '700', 
                        color: log.quantityChange > 0 ? 'var(--success)' : log.quantityChange < 0 ? 'var(--danger)' : 'inherit' 
                      }}>
                        {log.quantityChange > 0 ? `+${log.quantityChange}` : log.quantityChange === 0 ? '-' : log.quantityChange}
                      </td>
                      <td style={{ fontSize: '0.85rem', maxWidth: '220px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {log.note}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>

        {/* Right Side: Low Stock Warnings */}
        <div className="table-panel" style={{ padding: '1.5rem' }}>
          <h3 style={{ marginBottom: '1rem' }}>Low Stock Reminders</h3>
          
          {lowStockAlerts.length === 0 ? (
            <div style={{ 
              textAlign: 'center', 
              padding: '4rem 1rem', 
              color: 'var(--success)', 
              backgroundColor: 'rgba(16,185,129,0.03)',
              border: '1px solid rgba(16,185,129,0.1)',
              borderRadius: 'var(--radius-md)'
            }}>
              <div style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>✓</div>
              <p style={{ fontWeight: '500' }}>All stock levels healthy</p>
              <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginTop: '0.2rem' }}>No reorder points breached.</p>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.8rem' }}>
              {lowStockAlerts.map((alert) => (
                <div 
                  key={alert.id} 
                  style={{ 
                    padding: '0.8rem 1rem', 
                    backgroundColor: 'rgba(239,68,68,0.05)', 
                    border: '1px solid rgba(239,68,68,0.15)',
                    borderRadius: 'var(--radius-sm)',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center'
                  }}
                >
                  <div>
                    <div style={{ fontWeight: '600', fontSize: '0.9rem' }}>{alert.sku}</div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                      Location: Whs {alert.location.warehouse} | Aisle {alert.location.aisle}
                    </div>
                  </div>
                  <div style={{ textAlign: 'right' }}>
                    <div style={{ color: 'var(--danger)', fontWeight: '700', fontSize: '0.95rem' }}>
                      {alert.availableStock} available
                    </div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                      Reorder at: {alert.reorderPoint}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
