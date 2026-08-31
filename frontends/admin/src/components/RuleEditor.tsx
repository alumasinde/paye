import { useState } from "react";
import type { RuleSet, Component } from "../types";
import { api } from "../api";
import RulePreview from "./RulePreview";

const emptyComponent = (): Component => ({
  component_code: "",
  component_type: "STATUTORY_DEDUCTION",
  name: "",
  calculation_order: 100,
  reduces_taxable_income: false,
  reduces_net_pay: true,
  formula_type: "PERCENTAGE",
  payload: { rate: 0 },
  is_active: true,
});
const emptyRule = (): RuleSet => ({
  code: "KE_PAYROLL",
  name: "Kenya Payroll Rules",
  jurisdiction: "KE",
  effective_from: new Date().toISOString().slice(0, 10),
  components: [],
});

// Switching formula type resets the payload to that type's expected shape
// - the previous shape (e.g. a band list) wouldn't mean anything under a
// different formula, and publishing expects exactly one of these three
// shapes (see internal/payrollrules/service.Workflow's materialize step).
function defaultPayloadFor(formulaType: string): any {
  switch (formulaType) {
    case "FIXED":
      return { amount: 0 };
    case "PERCENTAGE":
      return { rate: 0 };
    case "PERCENTAGE_WITH_MINIMUM":
      return { rate: 0, minimum_amount: 0 };
    case "CAPPED_PERCENTAGE":
      return { rate: 0, upper_earnings_limit: 0, maximum_contribution: 0 };
    case "BANDS":
      return { bands: [{ from: 0, to: null, rate: 0 }] };
    case "TIERED_FIXED_AMOUNT":
      return { bands: [{ from: 0, to: null, fixed_amount: 0 }] };
    default:
      return {};
  }
}

export default function RuleEditor({ initial, onSaved }: { initial?: RuleSet; onSaved?: () => void }) {
  const [x, setX] = useState<RuleSet>(initial || emptyRule());
  const [validation, setValidation] = useState<any>(null);
  const [busy, setBusy] = useState(false);
  const [saveError, setSaveError] = useState("");

  const patch = (p: Partial<RuleSet>) => setX({ ...x, ...p });
  const add = () => patch({ components: [...x.components, emptyComponent()] });
  const component = (i: number, p: Partial<Component>) => patch({ components: x.components.map((c, n) => (n === i ? { ...c, ...p } : c)) });
  const remove = (i: number) => patch({ components: x.components.filter((_, n) => n !== i) });

  async function validate() {
    setBusy(true);
    try {
      setValidation(await api("/admin/rule-sets/validate", { method: "POST", body: JSON.stringify(x) }));
    } finally {
      setBusy(false);
    }
  }
  async function save() {
    setSaveError("");
    await validate();
    if (validation && !validation.valid) return;
    try {
      await api("/admin/rule-sets", { method: "POST", body: JSON.stringify(x) });
      onSaved?.();
    } catch (e: any) {
      setSaveError(e.message || "Could not save draft");
    }
  }

  return (
    <div className="editor">
      <div className="editorHead">
        <div>
          <h2>Rule Editor</h2>
          <p>Draft a version. Published rules must not be edited in place.</p>
        </div>
        <div className="actions">
          <button className="secondary" onClick={validate} disabled={busy}>Validate</button>
          <button onClick={save} disabled={busy}>Save Draft</button>
        </div>
      </div>

      <div className="formGrid">
        <label>Name<input value={x.name} onChange={(e) => patch({ name: e.target.value })} /></label>
        <label>Code<input value={x.code} onChange={(e) => patch({ code: e.target.value })} /></label>
        <label>Effective From<input type="date" value={x.effective_from} onChange={(e) => patch({ effective_from: e.target.value })} /></label>
        <label>Effective To<input type="date" value={x.effective_to || ""} onChange={(e) => patch({ effective_to: e.target.value || undefined })} /></label>
        <label className="wide">Official Source Reference<input value={x.source_reference || ""} onChange={(e) => patch({ source_reference: e.target.value })} /></label>
        <label className="wide">Notes<textarea value={x.source_notes || ""} onChange={(e) => patch({ source_notes: e.target.value })} /></label>
      </div>

      <h3>Components</h3>
      {x.components.map((c, i) => (
        <div className="component" key={i}>
          <div className="formGrid">
            <label>Code<input value={c.component_code} onChange={(e) => component(i, { component_code: e.target.value })} /></label>
            <label>Name<input value={c.name} onChange={(e) => component(i, { name: e.target.value })} /></label>
            <label>
              Type
              <select value={c.component_type} onChange={(e) => component(i, { component_type: e.target.value as any })}>
                <option>PAYE_BANDS</option>
                <option>STATUTORY_DEDUCTION</option>
                <option>RELIEF</option>
                <option>CONFIGURATION</option>
              </select>
            </label>
            <label>
              Calculation
              <select
                value={c.formula_type}
                onChange={(e) => component(i, { formula_type: e.target.value as any, payload: defaultPayloadFor(e.target.value) })}
              >
                <option value="FIXED">Fixed amount</option>
                <option value="PERCENTAGE">Percentage</option>
                <option value="PERCENTAGE_WITH_MINIMUM">Percentage with minimum (e.g. SHIF)</option>
                <option value="CAPPED_PERCENTAGE">Capped percentage (e.g. NSSF)</option>
                <option value="BANDS">Progressive bands (e.g. PAYE)</option>
                <option value="TIERED_FIXED_AMOUNT">Tiered fixed amount</option>
                <option value="JSON">Advanced (raw JSON)</option>
              </select>
            </label>
          </div>

          <CalculationEditor component={c} onChange={(p) => component(i, { payload: p })} />

          <div className="formGrid">
            <label>Calculation order<input type="number" value={c.calculation_order} onChange={(e) => component(i, { calculation_order: Number(e.target.value) })} /></label>
            <label className="checkboxLabel">
              <input type="checkbox" checked={c.reduces_taxable_income} onChange={(e) => component(i, { reduces_taxable_income: e.target.checked })} /> Reduces taxable income
            </label>
            <label className="checkboxLabel">
              <input type="checkbox" checked={c.reduces_net_pay} onChange={(e) => component(i, { reduces_net_pay: e.target.checked })} /> Reduces net pay
            </label>
            <label className="checkboxLabel">
              <input type="checkbox" checked={c.is_active} onChange={(e) => component(i, { is_active: e.target.checked })} /> Active
            </label>
          </div>

          <button className="danger" onClick={() => remove(i)}>Remove component</button>
        </div>
      ))}
      <button className="secondary" onClick={add}>+ Add component</button>

      {validation && (
        <div className={validation.valid ? "validation ok" : "validation bad"}>
          <strong>{validation.valid ? "Valid" : "Validation failed"}</strong>
          {validation.errors?.map((e: string) => <div key={e}>{e}</div>)}
          {validation.warnings?.map((e: string) => <div key={e}>Warning: {e}</div>)}
        </div>
      )}
      {saveError && <p className="error">{saveError}</p>}

      <RulePreview rule={x} />
    </div>
  );
}

// Renders the right structured editor for the component's calculation
// type - FIXED and PERCENTAGE are single-value inputs, BANDS is a real
// row-by-row table editor (add/remove rows), matching exactly what
// internal/payrollrules/service.Workflow expects when materializing a
// published rule set into the live calculator's tables. JSON stays a raw
// textarea fallback and cannot be published (server-side rejects it).
function CalculationEditor({ component: c, onChange }: { component: Component; onChange: (payload: any) => void }) {
  if (c.formula_type === "FIXED") {
    const amount = typeof c.payload?.amount === "number" ? c.payload.amount : 0;
    return (
      <div className="formGrid">
        <label className="wide">
          Amount (KES)
          <input type="number" step="0.01" value={amount} onChange={(e) => onChange({ amount: Number(e.target.value) })} />
        </label>
      </div>
    );
  }

  if (c.formula_type === "PERCENTAGE") {
    const rate = typeof c.payload?.rate === "number" ? c.payload.rate : 0;
    return (
      <div className="formGrid">
        <label className="wide">
          Rate (as a fraction - 0.06 = 6%)
          <input type="number" step="0.0001" value={rate} onChange={(e) => onChange({ rate: Number(e.target.value) })} />
        </label>
        <p className="hint">= {(rate * 100).toFixed(2)}%</p>
      </div>
    );
  }

  if (c.formula_type === "PERCENTAGE_WITH_MINIMUM") {
    const rate = typeof c.payload?.rate === "number" ? c.payload.rate : 0;
    const minimumAmount = typeof c.payload?.minimum_amount === "number" ? c.payload.minimum_amount : 0;
    return (
      <div className="formGrid">
        <label>
          Rate (as a fraction - 0.0275 = 2.75%)
          <input type="number" step="0.0001" value={rate} onChange={(e) => onChange({ rate: Number(e.target.value), minimum_amount: minimumAmount })} />
        </label>
        <label>
          Minimum amount (KES)
          <input type="number" step="0.01" value={minimumAmount} onChange={(e) => onChange({ rate, minimum_amount: Number(e.target.value) })} />
        </label>
        <p className="hint wide">= {(rate * 100).toFixed(2)}% of gross, or KES {minimumAmount || 0}, whichever is higher. This is how SHIF is calculated.</p>
      </div>
    );
  }

  if (c.formula_type === "CAPPED_PERCENTAGE") {
    const rate = typeof c.payload?.rate === "number" ? c.payload.rate : 0;
    const upperLimit = typeof c.payload?.upper_earnings_limit === "number" ? c.payload.upper_earnings_limit : 0;
    const maxContribution = typeof c.payload?.maximum_contribution === "number" ? c.payload.maximum_contribution : 0;
    const set = (patch: Partial<{ rate: number; upper_earnings_limit: number; maximum_contribution: number }>) =>
      onChange({ rate, upper_earnings_limit: upperLimit, maximum_contribution: maxContribution, ...patch });
    return (
      <div className="formGrid">
        <label>
          Rate (as a fraction - 0.06 = 6%)
          <input type="number" step="0.0001" value={rate} onChange={(e) => set({ rate: Number(e.target.value) })} />
        </label>
        <label>
          Upper earnings limit (KES) - the rate only applies up to this much gross pay
          <input type="number" step="0.01" value={upperLimit} onChange={(e) => set({ upper_earnings_limit: Number(e.target.value) })} />
        </label>
        <label className="wide">
          Maximum contribution (KES) - the hard cap regardless of gross pay
          <input type="number" step="0.01" value={maxContribution} onChange={(e) => set({ maximum_contribution: Number(e.target.value) })} />
        </label>
        <p className="hint wide">= min(gross, {upperLimit || 0}) × {(rate * 100).toFixed(2)}%, capped at KES {maxContribution || 0}. This is how NSSF is calculated.</p>
      </div>
    );
  }

  if (c.formula_type === "BANDS" || c.formula_type === "TIERED_FIXED_AMOUNT") {
    const isTiered = c.formula_type === "TIERED_FIXED_AMOUNT";
    const bands: Array<{ from: number; to: number | null; rate?: number; fixed_amount?: number }> = Array.isArray(c.payload?.bands) ? c.payload.bands : [];
    function updateBand(i: number, patch: Partial<{ from: number; to: number | null; rate: number; fixed_amount: number }>) {
      const next = bands.map((b, n) => (n === i ? { ...b, ...patch } : b));
      onChange({ bands: next });
    }
    function addBand() {
      const last = bands[bands.length - 1];
      const nextFrom = last ? last.to ?? 0 : 0;
      onChange({ bands: [...bands, isTiered ? { from: nextFrom, to: null, fixed_amount: 0 } : { from: nextFrom, to: null, rate: 0 }] });
    }
    function removeBand(i: number) {
      onChange({ bands: bands.filter((_, n) => n !== i) });
    }
    return (
      <div className="bandsEditor">
        <table>
          <thead>
            <tr><th>From (KES)</th><th>To (KES, blank = no limit)</th><th>{isTiered ? "Fixed amount (KES)" : "Rate (fraction)"}</th><th></th></tr>
          </thead>
          <tbody>
            {bands.map((b, i) => (
              <tr key={i}>
                <td><input type="number" step="0.01" value={b.from} onChange={(e) => updateBand(i, { from: Number(e.target.value) })} /></td>
                <td><input type="number" step="0.01" value={b.to ?? ""} placeholder="no limit" onChange={(e) => updateBand(i, { to: e.target.value === "" ? null : Number(e.target.value) })} /></td>
                <td>
                  {isTiered ? (
                    <input type="number" step="0.01" value={b.fixed_amount ?? 0} onChange={(e) => updateBand(i, { fixed_amount: Number(e.target.value) })} />
                  ) : (
                    <input type="number" step="0.0001" value={b.rate ?? 0} onChange={(e) => updateBand(i, { rate: Number(e.target.value) })} />
                  )}
                </td>
                <td><button className="danger" onClick={() => removeBand(i)}>Remove</button></td>
              </tr>
            ))}
          </tbody>
        </table>
        <button className="secondary" onClick={addBand}>+ Add band</button>
      </div>
    );
  }

  // JSON fallback - advanced/escape hatch, cannot be published as-is.
  return (
    <div>
      <p className="hint warning">Advanced: this formula type cannot be published. Switch to one of the calculation types above before submitting for review.</p>
      <label>
        Payload JSON
        <textarea
          className="json"
          value={JSON.stringify(c.payload, null, 2)}
          onChange={(e) => {
            try {
              onChange(JSON.parse(e.target.value));
            } catch {
              /* keep typing until it's valid JSON again */
            }
          }}
        />
      </label>
    </div>
  );
}
