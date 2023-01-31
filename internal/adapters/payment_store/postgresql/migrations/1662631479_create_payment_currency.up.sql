CREATE INDEX IF NOT EXISTS payment_currency ON payment_instructions USING btree ((body->'incomingInstruction'->'payment'->'currency'->>'isoCode'));
