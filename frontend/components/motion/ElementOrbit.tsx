export default function ElementOrbit({ compact = false }: { compact?: boolean }) {
  const nodes = [
    { key: "wood", label: "木", className: "element-orbit__node--wood", i: 0 },
    { key: "fire", label: "火", className: "element-orbit__node--fire", i: 1 },
    { key: "earth", label: "土", className: "element-orbit__node--earth", i: 2 },
    { key: "metal", label: "金", className: "element-orbit__node--metal", i: 3 },
    { key: "water", label: "水", className: "element-orbit__node--water", i: 4 },
  ] as const;

  return (
    <div
      className={`element-orbit ${compact ? "element-orbit--compact" : ""}`}
      aria-hidden
    >
      <div className="element-orbit__ring element-orbit__ring--outer" />
      <div className="element-orbit__ring element-orbit__ring--inner" />
      <div className="element-orbit__center">
        <span className="element-orbit__title">五行流转</span>
        <span className="element-orbit__subtitle">简化干支</span>
      </div>
      <div className="element-orbit__track">
        {nodes.map((node) => (
          <span
            key={node.key}
            className={`element-orbit__node ${node.className}`}
            style={{ ["--i" as string]: node.i }}
          >
            {node.label}
          </span>
        ))}
      </div>
    </div>
  );
}
