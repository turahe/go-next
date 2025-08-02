import { io, Socket } from 'socket.io-client';
import { WebSocketMessage } from '../types';

class WebSocketService {
  private socket: Socket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000; // Start with 1 second
  private isConnecting = false;
  private messageHandlers: Map<string, ((message: WebSocketMessage) => void)[]> = new Map();

  constructor() {
    this.setupSocket();
  }

  private setupSocket() {
    if (this.isConnecting || this.socket?.connected) {
      return;
    }

    this.isConnecting = true;
    
    // Get the API URL from environment or use default
    const apiUrl = (import.meta as any).env?.VITE_API_URL || 'http://localhost:8080';
    const wsUrl = apiUrl.replace('http', 'ws');
    
    try {
      this.socket = io(wsUrl, {
        path: '/api/v1/ws/connect',
        transports: ['websocket'],
        autoConnect: false,
        reconnection: false, // We'll handle reconnection manually
      });

      this.setupEventHandlers();
      this.connect();
    } catch (error) {
      console.error('Failed to setup WebSocket:', error);
      this.isConnecting = false;
      this.scheduleReconnect();
    }
  }

  private setupEventHandlers() {
    if (!this.socket) return;

    this.socket.on('connect', () => {
      console.log('WebSocket connected');
      this.isConnecting = false;
      this.reconnectAttempts = 0;
      this.reconnectDelay = 1000;
    });

    this.socket.on('disconnect', (reason) => {
      console.log('WebSocket disconnected:', reason);
      this.isConnecting = false;
      
      if (reason === 'io server disconnect') {
        // Server disconnected us, try to reconnect
        this.scheduleReconnect();
      }
    });

    this.socket.on('connect_error', (error) => {
      console.error('WebSocket connection error:', error);
      this.isConnecting = false;
      this.scheduleReconnect();
    });

    this.socket.on('message', (message: WebSocketMessage) => {
      this.handleMessage(message);
    });

    // Handle different message types
    this.socket.on('notification', (message: WebSocketMessage) => {
      this.handleMessage({ ...message, type: 'notification' });
    });

    this.socket.on('connected', (message: WebSocketMessage) => {
      console.log('WebSocket connection established:', message);
    });
  }

  private handleMessage(message: WebSocketMessage) {
    const handlers = this.messageHandlers.get(message.type);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(message);
        } catch (error) {
          console.error('Error in message handler:', error);
        }
      });
    }
  }

  private scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1); // Exponential backoff
    
    console.log(`Scheduling reconnection attempt ${this.reconnectAttempts} in ${delay}ms`);
    
    setTimeout(() => {
      this.setupSocket();
    }, delay);
  }

  public connect() {
    if (this.socket && !this.socket.connected && !this.isConnecting) {
      this.socket.connect();
    }
  }

  public disconnect() {
    if (this.socket) {
      this.socket.disconnect();
      this.socket = null;
    }
  }

  public isConnected(): boolean {
    return this.socket?.connected || false;
  }

  public onMessage(type: string, handler: (message: WebSocketMessage) => void) {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, []);
    }
    this.messageHandlers.get(type)!.push(handler);
  }

  public offMessage(type: string, handler: (message: WebSocketMessage) => void) {
    const handlers = this.messageHandlers.get(type);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  public sendMessage(type: string, data: any) {
    if (this.socket?.connected) {
      this.socket.emit('message', { type, data, timestamp: new Date().toISOString() });
    }
  }

  public getConnectionStatus(): 'connected' | 'connecting' | 'disconnected' {
    if (this.socket?.connected) return 'connected';
    if (this.isConnecting) return 'connecting';
    return 'disconnected';
  }
}

// Create a singleton instance
export const websocketService = new WebSocketService();
export default websocketService; 