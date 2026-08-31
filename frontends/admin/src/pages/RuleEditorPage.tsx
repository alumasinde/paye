import RuleEditor from "../components/RuleEditor";
import { useLocation, useNavigate } from "react-router-dom";
import type { RuleSet } from "../types";

export default function RuleEditorPage() {
  const nav = useNavigate();
  const location = useLocation();
  const prefill = (location.state as { prefill?: RuleSet } | null)?.prefill;
  return (
    <div>
      <RuleEditor initial={prefill} onSaved={() => nav("/rules")} />
    </div>
  );
}
