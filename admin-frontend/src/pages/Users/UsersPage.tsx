import React, { useState, useEffect } from 'react';
import { DataTable } from '@/components/DataTable/DataTable';
import { UserModal } from '@/components/Users/UserModal';
import { api } from '@/services/api';
import { formatDate, capitalizeFirst } from '@/utils/format';
import { useNotifications } from '@/context/NotificationContext';
import { useAuth } from '@/context/AuthContext';

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

interface TableColumn {
  key: string;
  label: string;
  sortable?: boolean;
  render?: (value: any, row: User) => React.ReactNode;
}

export const UsersPage: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalUsers, setTotalUsers] = useState(0);
  const [selectedRole, setSelectedRole] = useState<string>('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);

  const { addNotification } = useNotifications();
  const { hasRole } = useAuth();

  const columns: TableColumn[] = [
    {
      key: 'id',
      label: 'ID',
      sortable: true,
    },
    {
      key: 'username',
      label: 'Username',
      sortable: true,
    },
    {
      key: 'email',
      label: 'Email',
      sortable: true,
    },
    {
      key: 'roles',
      label: 'Roles',
      render: (value: Role[]) => (
        <div className="flex flex-wrap gap-1">
          {value.map((role) => (
            <span
              key={role.id}
              className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
            >
              {role.name}
            </span>
          ))}
        </div>
      ),
    },
    {
      key: 'email_verified',
      label: 'Email Verified',
      render: (value: string) => (
        <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
          value ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
        }`}>
          {value ? 'Yes' : 'No'}
        </span>
      ),
    },
    {
      key: 'created_at',
      label: 'Created',
      sortable: true,
      render: (value: string) => formatDate(value),
    },
    {
      key: 'actions',
      label: 'Actions',
      render: (_: any, row: User) => (
        <div className="flex space-x-2">
          <button
            onClick={() => handleEditUser(row)}
            className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
            disabled={!hasRole('admin')}
          >
            Edit
          </button>
          <button
            onClick={() => handleDeleteUser(row.id)}
            className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
            disabled={!hasRole('admin')}
          >
            Delete
          </button>
        </div>
      ),
    },
  ];

  useEffect(() => {
    loadUsers();
    loadRoles();
  }, [currentPage, searchTerm, selectedRole]);

  const loadUsers = async () => {
    try {
      setLoading(true);
      const response = await api.getUsers(currentPage, 10, searchTerm);
      setUsers(response.users);
      setTotalUsers(response.total);
    } catch (error) {
      addNotification({
        type: 'error',
        title: 'Error',
        message: 'Failed to load users',
      });
    } finally {
      setLoading(false);
    }
  };

  const loadRoles = async () => {
    try {
      const rolesData = await api.getRoles();
      setRoles(rolesData);
    } catch (error) {
      console.error('Failed to load roles:', error);
    }
  };

  const handleEditUser = (user: User) => {
    setSelectedUser(user);
    setIsModalOpen(true);
  };

  const handleCreateUser = () => {
    setSelectedUser(null);
    setIsModalOpen(true);
  };

  const handleModalSuccess = () => {
    loadUsers();
  };

  const handleDeleteUser = async (userId: number) => {
    if (!window.confirm('Are you sure you want to delete this user?')) {
      return;
    }

    try {
      await api.deleteUser(userId);
      addNotification({
        type: 'success',
        title: 'Success',
        message: 'User deleted successfully',
      });
      loadUsers();
    } catch (error) {
      addNotification({
        type: 'error',
        title: 'Error',
        message: 'Failed to delete user',
      });
    }
  };

  const handleSearch = (value: string) => {
    setSearchTerm(value);
    setCurrentPage(1);
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  const filteredUsers = users.filter(user => {
    if (selectedRole && !user.roles.some(role => role.name === selectedRole)) {
      return false;
    }
    return true;
  });

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">User Management</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Manage user accounts and permissions
          </p>
        </div>
        {hasRole('admin') && (
          <button onClick={handleCreateUser} className="btn-primary">
            Add New User
          </button>
        )}
      </div>

      {/* Filters */}
      <div className="card p-4">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <label htmlFor="search" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Search Users
            </label>
            <input
              id="search"
              type="text"
              placeholder="Search by username or email..."
              value={searchTerm}
              onChange={(e) => handleSearch(e.target.value)}
              className="input-field"
            />
          </div>
          <div className="sm:w-48">
            <label htmlFor="role-filter" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Filter by Role
            </label>
            <select
              id="role-filter"
              value={selectedRole}
              onChange={(e) => setSelectedRole(e.target.value)}
              className="input-field"
            >
              <option value="">All Roles</option>
              {roles.map((role) => (
                <option key={role.id} value={role.name}>
                  {capitalizeFirst(role.name)}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

             {/* Users Table */}
       <div className="card">
         <DataTable
           data={filteredUsers}
           columns={columns}
           isLoading={loading}
           pagination={{
             currentPage,
             totalItems: totalUsers,
             itemsPerPage: 10,
             totalPages: Math.ceil(totalUsers / 10),
           }}
           onPageChange={handlePageChange}
         />
       </div>

       {/* User Modal */}
       <UserModal
         isOpen={isModalOpen}
         onClose={() => setIsModalOpen(false)}
         onSuccess={handleModalSuccess}
         user={selectedUser}
       />
    </div>
  );
}; 