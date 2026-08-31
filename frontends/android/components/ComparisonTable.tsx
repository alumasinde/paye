import { ScrollView, StyleSheet, Text, View } from "react-native";
import type { SavedCalculation } from "../lib/accountApi";
import { kes } from "../lib/money";
import { colors, radius } from "../lib/theme";

function statutory(item: SavedCalculation, code: string): string | null {
  const match = item.payload.statutory_deductions?.find((deduction) => deduction.code === code);
  return match ? match.amount : null;
}

const ROWS: { label: string; emphasis?: boolean; get: (item: SavedCalculation) => string }[] = [
  { label: "Gross salary", get: (item) => item.gross_salary },
  { label: "PAYE", get: (item) => statutory(item, "PAYE") ?? item.payload.paye },
  { label: "NSSF", get: (item) => statutory(item, "NSSF") ?? "0" },
  { label: "SHIF", get: (item) => statutory(item, "SHIF") ?? statutory(item, "NHIF") ?? "0" },
  { label: "AHL", get: (item) => statutory(item, "AHL") ?? "0" },
  { label: "Net salary", get: (item) => item.net_salary, emphasis: true },
];

const ROW_HEIGHT = 52;
const COLUMN_WIDTH = 136;

// Fixed label column on the left + a horizontally-scrolling set of value
// columns on the right, so this works for 2 or 3 saved calculations
// without the row labels scrolling out of view.
export function ComparisonTable({ items }: { items: SavedCalculation[] }) {
  return (
    <View style={s.wrap}>
      <View style={s.labelColumn}>
        <View style={[s.headerCell, { height: ROW_HEIGHT }]} />
        {ROWS.map((row) => (
          <View key={row.label} style={[s.cell, { height: ROW_HEIGHT }]}>
            <Text style={s.rowLabel}>{row.label}</Text>
          </View>
        ))}
      </View>
      <ScrollView horizontal showsHorizontalScrollIndicator={false}>
        <View style={s.valueColumns}>
          {items.map((item) => (
            <View key={item.id} style={[s.column, { width: COLUMN_WIDTH }]}>
              <View style={[s.headerCell, s.columnHeader, { height: ROW_HEIGHT }]}>
                <Text style={s.columnTitle} numberOfLines={2}>
                  {item.label ?? item.calculation_date}
                </Text>
              </View>
              {ROWS.map((row) => (
                <View key={row.label} style={[s.cell, { height: ROW_HEIGHT }]}>
                  <Text style={[s.value, row.emphasis ? s.valueEmphasis : null]}>{kes(row.get(item))}</Text>
                </View>
              ))}
            </View>
          ))}
        </View>
      </ScrollView>
    </View>
  );
}

const s = StyleSheet.create({
  wrap: { flexDirection: "row", borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, overflow: "hidden", backgroundColor: colors.surface },
  labelColumn: { width: 124, borderRightWidth: 1, borderRightColor: colors.border },
  valueColumns: { flexDirection: "row" },
  column: { borderRightWidth: 1, borderRightColor: colors.border },
  headerCell: { justifyContent: "center", paddingHorizontal: 10, backgroundColor: "#F8FAFC", borderBottomWidth: 1, borderBottomColor: colors.border },
  columnHeader: { alignItems: "center" },
  columnTitle: { fontSize: 12, fontWeight: "800", color: colors.text, textAlign: "center" },
  cell: { justifyContent: "center", paddingHorizontal: 10, borderBottomWidth: 1, borderBottomColor: "#F2F4F7" },
  rowLabel: { fontSize: 12, fontWeight: "700", color: colors.textMuted },
  value: { fontSize: 13, fontWeight: "700", color: colors.text },
  valueEmphasis: { fontWeight: "900", fontSize: 14 },
});
