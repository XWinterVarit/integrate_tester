import React, { useRef, useState, useCallback } from 'react';
import { FloatingWindow as FloatingWindowType } from '../../types';

interface FloatingWindowProps {
  window: FloatingWindowType;
  onClose: (id: string) => void;
  onPopOut: (win: FloatingWindowType) => void;
  children: React.ReactNode;
}

const FloatingWindow: React.FC<FloatingWindowProps> = ({ window: win, onClose, onPopOut, children }) => {
  const [pos, setPos] = useState({ x: win.x, y: win.y });
  const [size, setSize] = useState({ w: win.width, h: win.height });
  const dragRef = useRef<{ startX: number; startY: number; origX: number; origY: number } | null>(null);
  const resizeRef = useRef<{ startX: number; startY: number; origW: number; origH: number } | null>(null);

  const onDragStart = useCallback((e: React.MouseEvent) => {
    dragRef.current = { startX: e.clientX, startY: e.clientY, origX: pos.x, origY: pos.y };
    const onMove = (ev: MouseEvent) => {
      if (!dragRef.current) return;
      setPos({
        x: dragRef.current.origX + ev.clientX - dragRef.current.startX,
        y: dragRef.current.origY + ev.clientY - dragRef.current.startY,
      });
    };
    const onUp = () => {
      dragRef.current = null;
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
    };
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  }, [pos]);

  const onResizeStart = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    resizeRef.current = { startX: e.clientX, startY: e.clientY, origW: size.w, origH: size.h };
    const onMove = (ev: MouseEvent) => {
      if (!resizeRef.current) return;
      setSize({
        w: Math.max(300, resizeRef.current.origW + ev.clientX - resizeRef.current.startX),
        h: Math.max(200, resizeRef.current.origH + ev.clientY - resizeRef.current.startY),
      });
    };
    const onUp = () => {
      resizeRef.current = null;
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
    };
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  }, [size]);

  const handlePopOut = () => {
    const content = document.getElementById(`floating-body-${win.id}`)?.innerHTML || '';
    const styles = Array.from(document.querySelectorAll('style, link[rel="stylesheet"]'))
      .map(el => el.outerHTML)
      .join('\n');
    const popup = window.open('', '_blank', `width=${size.w},height=${size.h},left=${pos.x},top=${pos.y}`);
    if (popup) {
      popup.document.write(`
        <!DOCTYPE html><html><head><title>${win.title}</title>
        ${styles}
        <style>body{margin:0;padding:16px;}</style>
        </head><body>${content}</body></html>
      `);
      popup.document.close();
    }
    onPopOut(win);
  };

  return (
    <div
      className="floating-window"
      style={{ left: pos.x, top: pos.y, width: size.w, height: size.h }}
    >
      <div className="floating-header" onMouseDown={onDragStart}>
        <span className="title">{win.title}</span>
        <div className="actions">
          <button onClick={handlePopOut}>↗ Pop Out</button>
          <button onClick={() => onClose(win.id)}>✕</button>
        </div>
      </div>
      <div className="floating-body" id={`floating-body-${win.id}`}>
        {children}
      </div>
      <div className="resize-handle" onMouseDown={onResizeStart} />
    </div>
  );
};

export default FloatingWindow;
