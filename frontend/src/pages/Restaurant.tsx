import { useState, useEffect } from 'react'
import { API_URL, authenticatedFetch } from '../config'
import { formatCurrency } from '../utils/currency'

interface RestaurantData {
  ID: number
  Name: string
  Address: string
  PhoneNumber: string
  LogoURL: string
}

export default function Restaurant() {
  const [restaurants, setRestaurants] = useState<RestaurantData[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    address: '',
    phone_number: '',
    logo_url: '',
  })

  // State for menu form
  const [showMenuForm, setShowMenuForm] = useState(false)
  const [selectedRestaurantId, setSelectedRestaurantId] = useState<number | null>(null)
  const [menuFormData, setMenuFormData] = useState({
    name: '',
    description: '',
    price: '',
    category: '',
    image_url: '',
    quantity: '0',
  })

  // State for showing menus
  const [showingMenuId, setShowingMenuId] = useState<number | null>(null)
  const [menuItems, setMenuItems] = useState<any[]>([])
  const [loadingMenu, setLoadingMenu] = useState(false)

  // State for editing menu items
  const [editingMenuItem, setEditingMenuItem] = useState<any>(null)
  const [editingMenuForm, setEditingMenuForm] = useState({
    name: '',
    description: '',
    price: '',
    category: '',
    image_url: '',
    quantity: '0',
  })


  const fetchRestaurants = async () => {
    try {
      const res = await authenticatedFetch(`${API_URL}/api/restaurant/`)
      if (res.ok) {
        const data = await res.json()
        setRestaurants(data || [])
      } else {
        setError('Failed to load restaurants')
      }
    } catch {
      setError('Failed to load restaurants')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchRestaurants()
  }, [])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const res = await authenticatedFetch(`${API_URL}/api/restaurant/`, {
        method: 'POST',
        body: JSON.stringify(formData),
      })
      if (res.ok) {
        setShowForm(false)
        setFormData({ name: '', address: '', phone_number: '', logo_url: '' })
        fetchRestaurants()
      } else {
        setError('Failed to create restaurant')
      }
    } catch {
      setError('Failed to create restaurant')
    }
  }

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editingId) return
    try {
      const res = await authenticatedFetch(`${API_URL}/api/restaurant/${editingId}`, {
        method: 'PUT',
        body: JSON.stringify(formData),
      })
      if (res.ok) {
        setEditingId(null)
        setFormData({ name: '', address: '', phone_number: '', logo_url: '' })
        fetchRestaurants()
      } else {
        setError('Failed to update restaurant')
      }
    } catch {
      setError('Failed to update restaurant')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this restaurant?')) return
    try {
      const res = await authenticatedFetch(`${API_URL}/api/restaurant/${id}`, {
        method: 'DELETE',
      })
      if (res.ok) {
        fetchRestaurants()
      } else {
        setError('Failed to delete restaurant')
      }
    } catch {
      setError('Failed to delete restaurant')
    }
  }

  const startEdit = (restaurant: RestaurantData) => {
    setFormData({
      name: restaurant.Name,
      address: restaurant.Address,
      phone_number: restaurant.PhoneNumber,
      logo_url: restaurant.LogoURL,
    })
    setEditingId(restaurant.ID)
    setShowForm(false)
  }

  const cancelEdit = () => {
    setEditingId(null)
    setFormData({ name: '', address: '', phone_number: '', logo_url: '' })
  }

  // Handle creating menu item
  const handleCreateMenu = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedRestaurantId) return

    try {
      const res = await authenticatedFetch(`${API_URL}/api/restaurant/${selectedRestaurantId}/menu`, {
        method: 'POST',
        body: JSON.stringify({
          name: menuFormData.name,
          description: menuFormData.description,
          price: parseFloat(menuFormData.price),
          category: menuFormData.category,
          image_url: menuFormData.image_url,
          quantity: parseInt(menuFormData.quantity),
        }),
      })
      if (res.ok) {
        setShowMenuForm(false)
        setMenuFormData({
          name: '',
          description: '',
          price: '',
          category: '',
          image_url: '',
          quantity: '0'
        })
        setSelectedRestaurantId(null)
        // Optionally show a success message
        alert('Menu item created successfully!')
      } else {
        setError('Failed to create menu item')
      }
    } catch {
      setError('Failed to create menu item')
    }
  }

  const cancelMenuCreation = () => {
    setShowMenuForm(false)
    setMenuFormData({
      name: '',
      description: '',
      price: '',
      category: '',
      image_url: '',
      quantity: '0'
    })
    setSelectedRestaurantId(null)
  }

  // Function to fetch and display menu items for a restaurant
  const fetchMenuItems = async (restaurantId: number) => {
    setLoadingMenu(true)
    try {
      const res = await authenticatedFetch(`${API_URL}/api/restaurant/${restaurantId}/menu`)
      if (res.ok) {
        const data = await res.json()
        setMenuItems(data || [])
        setShowingMenuId(restaurantId)
      } else {
        setError('Failed to load menu items')
        setMenuItems([])
        setShowingMenuId(null)
      }
    } catch {
      setError('Failed to load menu items')
      setMenuItems([])
      setShowingMenuId(null)
    } finally {
      setLoadingMenu(false)
    }
  }

  // Function to toggle showing menu items
  const toggleShowMenu = async (restaurantId: number) => {
    if (showingMenuId === restaurantId) {
      // If already showing menu for this restaurant, close it
      setShowingMenuId(null)
      setMenuItems([])
    } else {
      // Fetch and show menu for this restaurant
      await fetchMenuItems(restaurantId)
    }
  }

  // Function to start editing a menu item
  const startEditMenuItem = (item: any) => {
    setEditingMenuItem(item)
    setEditingMenuForm({
      name: item.Name || '',
      description: item.Description || '',
      price: item.Price?.toString() || '',
      category: item.Category || '',
      image_url: item.ImageURL || '',
      quantity: item.Quantity?.toString() || '0'
    })
  }

  // Function to cancel editing a menu item
  const cancelEditMenuItem = () => {
    setEditingMenuItem(null)
    setEditingMenuForm({
      name: '',
      description: '',
      price: '',
      category: '',
      image_url: '',
      quantity: '0'
    })
  }

  // Function to handle updating a menu item
  const handleUpdateMenu = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editingMenuItem || !showingMenuId) return

    try {
      const res = await authenticatedFetch(`${API_URL}/api/restaurant/${showingMenuId}/menu/${editingMenuItem.ID}`, {
        method: 'PUT',
        body: JSON.stringify({
          name: editingMenuForm.name,
          description: editingMenuForm.description,
          price: parseFloat(editingMenuForm.price),
          category: editingMenuForm.category,
          image_url: editingMenuForm.image_url,
          quantity: parseInt(editingMenuForm.quantity)
        }),
      })
      if (res.ok) {
        // Refresh the menu items for the current restaurant
        await fetchMenuItems(showingMenuId)
        cancelEditMenuItem()
        alert('Menu item updated successfully!')
      } else {
        setError('Failed to update menu item')
      }
    } catch {
      setError('Failed to update menu item')
    }
  }

  // Function to handle deleting a menu item
  const handleDeleteMenu = async (restaurantId: number, itemId: number) => {
    if (!confirm('Are you sure you want to delete this menu item?')) return

    try {
      const res = await authenticatedFetch(`${API_URL}/api/restaurant/${restaurantId}/menu/${itemId}`, {
        method: 'DELETE',
      })
      if (res.ok) {
        // Refresh the menu items for the current restaurant
        await fetchMenuItems(restaurantId)
        alert('Menu item deleted successfully!')
      } else {
        setError('Failed to delete menu item')
      }
    } catch {
      setError('Failed to delete menu item')
    }
  }


  if (loading) return <div className="page-content"><p>Loading...</p></div>

  return (
    <div className="page-content">
      <h1>Restaurants</h1>
      {error && <p className="error">{error}</p>}
      
      <button className="btn" onClick={() => { setShowForm(true); setEditingId(null); }}>
        + Add Restaurant
      </button>

      {showForm && (
        <div className="card" style={{ marginTop: '1rem' }}>
          <form onSubmit={handleCreate} className="form">
            <input
              type="text"
              placeholder="Restaurant Name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              required
            />
            <input
              type="text"
              placeholder="Address"
              value={formData.address}
              onChange={(e) => setFormData({ ...formData, address: e.target.value })}
            />
            <input
              type="text"
              placeholder="Phone Number"
              value={formData.phone_number}
              onChange={(e) => setFormData({ ...formData, phone_number: e.target.value })}
            />
            <input
              type="text"
              placeholder="Logo URL"
              value={formData.logo_url}
              onChange={(e) => setFormData({ ...formData, logo_url: e.target.value })}
            />
            <div className="btn-group">
              <button type="submit" className="btn">Save</button>
              <button type="button" className="btn btn-secondary" onClick={() => setShowForm(false)}>Cancel</button>
            </div>
          </form>
        </div>
      )}

      <div className="restaurant-list" style={{ marginTop: '1.5rem' }}>
        {restaurants.length === 0 ? (
          <div className="card">
            <p>No restaurants yet. Create one!</p>
          </div>
        ) : (
          restaurants.map((restaurant) => (
            <div className="card" key={restaurant.ID} style={{ marginBottom: '1rem' }}>
              {editingId === restaurant.ID ? (
                <form onSubmit={handleUpdate} className="form">
                  <input
                    type="text"
                    placeholder="Restaurant Name"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    required
                  />
                  <input
                    type="text"
                    placeholder="Address"
                    value={formData.address}
                    onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                  />
                  <input
                    type="text"
                    placeholder="Phone Number"
                    value={formData.phone_number}
                    onChange={(e) => setFormData({ ...formData, phone_number: e.target.value })}
                  />
                  <input
                    type="text"
                    placeholder="Logo URL"
                    value={formData.logo_url}
                    onChange={(e) => setFormData({ ...formData, logo_url: e.target.value })}
                  />
                  <div className="btn-group">
                    <button type="submit" className="btn">Update</button>
                    <button type="button" className="btn btn-secondary" onClick={cancelEdit}>Cancel</button>
                  </div>
                </form>
              ) : (
                <>
                  <p><strong>Name:</strong> {restaurant.Name}</p>
                  <p><strong>Address:</strong> {restaurant.Address}</p>
                  <p><strong>Phone:</strong> {restaurant.PhoneNumber}</p>
                  {restaurant.LogoURL && <p><strong>Logo:</strong> {restaurant.LogoURL}</p>}
                  <div className="btn-group">
                    <button className="btn" onClick={() => startEdit(restaurant)}>Edit</button>
                    <button className="btn btn-danger" onClick={() => handleDelete(restaurant.ID)}>Delete</button>
                    <button className="btn btn-success"
                      onClick={() => {
                        setSelectedRestaurantId(restaurant.ID);
                        setShowMenuForm(true);
                      }}>Add Menu</button>
                    <button className="btn btn-info"
                      onClick={() => toggleShowMenu(restaurant.ID)}>Show Menu</button>
                  </div>

                  {/* Show menu items for this restaurant if currently selected */}
                  {showingMenuId === restaurant.ID && (
                    <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid #eee' }}>
                      <h3>Menu Items:</h3>
                      {loadingMenu ? (
                        <p>Loading menu items...</p>
                      ) : menuItems.length === 0 ? (
                        <p>No menu items found.</p>
                      ) : (
                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: '1rem', marginTop: '1rem' }}>
                          {menuItems.map((item) => (
                            <div key={item.ID} className="card" style={{ padding: '1rem', textAlign: 'left' }}>
                              {/* Display mode for menu item */}
                              {editingMenuItem?.ID !== item.ID ? (
                                <>
                                  <h4>{item.Name}</h4>
                                  <p>{item.Description}</p>
                                  <p><strong>Price: {formatCurrency(item.Price)}</strong></p>
                                  <p><strong>Quantity: {item.Quantity}</strong></p>
                                  {item.Category && <p><em>Category: {item.Category}</em></p>}
                                  {item.ImageURL && <img src={item.ImageURL} alt={item.Name} style={{ maxWidth: '100%', height: 'auto', marginTop: '0.5rem' }} />}
                                  <div className="btn-group" style={{ marginTop: '0.5rem', justifyContent: 'center' }}>
                                    <button className="btn btn-warning" onClick={() => startEditMenuItem(item)}>Edit</button>
                                    <button className="btn btn-danger" onClick={() => handleDeleteMenu(restaurant.ID, item.ID)}>Delete</button>
                                  </div>
                                </>
                              ) : (
                                /* Edit mode for menu item */
                                <form onSubmit={handleUpdateMenu} className="form">
                                  <input
                                    type="text"
                                    placeholder="Name"
                                    value={editingMenuForm.name}
                                    onChange={(e) => setEditingMenuForm({ ...editingMenuForm, name: e.target.value })}
                                    required
                                    style={{ width: '100%', marginBottom: '0.5rem' }}
                                  />
                                  <textarea
                                    placeholder="Description"
                                    value={editingMenuForm.description}
                                    onChange={(e) => setEditingMenuForm({ ...editingMenuForm, description: e.target.value })}
                                    style={{ width: '100%', height: '60px', marginBottom: '0.5rem' }}
                                  />
                                  <input
                                    type="number"
                                    step="0.01"
                                    placeholder="Price"
                                    value={editingMenuForm.price}
                                    onChange={(e) => setEditingMenuForm({ ...editingMenuForm, price: e.target.value })}
                                    required
                                    style={{ width: '100%', marginBottom: '0.5rem' }}
                                  />
                                  <input
                                    type="text"
                                    placeholder="Category"
                                    value={editingMenuForm.category}
                                    onChange={(e) => setEditingMenuForm({ ...editingMenuForm, category: e.target.value })}
                                    style={{ width: '100%', marginBottom: '0.5rem' }}
                                  />
                                  <input
                                    type="text"
                                    placeholder="Image URL"
                                    value={editingMenuForm.image_url}
                                    onChange={(e) => setEditingMenuForm({ ...editingMenuForm, image_url: e.target.value })}
                                    style={{ width: '100%', marginBottom: '0.5rem' }}
                                  />
                                  <input
                                    type="number"
                                    placeholder="Quantity"
                                    value={editingMenuForm.quantity}
                                    onChange={(e) => setEditingMenuForm({ ...editingMenuForm, quantity: e.target.value })}
                                    style={{ width: '100%', marginBottom: '0.5rem' }}
                                  />
                                  <div className="btn-group" style={{ marginTop: '0.5rem', justifyContent: 'center' }}>
                                    <button type="submit" className="btn btn-success">Update</button>
                                    <button type="button" className="btn btn-secondary" onClick={cancelEditMenuItem}>Cancel</button>
                                  </div>
                                </form>
                              )}
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  )}

                </>
              )}
            </div>
          ))
        )}
      </div>
      {/* Modal for adding menu items */}
      {showMenuForm && selectedRestaurantId && (
        <div className="modal-overlay" style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          backgroundColor: 'rgba(0,0,0,0.5)',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          zIndex: 1000
        }}
        onClick={cancelMenuCreation}
        >
          <div className="card"
            style={{
              width: '90%',
              maxWidth: '500px',
              padding: '1.5rem',
              margin: '1rem'
            }}
            onClick={(e) => e.stopPropagation()}
          >
            <h2>Add Menu Item</h2>
            <form onSubmit={handleCreateMenu} className="form">
              <input
                type="text"
                placeholder="Menu Item Name"
                value={menuFormData.name}
                onChange={(e) => setMenuFormData({ ...menuFormData, name: e.target.value })}
                required
                style={{ marginBottom: '0.5rem' }}
              />
              <textarea
                placeholder="Description"
                value={menuFormData.description}
                onChange={(e) => setMenuFormData({ ...menuFormData, description: e.target.value })}
                style={{ marginBottom: '0.5rem', width: '100%', height: '80px' }}
              />
              <input
                type="number"
                step="0.01"
                placeholder="Price"
                value={menuFormData.price}
                onChange={(e) => setMenuFormData({ ...menuFormData, price: e.target.value })}
                required
                style={{ marginBottom: '0.5rem' }}
              />
              <input
                type="text"
                placeholder="Category"
                value={menuFormData.category}
                onChange={(e) => setMenuFormData({ ...menuFormData, category: e.target.value })}
                style={{ marginBottom: '0.5rem' }}
              />
              <input
                type="text"
                placeholder="Image URL"
                value={menuFormData.image_url}
                onChange={(e) => setMenuFormData({ ...menuFormData, image_url: e.target.value })}
                style={{ marginBottom: '0.5rem' }}
              />
              <input
                type="number"
                placeholder="Quantity"
                value={menuFormData.quantity}
                onChange={(e) => setMenuFormData({ ...menuFormData, quantity: e.target.value })}
                style={{ marginBottom: '1rem' }}
              />
              <div className="btn-group">
                <button type="submit" className="btn">Add Menu Item</button>
                <button type="button" className="btn btn-secondary" onClick={cancelMenuCreation}>Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}

    </div>
  )
}
