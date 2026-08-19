import React from 'react';
import { ShieldCheck, ShieldAlert, Shield } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';

export interface SSLStatusBadgeProps {
  status: 'active' | 'pending' | 'error';
  className?: string;
}

export function SSLStatusBadge({ status, className }: SSLStatusBadgeProps) {
  if (status === 'active') {
    return (
      <Badge variant="emerald" size="sm" dot className={className}>
        SSL Active
      </Badge>
    );
  }

  if (status === 'pending') {
    return (
      <Badge variant="amber" size="sm" dot className={className}>
        SSL Pending
      </Badge>
    );
  }

  return (
    <Badge variant="rose" size="sm" dot className={className}>
      SSL Error
    </Badge>
  );
}

export default SSLStatusBadge;
