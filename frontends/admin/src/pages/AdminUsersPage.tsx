import { useEffect, useState } from "react";
import { api } from "../api";

type AdminUser = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  status: string;
  roles: string[];
  created_at: string;
};

const ROLES = ["SUPER_ADMIN", "RULE_EDITOR", "RULE_APPROVER", "AUDITOR"];

export default function AdminUsersPage() {
  const [items, setItems] = useState<AdminUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [form, setForm] = useState({ Email: "", Password: "", FirstName: "", LastName: "", Role: "AUDITOR" });
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState("");

  function load() {
    setLoading(true);
    api("/admin/users")
      .then((b) => setItems(b.items || []))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }
  useEffect(load, []);

  async function createUser(e: React.FormEvent) {
    e.preventDefault();
    setCreateError("");
    setCreating(true);
    try {
      await api("/admin/users", { method: "POST", body: JSON.stringify(form) });
      setForm({ Email: "", Password: "", FirstName: "", LastName: "", Role: "AUDITOR" });
      load();
    } catch (e: any) {
      setCreateError(e.message || "Could not create admin");
    } finally {
      setCreating(false);
    }
  }

  async function toggleStatus(u: AdminUser) {
    const next = u.status === "ACTIVE" ? "DISABLED" : "ACTIVE";
    try {
      await api(`/admin/users/${u.id}/status`, { method: "PATCH", body: JSON.stringify({ Status: next }) });
      load();
    } catch (e: any) {
      setError(e.message || "Could not update status");
    }
  }

  return (
    <div>
      <div className="editorHead">
        <div>
          <h2>Admin Users</h2>
          <p>Accounts that can sign in to this panel, and what each role can do.</p>
        </div>
      </div>

      <section>
        <h3>Add an admin</h3>
        <form onSubmit={createUser} className="formGrid">
          <label>Email<input type="email" value={form.Email} onChange={(e) => setForm({ ...form, Email: e.target.value })} required /></label>
          <label>Password<input type="password" value={form.Password} onChange={(e) => setForm({ ...form, Password: e.target.value })} required minLength={12} /></label>
          <label>First name<input value={form.FirstName} onChange={(e) => setForm({ ...form, FirstName: e.target.value })} required /></label>
          <label>Last name<input value={form.LastName} onChange={(e) => setForm({ ...form, LastName: e.target.value })} required /></label>
          <label>
            Role
            <select value={form.Role} onChange={(e) => setForm({ ...form, Role: e.target.value })}>
              {ROLES.map((r) => <option key={r}>{r}</option>)}
            </select>
          </label>
          <div className="wide actions">
            <button disabled={creating}>{creating ? "Creating…" : "Create admin"}</button>
          </div>
        </form>
        {createError && <p className="error">{createError}</p>}
      </section>

      {loading && <p className="empty">Loading…</p>}
      {error && <p className="error">{error}</p>}
      {!loading && !error && (
        <table>
          <thead>
            <tr><th>Name</th><th>Email</th><th>Roles</th><th>Status</th><th></th></tr>
          </thead>
          <tbody>
            {items.map((u) => (
              <tr key={u.id}>
                <td>{u.first_name} {u.last_name}</td>
                <td>{u.email}</td>
                <td>{u.roles?.join(", ")}</td>
                <td><span className={u.status === "ACTIVE" ? "badge published" : "badge"}>{u.status}</span></td>
                <td><button className="secondary" onClick={() => toggleStatus(u)}>{u.status === "ACTIVE" ? "Disable" : "Enable"}</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
