import { Routes, Route, NavLink, useNavigate } from 'react-router-dom'
import Profile from './Profile'
import Restaurant from './Restaurant'
import Tables from './Tables'
import Orders from './Orders'
import '../styles/dashboard.css'

export default function Dashboard() {
  const navigate = useNavigate()

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('user')
    navigate('/login')
  }

  return (
    <div className="dashboard">
      <aside className="sidebar">
        <h2>Order System</h2>
        <nav>
          <NavLink to="/dashboard/profile">Profile</NavLink>
          <NavLink to="/dashboard/restaurant">Restaurant</NavLink>
          <NavLink to="/dashboard/tables">Tables</NavLink>
          <NavLink to="/dashboard/orders">Orders</NavLink>
        </nav>
        <button onClick={handleLogout} className="logout-btn">Logout</button>
      </aside>
      <main className="content">
        <Routes>
          <Route path="profile" element={<Profile />} />
          <Route path="restaurant" element={<Restaurant />} />
          <Route path="tables" element={<Tables />} />
          <Route path="orders" element={<Orders />} />
        </Routes>
      </main>
    </div>
  )
}

