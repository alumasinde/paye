import { useState } from "react";
import { api } from "../api";
import type { RuleSet } from "../types";

export default function RulePreview({ rule }: { rule: RuleSet }) {
  const [gross, setGross] = useState("100000");
  const [result, setResult] = useState<any>();
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function run() {
    setBusy(true);
    setError("");
    setResult(undefined);
    try {
      setResult(await api("/admin/rule-sets/preview", { method: "POST", body: JSON.stringify({ rule_set: rule, gross_salary: gross }) }));
    } catch (e: any) {
      setError(e.message || "Preview failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section>
      <h3>Sandbox Preview</h3>
      <p>Preview a draft before publication. This does not alter production calculations.</p>
      <div className="actions">
        <input value={gross} onChange={(e) => setGross(e.target.value)} placeholder="Gross salary" />
        <button onClick={run} disabled={busy}>{busy ? "Running…" : "Run preview"}</button>
      </div>
      {error && <p className="error">{error}</p>}
      {result && <pre className="json">{JSON.stringify(result, null, 2)}</pre>}
    </section>
  );
}
