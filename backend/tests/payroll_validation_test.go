package tests
import ("testing";"github.com/shopspring/decimal")
func TestDecimalPrecision(t *testing.T){a:=decimal.RequireFromString("0.10");b:=decimal.RequireFromString("0.20");if !a.Add(b).Equal(decimal.RequireFromString("0.30")){t.Fatal("decimal precision failed")}}
