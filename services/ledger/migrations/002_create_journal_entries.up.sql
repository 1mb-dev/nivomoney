-- Create journal entries table
CREATE TABLE IF NOT EXISTS journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_number VARCHAR(50) NOT NULL UNIQUE, -- Sequential number (e.g., "JE-2024-00001")
    type VARCHAR(20) NOT NULL DEFAULT 'standard',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    description TEXT NOT NULL,
    reference_type VARCHAR(50),  -- Type of referenced entity (e.g., "transaction", "invoice")
    reference_id VARCHAR(100),   -- ID of referenced entity
    posted_at TIMESTAMP WITH TIME ZONE,
    posted_by UUID,  -- User ID who posted the entry
    voided_at TIMESTAMP WITH TIME ZONE,
    voided_by UUID,  -- User ID who voided the entry
    void_reason TEXT,
    reversal_entry_id UUID REFERENCES journal_entries(id), -- Entry that reversed this one
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT journal_entries_type_check CHECK (type IN ('standard', 'opening', 'closing', 'adjusting', 'reversing')),
    CONSTRAINT journal_entries_status_check CHECK (status IN ('draft', 'posted', 'voided', 'reversed')),
    CONSTRAINT journal_entries_posted_check CHECK (
        (status = 'posted' AND posted_at IS NOT NULL AND posted_by IS NOT NULL) OR
        (status != 'posted' AND posted_at IS NULL AND posted_by IS NULL)
    ),
    CONSTRAINT journal_entries_voided_check CHECK (
        (status = 'voided' AND voided_at IS NOT NULL AND voided_by IS NOT NULL AND void_reason IS NOT NULL) OR
        (status != 'voided')
    )
);

-- Create indexes
CREATE INDEX idx_journal_entries_entry_number ON journal_entries(entry_number);
CREATE INDEX idx_journal_entries_status ON journal_entries(status);
CREATE INDEX idx_journal_entries_type ON journal_entries(type);
CREATE INDEX idx_journal_entries_posted_at ON journal_entries(posted_at);
CREATE INDEX idx_journal_entries_reference ON journal_entries(reference_type, reference_id);
CREATE INDEX idx_journal_entries_created_at ON journal_entries(created_at DESC);

-- Create trigger to update updated_at
CREATE TRIGGER update_journal_entries_updated_at
    BEFORE UPDATE ON journal_entries
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to generate entry number
CREATE OR REPLACE FUNCTION generate_entry_number()
RETURNS TEXT AS $$
DECLARE
    current_year TEXT;
    next_number INTEGER;
    new_entry_number TEXT;
BEGIN
    -- Get current year
    current_year := TO_CHAR(NOW(), 'YYYY');

    -- Get next number for this year
    SELECT COALESCE(MAX(
        CAST(
            SUBSTRING(entry_number FROM 'JE-' || current_year || '-(\d+)') AS INTEGER
        )
    ), 0) + 1
    INTO next_number
    FROM journal_entries
    WHERE entry_number LIKE 'JE-' || current_year || '-%';

    -- Format as JE-YYYY-00001
    new_entry_number := 'JE-' || current_year || '-' || LPAD(next_number::TEXT, 5, '0');

    RETURN new_entry_number;
END;
$$ LANGUAGE plpgsql;
