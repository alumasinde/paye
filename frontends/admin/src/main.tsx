import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Routes, Route, Navigate, NavLink, useNavigate } from "react-router-dom";
import { api, setToken, clearToken, isAuthenticated } from "./api";
import RuleEditorPage from "./pages/RuleEditorPage";
import WorkflowPage from "./pages/WorkflowPage";
import AdminUsersPage from "./pages/AdminUsersPage";
import AuditLogPage from "./pages/AuditLogPage";
import ChangePasswordPage from "./pages/ChangePasswordPage";
import { loadTheme } from "./theme";
import type { LiveRule, RuleSet } from "./types";
import "./styles/theme.css";
import "./styles/app.css";

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  if (!isAuthenticated()) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const nav = useNavigate();
  useEffect(() => {
    if (isAuthenticated()) nav("/", { replace: true });
  }, []);
  async function go(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      const b = await api("/admin/auth/login", { method: "POST", body: JSON.stringify({ email, password }) });
      setToken(b.tokens.access_token, b.tokens.refresh_token);
      nav("/");
    } catch (e: any) {
      setError(e.message || "Sign in failed");
    }
  }
  return (
    <div className="login">
      <form onSubmit={go}>
        <h1>Budget254 PAYE</h1>
        <p>Administrator sign in</p>
        <input placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
        <input type="password" placeholder="Password" value={password} onChange={(e) => setPassword(e.target.value)} />
        <button>Sign in</button>
        {error && <p className="error">{error}</p>}
      </form>
    </div>
  );
}

function Shell({ children }: { children: React.ReactNode }) {
  const nav = useNavigate();
  function logout() {
    clearToken();
    nav("/login");
  }
  return (
    <div className="shell">
      <aside>
        <h1>Budget254</h1>
        <p>PAYE Admin</p>
        <nav>
          <NavLink to="/" end>Dashboard</NavLink>
          <NavLink to="/live-rules">Live Payroll Rules</NavLink>
          <NavLink to="/rules">Rule Set Drafts</NavLink>
          <NavLink to="/rules/new">New Rule Set</NavLink>
          <NavLink to="/workflow">Publishing</NavLink>
          <NavLink to="/audit">Audit Log</NavLink>
          <NavLink to="/admins">Admin Users</NavLink>
          <NavLink to="/change-password">Change Password</NavLink>
        </nav>
        <div className="sidebarFooter">
          <button className="logoutButton" onClick={logout}>Sign out</button>
        </div>
      </aside>
      <main>{children}</main>
    </div>
  );
}

function todayISO() {
  return new Date().toISOString().slice(0, 10);
}

// The rules the calculator is actually using right now - GET
// /admin/live-rules reads from rule_definitions/rule_versions (full
// parameters and bands included), which is a completely different table
// set from payroll_rule_sets (the admin governance/draft workflow below).
// This is why "my seeded rules" only show up here, not under Rule Set
// Drafts - they're not the same data.
function useLiveRules() {
  const [rules, setRules] = useState<LiveRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  useEffect(() => {
    api(`/admin/live-rules?date=${todayISO()}`)
      .then((b) => setRules(b.rules || []))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);
  return { rules, loading, error };
}

// Converts a live rule's current values into a draft RuleSet with one
// pre-filled component, so tweaking an existing statutory rate (SHIF,
// NSSF, etc.) doesn't mean retyping everything from scratch. Only the
// formula types the publish bridge actually supports get a real
// conversion; anything else falls back to the JSON editor with the raw
// parameters visible, so nothing is silently lost.
function liveRuleToDraft(rule: LiveRule): RuleSet {
  const param = (name: string) => rule.parameters.find((p) => p.name === name)?.decimal;
  const num = (v?: string) => (v ? Number(v) : 0);

  const categoryToComponentType: Record<string, string> = {
    INCOME_TAX: "PAYE_BANDS",
    RELIEF: "RELIEF",
    STATUTORY_DEDUCTION: "STATUTORY_DEDUCTION",
    STATUTORY_CONTRIBUTION: "STATUTORY_DEDUCTION",
  };

  let formulaType = "JSON";
  let payload: any = { parameters: rule.parameters, bands: rule.bands };

  switch (rule.calculation_method) {
    case "FIXED_AMOUNT":
      formulaType = "FIXED";
      payload = { amount: num(param("amount")) };
      break;
    case "PERCENTAGE":
      formulaType = "PERCENTAGE";
      payload = { rate: num(param("rate")) };
      break;
    case "PERCENTAGE_WITH_MINIMUM":
      formulaType = "PERCENTAGE_WITH_MINIMUM";
      payload = { rate: num(param("rate")), minimum_amount: num(param("minimum_amount")) };
      break;
    case "CAPPED_PERCENTAGE":
      formulaType = "CAPPED_PERCENTAGE";
      payload = { rate: num(param("rate")), upper_earnings_limit: num(param("upper_earnings_limit")), maximum_contribution: num(param("maximum_contribution")) };
      break;
    case "PROGRESSIVE_BANDS":
      formulaType = "BANDS";
      payload = { bands: rule.bands.map((b) => ({ from: num(b.from), to: b.to ? num(b.to) : null, rate: num(b.rate) })) };
      break;
    case "TIERED_FIXED_AMOUNT":
      formulaType = "TIERED_FIXED_AMOUNT";
      payload = { bands: rule.bands.map((b) => ({ from: num(b.from), to: b.to ? num(b.to) : null, fixed_amount: num(b.fixed_amount) })) };
      break;
  }

  return {
    code: rule.code + "_UPDATE",
    name: rule.name,
    jurisdiction: "KE",
    effective_from: new Date().toISOString().slice(0, 10),
    source_notes: `Drafted from the live rule ${rule.code} (${rule.version_code}) as a starting point - adjust the values below and the effective date, then submit for review.`,
    components: [
      {
        component_code: rule.code,
        component_type: (categoryToComponentType[rule.category] || "STATUTORY_DEDUCTION") as any,
        name: rule.name,
        calculation_order: rule.calculation_order,
        reduces_taxable_income: rule.affects_taxable_income,
        reduces_net_pay: rule.affects_net_pay,
        formula_type: formulaType as any,
        payload,
        is_active: true,
      },
    ],
  };
}

function Dashboard() {
  const { rules, loading, error } = useLiveRules();
  return (
    <Shell>
      <h2>Dynamic Payroll Governance</h2>
      <div className="grid">
        <Card t="Live published rules" v={loading ? "…" : String(rules.length)} />
        <Card t="Effective as of" v={todayISO()} />
        <Card t="Draft rule sets" v="See Rule Set Drafts" />
      </div>
      <section>
        <h3>Safety rule</h3>
        <p>A published rule is historical evidence. Create a new version for changes instead of overwriting it.</p>
      </section>
      {error && <p className="error">Could not load live rules: {error}</p>}
    </Shell>
  );
}

function Card({ t, v }: { t: string; v: string }) {
  return (
    <div className="card">
      <small>{t}</small>
      <strong>{v}</strong>
    </div>
  );
}

function LiveRules() {
  const { rules, loading, error } = useLiveRules();
  const nav = useNavigate();
  function createVersion(rule: LiveRule) {
    nav("/rules/new", { state: { prefill: liveRuleToDraft(rule) } });
  }
  return (
    <Shell>
      <div className="editorHead">
        <div>
          <h2>Live Payroll Rules</h2>
          <p>What the calculator is actually using right now, read straight from the database - not a draft.</p>
        </div>
      </div>
      {loading && <p className="empty">Loading…</p>}
      {error && <p className="error">{error}</p>}
      {!loading && !error && rules.length === 0 && <p className="empty">No published rules are in effect today.</p>}
      {!loading && rules.length > 0 && (
        <table>
          <thead>
            <tr><th>Code</th><th>Name</th><th>Version</th><th>Method</th><th>Effective from</th><th>Effective to</th><th></th></tr>
          </thead>
          <tbody>
            {rules.map((x) => (
              <tr key={x.code}>
                <td>{x.code}</td>
                <td>{x.name}</td>
                <td>{x.version_code}</td>
                <td>{x.calculation_method}</td>
                <td>{x.effective_from}</td>
                <td>{x.effective_to || "—"}</td>
                <td><button className="secondary" onClick={() => createVersion(x)}>New version</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {!loading && rules.length > 0 && (
        <p className="hint">"New version" pre-fills the editor with this rule's current values - adjust the rate, cap, or bands, set a new effective date, then submit it for review to publish an update.</p>
      )}
    </Shell>
  );
}

function RuleSets() {
  const [items, setItems] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  useEffect(() => {
    api("/admin/rule-sets")
      .then((x) => setItems(x.items || []))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);
  return (
    <Shell>
      <div className="editorHead">
        <div>
          <h2>Rule Set Drafts</h2>
          <p>The versioned governance workflow for creating new rule sets - separate from the live rules above until one is published through here.</p>
        </div>
        <NavLink className="button" to="/rules/new">New Rule Set</NavLink>
      </div>
      {loading && <p className="empty">Loading…</p>}
      {error && <p className="error">{error}</p>}
      {!loading && !error && items.length === 0 && <p className="empty">No draft rule sets yet - this is a separate workspace from the live rules, so it starts empty. Use "New Rule Set" to start one.</p>}
      {!loading && items.length > 0 && (
        <table>
          <thead>
            <tr><th>Name</th><th>Code</th><th>Effective</th><th>Version</th><th>Status</th><th>ID</th></tr>
          </thead>
          <tbody>
            {items.map((x) => (
              <tr key={x.id}>
                <td>{x.name}</td>
                <td>{x.code}</td>
                <td>{x.effective_from}</td>
                <td>{x.version_number}</td>
                <td><span className={x.status === "PUBLISHED" ? "badge published" : "badge"}>{x.status}</span></td>
                <td><code className="idCell">{x.id}</code></td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {!loading && items.length > 0 && (
        <p className="hint">Copy a rule set's ID into the <NavLink to="/workflow">Publishing</NavLink> page to submit it for review, approve, or publish it.</p>
      )}
    </Shell>
  );
}

function Editor() {
  return <Shell><RuleEditorPage /></Shell>;
}
function Workflow() {
  return <Shell><WorkflowPage /></Shell>;
}
function Audit() {
  return <Shell><AuditLogPage /></Shell>;
}
function AdminUsers() {
  return <Shell><AdminUsersPage /></Shell>;
}
function ChangePassword() {
  return <Shell><ChangePasswordPage /></Shell>;
}

function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
      <Route path="/live-rules" element={<ProtectedRoute><LiveRules /></ProtectedRoute>} />
      <Route path="/rules" element={<ProtectedRoute><RuleSets /></ProtectedRoute>} />
      <Route path="/rules/new" element={<ProtectedRoute><Editor /></ProtectedRoute>} />
      <Route path="/workflow" element={<ProtectedRoute><Workflow /></ProtectedRoute>} />
      <Route path="/audit" element={<ProtectedRoute><Audit /></ProtectedRoute>} />
      <Route path="/admins" element={<ProtectedRoute><AdminUsers /></ProtectedRoute>} />
      <Route path="/change-password" element={<ProtectedRoute><ChangePassword /></ProtectedRoute>} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

loadTheme();
createRoot(document.getElementById("root")!).render(
  <BrowserRouter>
    <App />
  </BrowserRouter>
);
