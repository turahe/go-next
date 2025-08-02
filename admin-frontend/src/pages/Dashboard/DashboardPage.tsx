import React, { useEffect, useState } from 'react';
import { api } from '@/services/api';
import { DashboardStats } from '@/types';
import { formatCurrency, formatNumber, formatPercentage } from '@/utils/format';
import { useNotifications } from '@/context/NotificationContext';

interface StatCardProps {
  title: string;
  value: string | number;
  change?: string;
  changeType?: 'positive' | 'negative' | 'neutral';
  icon: string;
}

const StatCard: React.FC<StatCardProps> = ({ title, value, change, changeType, icon }) => {
  const getChangeColor = () => {
    switch (changeType) {
      case 'positive':
        return 'text-green-600 dark:text-green-400';
      case 'negative':
        return 'text-red-600 dark:text-red-400';
      default:
        return 'text-gray-600 dark:text-gray-400';
    }
  };

  return (
    <div className="card p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400">{title}</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">{value}</p>
          {change && (
            <p className={`text-sm font-medium mt-1 ${getChangeColor()}`}>
              {change}
            </p>
          )}
        </div>
        <div className="text-3xl">{icon}</div>
      </div>
    </div>
  );
};

export const DashboardPage: React.FC = () => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const { addNotification } = useNotifications();

  useEffect(() => {
    const fetchStats = async () => {
      try {
        setIsLoading(true);
              const response = await api.getDashboardStats();
      setStats(response);
      } catch (error) {
        addNotification({
          type: 'error',
          title: 'Failed to Load Dashboard',
          message: 'Unable to load dashboard statistics. Please try again.',
        });
      } finally {
        setIsLoading(false);
      }
    };

    fetchStats();
  }, [addNotification]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 dark:text-gray-400">Failed to load dashboard data</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
        <p className="text-gray-600 dark:text-gray-400 mt-1">
          Welcome to your admin dashboard. Here's an overview of your system.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Users"
          value={formatNumber(stats.totalUsers)}
          icon="👥"
        />
        <StatCard
          title="Active Users"
          value={formatNumber(stats.activeUsers)}
          icon="✅"
        />
        <StatCard
          title="Total Posts"
          value={formatNumber(stats.totalPosts)}
          icon="📝"
        />
        <StatCard
          title="Total Comments"
          value={formatNumber(stats.totalComments)}
          icon="💬"
        />
        <StatCard
          title="Revenue"
          value={formatCurrency(stats.revenue)}
          change={`${formatPercentage(stats.growthRate)} from last month`}
          changeType={stats.growthRate >= 0 ? 'positive' : 'negative'}
          icon="💰"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="card p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Recent Activity
          </h3>
          <div className="space-y-4">
            <div className="flex items-center space-x-3">
              <div className="w-2 h-2 bg-green-500 rounded-full"></div>
              <div className="flex-1">
                <p className="text-sm font-medium text-gray-900 dark:text-white">
                  New user registration
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  2 minutes ago
                </p>
              </div>
            </div>
            <div className="flex items-center space-x-3">
              <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
              <div className="flex-1">
                <p className="text-sm font-medium text-gray-900 dark:text-white">
                  System update completed
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  1 hour ago
                </p>
              </div>
            </div>
            <div className="flex items-center space-x-3">
              <div className="w-2 h-2 bg-yellow-500 rounded-full"></div>
              <div className="flex-1">
                <p className="text-sm font-medium text-gray-900 dark:text-white">
                  Database backup
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  3 hours ago
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="card p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Quick Actions
          </h3>
          <div className="space-y-3">
            <button className="w-full btn-primary text-left">
              👥 Add New User
            </button>
            <button className="w-full btn-secondary text-left">
              📊 View Reports
            </button>
            <button className="w-full btn-secondary text-left">
              ⚙️ System Settings
            </button>
            <button className="w-full btn-secondary text-left">
              📧 Send Notification
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}; 