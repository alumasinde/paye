import { Text, TextProps } from "react-native";
import { kes } from "../lib/money";

type Props = TextProps & { value: string | number };

export function MoneyText({ value, ...rest }: Props) {
  return <Text {...rest}>{kes(value)}</Text>;
}
