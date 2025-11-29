export const API_URL = 'http://localhost:3000'

// Token refresh function
export const refreshAccessToken = async (): Promise<boolean> => {
  try {
    const response = await fetch(`${API_URL}/api/user/refresh`, {
      method: 'POST',
      credentials: 'include', // Include cookies in the request
    })

    if (response.ok) {
      const data = await response.json()
      if (data.success) {
        return true
      } else {
        // If refresh fails, clear tokens and redirect to login
        localStorage.removeItem('user')
        window.location.href = '/login'
        return false
      }
    } else {
      // If refresh fails, clear tokens and redirect to login
      localStorage.removeItem('user')
      window.location.href = '/login'
      return false
    }
  } catch (error) {
    console.error('Error refreshing token:', error)
    localStorage.removeItem('user')
    window.location.href = '/login'
    return false
  }
}

// Function to make authenticated requests with automatic token refresh
export const authenticatedFetch = async (
  url: string,
  options: RequestInit = {}
): Promise<Response> => {
  const config: RequestInit = {
    ...options,
    credentials: 'include', // Include cookies in all requests
    headers: {
      ...options.headers,
      'Content-Type': 'application/json',
    },
  }

  let response = await fetch(url, config)

  // If the response is 401, try to refresh the token
  if (response.status === 401) {
    const refreshSuccess = await refreshAccessToken()

    if (refreshSuccess) {
      // Retry the original request with the refreshed token (in cookie)
      response = await fetch(url, config)
    }
  }

  return response
}

// Logout function
export const logout = async (): Promise<void> => {
  try {
    await fetch(`${API_URL}/api/user/logout`, {
      method: 'POST',
      credentials: 'include', // Include cookies in the logout request
    })
  } catch (error) {
    console.error('Error during logout:', error)
  } finally {
    // Clear local storage and redirect
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('user')
    window.location.href = '/login'
  }
}

