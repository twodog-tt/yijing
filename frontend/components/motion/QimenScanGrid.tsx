export default function QimenScanGrid({ compact = false }: { compact?: boolean }) {
  const sizeClass = compact ? "qimen-scan-grid--compact" : "qimen-scan-grid--hero";

  return (
    <div className={`qimen-scan-grid ${sizeClass}`} aria-hidden>
      <div className="qimen-scan-grid__glow" />
      <div className="qimen-scan-grid__scan-line" />
      <div className="qimen-scan-grid__scan-diagonal" />
      <div className="qimen-scan-grid__dot" />
      <div className="qimen-scan-grid__cells">
        {Array.from({ length: 9 }).map((_, index) => (
          <div
            key={index}
            className="qimen-scan-grid__cell"
            style={{ ["--cell-i" as string]: index }}
          />
        ))}
      </div>
    </div>
  );
}
