import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { API_URL, authenticatedFetch } from '../config'
import '../styles/auth.css'

export default function Login() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    try {
      const res = await fetch(`${API_URL}/api/user/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })

      const response = await res.json()

      if (response.success && response.data?.access_token) {
        const token = response.data.access_token
        const refreshToken = response.data.refresh_token

        localStorage.setItem('token', token)
        if (refreshToken) {
          localStorage.setItem('refreshToken', refreshToken)
        }
        localStorage.setItem('user', JSON.stringify(response.data))
        navigate('/dashboard/profile')
      } else {
        const errorMessage = response.error || 'Invalid credentials'
        setError(errorMessage)
      }
    } catch {
      setError('Login failed')
    }
  }

  return (
    <div className="auth-container">
      <form className="auth-form" onSubmit={handleSubmit}>
        <h1>Login</h1>
        {error && <p className="error">{error}</p>}
        <input
          type="text"
          placeholder="Username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          required
        />
        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />
        <button type="submit">Login</button>
        <p>
          Don't have an account? <Link to="/register">Register</Link>
        </p>
      </form>
    </div>
  )
}

