import React, { useState } from 'react';
import { Bell, Check, Trash2, Info, AlertTriangle, AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { NotificationItem } from '@/components/notifications/NotificationDrawer';
import { cn } from '@/lib/utils';

const INITIAL_NOTIFICATIONS: NotificationItem[] = [];

export function NotificationsPage() {
  const [notifications, setNotifications] =
    useState<NotificationItem[]>(INITIAL_NOTIFICATIONS);
  const [filter, setFilter] = useState<'all' | 'unread'>('all');

  const unreadCount = notifications.filter((n) => !n.isRead).length;

  const handleMarkAllAsRead = () => {
    setNotifications((prev) => prev.map((n) => ({ ...n, isRead: true })));
  };

  const handleMarkAsRead = (id: string) => {
    setNotifications((prev) =>
      prev.map((n) => (n.id === id ? { ...n, isRead: true } : n))
    );
  };

  const filteredNotifications = notifications.filter((n) =>
    filter === 'unread' ? !n.isRead : true
  );

  const getSeverityBadge = (severity: NotificationItem['severity']) => {
    switch (severity) {
      case 'alert':
        return <Badge variant="rose" size="sm">Alert</Badge>;
      case 'warning':
        return <Badge variant="amber" size="sm">Warning</Badge>;
      case 'info':
      default:
        return <Badge variant="blue" size="sm">Info</Badge>;
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Notification Center
            </h1>
            {unreadCount > 0 && (
              <Badge variant="blue" size="sm">
                {`${unreadCount} Unread`}
              </Badge>
            )}
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Real-time security alerts, usage threshold warnings, and domain SSL notifications.
          </p>
        </div>

        <div className="flex items-center gap-2">
          {unreadCount > 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleMarkAllAsRead}
              leftIcon={<Check className="h-3.5 w-3.5" />}
            >
              Mark all as read
            </Button>
          )}
        </div>
      </div>

      {/* Filter Tabs */}
      <div className="flex gap-2">
        <button
          type="button"
          onClick={() => setFilter('all')}
          className={`rounded-lg px-3 py-1.5 text-xs font-medium transition-colors ${
            filter === 'all'
              ? 'bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900'
              : 'bg-zinc-100 text-zinc-600 hover:bg-zinc-200/70 dark:bg-zinc-900 dark:text-zinc-400'
          }`}
        >
          All Notifications
        </button>
        <button
          type="button"
          onClick={() => setFilter('unread')}
          className={`rounded-lg px-3 py-1.5 text-xs font-medium transition-colors ${
            filter === 'unread'
              ? 'bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900'
              : 'bg-zinc-100 text-zinc-600 hover:bg-zinc-200/70 dark:bg-zinc-900 dark:text-zinc-400'
          }`}
        >
          Unread Only
        </button>
      </div>

      {/* Notifications List */}
      <div className="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
        <div className="divide-y divide-zinc-100 dark:divide-zinc-900">
          {filteredNotifications.length === 0 ? (
            <div className="flex flex-col items-center justify-center p-12 text-center text-zinc-400">
              <Bell className="h-8 w-8 stroke-1 opacity-40" />
              <p className="mt-2 text-xs font-medium">No notifications in this view</p>
            </div>
          ) : (
            filteredNotifications.map((n) => (
              <div
                key={n.id}
                onClick={() => handleMarkAsRead(n.id)}
                className={cn(
                  'flex items-start justify-between gap-4 p-5 transition-colors cursor-pointer',
                  n.isRead
                    ? 'hover:bg-zinc-50 dark:hover:bg-zinc-900/40 opacity-70'
                    : 'bg-zinc-50/70 hover:bg-zinc-100/70 dark:bg-zinc-900/40 dark:hover:bg-zinc-900'
                )}
              >
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    {!n.isRead && (
                      <span className="h-2 w-2 rounded-full bg-blue-600 dark:bg-blue-400 shrink-0" />
                    )}
                    <h3 className="text-xs font-bold text-zinc-900 dark:text-zinc-100">
                      {n.title}
                    </h3>
                    {getSeverityBadge(n.severity)}
                  </div>
                  <p className="text-xs text-zinc-600 dark:text-zinc-400">
                    {n.message}
                  </p>
                </div>

                <div className="font-mono text-[11px] text-zinc-400 shrink-0">
                  {new Date(n.createdAt).toLocaleDateString(undefined, {
                    month: 'short',
                    day: 'numeric',
                  })}
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}

export default NotificationsPage;
