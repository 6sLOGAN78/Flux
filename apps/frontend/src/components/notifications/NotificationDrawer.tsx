import React from 'react';
import { Bell, Check, Info, AlertTriangle, AlertCircle, X } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface NotificationItem {
  id: string;
  title: string;
  message: string;
  severity: 'info' | 'warning' | 'alert';
  isRead: boolean;
  createdAt: string;
  linkUrl?: string;
}

export interface NotificationDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  notifications: NotificationItem[];
  onMarkAllAsRead: () => void;
  onMarkAsRead: (id: string) => void;
  className?: string;
}

export function NotificationDrawer({
  isOpen,
  onClose,
  notifications,
  onMarkAllAsRead,
  onMarkAsRead,
  className,
}: NotificationDrawerProps) {
  if (!isOpen) return null;

  const unreadCount = notifications.filter((n) => !n.isRead).length;

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
    <div className="fixed inset-0 z-50 overflow-hidden">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/40 backdrop-blur-xs transition-opacity animate-in fade-in"
        onClick={onClose}
      />

      <div className="fixed inset-y-0 right-0 max-w-full flex pl-10">
        <div className="w-screen max-w-md border-l border-zinc-200 bg-white shadow-2xl dark:border-zinc-800 dark:bg-zinc-950 flex flex-col animate-in slide-in-from-right duration-200">
          {/* Drawer Header */}
          <div className="flex items-center justify-between border-b border-zinc-100 p-5 dark:border-zinc-900">
            <div className="flex items-center gap-2">
              <Bell className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
              <h2 className="text-sm font-bold text-zinc-900 dark:text-zinc-100">
                Notifications
              </h2>
              {unreadCount > 0 && (
                <Badge variant="zinc" size="sm">
                  {`${unreadCount} new`}
                </Badge>
              )}
            </div>

            <div className="flex items-center gap-2">
              {unreadCount > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={onMarkAllAsRead}
                  className="text-xs"
                >
                  Mark all as read
                </Button>
              )}
              <button
                type="button"
                onClick={onClose}
                className="rounded-lg p-1.5 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-900 dark:hover:bg-zinc-900 dark:hover:text-zinc-100"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          </div>

          {/* Notifications Feed */}
          <div className="flex-1 overflow-y-auto divide-y divide-zinc-100 dark:divide-zinc-900 p-2">
            {notifications.length === 0 ? (
              <div className="flex h-full flex-col items-center justify-center p-8 text-center text-zinc-400">
                <Bell className="h-8 w-8 stroke-1 opacity-40" />
                <p className="mt-2 text-xs">No notifications yet</p>
              </div>
            ) : (
              notifications.map((n) => (
                <div
                  key={n.id}
                  onClick={() => onMarkAsRead(n.id)}
                  className={cn(
                    'p-4 rounded-xl transition-colors cursor-pointer',
                    n.isRead
                      ? 'hover:bg-zinc-50 dark:hover:bg-zinc-900/40 opacity-70'
                      : 'bg-zinc-50/80 hover:bg-zinc-100/70 dark:bg-zinc-900/50 dark:hover:bg-zinc-900'
                  )}
                >
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      {!n.isRead && (
                        <span className="h-2 w-2 rounded-full bg-blue-600 dark:bg-blue-400 shrink-0" />
                      )}
                      <span className="text-xs font-bold text-zinc-900 dark:text-zinc-100">
                        {n.title}
                      </span>
                    </div>
                    {getSeverityBadge(n.severity)}
                  </div>

                  <p className="mt-1 text-xs text-zinc-600 dark:text-zinc-400 leading-relaxed">
                    {n.message}
                  </p>

                  <div className="mt-2 font-mono text-[10px] text-zinc-400">
                    {new Date(n.createdAt).toLocaleTimeString([], {
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default NotificationDrawer;
