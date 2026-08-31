-- Phase C: lets a person name a saved calculation ("Job Offer A", "My
-- current salary") instead of only seeing a bare date, and enables the
-- salary comparison feature. Nullable so existing saved rows (all saved
-- before this column existed) just fall back to their date in the UI.
ALTER TABLE saved_calculations
    ADD COLUMN label VARCHAR(80) NULL AFTER calculation_date;
