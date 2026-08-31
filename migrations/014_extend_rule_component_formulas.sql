-- The admin governance schema (payroll_rule_components.formula_type) only
-- supported BANDS/PERCENTAGE/FIXED/JSON - but several real, currently-live
-- statutory rules use calculation methods outside that set: NSSF uses
-- CAPPED_PERCENTAGE, SHIF uses PERCENTAGE_WITH_MINIMUM. There was no way
-- to draft a new version of either through the admin panel. Extends the
-- enum to cover every method the live engine (internal/payroll/engine)
-- actually understands.

ALTER TABLE payroll_rule_components
    MODIFY COLUMN formula_type ENUM('BANDS','PERCENTAGE','PERCENTAGE_WITH_MINIMUM','CAPPED_PERCENTAGE','TIERED_FIXED_AMOUNT','FIXED','JSON') NOT NULL;
