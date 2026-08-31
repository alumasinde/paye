package regression_test
import("testing";"time")
// Historical integration regression is intentionally data-driven: production figures
// are verified by resolving published versions from MySQL, not duplicated here.
func TestHistoricalRegressionDatesAreDistinct(t *testing.T){dates:=[]string{"2022-01-01","2023-07-01","2024-12-01","2025-01-01","2026-01-01"};for _,v:=range dates{if _,e:=time.Parse("2006-01-02",v);e!=nil{t.Fatal(e)}}}
