export type DeductionType="NET_PAY"|"TAXABLE_INCOME";
export type CustomDeduction={name:string;amount:string;type:DeductionType};
export type CalculateRequest={gross_salary:string;calculation_date:string;explain:boolean;custom_deductions:CustomDeduction[]};
export type Deduction={code:string;name:string;amount:string;reduces_taxable_income:boolean};
export type BandTrace={from:string;to?:string;rate:string;tax:string};
export type Calculation={calculation_date:string;gross_salary:string;taxable_income:string;paye_before_relief:string;relief:string;paye:string;statutory_deductions:Deduction[];custom_deductions:Deduction[];total_deductions:string;net_salary:string;rule_versions:Record<string,string>;trace?:BandTrace[]};
export type APIError={code:string;message:string;fields?:Record<string,string>;request_id?:string};
