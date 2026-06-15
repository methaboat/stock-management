import React, { useState, useEffect } from 'react';

export default function StockLevels({ userRole, apiHeaders, triggerToast }) {
  const [stockLevels, setStockLevels] = useState([]);
  const [products, setProducts] = useState({});
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);

  // Adjustment Modal
  const [isOpen, setIsOpen] = useState(false);
  const [adjustData, setAdjustData] = useState({
    stockId: '',
    sku: '',
    stockOnHand: 0,
    reorderPoint: 10,
    warehouse: 'Main Warehouse',
    aisle: 'A1',
    bin: '01',
    note: ''
  });

  const fetchData = async () => {
    try {
      setLoading(true);
      const [stockRes, productsRes] = await Promise.all([
        fetch('http://localhost:8080/api/stock', { headers: apiHeaders }),
        fetch('http://localhost:8080/api/products', { headers: apiHeaders })
      ]);

      if (stockRes.ok && productsRes.ok) {
        const stockData = await stockRes.json();
        const productsData = await productsRes.json();

        // Create a product lookup map (ID -> Name)
        const pMap = {};
        productsData.forEach(p => {
          pMap[p.id] = p.name;
        });

        setProducts(pMap);
        setStockLevels(stockData);
      }
    } catch (err) {
      console.error(err);
      triggerToast('Error loading stock levels');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [userRole]);

  const handleOpenAdjust = (s) => {
    if (userRole !== 'Admin' && userRole !== 'WarehouseStaff') {
      triggerToast('Access denied: Manual stock adjustments require Admin or WarehouseStaff roles.');
      return;
    }
    setAdjustData({
      stockId: s.id,
      sku: s.sku,
      stockOnHand: s.stockOnHand,
      reorderPoint: s.reorderPoint,
      warehouse: s.location.warehouse || 'Main Warehouse',
      aisle: s.location.aisle || 'A1',
      bin: s.location.bin || '01',
      note: ''
    });
    setIsOpen(true);
  };

  const handleAdjustSubmit = async (e) => {
    e.preventDefault();
    if (adjustData.stockOnHand < 0) {
      alert('Stock on Hand cannot be negative.');
      return;
    }

    const payload = {
      stockId: adjustData.stockId,
      stockOnHand: parseFloat(adjustData.stockOnHand),
      reorderPoint: parseFloat(adjustData.reorderPoint),
      location: {
        warehouse: adjustData.warehouse,
        aisle: adjustData.aisle,
        bin: adjustData.bin
      },
      note: adjustData.note
    };

    try {
      const res = await fetch('http://localhost:8080/api/stock/adjust', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...apiHeaders
        },
        body: JSON.stringify(payload)
      });

      if (res.ok) {
        triggerToast('Stock adjusted successfully');
        setIsOpen(false);
        fetchData();
      } else {
        const text = await res.text();
        alert(text || 'Failed to adjust stock');
      }
    } catch (err) {
      console.error(err);
      alert('Server error adjusting stock');
    }
  };

  const filteredStock = stockLevels.filter(s => 
    s.sku.toLowerCase().includes(search.toLowerCase()) ||
    (products[s.productId] && products[s.productId].toLowerCase().includes(search.toLowerCase())) ||
    (s.location.warehouse && s.location.warehouse.toLowerCase().includes(search.toLowerCase()))
  );

  return (
    <div>
      <div className="actions-panel">
        <div>
          <h2>Real-Time Stock levels & Storage Coordinates</h2>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            Physical quantities on hand, committed, available, and layout coordinates.
          </p>
        </div>
        
        <div style={{ display: 'flex', gap: '1rem' }}>
          <input 
            type="text" 
            className="search-input" 
            placeholder="Search SKU, product name, warehouse..." 
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <button className="btn btn-outline" onClick={fetchData}>
            🔄 Refresh
          </button>
        </div>
      </div>

      <div className="table-panel">
        <div className="table-wrapper">
          {loading ? (
            <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>Loading stock balances...</div>
          ) : filteredStock.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>
              No stock entries found matching your query.
            </div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>SKU</th>
                  <th>Product Name</th>
                  <th>Warehouse Location</th>
                  <th>Stock On Hand (Physical)</th>
                  <th>Committed (Reserved)</th>
                  <th>Available (Saleable)</th>
                  <th>Reorder Point</th>
                  <th>Status Alert</th>
                  <th style={{ textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredStock.map((s) => {
                  const isLow = s.availableStock <= s.reorderPoint;
                  return (
                    <tr key={s.id}>
                      <td style={{ fontFamily: 'monospace', fontWeight: '600', color: 'var(--text-primary)' }}>
                        {s.sku}
                      </td>
                      <td>{products[s.productId] || 'Unknown Product'}</td>
                      <td>
                        <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
                          Whse: {s.location.warehouse} | Aisle: {s.location.aisle} | Bin: {s.location.bin}
                        </span>
                      </td>
                      <td style={{ fontWeight: '600' }}>{s.stockOnHand}</td>
                      <td style={{ color: 'var(--warning)', fontWeight: '500' }}>{s.committedStock}</td>
                      <td style={{ 
                        fontWeight: '700', 
                        color: isLow ? 'var(--danger)' : 'var(--success)' 
                      }}>
                        {s.availableStock}
                      </td>
                      <td>{s.reorderPoint}</td>
                      <td>
                        <span className={`stock-alert-pill ${isLow ? 'alert-danger' : 'alert-success'}`}>
                          {isLow ? 'LOW STOCK' : 'HEALTHY'}
                        </span>
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        <button 
                          className="btn btn-primary btn-sm"
                          onClick={() => handleOpenAdjust(s)}
                          disabled={userRole !== 'Admin' && userRole !== 'WarehouseStaff'}
                        >
                          Adjust Stock
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* Adjust Stock Modal */}
      {isOpen && (
        <div className="modal-overlay" onClick={() => setIsOpen(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Manual Inventory Count Override: {adjustData.sku}</h3>
              <button className="modal-close" onClick={() => setIsOpen(false)}>&times;</button>
            </div>
            
            <form onSubmit={handleAdjustSubmit}>
              <div className="modal-body">
                <div className="form-grid">
                  <div className="form-group">
                    <label className="form-label">Physical Stock on Hand *</label>
                    <input 
                      type="number" 
                      className="form-input"
                      value={adjustData.stockOnHand}
                      onChange={e => setAdjustData({...adjustData, stockOnHand: e.target.value})}
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">Reorder Alert Threshold *</label>
                    <input 
                      type="number" 
                      className="form-input"
                      value={adjustData.reorderPoint}
                      onChange={e => setAdjustData({...adjustData, reorderPoint: e.target.value})}
                      required
                    />
                  </div>
                </div>

                <div className="form-group" style={{ marginBottom: '1rem' }}>
                  <label className="form-label">Warehouse Name</label>
                  <input 
                    type="text" 
                    className="form-input"
                    value={adjustData.warehouse}
                    onChange={e => setAdjustData({...adjustData, warehouse: e.target.value})}
                  />
                </div>

                <div className="form-grid">
                  <div className="form-group">
                    <label className="form-label">Aisle Code</label>
                    <input 
                      type="text" 
                      className="form-input"
                      placeholder="e.g. A3"
                      value={adjustData.aisle}
                      onChange={e => setAdjustData({...adjustData, aisle: e.target.value})}
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">Bin Code</label>
                    <input 
                      type="text" 
                      className="form-input"
                      placeholder="e.g. 14"
                      value={adjustData.bin}
                      onChange={e => setAdjustData({...adjustData, bin: e.target.value})}
                    />
                  </div>
                </div>

                <div className="form-group form-group-full">
                  <label className="form-label">Discrepancy Note (Reason for change) *</label>
                  <textarea 
                    className="form-textarea"
                    placeholder="e.g. Weekly physical audit correction, damaged items disposal, shrinkage correction..."
                    value={adjustData.note}
                    onChange={e => setAdjustData({...adjustData, note: e.target.value})}
                    required
                  />
                </div>
              </div>
              
              <div className="modal-footer">
                <button type="button" className="btn btn-outline" onClick={() => setIsOpen(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Override Counts
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
