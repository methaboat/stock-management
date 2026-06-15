import React, { useState, useEffect } from 'react';

export default function Catalog({ userRole, apiHeaders, triggerToast }) {
  const [products, setProducts] = useState([]);
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);
  
  // Modal state
  const [isOpen, setIsOpen] = useState(false);
  const [currentProduct, setCurrentProduct] = useState({
    sku: '',
    name: '',
    barcode: '',
    category: '',
    brand: '',
    costPrice: '',
    sellingPrice: '',
    supplierIds: ''
  });
  const [isEdit, setIsEdit] = useState(false);

  const fetchProducts = async () => {
    try {
      setLoading(true);
      const res = await fetch('http://localhost:8080/api/products', { headers: apiHeaders });
      if (res.ok) {
        const data = await res.json();
        setProducts(data);
      }
    } catch (err) {
      console.error(err);
      triggerToast('Error loading catalog');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProducts();
  }, [userRole]);

  const handleOpenAdd = () => {
    if (userRole !== 'Admin' && userRole !== 'SalesRep') {
      triggerToast('Access denied: Catalog additions require Admin or SalesRep roles.');
      return;
    }
    setIsEdit(false);
    setCurrentProduct({
      sku: '',
      name: '',
      barcode: '',
      category: '',
      brand: '',
      costPrice: '',
      sellingPrice: '',
      supplierIds: ''
    });
    setIsOpen(true);
  };

  const handleOpenEdit = (p) => {
    if (userRole !== 'Admin') {
      triggerToast('Access denied: Product profile editing requires the Admin role.');
      return;
    }
    setIsEdit(true);
    setCurrentProduct({
      id: p.id,
      sku: p.sku,
      name: p.name,
      barcode: p.barcode || '',
      category: p.category || '',
      brand: p.brand || '',
      costPrice: p.costPrice,
      sellingPrice: p.sellingPrice,
      supplierIds: p.supplierIds ? p.supplierIds.join(', ') : ''
    });
    setIsOpen(true);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!currentProduct.sku || !currentProduct.name || !currentProduct.costPrice || !currentProduct.sellingPrice) {
      alert('SKU, Name, Cost, and Selling Price are required.');
      return;
    }

    const supplierArr = currentProduct.supplierIds
      ? currentProduct.supplierIds.split(',').map(s => s.trim()).filter(s => s.length > 0)
      : [];

    const payload = {
      sku: currentProduct.sku,
      name: currentProduct.name,
      barcode: currentProduct.barcode,
      category: currentProduct.category,
      brand: currentProduct.brand,
      costPrice: parseFloat(currentProduct.costPrice),
      sellingPrice: parseFloat(currentProduct.sellingPrice),
      supplierIds: supplierArr
    };

    try {
      const url = isEdit 
        ? `http://localhost:8080/api/products/${currentProduct.id}` 
        : 'http://localhost:8080/api/products';
      
      const method = isEdit ? 'PUT' : 'POST';

      const res = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          ...apiHeaders
        },
        body: JSON.stringify(payload)
      });

      if (res.ok) {
        triggerToast(isEdit ? 'Product updated successfully' : 'Product created successfully');
        setIsOpen(false);
        fetchProducts();
      } else {
        const text = await res.text();
        alert(text || 'Failed to submit product profile');
      }
    } catch (err) {
      console.error(err);
      alert('Server error connecting to API');
    }
  };

  const handleDelete = async (id, name) => {
    if (userRole !== 'Admin') {
      triggerToast('Access denied: Product removal requires the Admin role.');
      return;
    }

    if (!confirm(`Are you sure you want to delete "${name}" from catalog? This will delete associated stock records too.`)) {
      return;
    }

    try {
      const res = await fetch(`http://localhost:8080/api/products/${id}`, {
        method: 'DELETE',
        headers: apiHeaders
      });

      if (res.ok) {
        triggerToast('Product deleted successfully');
        fetchProducts();
      } else {
        const text = await res.text();
        alert(text || 'Failed to delete product');
      }
    } catch (err) {
      console.error(err);
      alert('Server error deleting product');
    }
  };

  const filteredProducts = products.filter(p => 
    p.sku.toLowerCase().includes(search.toLowerCase()) ||
    p.name.toLowerCase().includes(search.toLowerCase()) ||
    (p.category && p.category.toLowerCase().includes(search.toLowerCase())) ||
    (p.brand && p.brand.toLowerCase().includes(search.toLowerCase()))
  );

  const formatPrice = (price) => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(price);
  };

  return (
    <div>
      <div className="actions-panel">
        <div>
          <h2>Product Catalog Profiles</h2>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            Listings of all SKU models and pricing profiles.
          </p>
        </div>
        
        <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
          <input 
            type="text" 
            className="search-input" 
            placeholder="Search catalog SKU, name, category..." 
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <button className="btn btn-primary" onClick={handleOpenAdd}>
            + Add SKU Product
          </button>
        </div>
      </div>

      <div className="table-panel">
        <div className="table-wrapper">
          {loading ? (
            <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>Loading products list...</div>
          ) : filteredProducts.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>
              No products found matching your search.
            </div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>SKU</th>
                  <th>Product Name</th>
                  <th>Barcode</th>
                  <th>Category</th>
                  <th>Brand</th>
                  <th>Cost Price</th>
                  <th>Selling Price</th>
                  <th>Suppliers</th>
                  <th style={{ textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredProducts.map((p) => (
                  <tr key={p.id}>
                    <td style={{ fontFamily: 'monospace', fontWeight: '600', color: 'var(--text-primary)' }}>
                      {p.sku}
                    </td>
                    <td>{p.name}</td>
                    <td>{p.barcode || '-'}</td>
                    <td>{p.category || '-'}</td>
                    <td>{p.brand || '-'}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{formatPrice(p.costPrice)}</td>
                    <td style={{ fontWeight: '600' }}>{formatPrice(p.sellingPrice)}</td>
                    <td>{p.supplierIds && p.supplierIds.length > 0 ? p.supplierIds.join(', ') : '-'}</td>
                    <td style={{ textAlign: 'right' }}>
                      <div style={{ display: 'inline-flex', gap: '0.5rem' }}>
                        <button 
                          className="btn btn-outline btn-sm"
                          onClick={() => handleOpenEdit(p)}
                          disabled={userRole !== 'Admin'}
                        >
                          Edit
                        </button>
                        <button 
                          className="btn btn-danger btn-sm"
                          onClick={() => handleDelete(p.id, p.name)}
                          disabled={userRole !== 'Admin'}
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* Product Form Modal */}
      {isOpen && (
        <div className="modal-overlay" onClick={() => setIsOpen(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>{isEdit ? 'Edit Product Profile' : 'Add Product to Catalog'}</h3>
              <button className="modal-close" onClick={() => setIsOpen(false)}>&times;</button>
            </div>
            
            <form onSubmit={handleSubmit}>
              <div className="modal-body">
                <div className="form-grid">
                  <div className="form-group">
                    <label className="form-label">SKU (Stock Keeping Unit) *</label>
                    <input 
                      type="text" 
                      className="form-input"
                      placeholder="e.g. ELEC-LAP-05"
                      value={currentProduct.sku}
                      onChange={e => setCurrentProduct({...currentProduct, sku: e.target.value})}
                      disabled={isEdit} // Disable SKU edits
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">Barcode / EAN</label>
                    <input 
                      type="text" 
                      className="form-input"
                      placeholder="e.g. 8809123456"
                      value={currentProduct.barcode}
                      onChange={e => setCurrentProduct({...currentProduct, barcode: e.target.value})}
                    />
                  </div>
                </div>

                <div className="form-group form-group-full" style={{ marginBottom: '1rem' }}>
                  <label className="form-label">Product Name *</label>
                  <input 
                    type="text" 
                    className="form-input"
                    placeholder="e.g. Core i7 Laptop 15 inch"
                    value={currentProduct.name}
                    onChange={e => setCurrentProduct({...currentProduct, name: e.target.value})}
                    required
                  />
                </div>

                <div className="form-grid">
                  <div className="form-group">
                    <label className="form-label">Category</label>
                    <input 
                      type="text" 
                      className="form-input"
                      placeholder="e.g. Electronics"
                      value={currentProduct.category}
                      onChange={e => setCurrentProduct({...currentProduct, category: e.target.value})}
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">Brand</label>
                    <input 
                      type="text" 
                      className="form-input"
                      placeholder="e.g. Asus"
                      value={currentProduct.brand}
                      onChange={e => setCurrentProduct({...currentProduct, brand: e.target.value})}
                    />
                  </div>
                </div>

                <div className="form-grid">
                  <div className="form-group">
                    <label className="form-label">Cost Price (USD) *</label>
                    <input 
                      type="number" 
                      step="0.01"
                      className="form-input"
                      placeholder="e.g. 450.00"
                      value={currentProduct.costPrice}
                      onChange={e => setCurrentProduct({...currentProduct, costPrice: e.target.value})}
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">Selling Price (USD) *</label>
                    <input 
                      type="number" 
                      step="0.01"
                      className="form-input"
                      placeholder="e.g. 699.99"
                      value={currentProduct.sellingPrice}
                      onChange={e => setCurrentProduct({...currentProduct, sellingPrice: e.target.value})}
                      required
                    />
                  </div>
                </div>

                <div className="form-group form-group-full">
                  <label className="form-label">Suppliers (Comma separated IDs)</label>
                  <input 
                    type="text" 
                    className="form-input"
                    placeholder="e.g. Intel, AsusDistributor, Newegg"
                    value={currentProduct.supplierIds}
                    onChange={e => setCurrentProduct({...currentProduct, supplierIds: e.target.value})}
                  />
                </div>
              </div>
              
              <div className="modal-footer">
                <button type="button" className="btn btn-outline" onClick={() => setIsOpen(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  {isEdit ? 'Update Profile' : 'Add to Catalog'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
