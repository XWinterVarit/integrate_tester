import React from 'react';

interface Props {
  open: boolean;
  onClose: () => void;
}

const AboutModal: React.FC<Props> = ({ open, onClose }) => {
  if (!open) return null;

  return (
    <div className="cm-overlay" onClick={onClose}>
      <div className="cm-about-box" onClick={e => e.stopPropagation()}>
        <div className="cm-about-title">DB Viewer</div>
        <div className="cm-about-version">Version 1.0.0</div>
        <div className="cm-about-desc">
          A lightweight Oracle DB viewer for internal use.
        </div>
        <div className="cm-about-footer">
          <button className="cm-btn cm-btn-ghost" onClick={onClose}>Close</button>
        </div>
      </div>
    </div>
  );
};

export default AboutModal;
