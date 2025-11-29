import { useState, useEffect } from 'react'
import { API_URL, authenticatedFetch } from '../config'
import { handleApiResponse, isResponseSuccess } from '../utils/api'

interface ProfileData {
  username: string
  email: string
  role: string
}

export default function Profile() {
  const [profile, setProfile] = useState<ProfileData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const res = await authenticatedFetch(`${API_URL}/api/user/profile`)
        const response = await handleApiResponse(res)

        if (res.ok && isResponseSuccess(response)) {
          setProfile(response.data)
        } else {
          const errorMessage = response.error || 'Failed to load profile'
          setError(errorMessage)
        }
      } catch {
        setError('Failed to load profile')
      } finally {
        setLoading(false)
      }
    }
    fetchProfile()
  }, [])

  if (loading) return <div className="page-content"><p>Loading...</p></div>
  if (error) return <div className="page-content"><p className="error">{error}</p></div>

  return (
    <div className="page-content">
      <h1>Profile</h1>
      {profile && (
        <div className="card">
          <p><strong>Username:</strong> {profile.username}</p>
          <p><strong>Email:</strong> {profile.email}</p>
          <p><strong>Role:</strong> {profile.role}</p>
        </div>
      )}
    </div>
  )
}

