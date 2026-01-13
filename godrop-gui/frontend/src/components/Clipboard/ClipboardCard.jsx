import React from 'react';

export const ClipboardCard = ({ item, index, total, onCopy }) => {
    const truncate = (text, limit = 200) => {
        if (text.length <= limit) return text;
        return text.substring(0, limit - 3) + "...";
    };

    return (
        <div className="clipboard-card vibrant-anim" onClick={() => onCopy(item)}>
            <div className="card-header">
                <span className="item-id"># {total - index}</span>
                <span className="item-type">TEXT</span>
            </div>
            <div className="card-body">
                {truncate(item)}
            </div>
            <div className="card-footer">
                <span className="action-hint">CLICK TO COPY</span>
            </div>
        </div>
    );
};
