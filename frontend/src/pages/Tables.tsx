import { useState, useEffect } from 'react'
import { API_URL, authenticatedFetch } from '../config'
import { handleApiResponse, isResponseSuccess } from '../utils/api'

interface TableData {
  ID: number
  RestaurantID: number
  TableNumber: number
  QRCodeURL: string
  RestaurantName?: string
  editing?: boolean
}

interface Restaurant {
  ID: number;
  Name: string;
  Address: string;
  PhoneNumber: string;
  LogoURL: string;
  CreatedAt: string;
  UpdatedAt: string;
}

export default function Tables() {
  const [restaurants, setRestaurants] = useState<Restaurant[]>([]);
  const [tables, setTables] = useState<TableData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Fetch all restaurants and all tables for the user
  const fetchRestaurantsAndTables = async () => {
    try {
      // Fetch all restaurants
      const resRestaurants = await authenticatedFetch(`${API_URL}/api/restaurant/`);
      const restaurantResponse = await handleApiResponse(resRestaurants);
      if (resRestaurants.ok && isResponseSuccess(restaurantResponse)) {
        setRestaurants(restaurantResponse.data);
      } else {
        const errorMessage = restaurantResponse.error || 'Failed to load restaurants';
        setError(errorMessage);
        return;
      }

      // Fetch all tables
      const resTables = await authenticatedFetch(`${API_URL}/api/table`);
      const tablesResponse = await handleApiResponse(resTables);
      if (resTables.ok && isResponseSuccess(tablesResponse)) {
        setTables(tablesResponse.data);
      } else {
        const errorMessage = tablesResponse.error || 'Failed to load tables';
        setError(errorMessage);
      }
    } catch {
      setError('Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRestaurantsAndTables();
  }, []);

  // Group tables by restaurant
  const tablesByRestaurant = tables && tables.length > 0 ? tables.reduce((acc, table) => {
    const restaurantId = table.RestaurantID;
    if (!acc[restaurantId]) {
      acc[restaurantId] = {
        restaurantId: restaurantId,
        restaurantName: table.RestaurantName || `Restaurant ${restaurantId}`,
        tables: []
      };
    }
    acc[restaurantId].tables.push(table);
    return acc;
  }, {} as Record<number, { restaurantId: number; restaurantName: string; tables: TableData[] }>) : {};

  // Function to handle creating a new table
  const handleCreateTable = async (e: React.FormEvent, restaurantId: number) => {
    e.preventDefault()

    const tableNumberInput = prompt('Enter table number:');
    if (!tableNumberInput) return;

    const tableNumber = parseInt(tableNumberInput);
    if (isNaN(tableNumber)) {
      alert('Please enter a valid table number');
      return;
    }

    try {
      const res = await authenticatedFetch(`${API_URL}/api/restaurant/${restaurantId}/table`, {
        method: 'POST',
        body: JSON.stringify({
          table_number: tableNumber,
        }),
      })
      const response = await handleApiResponse(res)
      if (res.ok && isResponseSuccess(response)) {
        // Refresh all tables after creating a new one
        fetchRestaurantsAndTables();
      } else {
        const errorMessage = response.error || 'Failed to create table'
        setError(errorMessage)
      }
    } catch {
      setError('Failed to create table')
    }
  }

  // Function to handle updating a table
  const handleUpdateTable = async (e: React.FormEvent, tableId: number, restaurantId: number) => {
    e.preventDefault()

    // Find the table being edited to get its current values from state
    const tableToEdit = tables.find(t => t.ID === tableId);
    if (!tableToEdit) return;

    try {
      const res = await authenticatedFetch(`${API_URL}/api/restaurant/${restaurantId}/table/${tableId}`, {
        method: 'PUT',
        body: JSON.stringify({
          table_number: tableToEdit.TableNumber,
        }),
      })
      const response = await handleApiResponse(res)
      if (res.ok && isResponseSuccess(response)) {
        // Update the table in the local state
        setTables(prevTables =>
          prevTables.map(table =>
            table.ID === tableId
              ? { ...table, editing: false } // Turn off editing mode after successful update
              : table
          )
        )
      } else {
        const errorMessage = response.error || 'Failed to update table'
        setError(errorMessage)
      }
    } catch {
      setError('Failed to update table')
    }
  }

  // Function to handle deleting a table
  const handleDeleteTable = async (tableId: number, restaurantId: number) => {
    if (!confirm('Are you sure you want to delete this table?')) return

    try {
      const res = await authenticatedFetch(`${API_URL}/api/restaurant/${restaurantId}/table/${tableId}`, {
        method: 'DELETE',
      })
      const response = await handleApiResponse(res)
      if (res.ok && isResponseSuccess(response)) {
        // Remove the deleted table from the local state
        setTables(prevTables => prevTables.filter(t => t.ID !== tableId))
      } else {
        const errorMessage = response.error || 'Failed to delete table'
        setError(errorMessage)
      }
    } catch {
      setError('Failed to delete table')
    }
  }

  // Function to start editing a table
  const startEditTable = (tableId: number) => {
    setTables(prevTables =>
      prevTables.map(table =>
        table.ID === tableId
          ? { ...table, editing: true }
          : { ...table, editing: false } // Ensure only one table is being edited at a time
      )
    )
  }

  // Function to cancel editing a table
  const cancelEditTable = (tableId: number) => {
    setTables(prevTables =>
      prevTables.map(table =>
        table.ID === tableId
          ? { ...table, editing: false }
          : table
      )
    )
  }

  // Function to update table number while editing
  const updateTableNumber = (tableId: number, newNumber: number) => {
    setTables(prevTables =>
      prevTables.map(table =>
        table.ID === tableId
          ? { ...table, TableNumber: newNumber }
          : table
      )
    )
  }

  if (loading) return <div className="page-content"><p>Loading...</p></div>

  return (
    <div className="page-content">
      <h1>Tables</h1>
      {error && <p className="error">{error}</p>}

      {restaurants.length === 0 ? (
        <div className="card">
          <p>No restaurants found. Create a restaurant first to add tables.</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
          {restaurants.map((restaurant) => {
            const restaurantTables = tablesByRestaurant[restaurant.ID] ? tablesByRestaurant[restaurant.ID].tables : [];
            return (
              <div key={restaurant.ID} className="card" style={{ padding: '1.5rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                  <h2>{restaurant.Name}</h2>
                </div>

                <button className="btn" onClick={(e) => { e.preventDefault(); handleCreateTable(e, restaurant.ID); }} style={{ marginBottom: '1rem' }}>
                  + Add Table
                </button>

                <div>
                  {restaurantTables.length === 0 ? (
                    <p>No tables for this restaurant yet.</p>
                  ) : (
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '1rem' }}>
                      {restaurantTables.map((table) => (
                        <div key={table.ID} className="card" style={{ padding: '1rem', textAlign: 'left' }}>
                          {table.editing ? (
                            <form onSubmit={(e) => handleUpdateTable(e, table.ID, table.RestaurantID)} className="form">
                              <input
                                type="number"
                                placeholder="Table Number"
                                value={table.TableNumber}
                                onChange={(e) => updateTableNumber(table.ID, parseInt(e.target.value) || 0)}
                                required
                                style={{ width: '100%', marginBottom: '0.5rem' }}
                              />
                              <div style={{ width: '100%', marginBottom: '0.5rem' }}>
                                <label>QR Code:</label>
                                {table.QRCodeURL && (
                                  <div style={{ marginTop: '0.5rem' }}>
                                    <img
                                      src={table.QRCodeURL}
                                      alt={`QR code for Table ${table.TableNumber}`}
                                      style={{ width: '150px', height: '150px' }}
                                    />
                                  </div>
                                )}
                              </div>
                              <div className="btn-group">
                                <button type="submit" className="btn">Update</button>
                                <button type="button" className="btn btn-secondary"
                                  onClick={() => cancelEditTable(table.ID)}>
                                  Cancel
                              </button>
                              </div>
                            </form>
                          ) : (
                            <>
                              <h4>Table #{table.TableNumber}</h4>
                              {table.QRCodeURL && (
                                <div>
                                  <p><strong>QR Code:</strong></p>
                                  <img
                                    src={table.QRCodeURL}
                                    alt={`QR code for Table ${table.TableNumber}`}
                                    style={{ width: '150px', height: '150px', marginTop: '0.5rem' }}
                                  />
                                </div>
                              )}
                              <div className="btn-group" style={{ marginTop: '0.5rem', justifyContent: 'center' }}>
                                <button className="btn btn-warning" onClick={() => startEditTable(table.ID)}>Edit</button>
                                <button className="btn btn-danger" onClick={() => handleDeleteTable(table.ID, table.RestaurantID)}>Delete</button>
                              </div>
                            </>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  )
}
