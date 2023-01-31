CREATE INDEX IF NOT EXISTS payment_status ON payment_instructions USING btree ((body->>'status'));
CREATE INDEX IF NOT EXISTS payment_amount ON payment_instructions USING btree (((body->'incomingInstruction'->'payment'->>'amount')));
CREATE INDEX IF NOT EXISTS payment_execution_date on payment_instructions using btree ((body->'incomingInstruction'->'payment'->>'executionDate'));
