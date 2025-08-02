// API service for backend integration
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api';



interface LoginResponse {
  token: string;
  refresh_token: string;
}

interface User {
  id: number;
  username: string;
  email: string;
  phone?: string;
  email_verified?: string;
  phone_verified?: string;
  roles: Role[];
  created_at: string;
  updated_at: string;
}

interface Role {
  id: number;
  name: string;
  created_at: string;
  updated_at: string;
}

interface DashboardStats {
  totalUsers: number;
  activeUsers: number;
  totalPosts: number;
  totalComments: number;
  revenue: number;
  growthRate: number;
}

class ApiService {
  private token: string | null = null;

  setToken(token: string) {
    this.token = token;
    localStorage.setItem('authToken', token);
  }

  getToken(): string | null {
    if (!this.token) {
      this.token = localStorage.getItem('authToken');
    }
    return this.token;
  }

  clearToken() {
    this.token = null;
    localStorage.removeItem('authToken');
  }

  async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string> || {}),
    };

    const token = this.getToken();
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  // Authentication
  async login(credentials: { username: string; password: string }): Promise<LoginResponse> {
    const response = await this.request<LoginResponse>('/login', {
      method: 'POST',
      body: JSON.stringify({
        username: credentials.username,
        email: credentials.username, // Backend expects both username and email
        password: credentials.password,
      }),
    });
    return response;
  }

  async register(userData: { username: string; email: string; password: string; role?: string }): Promise<LoginResponse> {
    const response = await this.request<LoginResponse>('/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
    return response;
  }

  async refreshToken(): Promise<LoginResponse> {
    const response = await this.request<LoginResponse>('/v1/auth/refresh', {
      method: 'POST',
    });
    return response;
  }

  // Users
  async getUsers(page: number = 1, limit: number = 10, search?: string): Promise<{ users: User[]; total: number }> {
    let endpoint = `/v1/users?page=${page}&limit=${limit}`;
    if (search) {
      endpoint += `&search=${encodeURIComponent(search)}`;
    }
    
    const response = await this.request<{ users: User[]; total: number; page: number; limit: number; pages: number }>(endpoint);
    return {
      users: response.users,
      total: response.total,
    };
  }

  async getUser(id: number): Promise<User> {
    const response = await this.request<User>(`/v1/users/${id}`);
    return response;
  }

  async updateUser(id: number, userData: Partial<User>): Promise<User> {
    const response = await this.request<User>(`/v1/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(userData),
    });
    return response;
  }

  async deleteUser(id: number): Promise<void> {
    await this.request(`/v1/users/${id}`, {
      method: 'DELETE',
    });
  }

  async updateUserRole(userId: number, roleId: number): Promise<User> {
    const response = await this.request<User>(`/v1/users/${userId}/role`, {
      method: 'PUT',
      body: JSON.stringify({ role_id: roleId }),
    });
    return response;
  }

  // Roles
  async getRoles(): Promise<Role[]> {
    const response = await this.request<Role[]>('/v1/roles');
    return response;
  }

  async createRole(roleData: { name: string }): Promise<Role> {
    const response = await this.request<Role>('/v1/roles', {
      method: 'POST',
      body: JSON.stringify(roleData),
    });
    return response;
  }

  async updateRole(id: number, roleData: { name: string }): Promise<Role> {
    const response = await this.request<Role>(`/v1/roles/${id}`, {
      method: 'PUT',
      body: JSON.stringify(roleData),
    });
    return response;
  }

  async deleteRole(id: number): Promise<void> {
    await this.request(`/v1/roles/${id}`, {
      method: 'DELETE',
    });
  }

  // Dashboard Stats
  async getDashboardStats(): Promise<DashboardStats> {
    const response = await this.request<DashboardStats>('/v1/dashboard/stats');
    return response;
  }

  // Health check
  async healthCheck(): Promise<{ status: string }> {
    const response = await fetch(`${API_BASE_URL.replace('/api', '')}/health`);
    if (!response.ok) {
      throw new Error('Backend is not available');
    }
    return { status: 'OK' };
  }
}

export const api = new ApiService(); 