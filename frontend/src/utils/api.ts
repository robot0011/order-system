// Utility functions to handle the new API response format consistently

interface APIResponse<T = any> {
  success: boolean;
  data: T | null;
  error: string | null;
}

// Helper function to handle API responses according to the new format
export const handleApiResponse = async <T = any>(
  response: Response
): Promise<APIResponse<T>> => {
  try {
    const data: APIResponse<T> = await response.json();
    return data;
  } catch (error) {
    // If JSON parsing fails, return a generic error response
    return {
      success: false,
      data: null,
      error: response.statusText || 'Response parsing error',
    };
  }
};

// Helper function to check if a response is successful according to the new format
export const isResponseSuccess = <T = any>(
  response: APIResponse<T>
): response is APIResponse<NonNullable<T>> => {
  return response.success && response.data !== null;
};

// Helper function to extract data from a successful response or return null
export const extractData = <T = any>(response: APIResponse<T>): T | null => {
  if (isResponseSuccess(response)) {
    return response.data;
  }
  return null;
};

// Helper function to get error message from response
export const getErrorMessage = (response: APIResponse): string => {
  if (response.error) {
    return typeof response.error === 'string' ? response.error : 'Unknown error occurred';
  }
  return 'Unknown error occurred';
};