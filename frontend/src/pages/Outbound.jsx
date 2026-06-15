import React, { useState, useEffect } from 'react';

export default function Outbound({ userRole, apiHeaders, triggerToast }) {
  const [salesOrders, setSalesOrders] = useState([]);
  const [products, setProducts] = useState([]);
  const [stockLevels, setStockLevels] = useState({}); // SKU -> Available stock map for client-side help
  const [rmas, setRmas] = useState([]);
  const [loading, setLoading] = useState(true);

  // Tabs: 'orders' or 'returns'
  const [activeTab, setActiveTab] = useState('orders');

  // SO Form State
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [customerName, setCustomerName] = useState('');
  const [soItems, setSoItems] = useState([
    { productId: '', sku: '', quantity: 1, unitPrice: 0 }
  ]);

  // RMA Form State
  const [isRmaOpen, setIsRmaOpen] = useState(false);
  const [rmaData, setRmaData] = useState({
    salesOrderId: '',
    productId: '',
    sku: '',
    quantity: 1,
    reason: '',
    status: 'Restocked' // 'Restocked' or 'Damaged'
  });

  const fetchData = async () => {
    try {
      setLoading(true);
      const [soRes, productsRes, stockRes, rmaRes] = await Promise.all([
        fetch('http://localhost:8080/api/outbound/so', { headers: apiHeaders }),
        fetch('http://localhost:8080/api/products', { headers: apiHeaders }),
        fetch('http://localhost:8080/api/stock', { headers: apiHeaders }),
        fetch('http://localhost:8080/api/outbound/rma', { headers: apiHeaders })
      ]);

      if (soRes.ok && productsRes.ok && stockRes.ok && rmaRes.ok) {
        const soData = await soRes.json();
        const productsData = await productsRes.json();
        const stockData = await stockRes.json();
        const rmaData = await rmaRes.json();

        setSalesOrders(soData.sort((a,b) => new Date(b.createdAt) - new Date(a.createdAt)));
        setProducts(productsData);
        setRmas(rmaData.sort((a,b) => new Date(b.createdAt) - new Date(a.createdAt)));

        const stockMap = {};
        stockData.forEach(s => {
          stockMap[s.productId] = s.availableStock;
        });
        setStockLevels(stockMap);
      }
    } catch (err) {
      console.error(err);
      triggerToast('Error loading outbound logistics');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [userRole]);

  // Items rows adding in SO creation
  const handleAddItemRow = () => {
    setSoItems([...soItems, { productId: '', sku: '', quantity: 1, unitPrice: 0 }]);
  };

  const handleRemoveItemRow = (index) => {
    const list = [...soItems];
    list.splice(index, 1);
    setSoItems(list);
  };

  const handleItemRowChange = (index, field, value) => {
    const list = [...soItems];
    if (field === 'productId') {
      const p = products.find(prod => prod.id === value);
      list[index].productId = value;
      list[index].sku = p ? p.sku : '';
      list[index].unitPrice = p ? p.sellingPrice : 0;
    } else if (field === 'quantity') {
      list[index].quantity = parseFloat(value) || 0;
    } else if (field === 'unitPrice') {
      list[index].unitPrice = parseFloat(value) || 0;
    }
    setSoItems(list);
  };

  const handleOpenAddSO = () => {
    if (userRole !== 'Admin' && userRole !== 'SalesRep') {
      triggerToast('Access denied: Creating Sales Orders requires Admin or SalesRep roles.');
      return;
    }
    setCustomerName('');
    setSoItems([{ productId: '', sku: '', quantity: 1, unitPrice: 0 }]);
    setIsAddOpen(true);
  };

  const handleCreateSOSubmit = async (e) => {
    e.preventDefault();
    if (!customerName || soItems.some(item => !item.productId || item.quantity <= 0)) {
      alert('Customer Name is required, and all items must have a product and a quantity > 0.');
      return;
    }

    // Client-side quick availability check
    for (const item of soItems) {
      const avail = stockLevels[item.productId] || 0;
      if (item.quantity > avail) {
        alert(`Insufficient stock for SKU ${item.sku}. Available stock is ${avail}, but you requested ${item.quantity}.`);
        return;
      }
    }

    const payload = {
      customerName,
      items: soItems.map(item => ({
        productId: item.productId,
        sku: item.sku,
        quantity: item.quantity,
        unitPrice: item.unitPrice
      }))
    };

    try {
      const res = await fetch('http://localhost:8080/api/outbound/so', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...apiHeaders
        },
        body: JSON.stringify(payload)
      });

      if (res.ok) {
        triggerToast('Sales Order created and inventory reserved.');
        setIsAddOpen(false);
        fetchData();
      } else {
        const text = await res.text();
        alert(text || 'Failed to create Sales Order');
      }
    } catch (err) {
      console.error(err);
      alert('Error creating Sales Order');
    }
  };

  const handleShipOrder = async (id, soNum) => {
    if (userRole !== 'Admin' && userRole !== 'WarehouseStaff') {
      triggerToast('Access denied: Picking, packing and shipping orders requires Admin or WarehouseStaff roles.');
      return;
    }

    if (!confirm(`Mark ${soNum} as Picked, Packed, and Shipped? This deducts physical stock on hand.`)) {
      return;
    }

    try {
      const res = await fetch(`http://localhost:8080/api/outbound/so/ship/${id}`, {
        method: 'POST',
        headers: apiHeaders
      });

      if (res.ok) {
        triggerToast(`Order ${soNum} shipped successfully`);
        fetchData();
      } else {
        const text = await res.text();
        alert(text || 'Failed to ship order');
      }
    } catch (err) {
      console.error(err);
      alert('Error shipping order');
    }
  };

  const handleOpenRMA = () => {
    if (userRole !== 'Admin' && userRole !== 'WarehouseStaff') {
      triggerToast('Access denied: Returns processing (RMA) requires Admin or WarehouseStaff roles.');
      return;
    }
    
    // Choose the first shipped order and first item for defaults if available
    const shippedOrders = salesOrders.filter(so => so.status === 'Shipped');
    const firstSO = shippedOrders[0];
    const firstItem = firstSO ? firstSO.items[0] : null;

    setRmaData({
      salesOrderId: firstSO ? firstSO.id : '',
      productId: firstItem ? firstItem.productId : '',
      sku: firstItem ? firstItem.sku : '',
      quantity: 1,
      reason: '',
      status: 'Restocked'
    });
    setIsRmaOpen(true);
  };

  const handleRMAOrderChange = (soId) => {
    const so = salesOrders.find(o => o.id === soId);
    const firstItem = so ? so.items[0] : null;
    setRmaData({
      ...rmaData,
      salesOrderId: soId,
      productId: firstItem ? firstItem.productId : '',
      sku: firstItem ? firstItem.sku : '',
      quantity: 1
    });
  };

  const handleRMAItemChange = (prodId) => {
    const p = products.find(prod => prod.id === prodId);
    setRmaData({
      ...rmaData,
      productId: prodId,
      sku: p ? p.sku : ''
    });
  };

  const handleRMASubmit = async (e) => {
    e.preventDefault();
    if (!rmaData.salesOrderId || !rmaData.productId || rmaData.quantity <= 0) {
      alert('Order, Item, and Return Quantity are required.');
      return;
    }

    const payload = {
      salesOrderId: rmaData.salesOrderId,
      productId: rmaData.productId,
      sku: rmaData.sku,
      quantity: parseFloat(rmaData.quantity),
      reason: rmaData.reason || 'Customer returned item',
      status: rmaData.status
    };

    try {
      const res = await fetch('http://localhost:8080/api/outbound/rma', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...apiHeaders
        },
        body: JSON.stringify(payload)
      });

      if (res.ok) {
        triggerToast(rmaData.status === 'Restocked' 
          ? 'Customer return processed and item restocked' 
          : 'Customer return logged as Damaged (not restocked)'
        );
        setIsRmaOpen(false);
        fetchData();
      } else {
        const text = await res.text();
        alert(text || 'Failed to process return');
      }
    } catch (err) {
      console.error(err);
      alert('Error processing customer return');
    }
  };

  const formatPrice = (price) => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(price);
  };

  const formatDate = (dateStr) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  };

  const getStatusClass = (status) => {
    switch (status) {
      case 'Pending': return 'alert-warning';
      case 'Shipped': return 'alert-success';
      case 'Cancelled': return 'alert-danger';
      default: return 'alert-warning';
    }
  };

  // Get shipped orders to show in returns drop down selection
  const shippedOrders = salesOrders.filter(so => so.status === 'Shipped');
  const selectedSOObj = salesOrders.find(o => o.id === rmaData.salesOrderId);

  return (
    <div>
      {/* Tab Switcher */}
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem' }}>
        <button 
          className={`btn ${activeTab === 'orders' ? 'btn-primary' : 'btn-outline'}`}
          onClick={() => setActiveTab('orders')}
        >
          📦 Sales & Outbound Orders
        </button>
        <button 
          className={`btn ${activeTab === 'returns' ? 'btn-primary' : 'btn-outline'}`}
          onClick={() => setActiveTab('returns')}
        >
          🔄 Customer Returns (RMA)
        </button>
      </div>

      {activeTab === 'orders' && (
        <div>
          <div className="actions-panel">
            <div>
              <h2>Outbound Sales Orders</h2>
              <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
                Stock reservations and Pick, Pack, & Ship dispatch tracking.
              </p>
            </div>
            
            <button className="btn btn-primary" onClick={handleOpenAddSO}>
              + Create Sales Order
            </button>
          </div>

          <div className="table-panel">
            <div className="table-wrapper">
              {loading ? (
                <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>Loading sales orders...</div>
              ) : salesOrders.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>
                  No sales orders found. Click "+ Create Sales Order" to reserve inventory for dispatch.
                </div>
              ) : (
                <table className="table">
                  <thead>
                    <tr>
                      <th>SO Number</th>
                      <th>Customer Name</th>
                      <th>Order Date</th>
                      <th>Items (Reserved / Dispatched)</th>
                      <th>Total Value</th>
                      <th>Status</th>
                      <th style={{ textAlign: 'right' }}>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {salesOrders.map((so) => {
                      const itemsSummary = so.items.map(item => 
                        `${item.sku}: ${item.quantity} reserved`
                      ).join(', ');

                      return (
                        <tr key={so.id}>
                          <td style={{ fontFamily: 'monospace', fontWeight: '600', color: 'var(--text-primary)' }}>
                            {so.soNumber}
                          </td>
                          <td>{so.customerName}</td>
                          <td>{formatDate(so.createdAt)}</td>
                          <td style={{ fontSize: '0.85rem', maxWidth: '300px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }} title={itemsSummary}>
                            {itemsSummary}
                          </td>
                          <td style={{ fontWeight: '600' }}>{formatPrice(so.totalAmount)}</td>
                          <td>
                            <span className={`stock-alert-pill ${getStatusClass(so.status)}`}>
                              {so.status}
                            </span>
                          </td>
                          <td style={{ textAlign: 'right' }}>
                            {so.status === 'Pending' ? (
                              <button 
                                className="btn btn-primary btn-sm"
                                onClick={() => handleShipOrder(so.id, so.soNumber)}
                                disabled={userRole !== 'Admin' && userRole !== 'WarehouseStaff'}
                              >
                                Pick, Pack & Ship
                              </button>
                            ) : (
                              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Dispatched</span>
                            )}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        </div>
      )}

      {activeTab === 'returns' && (
        <div>
          <div className="actions-panel">
            <div>
              <h2>Returns Management (RMA)</h2>
              <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
                Log customer returns, log damaged items, and decide restocking locations.
              </p>
            </div>
            
            <button 
              className="btn btn-primary" 
              onClick={handleOpenRMA}
              disabled={shippedOrders.length === 0}
              title={shippedOrders.length === 0 ? 'No shipped orders exist to return' : ''}
            >
              + Log Returned Merchandise
            </button>
          </div>

          <div className="table-panel">
            <div className="table-wrapper">
              {loading ? (
                <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>Loading return records...</div>
              ) : rmas.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>
                  No customer returns recorded yet.
                </div>
              ) : (
                <table className="table">
                  <thead>
                    <tr>
                      <th>Return Date</th>
                      <th>SKU</th>
                      <th>Quantity Returned</th>
                      <th>Return Status</th>
                      <th>Reason Note</th>
                    </tr>
                  </thead>
                  <tbody>
                    {rmas.map((r) => (
                      <tr key={r.id}>
                        <td>{formatDate(r.createdAt)}</td>
                        <td style={{ fontFamily: 'monospace', fontWeight: '600', color: 'var(--text-primary)' }}>{r.sku}</td>
                        <td style={{ fontWeight: '600' }}>{r.quantity}</td>
                        <td>
                          <span className={`stock-alert-pill ${r.status === 'Restocked' ? 'alert-success' : 'alert-danger'}`}>
                            {r.status.toUpperCase()}
                          </span>
                        </td>
                        <td style={{ fontSize: '0.85rem' }}>{r.reason}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Create SO Modal */}
      {isAddOpen && (
        <div className="modal-overlay" onClick={() => setIsAddOpen(false)}>
          <div className="modal-content" style={{ maxWidth: '650px' }} onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Create Sales Order</h3>
              <button className="modal-close" onClick={() => setIsAddOpen(false)}>&times;</button>
            </div>
            
            <form onSubmit={handleCreateSOSubmit}>
              <div className="modal-body">
                <div className="form-group" style={{ marginBottom: '1.2rem' }}>
                  <label className="form-label">Customer Name *</label>
                  <input 
                    type="text" 
                    className="form-input" 
                    placeholder="e.g. Walmart Logistics"
                    value={customerName}
                    onChange={e => setCustomerName(e.target.value)}
                    required
                  />
                </div>

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.8rem' }}>
                  <h4 style={{ fontSize: '0.95rem' }}>Reserve Items *</h4>
                  <button type="button" className="btn btn-outline btn-sm" onClick={handleAddItemRow}>
                    + Add Row
                  </button>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.8rem', maxHeight: '250px', overflowY: 'auto', paddingRight: '0.2rem' }}>
                  {soItems.map((item, index) => {
                    const avail = stockLevels[item.productId] || 0;
                    return (
                      <div key={index} style={{ display: 'flex', gap: '0.6rem', alignItems: 'flex-end' }}>
                        <div className="form-group" style={{ flex: 2 }}>
                          <label className="form-label">Product SKU</label>
                          <select 
                            className="form-select"
                            value={item.productId}
                            onChange={e => handleItemRowChange(index, 'productId', e.target.value)}
                            required
                          >
                            <option value="">-- Choose Product --</option>
                            {products.map(p => (
                              <option key={p.id} value={p.id}>
                                {p.sku} (Avail: {stockLevels[p.id] || 0})
                              </option>
                            ))}
                          </select>
                        </div>

                        <div className="form-group" style={{ flex: 1 }}>
                          <label className="form-label">Qty</label>
                          <input 
                            type="number" 
                            className="form-input"
                            min="1"
                            max={avail}
                            value={item.quantity}
                            onChange={e => handleItemRowChange(index, 'quantity', e.target.value)}
                            required
                          />
                        </div>

                        <div className="form-group" style={{ flex: 1 }}>
                          <label className="form-label">Price (USD)</label>
                          <input 
                            type="number" 
                            step="0.01"
                            className="form-input"
                            value={item.unitPrice}
                            onChange={e => handleItemRowChange(index, 'unitPrice', e.target.value)}
                            required
                          />
                        </div>

                        <button 
                          type="button" 
                          className="btn btn-danger btn-sm" 
                          style={{ height: '36px', width: '36px', padding: 0 }}
                          onClick={() => handleRemoveItemRow(index)}
                          disabled={soItems.length === 1}
                        >
                          🗑
                        </button>
                      </div>
                    );
                  })}
                </div>
              </div>
              
              <div className="modal-footer">
                <button type="button" className="btn btn-outline" onClick={() => setIsAddOpen(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Reserve & Place Order
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Log RMA Modal */}
      {isRmaOpen && (
        <div className="modal-overlay" onClick={() => setIsRmaOpen(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Log Customer Return (RMA)</h3>
              <button className="modal-close" onClick={() => setIsRmaOpen(false)}>&times;</button>
            </div>
            
            <form onSubmit={handleRMASubmit}>
              <div className="modal-body">
                <div className="form-group" style={{ marginBottom: '1rem' }}>
                  <label className="form-label">Select Shipped Sales Order *</label>
                  <select 
                    className="form-select"
                    value={rmaData.salesOrderId}
                    onChange={e => handleRMAOrderChange(e.target.value)}
                    required
                  >
                    <option value="">-- Choose Order --</option>
                    {shippedOrders.map(so => (
                      <option key={so.id} value={so.id}>{so.soNumber} ({so.customerName})</option>
                    ))}
                  </select>
                </div>

                {selectedSOObj && (
                  <div className="form-grid">
                    <div className="form-group">
                      <label className="form-label">Returned Item SKU *</label>
                      <select 
                        className="form-select"
                        value={rmaData.productId}
                        onChange={e => handleRMAItemChange(e.target.value)}
                        required
                      >
                        {selectedSOObj.items.map(item => (
                          <option key={item.productId} value={item.productId}>{item.sku}</option>
                        ))}
                      </select>
                    </div>

                    <div className="form-group">
                      <label className="form-label">Return Quantity *</label>
                      <input 
                        type="number" 
                        className="form-input"
                        min="1"
                        value={rmaData.quantity}
                        onChange={e => setRmaData({...rmaData, quantity: e.target.value})}
                        required
                      />
                    </div>
                  </div>
                )}

                <div className="form-group" style={{ marginBottom: '1rem' }}>
                  <label className="form-label">Restocking Decision *</label>
                  <select 
                    className="form-select"
                    value={rmaData.status}
                    onChange={e => setRmaData({...rmaData, status: e.target.value})}
                    required
                  >
                    <option value="Restocked">Restocked (Return to inventory shelves)</option>
                    <option value="Damaged">Damaged / Scrap (Discard, do not restock)</option>
                  </select>
                </div>

                <div className="form-group form-group-full">
                  <label className="form-label">Reason for Return</label>
                  <input 
                    type="text" 
                    className="form-input" 
                    placeholder="e.g. Client ordered wrong size, shipping package damaged..."
                    value={rmaData.reason}
                    onChange={e => setRmaData({...rmaData, reason: e.target.value})}
                  />
                </div>
              </div>
              
              <div className="modal-footer">
                <button type="button" className="btn btn-outline" onClick={() => setIsRmaOpen(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Confirm Return
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
