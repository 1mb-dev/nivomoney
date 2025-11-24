import { useEffect, useRef } from 'react';

interface SSEEvent {
  topic: string;
  event_type: string;
  data: Record<string, unknown>;
  timestamp: string;
}

interface UseSSEOptions {
  topics: string[];
  onEvent: (event: SSEEvent) => void;
  onError?: (error: Event) => void;
  enabled?: boolean;
}

export function useSSE({ topics, onEvent, onError, enabled = true }: UseSSEOptions) {
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!enabled || topics.length === 0) {
      return;
    }

    const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8000';
    const token = localStorage.getItem('auth_token');

    if (!token) {
      console.warn('No auth token found, skipping SSE connection');
      return;
    }

    // Build the SSE URL with topics
    const topicsParam = topics.join(',');
    const url = `${API_BASE_URL}/api/v1/events?topics=${topicsParam}`;

    console.log('Connecting to SSE:', url);

    // Create EventSource connection
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      console.log('SSE connection opened');
    };

    eventSource.onmessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data);
        console.log('SSE event received:', data);
        onEvent(data as SSEEvent);
      } catch (error) {
        console.error('Failed to parse SSE event:', error);
      }
    };

    eventSource.onerror = (error: Event) => {
      console.error('SSE connection error:', error);
      if (onError) {
        onError(error);
      }
      // EventSource will automatically try to reconnect
    };

    // Cleanup on unmount
    return () => {
      console.log('Closing SSE connection');
      eventSource.close();
      eventSourceRef.current = null;
    };
  }, [topics, enabled, onEvent, onError]);

  return {
    close: () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
    },
  };
}
