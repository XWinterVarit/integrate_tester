import React, { useEffect } from 'react';

export interface ToastMessage {
  id: number;
  type: 'success' | 'error';
  text: string;
  duration?: string;
}

interface ToastProps {
  messages: ToastMessage[];
  onDismiss: (id: number) => void;
}

const Toast: React.FC<ToastProps> = ({ messages, onDismiss }) => {
  return (
    <div className="toast-container">
      {messages.map((msg) => (
        <ToastItem key={msg.id} msg={msg} onDismiss={onDismiss} />
      ))}
    </div>
  );
};

const ToastItem: React.FC<{ msg: ToastMessage; onDismiss: (id: number) => void }> = ({ msg, onDismiss }) => {
  useEffect(() => {
    const t = setTimeout(() => onDismiss(msg.id), 3000);
    return () => clearTimeout(t);
  }, [msg.id, onDismiss]);

  return (
    <div className={`toast-item toast-${msg.type}`} onClick={() => onDismiss(msg.id)}>
      <span className="toast-icon">{msg.type === 'success' ? '✓' : '✕'}</span>
      <span className="toast-text">{msg.text}</span>
      {msg.duration && <span className="toast-duration">{msg.duration}</span>}
    </div>
  );
};

export default Toast;
