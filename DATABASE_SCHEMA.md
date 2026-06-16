# Database Schema — stock_management (MongoDB)

## Collections Overview

| Collection         | Description                              |
|--------------------|------------------------------------------|
| `users`            | System users with roles                  |
| `products`         | Product catalog                          |
| `stock_levels`     | Current inventory per product/location   |
| `purchase_orders`  | Inbound purchase orders                  |
| `sales_orders`     | Outbound sales orders                    |
| `returns_rma`      | Return/RMA records                       |
| `audit_logs`       | Immutable inventory action history       |
| `reconciliations`  | Stock count reconciliation sessions      |

---

## `users`

| Field          | Type     | Notes                                      |
|----------------|----------|--------------------------------------------|
| `_id`          | ObjectId | Primary key                                |
| `username`     | string   | Unique index                               |
| `passwordHash` | string   | bcrypt hash, never returned in JSON        |
| `role`         | string   | `"Admin"` \| `"WarehouseStaff"` \| `"SalesRep"` |

---

## `products`

| Field          | Type       | Notes                     |
|----------------|------------|---------------------------|
| `_id`          | ObjectId   | Primary key               |
| `sku`          | string     | Unique index              |
| `name`         | string     |                           |
| `barcode`      | string     |                           |
| `category`     | string     |                           |
| `brand`        | string     |                           |
| `costPrice`    | float64    |                           |
| `sellingPrice` | float64    |                           |
| `supplierIds`  | []string   | References to suppliers   |
| `createdAt`    | timestamp  |                           |
| `updatedAt`    | timestamp  |                           |

---

## `stock_levels`

| Field            | Type     | Notes                                      |
|------------------|----------|--------------------------------------------|
| `_id`            | ObjectId | Primary key                                |
| `productId`      | ObjectId | References `products._id`                  |
| `sku`            | string   | Unique index                               |
| `stockOnHand`    | float64  | Total physical quantity                    |
| `committedStock` | float64  | Reserved for open sales orders             |
| `availableStock` | float64  | `stockOnHand - committedStock`             |
| `reorderPoint`   | float64  | Triggers low-stock alert when `availableStock <= reorderPoint` |
| `location`       | object   | See **Location** sub-document below        |

### Location (embedded)

| Field       | Type   |
|-------------|--------|
| `warehouse` | string |
| `aisle`     | string |
| `bin`       | string |

---

## `purchase_orders`

| Field          | Type       | Notes                                          |
|----------------|------------|------------------------------------------------|
| `_id`          | ObjectId   | Primary key                                    |
| `poNumber`     | string     | Unique index                                   |
| `supplierName` | string     |                                                |
| `status`       | string     | `"Draft"` \| `"Sent"` \| `"PartialReceived"` \| `"Received"` \| `"Cancelled"` |
| `items`        | []POItem   | See **POItem** below                           |
| `totalAmount`  | float64    |                                                |
| `createdAt`    | timestamp  |                                                |
| `updatedAt`    | timestamp  |                                                |

### POItem (embedded array)

| Field              | Type     |
|--------------------|----------|
| `productId`        | ObjectId |
| `sku`              | string   |
| `quantity`         | float64  |
| `unitCost`         | float64  |
| `receivedQuantity` | float64  |

---

## `sales_orders`

| Field          | Type       | Notes                                          |
|----------------|------------|------------------------------------------------|
| `_id`          | ObjectId   | Primary key                                    |
| `soNumber`     | string     | Unique index                                   |
| `customerName` | string     |                                                |
| `status`       | string     | `"Draft"` \| `"Confirmed"` \| `"PartialShipped"` \| `"Shipped"` \| `"Cancelled"` |
| `items`        | []SOItem   | See **SOItem** below                           |
| `totalAmount`  | float64    |                                                |
| `createdAt`    | timestamp  |                                                |
| `updatedAt`    | timestamp  |                                                |

### SOItem (embedded array)

| Field             | Type     |
|-------------------|----------|
| `productId`       | ObjectId |
| `sku`             | string   |
| `quantity`        | float64  |
| `unitPrice`       | float64  |
| `shippedQuantity` | float64  |

---

## `returns_rma`

| Field          | Type      | Notes                                      |
|----------------|-----------|--------------------------------------------|
| `_id`          | ObjectId  | Primary key                                |
| `salesOrderId` | ObjectId  | References `sales_orders._id`              |
| `productId`    | ObjectId  | References `products._id`                  |
| `sku`          | string    |                                            |
| `quantity`     | float64   |                                            |
| `reason`       | string    |                                            |
| `status`       | string    | `"Pending"` \| `"Approved"` \| `"Rejected"` |
| `createdAt`    | timestamp |                                            |

---

## `audit_logs`

| Field            | Type      | Notes                                               |
|------------------|-----------|-----------------------------------------------------|
| `_id`            | ObjectId  | Primary key                                         |
| `userId`         | string    | Username of actor                                   |
| `username`       | string    |                                                     |
| `action`         | string    | `"Create"` \| `"Receive"` \| `"Reserve"` \| `"Ship"` \| `"Adjust"` \| `"Return"` \| `"Scan-in"` \| `"Scan-out"` |
| `entityType`     | string    | `"Product"` \| `"Stock"` \| `"PO"` \| `"SO"` \| `"RMA"` |
| `entityId`       | string    | ObjectId string of the affected document            |
| `sku`            | string    |                                                     |
| `quantityChange` | float64   | Positive = added, negative = removed                |
| `note`           | string    |                                                     |
| `timestamp`      | timestamp |                                                     |

---

## `reconciliations`

| Field         | Type                  | Notes                              |
|---------------|-----------------------|------------------------------------|
| `_id`         | ObjectId              | Primary key                        |
| `date`        | timestamp             |                                    |
| `status`      | string                | `"Draft"` \| `"Approved"`         |
| `items`       | []ReconciliationItem  | See **ReconciliationItem** below   |
| `submittedBy` | string                | Username                           |
| `approvedBy`  | string                | Username, optional                 |
| `approvedAt`  | timestamp             | Optional                           |
| `note`        | string                |                                    |
| `createdAt`   | timestamp             |                                    |

### ReconciliationItem (embedded array)

| Field         | Type     |
|---------------|----------|
| `productId`   | ObjectId |
| `sku`         | string   |
| `productName` | string   |
| `expectedQty` | float64  |
| `actualQty`   | float64  |
| `difference`  | float64  |

---

## Indexes Summary

| Collection        | Field      | Type   |
|-------------------|------------|--------|
| `products`        | `sku`      | Unique |
| `stock_levels`    | `sku`      | Unique |
| `purchase_orders` | `poNumber` | Unique |
| `sales_orders`    | `soNumber` | Unique |
| `users`           | `username` | Unique |

---

## Relationships Diagram

```
users ──────────────────────────────── audit_logs
                                           │ (entityId → any collection)

products ──┬──── stock_levels (productId)
           ├──── purchase_orders.items (productId)
           ├──── sales_orders.items (productId)
           └──── returns_rma (productId)

sales_orders ──── returns_rma (salesOrderId)

stock_levels ──── reconciliations.items (via SKU)
```
