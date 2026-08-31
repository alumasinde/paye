package handler
import ("net/http";"time";"github.com/alumasinde/budget254-paye-api/internal/payroll/dto";"github.com/alumasinde/budget254-paye-api/internal/rules/repository";"github.com/alumasinde/budget254-paye-api/internal/response")
type Handler struct{ Repo repository.Repository }
func(h Handler) Applicable(w http.ResponseWriter,r *http.Request){d:=r.URL.Query().Get("date");date,err:=time.Parse("2006-01-02",d);if err!=nil{response.JSON(w,400,map[string]any{"code":"VALIDATION_ERROR","message":"date must be YYYY-MM-DD"});return};rules,err:=h.Repo.Applicable(r.Context(),date);if err!=nil{response.JSON(w,500,map[string]any{"code":"RULES_UNAVAILABLE","message":"rules could not be resolved"});return};out:=dto.RulesResponse{CalculationDate:d};for _,x:=range rules{var e *string;if x.EffectiveTo!=nil{s:=x.EffectiveTo.Format("2006-01-02");e=&s};out.Rules=append(out.Rules,dto.Rule{x.Code,x.Name,x.Version,x.Method,x.EffectiveFrom.Format("2006-01-02"),e})};response.JSON(w,200,out)}

// AdminDetail exposes the same fully-resolved rule data (parameters and
// bands included, not just the summary Applicable() returns) to the admin
// panel - used to pre-fill a new draft rule set from an existing live
// rule's current values, so tweaking an existing statutory rate doesn't
// mean retyping everything from scratch.
func (h Handler) AdminDetail(w http.ResponseWriter, r *http.Request) {
	d := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", d)
	if err != nil {
		response.JSON(w, 400, map[string]any{"code": "VALIDATION_ERROR", "message": "date must be YYYY-MM-DD"})
		return
	}
	rules, err := h.Repo.ResolvedApplicable(r.Context(), date)
	if err != nil {
		response.JSON(w, 500, map[string]any{"code": "RULES_UNAVAILABLE", "message": "rules could not be resolved"})
		return
	}
	response.JSON(w, 200, map[string]any{"calculation_date": d, "rules": rules})
}
