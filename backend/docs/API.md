# Budget254 PAYE API v1 — Phase 5

## POST /api/v1/calculator/paye
Public endpoint. Authentication is not required.

### Request
```json
{
  "gross_salary":"100000.00",
  "calculation_date":"2026-08-01",
  "explain":true,
  "custom_deductions":[
    {"name":"SACCO Savings","amount":"5000.00","type":"NET_PAY"}
  ]
}
```

Custom deduction types:
- `NET_PAY`: affects net salary only.
- `TAXABLE_INCOME`: may reduce taxable income and net salary.

The API resolves the effective rule versions for the requested date. Kenya rates are not hardcoded in handlers.

## GET /api/v1/payroll/rules?date=YYYY-MM-DD
Returns the effective published rule metadata for the selected date.

## Error contract
`code`, `message`, optional `fields`, and `request_id` are returned for validation and application errors.
