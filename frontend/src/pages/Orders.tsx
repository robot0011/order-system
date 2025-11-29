import { useState, useEffect, useCallback } from 'react';
import { API_URL, authenticatedFetch } from '../config';
import { formatCurrency } from '../utils/currency';

interface OrderItem {
  ID: number;
  MenuItemID: number;
  Quantity: number;
  SpecialInstructions: string;
  MenuItem?: {
    Name: string;
    Price: number;
  };
}

interface Order {
  ID: number;
  TableID: number;
  CustomerName: string;
  Status: string;
  TotalAmount: number;
  OrderItems?: OrderItem[];
  CreatedAt: string;
  UpdatedAt: string;
  restaurant_name?: string;
  restaurant_id: number;
}

interface OrderEvent {
  type: 'order_created' | 'order_updated';
  order: OrderResponsePayload;
}

const isTransactionStatus = (status?: string) =>
  status === 'paid' || status === 'cancelled';

// Synchronous mirror of what the backend returns so we can normalize once
interface OrderItemResponse {
  id: number;
  order_id: number;
  menu_item_id: number;
  quantity: number;
  special_instructions: string;
  menu_item?: {
    name: string;
    price: number;
  };
}

interface OrderResponsePayload {
  id: number;
  table_id: number;
  customer_name: string;
  status: string;
  total_amount: number;
  order_items?: OrderItemResponse[];
  created_at?: string;
  updated_at?: string;
  restaurant_name?: string;
  restaurant_id?: number;
}

const upsertOrder = (orders: Order[], incoming: Order) => {
  const next = [...orders];
  const idx = next.findIndex((order) => order.ID === incoming.ID);
  if (idx >= 0) {
    next[idx] = incoming;
    return next;
  }
  return [incoming, ...orders];
};

const normalizeOrderItem = (item: OrderItemResponse): OrderItem => ({
  ID: item.id,
  OrderID: item.order_id,
  MenuItemID: item.menu_item_id,
  Quantity: item.quantity,
  SpecialInstructions: item.special_instructions,
  MenuItem: item.menu_item
    ? {
        Name: item.menu_item.name,
        Price: item.menu_item.price,
      }
    : undefined,
});

const normalizeOrder = (order: OrderResponsePayload): Order => ({
  ID: order.id,
  TableID: order.table_id,
  CustomerName: order.customer_name,
  Status: order.status,
  TotalAmount: order.total_amount,
  OrderItems: (order.order_items ?? []).map(normalizeOrderItem),
  CreatedAt: order.created_at ?? '',
  UpdatedAt: order.updated_at ?? '',
  restaurant_name: order.restaurant_name,
  restaurant_id: order.restaurant_id ?? 0,
});

export default function Orders() {
  const [activeOrders, setActiveOrders] = useState<Order[]>([]);
  const [paidOrders, setPaidOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [activeTab, setActiveTab] = useState('active'); // 'active' or 'paid'

  const fetchOrders = useCallback(async () => {
    try {
      const res = await authenticatedFetch(`${API_URL}/api/order`);
      if (res.ok) {
        const data: OrderResponsePayload[] = await res.json();
        const normalized = data.map(normalizeOrder);
        // Separate active and transaction orders
        const active = normalized.filter((order) => order.Status && !isTransactionStatus(order.Status));
        const paid = normalized.filter((order) => order.Status && isTransactionStatus(order.Status));
        setActiveOrders(active);
        setPaidOrders(paid);
      } else {
        setError('Failed to load orders');
      }
    } catch {
      setError('Failed to load orders');
    } finally {
      setLoading(false);
    }
  }, []);

  const addOrderToTransactionLog = (orderId: number, newStatus: string) => {
    const sourceOrder =
      activeOrders.find((order) => order.ID === orderId) ||
      paidOrders.find((order) => order.ID === orderId);
    if (!sourceOrder) return;

    setPaidOrders((prev) => {
      const filtered = prev.filter((order) => order.ID !== orderId);
      return [...filtered, { ...sourceOrder, Status: newStatus }];
    });
  };

  const handleIncomingOrder = useCallback((incomingOrder: OrderResponsePayload) => {
    const normalizedOrder = normalizeOrder(incomingOrder);

    if (isTransactionStatus(normalizedOrder.Status)) {
      setPaidOrders((prev) => upsertOrder(prev, normalizedOrder));
      setActiveOrders((prev) => prev.filter((order) => order.ID !== normalizedOrder.ID));
    } else {
      setActiveOrders((prev) => upsertOrder(prev, normalizedOrder));
      setPaidOrders((prev) => prev.filter((order) => order.ID !== normalizedOrder.ID));
    }
  }, []);

  useEffect(() => {
    let socket: WebSocket | null = null;
    let messageHandler: ((event: MessageEvent) => void) | null = null;
    let errorHandler: ((event: Event) => void) | null = null;
    let isActive = true;

    const initializeSocket = async () => {
      await fetchOrders();
      if (!isActive) return;

      const token = localStorage.getItem('token');
      if (!token) return;

      const wsProtocol = API_URL.startsWith('https') ? 'wss' : 'ws';
      const host = API_URL.replace(/^https?:\/\//, '');
      socket = new WebSocket(`${wsProtocol}://${host}/ws/orders?token=${encodeURIComponent(token)}`);

      messageHandler = (event: MessageEvent) => {
        try {
          const orderEvent: OrderEvent = JSON.parse(event.data);
          handleIncomingOrder(orderEvent.order);
        } catch (err) {
          console.error('Failed to parse order event', err);
        }
      };

      errorHandler = (event: Event) => console.warn('Order websocket encountered an error', event);

      socket.addEventListener('message', messageHandler);
      socket.addEventListener('error', errorHandler);
    };

    initializeSocket();

    return () => {
      isActive = false;
      if (socket) {
        if (messageHandler) socket.removeEventListener('message', messageHandler);
        if (errorHandler) socket.removeEventListener('error', errorHandler);
        socket.close();
      }
    };
  }, [fetchOrders, handleIncomingOrder]);

  // Function to update order status
  const updateOrderStatus = async (orderId: number, restaurantId: number, newStatus: string) => {
    try {
      const res = await authenticatedFetch(`${API_URL}/api/restaurant/${restaurantId}/order/${orderId}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ status: newStatus }),
      });

      if (res.ok) {
        // Update the local state based on the new status
        if (newStatus === 'paid' || newStatus === 'cancelled') {
          setActiveOrders((prev) => prev.filter((order) => order.ID !== orderId));
          addOrderToTransactionLog(orderId, newStatus);
        } else {
          // Move order to active orders
          setPaidOrders((prev) => prev.filter((order) => order.ID !== orderId));
          const updatedOrder = paidOrders.find((order) => order.ID === orderId);
          if (updatedOrder) {
            setActiveOrders((prev) => [...prev, { ...updatedOrder, Status: newStatus }]);
          } else {
            setActiveOrders((prev) =>
              prev.map((order) =>
                order.ID === orderId ? { ...order, Status: newStatus } : order
              )
            );
          }
        }
        alert('Order status updated successfully');
      } else {
        const errorData = await res.json();
        alert(`Failed to update order status: ${errorData.message || 'Unknown error'}`);
      }
    } catch (err) {
      console.error('Error updating order status:', err);
      alert('An error occurred while updating the order status');
    }
  };

  const handleMarkAsDelivered = (order: Order) => {
    updateOrderStatus(order.ID, order.restaurant_id, 'delivered');
  };

  const handleMarkAsPaid = (order: Order) => {
    updateOrderStatus(order.ID, order.restaurant_id, 'paid');
  };

  const handleCancelOrder = (order: Order) => {
    updateOrderStatus(order.ID, order.restaurant_id, 'cancelled');
  };

  if (loading) return <div className="page-content"><p>Loading orders...</p></div>;

  return (
    <div className="page-content">
      <h1>Orders</h1>
      {error && <p className="error">{error}</p>}

      {/* Tab navigation */}
      <div style={{ marginBottom: '1.5rem', borderBottom: '1px solid #dee2e6' }}>
        <button
          className={`btn ${activeTab === 'active' ? 'btn-secondary' : ''}`}
          style={{
            backgroundColor: activeTab === 'active' ? '#6c757d' : '#e9ecef',
            color: activeTab === 'active' ? 'white' : '#495057',
            marginRight: '0.5rem',
            borderRadius: '4px 4px 0 0',
            borderBottom: activeTab === 'active' ? 'none' : '1px solid #dee2e6'
          }}
          onClick={() => setActiveTab('active')}
        >
          Active Orders
        </button>
        <button
          className={`btn ${activeTab === 'paid' ? 'btn-secondary' : ''}`}
          style={{
            backgroundColor: activeTab === 'paid' ? '#6c757d' : '#e9ecef',
            color: activeTab === 'paid' ? 'white' : '#495057',
            borderRadius: '4px 4px 0 0',
            borderBottom: activeTab === 'paid' ? 'none' : '1px solid #dee2e6'
          }}
          onClick={() => setActiveTab('paid')}
        >
          Transaction Logs
        </button>
      </div>

      {/* Active Orders Tab */}
      {activeTab === 'active' && (
        <>
          <h2>Active Orders</h2>
          {activeOrders.length === 0 ? (
            <div className="card">
              <p>No active orders found.</p>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              {activeOrders.map((order) => (
                <div key={order.ID} className="card" style={{ padding: '1rem', textAlign: 'left' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                    <h3>Order #{order.ID}</h3>
                    <span style={{
                      padding: '0.25rem 0.5rem',
                      borderRadius: '4px',
                      backgroundColor:
                        !order.Status || order.Status === 'active' ? '#d4edda' :
                        order.Status === 'delivered' ? '#cce5ff' :
                        '#f8d7da',
                      color:
                        !order.Status || order.Status === 'active' ? '#155724' :
                        order.Status === 'delivered' ? '#004085' :
                        '#721c24'
                    }}>
                      {order.Status ? order.Status.charAt(0).toUpperCase() + order.Status.slice(1) : 'Unknown'}
                    </span>
                  </div>

                  {order.restaurant_name && (
                    <p><strong>Restaurant:</strong> {order.restaurant_name}</p>
                  )}

                  <p><strong>Table:</strong> {order.TableID}</p>
                  <p><strong>Customer:</strong> {order.CustomerName}</p>
                  <p><strong>Total:</strong> {formatCurrency(order.TotalAmount)}</p>

                  <div style={{ marginTop: '1rem' }}>
                    <p><strong>Items:</strong></p>
                      <ul style={{ paddingLeft: '1.5rem', margin: '0.5rem 0' }}>
                      {(order.OrderItems ?? []).map((item, index) => (
                        <li key={`order-${order.ID}-item-${item.ID ?? `${item.MenuItemID}-${index}`}`}>
                          {item.MenuItem?.Name || `Item ${item.MenuItemID}`} - Qty: {item.Quantity}
                          {item.SpecialInstructions && ` (${item.SpecialInstructions})`}
                        </li>
                      ))}
                    </ul>
                  </div>

                  <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem' }}>
                    {order.Status && order.Status === 'active' ? (
                      <button
                        className="btn"
                        style={{ backgroundColor: '#28a745', padding: '0.5rem 1rem' }}
                        onClick={() => handleMarkAsDelivered(order)}
                      >
                        Mark as Delivered
                      </button>
                    ) : order.Status && order.Status === 'delivered' ? (
                      <button
                        className="btn"
                        style={{ backgroundColor: '#007bff', padding: '0.5rem 1rem' }}
                        onClick={() => handleMarkAsPaid(order)}
                      >
                        Mark as Paid
                      </button>
                    ) : null}
                    {order.Status && order.Status !== 'paid' && order.Status !== 'cancelled' && (
                      <button
                        className="btn"
                        style={{ backgroundColor: '#dc3545', padding: '0.5rem 1rem' }}
                        onClick={() => handleCancelOrder(order)}
                      >
                        Cancel Order
                      </button>
                    )}
                  </div>

                  <p><strong>Created:</strong> {new Date(order.CreatedAt).toLocaleString()}</p>
                </div>
              ))}
            </div>
          )}
        </>
      )}

      {/* Paid Orders Tab (Transaction Logs) */}
      {activeTab === 'paid' && (
        <>
          <h2>Transaction Logs</h2>
          {paidOrders.length === 0 ? (
            <div className="card">
              <p>No transactions found.</p>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              {paidOrders.map((order) => (
                <div key={order.ID} className="card" style={{ padding: '1rem', textAlign: 'left' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                    <h3>Order #{order.ID}</h3>
                    {(() => {
                      const statusLabel = order.Status ? order.Status.charAt(0).toUpperCase() + order.Status.slice(1) : 'Paid';
                      const isCancelled = order.Status === 'cancelled';
                      return (
                        <span style={{
                          padding: '0.25rem 0.5rem',
                          borderRadius: '4px',
                          backgroundColor: isCancelled ? '#f8d7da' : '#d4edda',
                          color: isCancelled ? '#721c24' : '#155724'
                        }}>
                          {statusLabel}
                        </span>
                      );
                    })()}
                  </div>

                  {order.restaurant_name && (
                    <p><strong>Restaurant:</strong> {order.restaurant_name}</p>
                  )}

                  <p><strong>Table:</strong> {order.TableID}</p>
                  <p><strong>Customer:</strong> {order.CustomerName}</p>
                  <p><strong>Total:</strong> {formatCurrency(order.TotalAmount)}</p>

                  <div style={{ marginTop: '1rem' }}>
                    <p><strong>Items:</strong></p>
                    <ul style={{ paddingLeft: '1.5rem', margin: '0.5rem 0' }}>
                      {(order.OrderItems ?? []).map((item, index) => (
                        <li key={`order-${order.ID}-item-${item.ID ?? `${item.MenuItemID}-${index}`}`}>
                          {item.MenuItem?.Name || `Item ${item.MenuItemID}`} - Qty: {item.Quantity}
                          {item.SpecialInstructions && ` (${item.SpecialInstructions})`}
                        </li>
                      ))}
                    </ul>
                  </div>

                  <p><strong>Created:</strong> {new Date(order.CreatedAt).toLocaleString()}</p>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}
