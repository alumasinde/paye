import { useState } from "react";
import { api } from "../api";

export default function ChangePasswordPage() {
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setStatus("");
    setError("");
    setBusy(true);
    try {
      await api("/admin/auth/change-password", { method: "POST", body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }) });
      setStatus("Password changed.");
      setOldPassword("");
      setNewPassword("");
    } catch (e: any) {
      setError(e.message || "Could not change password");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div>
      <h2>Change Password</h2>
      <section>
        <form onSubmit={submit} className="formGrid">
          <label className="wide">Current password<input type="password" value={oldPassword} onChange={(e) => setOldPassword(e.target.value)} required /></label>
          <label className="wide">New password (12+ characters)<input type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} required minLength={12} /></label>
          <div className="wide actions">
            <button disabled={busy}>{busy ? "Saving…" : "Change password"}</button>
          </div>
        </form>
        {status && <p>{status}</p>}
        {error && <p className="error">{error}</p>}
      </section>
    </div>
  );
}
