export const FileSelection = ({ selectedFiles, onRemove }) => {
    const basename = (path) => path.split(/[/\\]/).pop();

    return (
        <>
            <div className="section-label">
                <span>Selected Files</span>
                <span>{selectedFiles.length}</span>
            </div>
            <div className="list-box">
                {selectedFiles.map(path => (
                    <div key={path} className="list-item">
                        <span style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>{basename(path)}</span>
                        <span style={{ opacity: 0.5, cursor: 'pointer' }} onClick={() => onRemove(path)}>×</span>
                    </div>
                ))}
            </div>
        </>
    );
};
