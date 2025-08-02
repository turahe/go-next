import React, { useState, useEffect } from 'react';
import { DataTable } from '@/components/DataTable/DataTable';
import { formatDate } from '@/utils/format';
import { useNotifications } from '@/context/NotificationContext';
import { useAuth } from '@/context/AuthContext';

interface Post {
  id: number;
  title: string;
  content: string;
  status: string;
  author: {
    id: number;
    username: string;
  };
  category?: {
    id: number;
    name: string;
  };
  created_at: string;
  updated_at: string;
}

interface TableColumn {
  key: string;
  label: string;
  sortable?: boolean;
  render?: (value: any, row: Post) => React.ReactNode;
}

export const PostsPage: React.FC = () => {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPosts, setTotalPosts] = useState(0);
  const [selectedStatus, setSelectedStatus] = useState<string>('');

  const { addNotification } = useNotifications();
  const { hasRole } = useAuth();

  const columns: TableColumn[] = [
    {
      key: 'id',
      label: 'ID',
      sortable: true,
    },
    {
      key: 'title',
      label: 'Title',
      sortable: true,
      render: (value: string) => (
        <div className="max-w-xs truncate" title={value}>
          {value}
        </div>
      ),
    },
    {
      key: 'author',
      label: 'Author',
      render: (value: { username: string }) => value?.username || 'Unknown',
    },
    {
      key: 'category',
      label: 'Category',
      render: (value: { name: string }) => (
        value ? (
          <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200">
            {value.name}
          </span>
        ) : (
          <span className="text-gray-400">No category</span>
        )
      ),
    },
    {
      key: 'status',
      label: 'Status',
      render: (value: string) => {
        const statusColors = {
          published: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
          draft: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
          pending: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
          archived: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200',
        };
        
        return (
          <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
            statusColors[value as keyof typeof statusColors] || 'bg-gray-100 text-gray-800'
          }`}>
            {value.charAt(0).toUpperCase() + value.slice(1)}
          </span>
        );
      },
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
      render: (_: any, row: Post) => (
        <div className="flex space-x-2">
          <button
            onClick={() => handleViewPost(row)}
            className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
          >
            View
          </button>
          {(hasRole('admin') || hasRole('editor')) && (
            <button
              onClick={() => handleEditPost(row)}
              className="text-green-600 hover:text-green-900 dark:text-green-400 dark:hover:text-green-300"
            >
              Edit
            </button>
          )}
          {hasRole('admin') && (
            <button
              onClick={() => handleDeletePost(row.id)}
              className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
            >
              Delete
            </button>
          )}
        </div>
      ),
    },
  ];

  useEffect(() => {
    loadPosts();
  }, [currentPage, searchTerm, selectedStatus]);

  const loadPosts = async () => {
    try {
      setLoading(true);
      // For now, we'll use mock data since the backend posts endpoint might not be fully implemented
      const mockPosts: Post[] = [
        {
          id: 1,
          title: 'Getting Started with React Admin Panel',
          content: 'This is a comprehensive guide...',
          status: 'published',
          author: { id: 1, username: 'admin' },
          category: { id: 1, name: 'Tutorial' },
          created_at: '2024-01-15T10:00:00Z',
          updated_at: '2024-01-15T10:00:00Z',
        },
        {
          id: 2,
          title: 'Advanced TypeScript Patterns',
          content: 'Learn advanced TypeScript...',
          status: 'draft',
          author: { id: 2, username: 'editor' },
          category: { id: 2, name: 'Development' },
          created_at: '2024-01-14T15:30:00Z',
          updated_at: '2024-01-14T15:30:00Z',
        },
        {
          id: 3,
          title: 'Building Scalable APIs with Go',
          content: 'Best practices for building...',
          status: 'pending',
          author: { id: 1, username: 'admin' },
          created_at: '2024-01-13T09:15:00Z',
          updated_at: '2024-01-13T09:15:00Z',
        },
      ];

      // Filter by search
      let filteredPosts = mockPosts;
      if (searchTerm) {
        filteredPosts = mockPosts.filter(post =>
          post.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
          post.author.username.toLowerCase().includes(searchTerm.toLowerCase())
        );
      }

      // Filter by status
      if (selectedStatus) {
        filteredPosts = filteredPosts.filter(post => post.status === selectedStatus);
      }

      // Simulate pagination
      const start = (currentPage - 1) * 10;
      const end = start + 10;
      const paginatedPosts = filteredPosts.slice(start, end);

      setPosts(paginatedPosts);
      setTotalPosts(filteredPosts.length);
    } catch (error) {
      addNotification({
        type: 'error',
        title: 'Error',
        message: 'Failed to load posts',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleViewPost = (post: Post) => {
    addNotification({
      type: 'info',
      title: 'View Post',
      message: `Viewing post: ${post.title}`,
    });
    // TODO: Implement post viewer modal
  };

  const handleEditPost = (post: Post) => {
    addNotification({
      type: 'info',
      title: 'Edit Post',
      message: `Editing post: ${post.title}`,
    });
    // TODO: Implement post editor modal
  };

  const handleDeletePost = async (postId: number) => {
    if (!window.confirm('Are you sure you want to delete this post?')) {
      return;
    }

    try {
      // TODO: Implement actual delete API call
      console.log('Deleting post with ID:', postId);
      addNotification({
        type: 'success',
        title: 'Success',
        message: 'Post deleted successfully',
      });
      loadPosts();
    } catch (error) {
      addNotification({
        type: 'error',
        title: 'Error',
        message: 'Failed to delete post',
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

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Posts Management</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Manage blog posts and content
          </p>
        </div>
        {(hasRole('admin') || hasRole('editor')) && (
          <button className="btn-primary">
            Create New Post
          </button>
        )}
      </div>

      {/* Filters */}
      <div className="card p-4">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <label htmlFor="search" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Search Posts
            </label>
            <input
              id="search"
              type="text"
              placeholder="Search by title or author..."
              value={searchTerm}
              onChange={(e) => handleSearch(e.target.value)}
              className="input-field"
            />
          </div>
          <div className="sm:w-48">
            <label htmlFor="status-filter" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Filter by Status
            </label>
            <select
              id="status-filter"
              value={selectedStatus}
              onChange={(e) => setSelectedStatus(e.target.value)}
              className="input-field"
            >
              <option value="">All Statuses</option>
              <option value="published">Published</option>
              <option value="draft">Draft</option>
              <option value="pending">Pending</option>
              <option value="archived">Archived</option>
            </select>
          </div>
        </div>
      </div>

      {/* Posts Table */}
      <div className="card">
        <DataTable
          data={posts}
          columns={columns}
          isLoading={loading}
          pagination={{
            currentPage,
            totalItems: totalPosts,
            itemsPerPage: 10,
            totalPages: Math.ceil(totalPosts / 10),
          }}
          onPageChange={handlePageChange}
        />
      </div>
    </div>
  );
}; 