package handler
import ("github.com/shopspring/decimal"; rules "github.com/alumasinde/budget254-paye-api/internal/payrollrules/model"; svc "github.com/alumasinde/budget254-paye-api/internal/payrollrules/service")
func servicePreview(x rules.RuleSet,g decimal.Decimal)(svc.PreviewResult,error){return svc.Preview(x,g)}
