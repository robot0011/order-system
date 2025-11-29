export const API_URL = 'http://localhost:3000'

// Token refresh function
export const refreshAccessToken = async (): Promise<string | null> => {
  try {
    const refreshToken = localStorage.getItem('refreshToken')
    if (!refreshToken) {
      return null
    }

    const response = await fetch(`${API_URL}/api/user/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    if (response.ok) {
      const data = await response.json()
      localStorage.setItem('token', data.access_token) // Update access token
      return data.access_token
    } else {
      // If refresh fails, clear tokens and redirect to login
      localStorage.removeItem('token')
      localStorage.removeItem('refreshToken')
      window.location.href = '/login'
      return null
    }
  } catch (error) {
    console.error('Error refreshing token:', error)
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    window.location.href = '/login'
    return null
  }
}

// Function to make authenticated requests with automatic token refresh
export const authenticatedFetch = async (
  url: string,
  options: RequestInit = {}
): Promise<Response> => {
  const token = localStorage.getItem('token')

  const config = {
    ...options,
    headers: {
      ...options.headers,
      ...(token && { Authorization: `Bearer ${token}` }),
      'Content-Type': 'application/json',
    },
  }

  let response = await fetch(url, config)

  // If the response is 401, try to refresh the token
  if (response.status === 401) {
    const newToken = await refreshAccessToken()

    if (newToken) {
      // Retry the original request with the new token
      config.headers.Authorization = `Bearer ${newToken}`
      response = await fetch(url, config)
    }
  }

  return response
}

