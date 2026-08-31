export type Band={lower:number;upper?:number;rate:number};
export type Component={id?:string;component_code:string;component_type:"PAYE_BANDS"|"STATUTORY_DEDUCTION"|"RELIEF"|"CONFIGURATION";name:string;calculation_order:number;reduces_taxable_income:boolean;reduces_net_pay:boolean;formula_type:"BANDS"|"PERCENTAGE"|"PERCENTAGE_WITH_MINIMUM"|"CAPPED_PERCENTAGE"|"TIERED_FIXED_AMOUNT"|"FIXED"|"JSON";payload:any;is_active:boolean};
export type RuleSet={id?:string;code:string;name:string;jurisdiction:string;effective_from:string;effective_to?:string;status?:string;version_number?:number;source_reference?:string;source_notes?:string;components:Component[]};

// A fully-resolved live rule, as returned by GET /admin/live-rules - used
// to pre-fill a new draft from an existing statutory rule's current
// values (see main.tsx's liveRuleToDraft).
export type LiveRuleParameter={name:string;type:string;decimal?:string;integer?:number;boolean?:boolean;text?:string};
export type LiveRuleBand={from:string;to?:string;rate?:string;fixed_amount?:string;order:number;label?:string};
export type LiveRule={
  id:number;code:string;name:string;category:string;description?:string;
  version_id:number;version_code:string;version_name:string;
  calculation_method:string;calculation_order:number;
  affects_taxable_income:boolean;affects_net_pay:boolean;
  effective_from:string;effective_to?:string;
  parameters:LiveRuleParameter[];bands:LiveRuleBand[];
};
