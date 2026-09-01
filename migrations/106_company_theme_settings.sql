ALTER TABLE companies
    ADD COLUMN primary_color CHAR(7) NOT NULL DEFAULT '#15803D' AFTER payroll_frequency,
    ADD COLUMN secondary_color CHAR(7) NOT NULL DEFAULT '#166534' AFTER primary_color;
