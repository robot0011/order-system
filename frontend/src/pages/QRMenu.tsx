import { useState, useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { API_URL } from '../config'
import { formatCurrency } from '../utils/currency'
import { handleApiResponse, isResponseSuccess } from '../utils/api'

interface MenuItem {
  ID: number
  Name: string
  Description: string
  Price: number
  Category: string
  ImageURL: string
  Quantity: number
}

interface CartItem {
  menuItemId: number
  name: string
  price: number
  quantity: number
}

interface Order {
  ID: number
  TableID: number
  CustomerName: string
  Status: string
  TotalAmount: number
  OrderItems: OrderItem[]
  CreatedAt: string
  UpdatedAt: string
}

interface OrderItem {
  ID: number
  MenuItemID: number
  Quantity: number
  SpecialInstructions: string
}

interface Restaurant {
  ID: number
  Name: string
  Address: string
  PhoneNumber: string
  LogoURL: string
  CreatedAt: string
  UpdatedAt: string
}

export default function QRMenu() {
  const { restaurantId, tableId } = useParams<{ restaurantId: string; tableId: string }>()
  const [menuItems, setMenuItems] = useState<MenuItem[]>([])
  const [restaurant, setRestaurant] = useState<Restaurant | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [cart, setCart] = useState<CartItem[]>([])
  const [customerName, setCustomerName] = useState('')
  const [isOrderPlaced, setIsOrderPlaced] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')

  useEffect(() => {
    const fetchRestaurantAndMenu = async () => {
      try {
        // Fetch restaurant details
        const restaurantRes = await fetch(`${API_URL}/api/restaurant/${restaurantId}`)
        const restaurantResponse = await handleApiResponse(restaurantRes)
        if (restaurantRes.ok && isResponseSuccess(restaurantResponse)) {
          setRestaurant(restaurantResponse.data)
        } else {
          const errorMessage = restaurantResponse.error || 'Failed to load restaurant details'
          setError(errorMessage)
        }

        // Fetch menu items for the restaurant (public endpoint)
        const menuRes = await fetch(`${API_URL}/api/restaurants/${restaurantId}/menu`)
        const menuResponse = await handleApiResponse(menuRes)
        if (menuRes.ok && isResponseSuccess(menuResponse)) {
          setMenuItems(menuResponse.data || [])
        } else {
          const errorMessage = menuResponse.error || 'Failed to load menu'
          setError(errorMessage)
        }
      } catch (err) {
        setError('Failed to load restaurant and menu')
        console.error('Error fetching restaurant and menu:', err)
      } finally {
        setLoading(false)
      }
    }

    if (restaurantId) {
      fetchRestaurantAndMenu()
    }
  }, [restaurantId])

  const filteredMenuItems = useMemo(() => {
    if (!searchTerm.trim()) return menuItems
    const term = searchTerm.toLowerCase()
    return menuItems.filter(
      (item) =>
        item.Name.toLowerCase().includes(term) ||
        item.Description.toLowerCase().includes(term) ||
        item.Category.toLowerCase().includes(term)
    )
  }, [menuItems, searchTerm])

  const groupedMenu = useMemo(() => {
    return filteredMenuItems.reduce((acc, menuItem) => {
      const category = menuItem.Category || 'Uncategorized'
      if (!acc[category]) acc[category] = []
      acc[category].push(menuItem)
      return acc
    }, {} as Record<string, MenuItem[]>)
  }, [filteredMenuItems])

  const sortedCategories = useMemo(() => {
    return Object.keys(groupedMenu).sort((a, b) => a.localeCompare(b))
  }, [groupedMenu])

  const addToCart = (item: MenuItem) => {
    setCart(prevCart => {
      const existingItem = prevCart.find(cartItem => cartItem.menuItemId === item.ID)
      if (existingItem) {
        // Check if we can add one more without exceeding available quantity
        const currentItem = menuItems.find(mi => mi.ID === item.ID);
        if (currentItem && existingItem.quantity >= currentItem.Quantity) {
          alert(`Cannot add more ${item.Name}. Only ${currentItem.Quantity} available.`);
          return prevCart;
        }
        return prevCart.map(cartItem =>
          cartItem.menuItemId === item.ID
            ? { ...cartItem, quantity: cartItem.quantity + 1 }
            : cartItem
        )
      } else {
        // Check if available quantity allows adding
        if (item.Quantity < 1) {
          alert(`${item.Name} is out of stock.`);
          return prevCart;
        }
        return [
          ...prevCart,
          {
            menuItemId: item.ID,
            name: item.Name,
            price: item.Price,
            quantity: 1
          }
        ]
      }
    })
  }

  const removeFromCart = (menuItemId: number) => {
    setCart(prevCart => prevCart.filter(item => item.menuItemId !== menuItemId))
  }

  const updateQuantity = (menuItemId: number, newQuantity: number) => {
    if (newQuantity <= 0) {
      removeFromCart(menuItemId)
      return
    }

    // Check if the new quantity exceeds available quantity
    const menuItem = menuItems.find(mi => mi.ID === menuItemId);
    if (menuItem && newQuantity > menuItem.Quantity) {
      alert(`Cannot set quantity to ${newQuantity}. Only ${menuItem.Quantity} ${menuItem.Name} available.`);
      return
    }

    setCart(prevCart =>
      prevCart.map(item =>
        item.menuItemId === menuItemId ? { ...item, quantity: newQuantity } : item
      )
    )
  }

  const getCartTotalPrice = () => {
    return cart.reduce((total, item) => total + (item.price * item.quantity), 0)
  }

  const placeOrder = async () => {
    if (cart.length === 0) {
      alert('Please add items to your cart')
      return
    }

    if (!customerName.trim()) {
      alert('Please enter your name')
      return
    }

    try {
      // Create the order payload
      const orderPayload = {
        table_id: parseInt(tableId || "0"),
        customer_name: customerName,
        order_items: cart.map(item => ({
          menu_item_id: item.menuItemId,
          quantity: item.quantity
        }))
      }

      // Send the order to the API (public endpoint)
      const res = await fetch(`${API_URL}/api/restaurants/${restaurantId}/order`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(orderPayload)
      })

      const response = await handleApiResponse(res)
      if (res.ok && isResponseSuccess(response)) {
        setIsOrderPlaced(true)
        setCart([]) // Clear the cart
        alert('Order placed successfully!')
      } else {
        const errorMessage = response.error || 'Failed to place order'
        alert(`Failed to place order: ${errorMessage}`)
      }
    } catch (error) {
      console.error('Error placing order:', error)
      alert('An error occurred while placing your order')
    }
  }

  if (loading) return (
    <div style={{
      minHeight: '100vh',
      background: '#f8f9fa',
      padding: '2rem'
    }}>
      <div style={{
        maxWidth: '1200px',
        margin: '0 auto',
        padding: '0 1rem',
        textAlign: 'center'
      }}>
        <h1 style={{
          color: '#e94560',
          marginBottom: '2rem'
        }}>Restaurant Menu</h1>
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '50vh'
        }}>
          <p style={{ fontSize: '1.2rem', color: '#6c757d' }}>Loading menu...</p>
        </div>
      </div>
    </div>
  )
  if (error) return (
    <div style={{
      minHeight: '100vh',
      background: '#f8f9fa',
      padding: '2rem'
    }}>
      <div style={{
        maxWidth: '1200px',
        margin: '0 auto',
        padding: '0 1rem',
        textAlign: 'center'
      }}>
        <h1 style={{
          color: '#e94560',
          marginBottom: '2rem'
        }}>Restaurant Menu</h1>
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '50vh'
        }}>
          <p className="error" style={{ fontSize: '1.2rem', color: '#dc3545' }}>{error}</p>
        </div>
      </div>
    </div>
  )

  return (
    <div style={{
      minHeight: '100vh',
      background: '#f8f9fa',
      padding: '2rem',
      color: '#333'
    }}>
      <div style={{
        maxWidth: '1200px',
        margin: '0 auto',
        padding: '0 1rem'
      }}>
        {restaurant && (
          <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
            {restaurant.LogoURL && (
              <img
                src={restaurant.LogoURL}
                alt={restaurant.Name}
                style={{
                  maxWidth: '120px',
                  height: 'auto',
                  margin: '0 auto 1rem',
                  borderRadius: '8px'
                }}
              />
            )}
            <h1 style={{
              color: '#e94560',
              marginBottom: '0.5rem',
              fontSize: '2rem'
            }}>{restaurant.Name}</h1>
            <div style={{
              display: 'flex',
              justifyContent: 'center',
              gap: '1.5rem',
              marginTop: '1rem',
              flexWrap: 'wrap'
            }}>
              {restaurant.PhoneNumber && (
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span style={{ color: '#495057', fontWeight: 'bold' }}>📞</span>
                  <span style={{ color: '#495057' }}>{restaurant.PhoneNumber}</span>
                </div>
              )}
              {restaurant.Address && (
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span style={{ color: '#495057', fontWeight: 'bold' }}>📍</span>
                  <span style={{ color: '#495057' }}>{restaurant.Address}</span>
                </div>
              )}
            </div>
          </div>
        )}
        <div style={{
          textAlign: 'center',
          marginBottom: '2rem',
          padding: '1.5rem',
          backgroundColor: '#fff',
          borderRadius: '12px',
          boxShadow: '0 4px 16px rgba(0, 0, 0, 0.1)'
        }}>
          <h2 style={{
            color: '#495057',
            marginBottom: '1rem'
          }}>Welcome to our restaurant!</h2>
          <p style={{
            fontSize: '1.2rem',
            color: '#6c757d',
            marginBottom: '0.5rem'
          }}>Please browse our menu and place your order</p>
          <p style={{
            fontSize: '1rem',
            color: '#888'
          }}>Select items to add to your cart</p>
        </div>

        {isOrderPlaced ? (
          <div style={{
            background: '#fff',
            padding: '2rem',
            borderRadius: '12px',
            boxShadow: '0 4px 16px rgba(0, 0, 0, 0.1)',
            textAlign: 'center',
            maxWidth: '600px',
            margin: '0 auto'
          }}>
            <h2 style={{ color: '#28a745', marginBottom: '1rem' }}>Thank you for your order!</h2>
            <p style={{ marginBottom: '1.5rem' }}>Your order has been placed successfully. Please wait for your food to be prepared.</p>
            <button
              className="btn"
              onClick={() => setIsOrderPlaced(false)}
              style={{ backgroundColor: '#6c757d' }}
            >
              Back to Menu
            </button>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
            {/* Menu Items */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                <h2 style={{ color: '#495057', marginBottom: 0 }}>Menu Items</h2>
                <input
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="Search by name, category, or description"
                  style={{
                    padding: '0.85rem 1rem',
                    borderRadius: '12px',
                    border: '1px solid #d1d5db',
                    fontSize: '1rem',
                    boxShadow: '0 1px 4px rgba(15, 23, 42, 0.08)'
                  }}
                />
              </div>

              {filteredMenuItems.length === 0 ? (
                <div style={{
                  textAlign: 'center',
                  padding: '2rem',
                  color: '#6c757d',
                  fontSize: '1.1rem',
                  background: '#fff',
                  borderRadius: '12px',
                  boxShadow: '0 4px 16px rgba(0, 0, 0, 0.08)'
                }}>
                  No menu items match your search.
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
                  {sortedCategories.map((category) => (
                    <section key={category} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                      <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                        <span style={{ fontSize: '0.85rem', color: '#2563eb', fontWeight: 600 }}>{category}</span>
                        <span style={{ color: '#94a3b8' }}>({groupedMenu[category].length} items)</span>
                      </div>
                      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '1rem' }}>
                        {groupedMenu[category].map((item) => (
                          <div
                            key={item.ID}
                            style={{
                              background: '#fff',
                              padding: '1rem',
                              textAlign: 'left',
                              borderRadius: '12px',
                              boxShadow: '0 4px 16px rgba(0,0,0,0.08)',
                              border: '1px solid #e9ecef',
                              display: 'flex',
                              flexDirection: 'column',
                              gap: '0.5rem'
                            }}
                          >
                            <h3 style={{ color: '#0f172a', margin: 0 }}>{item.Name}</h3>
                            <p style={{ color: '#6c757d', margin: 0 }}>{item.Description}</p>
                            <p style={{ fontWeight: 'bold', color: '#ef4444', margin: 0 }}>{formatCurrency(item.Price)}</p>
                            {item.ImageURL && (
                              <img
                                src={item.ImageURL}
                                alt={item.Name}
                                style={{ maxWidth: '100%', height: 'auto', borderRadius: '8px' }}
                              />
                            )}
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 'auto' }}>
                              <span style={{ color: '#10b981' }}>Available: {item.Quantity}</span>
                              <button
                                className="btn"
                                onClick={() => addToCart(item)}
                                style={{ backgroundColor: '#10b981' }}
                              >
                                Add to Cart
                              </button>
                            </div>
                          </div>
                        ))}
                      </div>
                    </section>
                  ))}
                </div>
              )}
            </div>

            {/* Cart Summary */}
            {cart.length > 0 && (
              <div style={{
                background: '#fff',
                padding: '1.5rem',
                borderRadius: '12px',
                boxShadow: '0 4px 16px rgba(0, 0, 0, 0.1)',
                border: '1px solid #e9ecef'
              }}>
                <h2 style={{ color: '#495057', marginBottom: '1rem' }}>Your Cart</h2>
                <div style={{ marginBottom: '1rem' }}>
                  {cart.map((item) => (
                    <div key={item.menuItemId} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid #e9ecef' }}>
                      <div>
                        <span>{item.name} - {formatCurrency(item.price)} x {item.quantity}</span>
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <button
                          className="btn"
                          onClick={() => updateQuantity(item.menuItemId, item.quantity - 1)}
                          style={{ padding: '0.25rem 0.5rem', fontSize: '0.8rem', backgroundColor: '#6c757d' }}
                        >
                          -
                        </button>
                        <input
                          type="number"
                          min="1"
                          max={menuItems.find(mi => mi.ID === item.menuItemId)?.Quantity || item.quantity}
                          value={item.quantity}
                          onChange={(e) => {
                            const newQuantity = parseInt(e.target.value);
                            if (!isNaN(newQuantity)) {
                              updateQuantity(item.menuItemId, newQuantity);
                            }
                          }}
                          style={{
                            width: '50px',
                            textAlign: 'center',
                            padding: '0.25rem',
                            border: '1px solid #ced4da',
                            borderRadius: '4px'
                          }}
                        />
                        <button
                          className="btn"
                          onClick={() => updateQuantity(item.menuItemId, item.quantity + 1)}
                          style={{ padding: '0.25rem 0.5rem', fontSize: '0.8rem', backgroundColor: '#28a745' }}
                        >
                          +
                        </button>
                        <button
                          className="btn btn-danger"
                          onClick={() => removeFromCart(item.menuItemId)}
                          style={{ padding: '0.25rem 0.5rem', fontSize: '0.8rem', marginLeft: '0.5rem', backgroundColor: '#dc3545' }}
                        >
                          Remove
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
                <div style={{ fontWeight: 'bold', fontSize: '1.2rem', marginBottom: '1rem', color: '#495057' }}>
                  Total: {formatCurrency(getCartTotalPrice())}
                </div>

                <div style={{ marginTop: '1rem' }}>
                  <label htmlFor="customer-name" style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>Your Name: </label>
                  <input
                    id="customer-name"
                    type="text"
                    value={customerName}
                    onChange={(e) => setCustomerName(e.target.value)}
                    placeholder="Enter your name"
                    style={{
                      width: '100%',
                      padding: '0.75rem',
                      border: '1px solid #ced4da',
                      borderRadius: '8px',
                      fontSize: '1rem'
                    }}
                  />
                </div>

                <button
                  className="btn"
                  onClick={placeOrder}
                  style={{ backgroundColor: '#e94560', fontSize: '1.1rem', padding: '0.75rem 1.5rem' }}
                >
                  Place Order
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
