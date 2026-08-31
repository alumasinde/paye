import { Fragment, useEffect, useState } from "react";
import { api } from "../api";

type AuditEntry = {
  id: string;
  admin_email?: string;
  action: string;
  entity_type: string;
  entity_id?: string;
  created_at: string;
  before?: any;
  after?: any;
};

export default function AuditLogPage() {
  const [items, setItems] = useState<AuditEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [expanded, setExpanded] = useState<string | null>(null);

  useEffect(() => {
    api("/admin/audit-logs?limit=100")
      .then((b) => setItems(b.items || []))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div>
      <h2>Audit Log</h2>
      <p>Every rule set creation and publish is recorded here with who did it and when.</p>
      {loading && <p className="empty">Loading…</p>}
      {error && <p className="error">{error}</p>}
      {!loading && !error && items.length === 0 && <p className="empty">No audit entries yet.</p>}
      {!loading && items.length > 0 && (
        <table>
          <thead>
            <tr><th>When</th><th>Who</th><th>Action</th><th>Entity</th><th></th></tr>
          </thead>
          <tbody>
            {items.map((e) => (
              <Fragment key={e.id}>
                <tr>
                  <td>{new Date(e.created_at).toLocaleString()}</td>
                  <td>{e.admin_email || "—"}</td>
                  <td>{e.action}</td>
                  <td>{e.entity_type} {e.entity_id ? `(${e.entity_id.slice(0, 8)}…)` : ""}</td>
                  <td>
                    <button className="secondary" onClick={() => setExpanded(expanded === e.id ? null : e.id)}>
                      {expanded === e.id ? "Hide" : "Details"}
                    </button>
                  </td>
                </tr>
                {expanded === e.id && (
                  <tr>
                    <td colSpan={5}>
                      <pre className="json">{JSON.stringify({ before: e.before, after: e.after }, null, 2)}</pre>
                    </td>
                  </tr>
                )}
              </Fragment>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
